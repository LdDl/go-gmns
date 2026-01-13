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
func GenerateMicroscopic(macroNet *macro.Net, mesoNet *meso.Net, movements movement.MovementsStorage, opts ...MicroGenOptions) (*micro.Net, error) {
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

	// Process each macro link (sorted for deterministic output)
	sortedMacroLinkIDs := make([]gmns.LinkID, 0, len(macroNet.Links))
	for linkID := range macroNet.Links {
		sortedMacroLinkIDs = append(sortedMacroLinkIDs, linkID)
	}
	sort.Slice(sortedMacroLinkIDs, func(i, j int) bool {
		return sortedMacroLinkIDs[i] < sortedMacroLinkIDs[j]
	})
	for _, macroLinkID := range sortedMacroLinkIDs {
		macroLink := macroNet.Links[macroLinkID]
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

	// Compute movement flags and fix gaps
	movementFlags := ComputeMovementFlags(macroNet, movements)
	err = fixGaps(macroNet, mesoNet, microNet, macroLinkMesoLinks, movementFlags, movements)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fix gaps")
	}

	if options.Verbose {
		fmt.Printf(" Done. Nodes: %d, Links: %d\n", len(microNet.Nodes), len(microNet.Links))
	}

	return microNet, nil
}

// buildMacroToMesoMapping builds a mapping from macro link IDs to their meso link IDs
func buildMacroToMesoMapping(mesoNet *meso.Net) map[gmns.LinkID][]gmns.LinkID {
	result := make(map[gmns.LinkID][]gmns.LinkID)
	// Sort meso link IDs for deterministic iteration
	sortedMesoLinkIDs := make([]gmns.LinkID, 0, len(mesoNet.Links))
	for linkID := range mesoNet.Links {
		sortedMesoLinkIDs = append(sortedMesoLinkIDs, linkID)
	}
	sort.Slice(sortedMesoLinkIDs, func(i, j int) bool {
		return sortedMesoLinkIDs[i] < sortedMesoLinkIDs[j]
	})
	for _, mesoLinkID := range sortedMesoLinkIDs {
		mesoLink := mesoNet.Links[mesoLinkID]
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

	// Sort micro node IDs for deterministic iteration
	sortedMicroNodeIDs := make([]gmns.NodeID, 0, len(microNet.Nodes))
	for nodeID := range microNet.Nodes {
		sortedMicroNodeIDs = append(sortedMicroNodeIDs, nodeID)
	}
	sort.Slice(sortedMicroNodeIDs, func(i, j int) bool {
		return sortedMicroNodeIDs[i] < sortedMicroNodeIDs[j]
	})

	for _, nodeID := range sortedMicroNodeIDs {
		node := microNet.Nodes[nodeID]
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

	// Sort by cellIndex and extract nodeIDs (sort keys for deterministic iteration)
	sortedMesoLinkIDs := make([]gmns.LinkID, 0, len(temp))
	for mesoLinkID := range temp {
		sortedMesoLinkIDs = append(sortedMesoLinkIDs, mesoLinkID)
	}
	sort.Slice(sortedMesoLinkIDs, func(i, j int) bool {
		return sortedMesoLinkIDs[i] < sortedMesoLinkIDs[j]
	})

	for _, mesoLinkID := range sortedMesoLinkIDs {
		lanes := temp[mesoLinkID]
		result[mesoLinkID] = make(map[int][]gmns.NodeID)

		// Sort lane IDs for deterministic iteration
		sortedLaneIDs := make([]int, 0, len(lanes))
		for laneID := range lanes {
			sortedLaneIDs = append(sortedLaneIDs, laneID)
		}
		sort.Ints(sortedLaneIDs)

		for _, laneID := range sortedLaneIDs {
			nodes := lanes[laneID]
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
		// Sort lane IDs for deterministic iteration
		sortedLaneIDs := make([]int, 0, len(lanes))
		for laneID := range lanes {
			sortedLaneIDs = append(sortedLaneIDs, laneID)
		}
		sort.Ints(sortedLaneIDs)

		for _, laneID := range sortedLaneIDs {
			nodeIDs := lanes[laneID]
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
		// Sort lane IDs for deterministic iteration
		sortedLaneIDs := make([]int, 0, len(lanes))
		for laneID := range lanes {
			sortedLaneIDs = append(sortedLaneIDs, laneID)
		}
		sort.Ints(sortedLaneIDs)

		for _, laneID := range sortedLaneIDs {
			nodeIDs := lanes[laneID]
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

	// Sort meso link IDs for deterministic iteration
	sortedMesoLinkIDs := make([]gmns.LinkID, 0, len(mesoNet.Links))
	for linkID := range mesoNet.Links {
		sortedMesoLinkIDs = append(sortedMesoLinkIDs, linkID)
	}
	sort.Slice(sortedMesoLinkIDs, func(i, j int) bool {
		return sortedMesoLinkIDs[i] < sortedMesoLinkIDs[j]
	})

	for _, mesoLinkID := range sortedMesoLinkIDs {
		mesoLink := mesoNet.Links[mesoLinkID]
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

// fixGaps removes duplicate micro nodes when movements are not needed at certain macro nodes.
// When movementIsNeeded == false for a macro node, there are duplicate micro nodes at the
// boundary between incoming and outcoming meso links. This function removes the duplicates
// based on the downstreamIsTarget and upstreamIsTarget flags on macro links.
func fixGaps(macroNet *macro.Net, mesoNet *meso.Net, microNet *micro.Net, macroLinkMesoLinks map[gmns.LinkID][]gmns.LinkID, flags *MovementFlags, movements movement.MovementsStorage) error {
	// Build global mapping of meso link to micro nodes per lane
	globalMapping := buildMesoMicroMapping(microNet)

	// Group movements by macro node (sorted for deterministic iteration)
	movementsByNode := make(map[gmns.NodeID][]*movement.Movement)
	sortedMvmtIDs := make([]gmns.MovementID, 0, len(movements))
	for mvmtID := range movements {
		sortedMvmtIDs = append(sortedMvmtIDs, mvmtID)
	}
	sort.Slice(sortedMvmtIDs, func(i, j int) bool {
		return sortedMvmtIDs[i] < sortedMvmtIDs[j]
	})
	for _, mvmtID := range sortedMvmtIDs {
		mvmt := movements[mvmtID]
		movementsByNode[mvmt.MacroNode()] = append(movementsByNode[mvmt.MacroNode()], mvmt)
	}

	// Process each macro node (sorted for deterministic iteration)
	sortedMacroNodeIDs := make([]gmns.NodeID, 0, len(macroNet.Nodes))
	for nodeID := range macroNet.Nodes {
		sortedMacroNodeIDs = append(sortedMacroNodeIDs, nodeID)
	}
	sort.Slice(sortedMacroNodeIDs, func(i, j int) bool {
		return sortedMacroNodeIDs[i] < sortedMacroNodeIDs[j]
	})

	for _, macroNodeID := range sortedMacroNodeIDs {
		macroNode := macroNet.Nodes[macroNodeID]
		// Skip nodes where movements are needed (intersections)
		if flags.NodesNeedMovement[macroNode.ID] {
			continue
		}

		// Get movements for this node
		nodeMvmts := movementsByNode[macroNode.ID]

		// Process each movement
		for _, mvmt := range nodeMvmts {
			incomingMacroLinkID := mvmt.IncomeMacroLink()
			outcomingMacroLinkID := mvmt.OutcomeMacroLink()

			if _, ok := macroNet.Links[incomingMacroLinkID]; !ok {
				continue
			}
			if _, ok := macroNet.Links[outcomingMacroLinkID]; !ok {
				continue
			}

			// Collect lanes info from movement
			incomeLaneStart := mvmt.IncomeLaneStart()
			incomeLaneEnd := mvmt.IncomeLaneEnd()
			outcomeLaneStart := mvmt.OutcomeLaneStart()
			outcomeLaneEnd := mvmt.OutcomeLaneEnd()

			// Build lane arrays
			var incomeLanes, outcomeLanes []int
			for laneNo := incomeLaneStart; laneNo <= incomeLaneEnd; laneNo++ {
				incomeLanes = append(incomeLanes, laneNo)
			}
			for laneNo := outcomeLaneStart; laneNo <= outcomeLaneEnd; laneNo++ {
				outcomeLanes = append(outcomeLanes, laneNo)
			}

			// Skip if lane counts don't match
			if len(incomeLanes) != len(outcomeLanes) {
				continue
			}
			// Skip if any lane is 0 (invalid)
			hasZero := false
			for _, l := range incomeLanes {
				if l == 0 {
					hasZero = true
					break
				}
			}
			for _, l := range outcomeLanes {
				if l == 0 {
					hasZero = true
					break
				}
			}
			if hasZero {
				continue
			}

			// Get the last meso link of the incoming macro link
			incomingMesoLinkIDs := macroLinkMesoLinks[incomingMacroLinkID]
			if len(incomingMesoLinkIDs) == 0 {
				continue
			}
			incomingMesoLinkID := incomingMesoLinkIDs[len(incomingMesoLinkIDs)-1]
			incomingMesoLink, ok := mesoNet.Links[incomingMesoLinkID]
			if !ok {
				continue
			}

			// Get the first meso link of the outcoming macro link
			outcomingMesoLinkIDs := macroLinkMesoLinks[outcomingMacroLinkID]
			if len(outcomingMesoLinkIDs) == 0 {
				continue
			}
			outcomingMesoLinkID := outcomingMesoLinkIDs[0]
			outcomingMesoLink, ok := mesoNet.Links[outcomingMesoLinkID]
			if !ok {
				continue
			}

			// Get micro nodes for each meso link
			incomingMicroLanes := globalMapping[incomingMesoLinkID]
			outcomingMicroLanes := globalMapping[outcomingMesoLinkID]

			if incomingMicroLanes == nil || outcomingMicroLanes == nil {
				continue
			}

			// Calculate lane indices with lanes change offset
			incomingLanesChange := incomingMesoLink.LanesChange()
			outcomingLanesChange := outcomingMesoLink.LanesChange()

			incomeLaneStartIdx := incomingLanesChange[0] + incomeLanes[0]
			if incomeLanes[0] >= 0 {
				incomeLaneStartIdx--
			}
			outcomeLaneStartIdx := outcomingLanesChange[0] + outcomeLanes[0]
			if outcomeLanes[0] >= 0 {
				outcomeLaneStartIdx--
			}

			// Skip if lane indices are invalid
			if incomeLaneStartIdx < 0 || outcomeLaneStartIdx < 0 {
				continue
			}
			if incomeLaneStartIdx+len(incomeLanes)-1 > incomingMesoLink.LanesNum()-1 {
				continue
			}
			if outcomeLaneStartIdx+len(outcomeLanes)-1 > outcomingMesoLink.LanesNum()-1 {
				continue
			}

			lanesNum := len(incomeLanes)

			// Process according to downstream/upstream target flags
			if flags.DownstreamIsTarget[incomingMacroLinkID] && !flags.UpstreamIsTarget[outcomingMacroLinkID] {
				// Delete outcoming micro nodes (first nodes of outcoming lanes)
				for i := 0; i < lanesNum; i++ {
					incomeLaneIdx := incomeLaneStartIdx + i + 1 // Convert to 1-indexed for globalMapping
					outcomeLaneIdx := outcomeLaneStartIdx + i + 1

					incomeLaneNodes := incomingMicroLanes[incomeLaneIdx]
					outcomeLaneNodes := outcomingMicroLanes[outcomeLaneIdx]

					if len(incomeLaneNodes) == 0 || len(outcomeLaneNodes) == 0 {
						continue
					}

					incomeLastNodeID := incomeLaneNodes[len(incomeLaneNodes)-1]
					outcomeFirstNodeID := outcomeLaneNodes[0]

					outcomeFirstNode, ok := microNet.Nodes[outcomeFirstNodeID]
					if !ok {
						continue
					}

					// Redirect all outcoming links from outcomeFirstNode to start from incomeLastNode
					for el := outcomeFirstNode.OutcomingLinks().Front(); el != nil; el = el.Next() {
						microLinkID := el.Key.(gmns.LinkID)
						microLink, ok := microNet.Links[microLinkID]
						if !ok {
							continue
						}
						microLink.SetSourceNode(incomeLastNodeID)
						if incomeLastNode, ok := microNet.Nodes[incomeLastNodeID]; ok {
							incomeLastNode.AddOutcomingLink(microLinkID)
						}
					}

					// Delete the duplicate node
					delete(microNet.Nodes, outcomeFirstNodeID)
				}
			} else if !flags.DownstreamIsTarget[incomingMacroLinkID] && flags.UpstreamIsTarget[outcomingMacroLinkID] {
				// Delete incoming micro nodes (last nodes of incoming lanes)
				for i := 0; i < lanesNum; i++ {
					incomeLaneIdx := incomeLaneStartIdx + i + 1
					outcomeLaneIdx := outcomeLaneStartIdx + i + 1

					incomeLaneNodes := incomingMicroLanes[incomeLaneIdx]
					outcomeLaneNodes := outcomingMicroLanes[outcomeLaneIdx]

					if len(incomeLaneNodes) == 0 || len(outcomeLaneNodes) == 0 {
						continue
					}

					incomeLastNodeID := incomeLaneNodes[len(incomeLaneNodes)-1]
					outcomeFirstNodeID := outcomeLaneNodes[0]

					incomeLastNode, ok := microNet.Nodes[incomeLastNodeID]
					if !ok {
						continue
					}

					// Redirect all incoming links to incomeLastNode to target outcomeFirstNode
					for el := incomeLastNode.IncomingLinks().Front(); el != nil; el = el.Next() {
						microLinkID := el.Key.(gmns.LinkID)
						microLink, ok := microNet.Links[microLinkID]
						if !ok {
							continue
						}
						microLink.SetTargetNode(outcomeFirstNodeID)
						if outcomeFirstNode, ok := microNet.Nodes[outcomeFirstNodeID]; ok {
							outcomeFirstNode.AddIncomingLink(microLinkID)
						}
					}

					// Delete the duplicate node
					delete(microNet.Nodes, incomeLastNodeID)
				}
			}
		}
	}

	return nil
}
