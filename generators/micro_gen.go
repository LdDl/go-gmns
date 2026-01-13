package generators

import (
	"fmt"
	"math"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/micro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/pkg/errors"
)

const (
	defaultCellLength    = 4.5
	defaultLaneWidth     = 3.5
	defaultBikeLaneWidth = 0.5
	defaultWalkLaneWidth = 0.5
)

// MicroGenOptions contains options for microscopic network generation
type MicroGenOptions struct {
	CellLength       float64
	LaneWidth        float64
	BikeLaneWidth    float64
	WalkLaneWidth    float64
	SeparateBikeWalk bool
	Verbose          bool
}

// DefaultMicroGenOptions returns default options for micro generation
func DefaultMicroGenOptions() MicroGenOptions {
	return MicroGenOptions{
		CellLength:       defaultCellLength,
		LaneWidth:        defaultLaneWidth,
		BikeLaneWidth:    defaultBikeLaneWidth,
		WalkLaneWidth:    defaultWalkLaneWidth,
		SeparateBikeWalk: false,
		Verbose:          false,
	}
}

// mesoMicroMapping tracks micro node IDs for each meso link's lanes
// Structure: mesoLinkID -> laneID -> []nodeID (sorted by cellIndex)
type mesoMicroMapping map[gmns.LinkID]map[int][]gmns.NodeID

// GenerateMicroscopic generates microscopic network from macro and meso networks
func GenerateMicroscopic(macroNet *macro.Net, mesoNet *meso.Net, opts ...MicroGenOptions) (*micro.Net, error) {
	options := DefaultMicroGenOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	if options.Verbose {
		fmt.Print("Generating microscopic network...")
	}

	microNet := micro.NewNet()

	// Build reverse mapping: macro link ID -> meso link IDs
	macroLinkMesoLinks := buildMacroToMesoMapping(mesoNet)

	// Process each macro link
	for _, macroLink := range macroNet.Links {
		mesoLinkIDs := macroLinkMesoLinks[macroLink.ID]
		err := processMacroLink(macroNet, mesoNet, microNet, macroLink, mesoLinkIDs, options)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to process macro link %d", macroLink.ID)
		}
	}

	// Connect micro links through movements
	err := connectMicroLinks(mesoNet, microNet, options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect micro links")
	}

	if options.Verbose {
		fmt.Printf(" Done. Nodes: %d, Links: %d\n", len(microNet.Nodes), len(microNet.Links))
	}

	return microNet, nil
}

// buildMacroToMesoMapping builds a mapping from macro link IDs to their meso link IDs
func buildMacroToMesoMapping(mesoNet *meso.Net) map[gmns.LinkID][]gmns.LinkID {
	result := make(map[gmns.LinkID][]gmns.LinkID)
	for _, mesoLink := range mesoNet.Links {
		macroLinkID := mesoLink.MacroLink()
		if macroLinkID < 0 {
			continue // Skip connection links (i.e. movements)
		}
		if _, ok := result[macroLinkID]; !ok {
			result[macroLinkID] = make([]gmns.LinkID, 0, 1)
		}
		result[macroLinkID] = append(result[macroLinkID], mesoLink.ID)
	}
	return result
}

// buildMesoMicroMapping builds a mapping from micro nodes in microNet
// Returns: mesoLinkID -> laneID -> []nodeID (sorted by cellIndex)
func buildMesoMicroMapping(microNet *micro.Net) mesoMicroMapping {
	result := make(mesoMicroMapping)

	// Collect nodes by mesoLinkID and laneID
	type nodeWithCell struct {
		nodeID    gmns.NodeID
		cellIndex int
	}
	temp := make(map[gmns.LinkID]map[int][]nodeWithCell)

	for _, node := range microNet.Nodes {
		mesoLinkID := node.MesoLink()
		if mesoLinkID < 0 {
			continue
		}
		laneID := node.LaneID()
		if _, ok := temp[mesoLinkID]; !ok {
			temp[mesoLinkID] = make(map[int][]nodeWithCell)
		}
		if _, ok := temp[mesoLinkID][laneID]; !ok {
			temp[mesoLinkID][laneID] = make([]nodeWithCell, 0)
		}
		temp[mesoLinkID][laneID] = append(temp[mesoLinkID][laneID], nodeWithCell{
			nodeID:    node.ID,
			cellIndex: node.CellIndex(),
		})
	}

	// Sort by cellIndex and extract nodeIDs
	for mesoLinkID, lanes := range temp {
		result[mesoLinkID] = make(map[int][]gmns.NodeID)
		for laneID, nodes := range lanes {
			sort.Slice(nodes, func(i, j int) bool {
				return nodes[i].cellIndex < nodes[j].cellIndex
			})
			nodeIDs := make([]gmns.NodeID, len(nodes))
			for i, n := range nodes {
				nodeIDs[i] = n.nodeID
			}
			result[mesoLinkID][laneID] = nodeIDs
		}
	}

	return result
}

// processMacroLink processes a single macro link and creates micro nodes/links
func processMacroLink(macroNet *macro.Net, mesoNet *meso.Net, microNet *micro.Net, macroLink *macro.Link, mesoLinkIDs []gmns.LinkID, opts MicroGenOptions) error {
	if len(mesoLinkIDs) == 0 {
		return nil // Skip links without meso links
	}

	// Determine multimodal agent types
	agentTypes := macroLink.AllowedAgentTypes()
	mainAgents, hasBike, hasWalk := prepareBikeWalkAgents(agentTypes, opts.SeparateBikeWalk)

	originalLanesNum := float64(macroLink.LanesNum())

	// Local mapping to track micro nodes during generation
	localMapping := make(mesoMicroMapping)

	// Create micro nodes for each meso link
	for _, mesoLinkID := range mesoLinkIDs {
		mesoLink, ok := mesoNet.Links[mesoLinkID]
		if !ok {
			return fmt.Errorf("meso link %d not found", mesoLinkID)
		}

		err := createMicroNodesForMesoLink(microNet, mesoLink, originalLanesNum, hasBike, hasWalk, opts, localMapping)
		if err != nil {
			return errors.Wrapf(err, "failed to create micro nodes for meso link %d", mesoLinkID)
		}
	}

	// Mark upstream and downstream end nodes
	if len(mesoLinkIDs) > 0 {
		err := markEndNodes(macroNet, microNet, macroLink, mesoLinkIDs, hasBike, hasWalk, localMapping)
		if err != nil {
			return err
		}
	}

	// Merge nodes between adjacent meso links
	err := mergeAdjacentMesoLinks(mesoNet, microNet, mesoLinkIDs, hasBike, hasWalk, localMapping)
	if err != nil {
		return err
	}

	// Create micro links (cells)
	for _, mesoLinkID := range mesoLinkIDs {
		mesoLink, ok := mesoNet.Links[mesoLinkID]
		if !ok {
			continue
		}
		err := createMicroLinksForMesoLink(microNet, mesoLink, mainAgents, hasBike, hasWalk, localMapping)
		if err != nil {
			return errors.Wrapf(err, "failed to create micro links for meso link %d", mesoLinkID)
		}
	}

	return nil
}

// createMicroNodesForMesoLink creates micro nodes for a single meso link
func createMicroNodesForMesoLink(microNet *micro.Net, mesoLink *meso.Link, originalLanesNum float64, hasBike, hasWalk bool, opts MicroGenOptions, localMapping mesoMicroMapping) error {
	lanesChange := mesoLink.LanesChange()
	laneChangesLeft := float64(lanesChange[0])
	lanesNumOffset := -1 * (originalLanesNum/2 - 0.5 + laneChangesLeft)

	// Calculate number of cells
	cellsNum := int(math.Max(1.0, math.Round(mesoLink.LengthMeters()/opts.CellLength)))

	// Generate lane geometries with offset
	laneGeometries := make([]orb.LineString, mesoLink.LanesNum())
	var bikeGeometry, walkGeometry orb.LineString
	lastOffset := 0.0

	for i := 0; i < mesoLink.LanesNum(); i++ {
		laneOffset := (lanesNumOffset + float64(i)) * opts.LaneWidth
		lastOffset = laneOffset
		if math.Abs(laneOffset) > 1e-2 {
			laneGeometries[i] = geomath.OffsetCurve(mesoLink.GeomEuclidean(), -laneOffset)
			laneGeometries[i] = geomath.LineToSpherical(laneGeometries[i])
		} else {
			laneGeometries[i] = mesoLink.Geom().Clone()
		}
	}

	// Bike/walk geometries
	if hasBike {
		bikeOffset := lastOffset + opts.BikeLaneWidth
		if math.Abs(bikeOffset) > 1e-2 {
			bikeGeometry = geomath.OffsetCurve(mesoLink.GeomEuclidean(), -bikeOffset)
			bikeGeometry = geomath.LineToSpherical(bikeGeometry)
		} else {
			bikeGeometry = mesoLink.Geom().Clone()
		}
	}
	if hasWalk {
		walkOffset := lastOffset + opts.WalkLaneWidth
		if hasBike {
			walkOffset += opts.BikeLaneWidth
		}
		if math.Abs(walkOffset) > 1e-2 {
			walkGeometry = geomath.OffsetCurve(mesoLink.GeomEuclidean(), -walkOffset)
			walkGeometry = geomath.LineToSpherical(walkGeometry)
		} else {
			walkGeometry = mesoLink.Geom().Clone()
		}
	}

	// Initialize local mapping for this meso link
	if _, ok := localMapping[mesoLink.ID]; !ok {
		localMapping[mesoLink.ID] = make(map[int][]gmns.NodeID)
	}

	// Create micro nodes for each lane
	for laneIdx := 0; laneIdx < mesoLink.LanesNum(); laneIdx++ {
		laneID := laneIdx + 1
		laneNodeIDs := make([]gmns.NodeID, 0, cellsNum+1)

		for cellIdx := 0; cellIdx <= cellsNum; cellIdx++ {
			fraction := float64(cellIdx) / float64(cellsNum)
			distance := mesoLink.LengthMeters() * fraction
			point, _ := geo.PointAtDistanceAlongLine(laneGeometries[laneIdx], distance)
			pointEuc := geomath.PointToEuclidean(point)

			nodeID := microNet.MaxNodeID()
			node := micro.NewNodeFrom(nodeID,
				micro.WithPointGeom(point),
				micro.WithPointGeomEuclidean(pointEuc),
				micro.WithNodeMesoLinkID(mesoLink.ID),
				micro.WithNodeLaneID(laneID),
				micro.WithCellIndex(cellIdx),
				micro.WithBoundaryType(types.BOUNDARY_NONE),
			)
			microNet.AddNode(node)
			laneNodeIDs = append(laneNodeIDs, nodeID)
		}
		localMapping[mesoLink.ID][laneID] = laneNodeIDs
	}

	// Create micro nodes for bike lane
	if hasBike && len(bikeGeometry) > 0 {
		bikeLaneID := -1
		bikeNodeIDs := make([]gmns.NodeID, 0, cellsNum+1)

		for cellIdx := 0; cellIdx <= cellsNum; cellIdx++ {
			fraction := float64(cellIdx) / float64(cellsNum)
			distance := mesoLink.LengthMeters() * fraction
			point, _ := geo.PointAtDistanceAlongLine(bikeGeometry, distance)
			pointEuc := geomath.PointToEuclidean(point)

			nodeID := microNet.MaxNodeID()
			node := micro.NewNodeFrom(nodeID,
				micro.WithPointGeom(point),
				micro.WithPointGeomEuclidean(pointEuc),
				micro.WithNodeMesoLinkID(mesoLink.ID),
				micro.WithNodeLaneID(bikeLaneID),
				micro.WithCellIndex(cellIdx),
				micro.WithBoundaryType(types.BOUNDARY_NONE),
			)
			microNet.AddNode(node)
			bikeNodeIDs = append(bikeNodeIDs, nodeID)
		}
		localMapping[mesoLink.ID][bikeLaneID] = bikeNodeIDs
	}

	// Create micro nodes for walk lane
	if hasWalk && len(walkGeometry) > 0 {
		walkLaneID := -2
		walkNodeIDs := make([]gmns.NodeID, 0, cellsNum+1)

		for cellIdx := 0; cellIdx <= cellsNum; cellIdx++ {
			fraction := float64(cellIdx) / float64(cellsNum)
			distance := mesoLink.LengthMeters() * fraction
			point, _ := geo.PointAtDistanceAlongLine(walkGeometry, distance)
			pointEuc := geomath.PointToEuclidean(point)

			nodeID := microNet.MaxNodeID()
			node := micro.NewNodeFrom(nodeID,
				micro.WithPointGeom(point),
				micro.WithPointGeomEuclidean(pointEuc),
				micro.WithNodeMesoLinkID(mesoLink.ID),
				micro.WithNodeLaneID(walkLaneID),
				micro.WithCellIndex(cellIdx),
				micro.WithBoundaryType(types.BOUNDARY_NONE),
			)
			microNet.AddNode(node)
			walkNodeIDs = append(walkNodeIDs, nodeID)
		}
		localMapping[mesoLink.ID][walkLaneID] = walkNodeIDs
	}

	return nil
}

// markEndNodes marks upstream and downstream end nodes
func markEndNodes(macroNet *macro.Net, microNet *micro.Net, macroLink *macro.Link, mesoLinkIDs []gmns.LinkID, hasBike, hasWalk bool, localMapping mesoMicroMapping) error {
	// First meso link - upstream end
	firstMesoLinkID := mesoLinkIDs[0]
	macroSourceNode := macroNet.Nodes[macroLink.SourceNode()]

	if lanes, ok := localMapping[firstMesoLinkID]; ok {
		for laneID, nodeIDs := range lanes {
			if laneID > 0 || (hasBike && laneID == -1) || (hasWalk && laneID == -2) {
				if len(nodeIDs) > 0 {
					if node, ok := microNet.Nodes[nodeIDs[0]]; ok {
						node.SetUpstreamEnd(true)
						node.SetZoneID(macroSourceNode.Zone())
					}
				}
			}
		}
	}

	// Last meso link - downstream end
	lastMesoLinkID := mesoLinkIDs[len(mesoLinkIDs)-1]
	macroTargetNode := macroNet.Nodes[macroLink.TargetNode()]

	if lanes, ok := localMapping[lastMesoLinkID]; ok {
		for laneID, nodeIDs := range lanes {
			if laneID > 0 || (hasBike && laneID == -1) || (hasWalk && laneID == -2) {
				if len(nodeIDs) > 0 {
					if node, ok := microNet.Nodes[nodeIDs[len(nodeIDs)-1]]; ok {
						node.SetDownstreamEnd(true)
						node.SetZoneID(macroTargetNode.Zone())
					}
				}
			}
		}
	}

	return nil
}

// mergeAdjacentMesoLinks merges nodes between adjacent meso links
func mergeAdjacentMesoLinks(mesoNet *meso.Net, microNet *micro.Net, mesoLinkIDs []gmns.LinkID, hasBike, hasWalk bool, localMapping mesoMicroMapping) error {
	for i := 0; i < len(mesoLinkIDs)-1; i++ {
		upstreamMesoLinkID := mesoLinkIDs[i]
		downstreamMesoLinkID := mesoLinkIDs[i+1]

		upstreamMesoLink := mesoNet.Links[upstreamMesoLinkID]
		downstreamMesoLink := mesoNet.Links[downstreamMesoLinkID]

		upstreamLanesChange := upstreamMesoLink.LanesChange()
		downstreamLanesChange := downstreamMesoLink.LanesChange()

		minLeftLane := min(upstreamLanesChange[0], downstreamLanesChange[0])
		upstreamLaneStart := upstreamLanesChange[0] - minLeftLane
		downstreamLaneStart := downstreamLanesChange[0] - minLeftLane

		numConnections := min(
			upstreamMesoLink.LanesNum()-upstreamLaneStart,
			downstreamMesoLink.LanesNum()-downstreamLaneStart,
		)

		upstreamLanes := localMapping[upstreamMesoLinkID]
		downstreamLanes := localMapping[downstreamMesoLinkID]

		for j := 0; j < numConnections; j++ {
			upLaneID := upstreamLaneStart + j + 1 // 1-indexed
			downLaneID := downstreamLaneStart + j + 1

			upLane, upOk := upstreamLanes[upLaneID]
			downLane, downOk := downstreamLanes[downLaneID]

			if upOk && downOk && len(upLane) > 0 && len(downLane) > 0 {
				// Replace upstream's last node with downstream's first node
				oldNodeID := upLane[len(upLane)-1]
				newNodeID := downLane[0]
				upstreamLanes[upLaneID][len(upLane)-1] = newNodeID
				microNet.DeleteNode(oldNodeID)
			}
		}

		// Merge bike lanes
		if hasBike {
			upBike, upOk := upstreamLanes[-1]
			downBike, downOk := downstreamLanes[-1]
			if upOk && downOk && len(upBike) > 0 && len(downBike) > 0 {
				oldNodeID := upBike[len(upBike)-1]
				upstreamLanes[-1][len(upBike)-1] = downBike[0]
				microNet.DeleteNode(oldNodeID)
			}
		}

		// Merge walk lanes
		if hasWalk {
			upWalk, upOk := upstreamLanes[-2]
			downWalk, downOk := downstreamLanes[-2]
			if upOk && downOk && len(upWalk) > 0 && len(downWalk) > 0 {
				oldNodeID := upWalk[len(upWalk)-1]
				upstreamLanes[-2][len(upWalk)-1] = downWalk[0]
				microNet.DeleteNode(oldNodeID)
			}
		}
	}

	return nil
}

// createMicroLinksForMesoLink creates micro links (cells) for a meso link
func createMicroLinksForMesoLink(microNet *micro.Net, mesoLink *meso.Link, mainAgents []types.AgentType, hasBike, hasWalk bool, localMapping mesoMicroMapping) error {
	lanes, ok := localMapping[mesoLink.ID]
	if !ok {
		return nil
	}

	// Get sorted lane IDs (positive lanes only for main traffic)
	var regularLaneIDs []int
	for laneID := range lanes {
		if laneID > 0 {
			regularLaneIDs = append(regularLaneIDs, laneID)
		}
	}
	sort.Ints(regularLaneIDs)

	for _, laneID := range regularLaneIDs {
		laneNodes := lanes[laneID]

		// Forward links
		for cellIdx := 0; cellIdx < len(laneNodes)-1; cellIdx++ {
			err := createMicroLink(microNet, mesoLink, laneNodes[cellIdx], laneNodes[cellIdx+1],
				laneID, types.CELL_FORWARD, mainAgents)
			if err != nil {
				return err
			}
		}

		// Lane change left (to higher lane number)
		nextLaneID := laneID + 1
		if nextLaneNodes, ok := lanes[nextLaneID]; ok {
			for cellIdx := 0; cellIdx < len(laneNodes)-1 && cellIdx < len(nextLaneNodes)-1; cellIdx++ {
				err := createMicroLink(microNet, mesoLink, laneNodes[cellIdx], nextLaneNodes[cellIdx+1],
					laneID, types.CELL_LANE_CHANGE, mainAgents)
				if err != nil {
					return err
				}
			}
		}

		// Lane change right (to lower lane number)
		prevLaneID := laneID - 1
		if prevLaneID > 0 {
			if prevLaneNodes, ok := lanes[prevLaneID]; ok {
				for cellIdx := 0; cellIdx < len(laneNodes)-1 && cellIdx < len(prevLaneNodes)-1; cellIdx++ {
					err := createMicroLink(microNet, mesoLink, laneNodes[cellIdx], prevLaneNodes[cellIdx+1],
						laneID, types.CELL_LANE_CHANGE, mainAgents)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Bike lane links
	if hasBike {
		if bikeNodes, ok := lanes[-1]; ok {
			for cellIdx := 0; cellIdx < len(bikeNodes)-1; cellIdx++ {
				err := createMicroLink(microNet, mesoLink, bikeNodes[cellIdx], bikeNodes[cellIdx+1],
					-1, types.CELL_FORWARD, []types.AgentType{types.AGENT_BIKE})
				if err != nil {
					return err
				}
			}
		}
	}

	// Walk lane links
	if hasWalk {
		if walkNodes, ok := lanes[-2]; ok {
			for cellIdx := 0; cellIdx < len(walkNodes)-1; cellIdx++ {
				err := createMicroLink(microNet, mesoLink, walkNodes[cellIdx], walkNodes[cellIdx+1],
					-2, types.CELL_FORWARD, []types.AgentType{types.AGENT_WALK})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// createMicroLink creates a single micro link
func createMicroLink(microNet *micro.Net, mesoLink *meso.Link, sourceNodeID, targetNodeID gmns.NodeID,
	laneID int, cellType types.CellType, allowedAgents []types.AgentType) error {

	sourceNode, ok := microNet.Nodes[sourceNodeID]
	if !ok {
		return fmt.Errorf("source micro node %d not found", sourceNodeID)
	}
	targetNode, ok := microNet.Nodes[targetNodeID]
	if !ok {
		return fmt.Errorf("target micro node %d not found", targetNodeID)
	}

	geom := orb.LineString{sourceNode.Geom(), targetNode.Geom()}
	geomEuc := orb.LineString{sourceNode.GeomEuclidean(), targetNode.GeomEuclidean()}
	length := geo.LengthHaversine(geom)

	linkID := microNet.MaxLinkID()
	link := micro.NewLinkFrom(linkID, sourceNodeID, targetNodeID,
		micro.WithLineGeom(geom),
		micro.WithLineGeomEuclidean(geomEuc),
		micro.WithLengthMeters(length),
		micro.WithMesoLinkID(mesoLink.ID),
		micro.WithMacroLinkID(mesoLink.MacroLink()),
		micro.WithMacroNodeID(mesoLink.MacroNode()),
		micro.WithCellType(cellType),
		micro.WithLaneID(laneID),
		micro.WithMesoLinkType(mesoLink.LinkType()),
		micro.WithControlType(mesoLink.ControlType()),
		micro.WithFreeSpeed(mesoLink.FreeSpeed()),
		micro.WithCapacity(mesoLink.Capacity()),
		micro.WithAllowedAgentTypes(allowedAgents),
		micro.WithMovementCompositeType(movement.MOVEMENT_UNDEFINED),
	)

	microNet.AddLink(link)
	sourceNode.AddOutcomingLink(linkID)
	targetNode.AddIncomingLink(linkID)

	return nil
}

// connectMicroLinks connects micro links through movements
func connectMicroLinks(mesoNet *meso.Net, microNet *micro.Net, opts MicroGenOptions) error {
	// Build global mapping from all micro nodes
	globalMapping := buildMesoMicroMapping(microNet)

	for _, mesoLink := range mesoNet.Links {
		// Process only movement meso links
		if mesoLink.Movement() < 0 {
			continue
		}

		incomeMesoLinkID := mesoLink.MovementMesoLinkIncome()
		outcomeMesoLinkID := mesoLink.MovementMesoLinkOutcome()

		if incomeMesoLinkID < 0 || outcomeMesoLinkID < 0 {
			continue
		}

		incomeStartSeq := mesoLink.MovementIncomeLaneStartSeqID()
		outcomeStartSeq := mesoLink.MovementOutcomeLaneStartSeqID()

		if incomeStartSeq < 0 || outcomeStartSeq < 0 {
			continue
		}

		incomeLanes, incomeOk := globalMapping[incomeMesoLinkID]
		outcomeLanes, outcomeOk := globalMapping[outcomeMesoLinkID]

		if !incomeOk || !outcomeOk {
			continue
		}

		// Create connector cells for each lane in the movement
		for laneOffset := 0; laneOffset < mesoLink.LanesNum(); laneOffset++ {
			incomeLaneID := incomeStartSeq + laneOffset + 1 // Convert to 1-indexed
			outcomeLaneID := outcomeStartSeq + laneOffset + 1

			incomeLaneNodes, incOk := incomeLanes[incomeLaneID]
			outcomeLaneNodes, outOk := outcomeLanes[outcomeLaneID]

			if !incOk || !outOk || len(incomeLaneNodes) == 0 || len(outcomeLaneNodes) == 0 {
				continue
			}

			startNodeID := incomeLaneNodes[len(incomeLaneNodes)-1]
			endNodeID := outcomeLaneNodes[0]

			startNode, ok := microNet.Nodes[startNodeID]
			if !ok {
				continue
			}
			endNode, ok := microNet.Nodes[endNodeID]
			if !ok {
				continue
			}

			// Create connector geometry
			laneGeom := orb.LineString{startNode.Geom(), endNode.Geom()}
			laneLength := geo.LengthHaversine(laneGeom)
			cellsNum := int(math.Max(1.0, math.Round(laneLength/opts.CellLength)))

			// Create intermediate nodes and links
			lastNodeID := startNodeID
			isFirstMovement := true

			for cellIdx := 1; cellIdx < cellsNum; cellIdx++ {
				fraction := float64(cellIdx) / float64(cellsNum)
				distance := laneLength * fraction
				point, _ := geo.PointAtDistanceAlongLine(laneGeom, distance)
				pointEuc := geomath.PointToEuclidean(point)

				nodeID := microNet.MaxNodeID()
				node := micro.NewNodeFrom(nodeID,
					micro.WithPointGeom(point),
					micro.WithPointGeomEuclidean(pointEuc),
					micro.WithNodeMesoLinkID(mesoLink.ID),
					micro.WithNodeLaneID(laneOffset+1),
					micro.WithCellIndex(cellIdx),
				)
				microNet.AddNode(node)

				// Create link from last node to this node
				lastNode, _ := microNet.Nodes[lastNodeID]
				linkGeom := orb.LineString{lastNode.Geom(), node.Geom()}
				linkGeomEuc := orb.LineString{lastNode.GeomEuclidean(), node.GeomEuclidean()}
				linkLength := geo.LengthHaversine(linkGeom)

				linkID := microNet.MaxLinkID()
				link := micro.NewLinkFrom(linkID, lastNodeID, nodeID,
					micro.WithLineGeom(linkGeom),
					micro.WithLineGeomEuclidean(linkGeomEuc),
					micro.WithLengthMeters(linkLength),
					micro.WithMesoLinkID(mesoLink.ID),
					micro.WithMacroLinkID(mesoLink.MacroLink()),
					micro.WithMacroNodeID(mesoLink.MacroNode()),
					micro.WithCellType(types.CELL_FORWARD),
					micro.WithLaneID(laneOffset+1),
					micro.WithMesoLinkType(mesoLink.LinkType()),
					micro.WithControlType(mesoLink.ControlType()),
					micro.WithFreeSpeed(mesoLink.FreeSpeed()),
					micro.WithCapacity(mesoLink.Capacity()),
					micro.WithAllowedAgentTypes(mesoLink.AllowedAgentTypes()),
					micro.WithIsFirstMovementCell(isFirstMovement),
					micro.WithMovementCompositeType(mesoLink.MvmtTextID()),
				)
				microNet.AddLink(link)
				lastNode.AddOutcomingLink(linkID)
				node.AddIncomingLink(linkID)

				isFirstMovement = false
				lastNodeID = nodeID
			}

			// Create final link to end node
			lastNode, _ := microNet.Nodes[lastNodeID]
			linkGeom := orb.LineString{lastNode.Geom(), endNode.Geom()}
			linkGeomEuc := orb.LineString{lastNode.GeomEuclidean(), endNode.GeomEuclidean()}
			linkLength := geo.LengthHaversine(linkGeom)

			linkID := microNet.MaxLinkID()
			link := micro.NewLinkFrom(linkID, lastNodeID, endNodeID,
				micro.WithLineGeom(linkGeom),
				micro.WithLineGeomEuclidean(linkGeomEuc),
				micro.WithLengthMeters(linkLength),
				micro.WithMesoLinkID(mesoLink.ID),
				micro.WithMacroLinkID(mesoLink.MacroLink()),
				micro.WithMacroNodeID(mesoLink.MacroNode()),
				micro.WithCellType(types.CELL_FORWARD),
				micro.WithLaneID(laneOffset+1),
				micro.WithMesoLinkType(mesoLink.LinkType()),
				micro.WithControlType(mesoLink.ControlType()),
				micro.WithFreeSpeed(mesoLink.FreeSpeed()),
				micro.WithCapacity(mesoLink.Capacity()),
				micro.WithAllowedAgentTypes(mesoLink.AllowedAgentTypes()),
				micro.WithIsFirstMovementCell(isFirstMovement),
				micro.WithMovementCompositeType(mesoLink.MvmtTextID()),
			)
			microNet.AddLink(link)
			lastNode.AddOutcomingLink(linkID)
			endNode.AddIncomingLink(linkID)
		}
	}

	return nil
}

// prepareBikeWalkAgents separates bike/walk agents if needed
func prepareBikeWalkAgents(agentTypes []types.AgentType, separate bool) (main []types.AgentType, bike bool, walk bool) {
	if len(agentTypes) == 0 || !separate {
		main = make([]types.AgentType, len(agentTypes))
		copy(main, agentTypes)
		return main, false, false
	}

	hasAuto := false
	hasBike := false
	hasWalk := false

	for _, agent := range agentTypes {
		switch agent {
		case types.AGENT_AUTO:
			hasAuto = true
		case types.AGENT_BIKE:
			hasBike = true
		case types.AGENT_WALK:
			hasWalk = true
		}
	}

	if hasAuto && hasBike && hasWalk {
		return []types.AgentType{types.AGENT_AUTO}, true, true
	} else if hasAuto && hasBike {
		return []types.AgentType{types.AGENT_AUTO}, true, false
	} else if hasAuto && hasWalk {
		return []types.AgentType{types.AGENT_AUTO}, false, true
	} else if hasBike && hasWalk {
		return []types.AgentType{types.AGENT_BIKE}, false, true
	}

	return agentTypes, false, false
}
