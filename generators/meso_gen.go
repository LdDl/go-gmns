package generators

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
}

const (
	SHORTCUT_LENGTH  = 0.1
	MIN_CUT_LENGTH   = 2.0
	TOTAL_CUT_LENGTH = 2 * SHORTCUT_LENGTH * MIN_CUT_LENGTH
)

var (
	CUT_LENGTHS          = [100]float64{2.0, 8.0, 12.0, 14.0, 16.0, 18.0, 20, 22, 24, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25}
	ErrNotImplementedYet = fmt.Errorf("not implemented yet")
	ErrBadParentInfo     = fmt.Errorf("bad parent information")
	ErrBadInterface      = fmt.Errorf("bad interface")
	VERBOSE              = true
)

type macroLinkProcessing struct {
	needsOffset         bool
	offsetDirection     float64 // -1.0 or 1.0; determines which side of centerline to offset
	id                  gmns.LinkID
	offsetGeomEuclidean orb.LineString
	offsetGeom          orb.LineString
	lanesInfo           macro.LanesInfo
	lanesInfoCut        macro.LanesInfo
	lengthMetersOffset  float64

	/* For cuts */
	downstreamShortCut bool
	upstreamShortCut   bool

	downstreamIsTarget bool
	upstreamIsTarget   bool

	upstreamCutLen   float64
	downstreamCutLen float64

	offsetGeomEuclideanCut []orb.LineString
	offsetGeomCut          []orb.LineString

	/* For link generation */
	sourceMacroNodeID gmns.NodeID
	targetMacroNodeID gmns.NodeID
}

type mesoGenerationOptions struct {
	preserveConnectors bool
	standardUses       bool
	cutLengths         [100]float64
}

func GenerateMesoscopic(macroNet *macro.Net, movements movement.MovementsStorage) (*meso.Net, error) {
	options := mesoGenerationOptions{cutLengths: CUT_LENGTHS}
	return generateMesoscopic(macroNet, movements, &options)
}

func generateMesoscopic(macroNet *macro.Net, movements movement.MovementsStorage, options *mesoGenerationOptions) (*meso.Net, error) {
	err := requireDirectedMacroLinks(macroNet)
	if err != nil {
		return nil, err
	}
	err = requireMacroLinkLengths(macroNet)
	if err != nil {
		return nil, err
	}
	err = requireMacroLaneData(macroNet)
	if err != nil {
		return nil, err
	}
	err = macroNet.ValidateMovements(movements)
	if err != nil {
		return nil, err
	}
	kind, err := macroIdentifierType(macroNet)
	if err != nil {
		return nil, err
	}
	err = movements.ValidateIDType(kind)
	if err != nil {
		return nil, err
	}
	err = requireSupportedMacroDatasetUses(macroNet, options.standardUses)
	if err != nil {
		return nil, err
	}
	for _, mvmt := range movements {
		if mvmt.AllowedUses() != "" && !options.standardUses {
			return nil, fmt.Errorf("%w: authored movement allowed_uses mapping is not implemented", ErrUnsupportedGenerationInput)
		}
	}

	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Msg("Preparing mesoscopic network")
	}

	st := time.Now()
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Msg("Preparing geometries offsets")
	}

	macroLinks := macroLinksToSlice(macroNet.Links)
	needToObserve := make(map[gmns.LinkID]*macroLinkProcessing, len(macroLinks))
	hashesEuclidean := make(map[gmns.LinkID]string, len(macroLinks))
	for i := range macroLinks {
		hashesEuclidean[macroLinks[i].ID] = geomath.GeometryHash(macroLinks[i].GeomEuclidean())
	}

	// Detect bidirectional links (single-threaded to avoid race conditions)
	for i := 0; i < len(macroLinks); i++ {
		macroLink := macroLinks[i]
		macroLinkID := macroLink.ID
		if _, ok := needToObserve[macroLinkID]; ok {
			continue
		}

		reversedGeom := macroLink.GeomEuclidean().Clone()
		reversedGeom.Reverse()
		reversedGeomHash := geomath.GeometryHash(reversedGeom)
		reversedLinkExists := false

		// Compare only with subsequent links
		for j := i + 1; j < len(macroLinks); j++ {
			macroLinkCompare := macroLinks[j]
			macroLinkCompareID := macroLinkCompare.ID
			macroLinkGeomEuclidean := macroLinkCompare.GeomEuclidean()
			if len(reversedGeom) != len(macroLinkGeomEuclidean) {
				continue
			}
			hashedGeomEuclidean, ok := hashesEuclidean[macroLinkCompareID]
			if !ok {
				panic("Could not happen. Hashes are prepared for every macroscopic link")
			}
			if reversedGeomHash == hashedGeomEuclidean {
				reversedLinkExists = true
				// Both bidirectional links use the same negative offset (-1.0)
				// Since their geometries are reversed, they end up on opposite sides of centerline
				needToObserve[macroLinkID] = &macroLinkProcessing{id: macroLinkID, lanesInfo: macroLink.LanesInfo(), needsOffset: true, offsetDirection: -1.0, sourceMacroNodeID: macroLink.SourceNode(), targetMacroNodeID: macroLink.TargetNode()}
				needToObserve[macroLinkCompareID] = &macroLinkProcessing{id: macroLinkCompareID, lanesInfo: macroLinkCompare.LanesInfo(), needsOffset: true, offsetDirection: -1.0, sourceMacroNodeID: macroLinkCompare.SourceNode(), targetMacroNodeID: macroLinkCompare.TargetNode()}
				break
			}
		}
		if !reversedLinkExists {
			needToObserve[macroLinkID] = &macroLinkProcessing{id: macroLinkID, lanesInfo: macroLink.LanesInfo(), offsetDirection: -1.0, sourceMacroNodeID: macroLink.SourceNode(), targetMacroNodeID: macroLink.TargetNode()}
		}
	}

	sortedNeedToObserve := sortedLinkIDs(needToObserve)
	for _, macroLinkID := range sortedNeedToObserve {
		macroLinkProcess := needToObserve[macroLinkID]
		macroLink, ok := macroNet.Links[macroLinkID]
		if !ok {
			return nil, errors.Wrapf(macro.ErrLinkNotFound, "Offset Link ID: %v", macroLinkID)
		}
		if !macroLinkProcess.needsOffset {
			macroLinkProcess.offsetGeomEuclidean = macroLink.GeomEuclidean().Clone()
			macroLinkProcess.offsetGeom = macroLink.Geom().Clone()
			continue
		}
		offsetDistance := 2 * (float64(macroLink.MaxLanes())/2 + 0.5) * macro.LANE_WIDTH
		macroLinkProcess.offsetGeomEuclidean = geomath.OffsetCurve(macroLink.GeomEuclidean(), macroLinkProcess.offsetDirection*offsetDistance)
		macroLinkProcess.offsetGeom = geomath.LineToSpherical(macroLinkProcess.offsetGeomEuclidean)
	}
	// Update breakpoints since geometry has changed
	for _, macroLinkID := range sortedNeedToObserve {
		macroLinkProcess := needToObserve[macroLinkID]
		// Re-calcuate length for offset geometry and round to 2 decimal places
		macroLinkProcess.lengthMetersOffset = math.Round(geo.LengthHaversine(macroLinkProcess.offsetGeom)*100.0) / 100.0
		macroLink, ok := macroNet.Links[macroLinkID]
		if !ok {
			return nil, errors.Wrapf(macro.ErrLinkNotFound, "Offset Link ID: %v", macroLinkID)
		}
		// Keep rescaled breakpoints separate from the source link's slice.
		sourceChangePoints := macroLinkProcess.lanesInfo.LanesChangePoints
		// Required source lengths were validated before geometry preparation.
		lengthMeters, _ := macroLink.LengthMeters()
		macroLinkProcess.lanesInfo.LanesChangePoints = make([]float64, len(sourceChangePoints))
		for i, item := range sourceChangePoints {
			macroLinkProcess.lanesInfo.LanesChangePoints[i] = (item / lengthMeters) * macroLinkProcess.lengthMetersOffset
		}
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done preparing geometries offsets. Aggregate movements for nodes")
	}
	st = time.Now()
	macroNodesMovements := make(map[gmns.NodeID][]*movement.Movement, len(macroNet.Nodes))
	// Sort movement IDs for deterministic iteration
	sortedMovementIDs := make([]gmns.MovementID, 0, len(movements))
	for mvmtID := range movements {
		sortedMovementIDs = append(sortedMovementIDs, mvmtID)
	}
	sort.Slice(sortedMovementIDs, func(i, j int) bool {
		return sortedMovementIDs[i].Less(sortedMovementIDs[j])
	})
	for _, mvmtID := range sortedMovementIDs {
		mvmt := movements[mvmtID]
		macroNodeID := mvmt.Node()
		if _, ok := macroNet.Nodes[macroNodeID]; !ok {
			return nil, errors.Wrapf(macro.ErrNodeNotFound, "Agg movements; Node ID: %v", macroNodeID)
		}
		if _, ok := macroNodesMovements[macroNodeID]; !ok {
			macroNodesMovements[macroNodeID] = make([]*movement.Movement, 0, 1)
		}
		macroNodesMovements[macroNodeID] = append(macroNodesMovements[macroNodeID], mvmt)
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done aggregating movements. Process movements (check necessity)")
	}
	st = time.Now()
	macroNodesNeedMovement := make(map[gmns.NodeID]bool)
	// Assume that every node need movement (i.e. all nodes are intersections by default). We will filter this set later
	sortedMacroNodeIDsMain := sortedMacroNodeIDs(macroNet.Nodes)
	for _, nodeID := range sortedMacroNodeIDsMain {
		macroNodesNeedMovement[nodeID] = true
	}

	for _, nodeID := range sortedMacroNodeIDsMain {
		macroNode := macroNet.Nodes[nodeID]
		// Authored lane scopes require explicit connectors instead of merged road endpoints.
		if options.preserveConnectors {
			continue
		}
		// Preserve explicit junctions with signals, stop or yield control.
		if macroNode.ControlType().HasControl() {
			continue
		}
		incomingMacroLinks := macroNode.IncomingLinks()
		outcomingMacroLinks := macroNode.OutcomingLinks()
		if len(incomingMacroLinks) == 1 && len(outcomingMacroLinks) >= 1 {
			// Only one incoming link
			incomingMacroLinkID := incomingMacroLinks[0]
			incomingMacroLink, ok := macroNet.Links[incomingMacroLinkID]
			if !ok {
				return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 1: Incoming link ID: %v for macroscopic node %v", incomingMacroLinkID, macroNode.ID)
			}
			badAngle := true
			for j := range outcomingMacroLinks {
				outcomingMacroLinkID := outcomingMacroLinks[j]
				outcomingMacroLink, ok := macroNet.Links[outcomingMacroLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 1: Outcoming link ID: %v for macroscopic node %v", outcomingMacroLinkID, macroNode.ID)
				}
				angle := geomath.AngleBetweenLines(incomingMacroLink.GeomEuclidean(), outcomingMacroLink.GeomEuclidean())
				if angle > 0.75*math.Pi || angle < -0.75*math.Pi {
					badAngle = false
					break
				}
			}
			if !badAngle {
				continue
			}
			hasMultipleConnections := false
			outcomingMacroLinksObserved := make(map[gmns.LinkID]struct{})
			macroNodeMvmts, ok := macroNodesMovements[macroNode.ID]
			if ok {
				for j := range macroNodeMvmts {
					mvmt := macroNodeMvmts[j]
					outcomeMacroLinkID := mvmt.OutcomeLink()
					if _, ok := outcomingMacroLinksObserved[outcomeMacroLinkID]; ok {
						hasMultipleConnections = true
						break
					}
					outcomingMacroLinksObserved[outcomeMacroLinkID] = struct{}{}
				}
			}
			if hasMultipleConnections {
				continue
			}
			macroNodesNeedMovement[macroNode.ID] = false
			macroLinkProcess := needToObserve[incomingMacroLinkID]
			if macroLinkProcess == nil {
				panic("Should find incoming macroscopic link in observable data")
			}
			macroLinkProcess.downstreamShortCut = true
			macroLinkProcess.downstreamIsTarget = true
			for j := range outcomingMacroLinks {
				outcomingMacroLinkID := outcomingMacroLinks[j]
				outcomingMacroLink, ok := needToObserve[outcomingMacroLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Nested outcoming link ID: %v for macroscopic node %v", outcomingMacroLinkID, macroNode.ID)
				}
				outcomingMacroLink.upstreamShortCut = true
			}
		} else if len(incomingMacroLinks) >= 1 && len(outcomingMacroLinks) == 1 {
			// Only one outcoming link
			outcomingMacroLinkID := outcomingMacroLinks[0]
			outcomingMacroLink, ok := macroNet.Links[outcomingMacroLinkID]
			if !ok {
				return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 2: Outcoming link ID: %v for macroscopic node %v", outcomingMacroLinkID, macroNode.ID)
			}
			badAngle := true
			for j := range incomingMacroLinks {
				incomingMacroLinkID := incomingMacroLinks[j]
				incomingMacroLink, ok := macroNet.Links[incomingMacroLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 2: Incoming link ID: %v for macroscopic node %v", incomingMacroLinkID, macroNode.ID)
				}
				angle := geomath.AngleBetweenLines(incomingMacroLink.GeomEuclidean(), outcomingMacroLink.GeomEuclidean())
				if angle > 0.75*math.Pi || angle < -0.75*math.Pi {
					badAngle = false
					break
				}
			}
			if !badAngle {
				continue
			}
			hasMultipleConnections := false
			outcomingMacroLinksObserved := make(map[gmns.LinkID]struct{})
			macroNodeMvmts, ok := macroNodesMovements[macroNode.ID]
			if ok {
				for j := range macroNodeMvmts {
					mvmt := macroNodeMvmts[j]
					incomeMacroLinkID := mvmt.IncomeLink()
					if _, ok := outcomingMacroLinksObserved[incomeMacroLinkID]; ok {
						hasMultipleConnections = true
						break
					}
					outcomingMacroLinksObserved[incomeMacroLinkID] = struct{}{}
				}
			}
			if hasMultipleConnections {
				continue
			}
			macroNodesNeedMovement[macroNode.ID] = false
			macroLinkProcess := needToObserve[outcomingMacroLinkID]
			if macroLinkProcess == nil {
				panic("Should find outcoming macroscopic link in observable data")
			}
			macroLinkProcess.upstreamShortCut = true
			macroLinkProcess.upstreamIsTarget = true
			for j := range incomingMacroLinks {
				incomingMacroLinkID := incomingMacroLinks[j]
				incomingMacroLink, ok := needToObserve[incomingMacroLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Nested incoming link ID: %v for macroscopic node %v", incomingMacroLinkID, macroNode.ID)
				}
				incomingMacroLink.downstreamShortCut = true
			}
		}
	}

	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done checking necessity of movements. Process movements (calculate cuts' lengths and perform cuts)")
	}
	st = time.Now()
	for _, macroLinkID := range sortedNeedToObserve {
		macroLinkProcess := needToObserve[macroLinkID]
		macroLinkProcess.updateCutLength(options.cutLengths)
		if options.preserveConnectors {
			points := macroLinkProcess.lanesInfo.LanesChangePoints
			// Leave half of each terminal section when its interval is very short.
			macroLinkProcess.upstreamCutLen = math.Min(macroLinkProcess.upstreamCutLen, (points[1]-points[0])/2)
			macroLinkProcess.downstreamCutLen = math.Min(macroLinkProcess.downstreamCutLen, (points[len(points)-1]-points[len(points)-2])/2)
		}
		macroLinkProcess.performCut()
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done cuts. Build mesoscopic links")
	}
	st = time.Now()

	mesoNodes, mesoLinks, err := generateBaseNodesLinks(macroNet.Nodes, needToObserve, kind)
	if err != nil {
		return nil, errors.Wrap(err, "Can't generate base mesoscopic nodes and links")
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done building mesoscopic links. Connect mesoscopic links")
	}
	st = time.Now()

	err = connectMesoscopicLinks(mesoLinks, mesoNodes, macroNet.Nodes, macroNet.Links, needToObserve, macroNodesMovements, macroNodesNeedMovement, kind)
	if err != nil {
		return nil, errors.Wrap(err, "Can't prepare connections between mesoscopic links")
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done connecting links. Updating boundary type for mesoscopic nodes")
	}
	st = time.Now()

	err = updateBoundaryType(mesoNodes, macroNet.Nodes)
	if err != nil {
		return nil, errors.Wrap(err, "Can't update boundary types for mesoscopic nodes")
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done updating boundary type. Updating additional information for mesoscopic links")
	}
	st = time.Now()

	err = updateLinksProperties(mesoNodes, mesoLinks, macroNet.Nodes, macroNet.Links, movements)
	if err != nil {
		return nil, errors.Wrap(err, "Can't update additional information for mesoscopic links")
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done updating links additional information. Preparing mesoscopic network done!")
	}

	mesoNet := meso.Net{
		Nodes: mesoNodes,
		Links: mesoLinks,
	}
	err = copyMacroCatalogs(macroNet, &mesoNet)
	if err != nil {
		return nil, err
	}
	return &mesoNet, nil
}

func macroLinksToSlice(links map[gmns.LinkID]*macro.Link) []*macro.Link {
	ans := make([]*macro.Link, 0, len(links))
	for i := range links {
		ans = append(ans, links[i])
	}
	// Sort by ID for deterministic iteration order
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].ID.Less(ans[j].ID)
	})
	return ans
}

// sortedLinkIDs returns sorted slice of link IDs from a map
func sortedLinkIDs(m map[gmns.LinkID]*macroLinkProcessing) []gmns.LinkID {
	keys := make([]gmns.LinkID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})
	return keys
}

// sortedMacroNodeIDs returns sorted slice of node IDs from a map
func sortedMacroNodeIDs(m map[gmns.NodeID]*macro.Node) []gmns.NodeID {
	keys := make([]gmns.NodeID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})
	return keys
}

// sortedMesoNodeIDs returns sorted slice of meso node IDs from a map
func sortedMesoNodeIDs(m map[gmns.NodeID]*meso.Node) []gmns.NodeID {
	keys := make([]gmns.NodeID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})
	return keys
}

// sortedMesoLinkIDs returns sorted slice of meso link IDs from a map
func sortedMesoLinkIDs(m map[gmns.LinkID]*meso.Link) []gmns.LinkID {
	keys := make([]gmns.LinkID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})
	return keys
}

// sortedLinkIDsGeneric returns sorted slice of link IDs from a map with struct{} values
func sortedLinkIDsGeneric(m map[gmns.LinkID]struct{}) []gmns.LinkID {
	keys := make([]gmns.LinkID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})
	return keys
}

func (macroLinkProcess *macroLinkProcessing) updateCutLength(cutLengths [100]float64) {
	laneChangePoints := macroLinkProcess.lanesInfo.LanesChangePoints
	// Dodge potential change of number of lanes on two ends of the macroscopic link
	// @todo: Bound check for LanesChangePoints
	upstreamMaxCut := math.Max(SHORTCUT_LENGTH, laneChangePoints[1]-laneChangePoints[0]-3)
	// Defife a variable downstreamMaxCut which is the maximum length of a cut that can be made downstream of the link,
	// calculated as the maximum of the shortcutLen and the difference between the last two elements in the link.lanesChangePoints minus 3.
	// @todo: Bound check for LanesChangePoints
	downstreamMaxCut := math.Max(SHORTCUT_LENGTH, laneChangePoints[len(laneChangePoints)-1]-laneChangePoints[len(laneChangePoints)-2]-3)
	if macroLinkProcess.upstreamShortCut && macroLinkProcess.downstreamShortCut {
		if macroLinkProcess.lengthMetersOffset > TOTAL_CUT_LENGTH {
			macroLinkProcess.upstreamCutLen = SHORTCUT_LENGTH
			macroLinkProcess.downstreamCutLen = SHORTCUT_LENGTH
		} else {
			macroLinkProcess.upstreamCutLen = (macroLinkProcess.lengthMetersOffset / TOTAL_CUT_LENGTH) * SHORTCUT_LENGTH
			macroLinkProcess.downstreamCutLen = macroLinkProcess.upstreamCutLen
		}
	} else if macroLinkProcess.upstreamShortCut {
		cutIdx := 0
		cutPlaceFound := false
		for i := macroLinkProcess.lanesInfo.LanesList[len(macroLinkProcess.lanesInfo.LanesList)-1]; i >= 0; i-- {
			if macroLinkProcess.lengthMetersOffset > math.Min(downstreamMaxCut, cutLengths[i])+SHORTCUT_LENGTH+MIN_CUT_LENGTH {
				cutIdx = i
				cutPlaceFound = true
				break
			}
		}
		if cutPlaceFound {
			macroLinkProcess.upstreamCutLen = SHORTCUT_LENGTH
			macroLinkProcess.downstreamCutLen = math.Min(downstreamMaxCut, cutLengths[cutIdx])
		} else {
			downStreamCut := math.Min(downstreamMaxCut, cutLengths[0])
			totalLen := downStreamCut + SHORTCUT_LENGTH + MIN_CUT_LENGTH
			macroLinkProcess.upstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * SHORTCUT_LENGTH
			macroLinkProcess.downstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * downStreamCut
		}
	} else if macroLinkProcess.downstreamShortCut {
		cutIdx := 0
		cutPlaceFound := false
		for i := macroLinkProcess.lanesInfo.LanesList[len(macroLinkProcess.lanesInfo.LanesList)-1]; i >= 0; i-- {
			if macroLinkProcess.lengthMetersOffset > math.Min(upstreamMaxCut, cutLengths[i])+SHORTCUT_LENGTH+MIN_CUT_LENGTH {
				cutIdx = i
				cutPlaceFound = true
				break
			}
		}
		if cutPlaceFound {
			macroLinkProcess.upstreamCutLen = math.Min(upstreamMaxCut, cutLengths[cutIdx])
			macroLinkProcess.downstreamCutLen = SHORTCUT_LENGTH
		} else {
			upStreamCut := math.Min(upstreamMaxCut, cutLengths[0])
			totalLen := upStreamCut + SHORTCUT_LENGTH + MIN_CUT_LENGTH
			macroLinkProcess.upstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * cutLengths[0]
			macroLinkProcess.downstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * SHORTCUT_LENGTH
		}
	} else {
		cutIdx := 0
		cutPlaceFound := false
		for i := macroLinkProcess.lanesInfo.LanesList[len(macroLinkProcess.lanesInfo.LanesList)-1]; i >= 0; i-- {
			if macroLinkProcess.lengthMetersOffset > math.Min(upstreamMaxCut, cutLengths[i])+math.Min(downstreamMaxCut, cutLengths[i])+MIN_CUT_LENGTH {
				cutIdx = i
				cutPlaceFound = true
				break
			}
		}
		if cutPlaceFound {
			macroLinkProcess.upstreamCutLen = math.Min(upstreamMaxCut, cutLengths[cutIdx])
			macroLinkProcess.downstreamCutLen = math.Min(downstreamMaxCut, cutLengths[cutIdx])
		} else {
			upStreamCut := math.Min(upstreamMaxCut, cutLengths[0])
			downStreamCut := math.Min(downstreamMaxCut, cutLengths[0])
			totalLen := downStreamCut + upStreamCut + MIN_CUT_LENGTH
			macroLinkProcess.upstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * upStreamCut
			macroLinkProcess.downstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * downStreamCut
		}
	}
}

func (macroLinkProcess *macroLinkProcessing) performCut() {
	lanesInfo := macroLinkProcess.lanesInfo

	// Create copy for those since we will do mutations and want to keep original data
	lanesChangePoints := make([]float64, len(lanesInfo.LanesChangePoints))
	copy(lanesChangePoints, lanesInfo.LanesChangePoints)

	macroLinkProcess.lanesInfoCut.LanesList = make([]int, len(lanesInfo.LanesList))
	copy(macroLinkProcess.lanesInfoCut.LanesList, lanesInfo.LanesList)
	macroLinkProcess.lanesInfoCut.LanesChange = make([][2]int, len(lanesInfo.LanesChange))
	copy(macroLinkProcess.lanesInfoCut.LanesChange, lanesInfo.LanesChange)

	lanesChangePoints[0] = macroLinkProcess.upstreamCutLen
	lanesChangePoints[len(lanesChangePoints)-1] = macroLinkProcess.lengthMetersOffset - macroLinkProcess.downstreamCutLen
	// breakIdx := 1
	// for breakIdx = 1; breakIdx < len(lanesChangePoints); breakIdx++ {
	// 	if lanesChangePoints[breakIdx] > macroLinkProcess.upstreamCutLen {
	// 		break
	// 	}
	// }
	// lanesChangePoints = append(lanesChangePoints[breakIdx:])
	// lanesChangePoints = append([]float64{macroLinkProcess.upstreamCutLen}, lanesChangePoints...)
	// macroLinkProcess.lanesInfoCut.LanesList = macroLinkProcess.lanesInfoCut.LanesList[breakIdx-1:]
	// macroLinkProcess.lanesInfo.LanesChange = macroLinkProcess.lanesInfo.LanesChange[breakIdx-1:]

	// breakIdx = len(lanesChangePoints) - 2
	// for breakIdx := len(lanesChangePoints) - 2; breakIdx >= 0; breakIdx-- {
	// 	if macroLinkProcess.lengthMetersOffset-lanesChangePoints[breakIdx] > macroLinkProcess.downstreamCutLen {
	// 		break
	// 	}
	// }
	// lanesChangePoints = lanesChangePoints[:breakIdx+1]
	// lanesChangePoints = append(lanesChangePoints, macroLinkProcess.lengthMetersOffset-macroLinkProcess.downstreamCutLen)
	// macroLinkProcess.lanesInfoCut.LanesList = macroLinkProcess.lanesInfoCut.LanesList[:breakIdx+1]
	// macroLinkProcess.lanesInfo.LanesChange = macroLinkProcess.lanesInfo.LanesChange[:breakIdx+1]

	for i := range macroLinkProcess.lanesInfoCut.LanesList {
		start := lanesChangePoints[i]
		end := lanesChangePoints[i+1]
		geomCut := geomath.SubstringHaversine(macroLinkProcess.offsetGeom, start, end)
		geomEuclideanCut := geomath.LineToEuclidean(geomCut)
		macroLinkProcess.offsetGeomCut = append(macroLinkProcess.offsetGeomCut, geomCut)
		macroLinkProcess.offsetGeomEuclideanCut = append(macroLinkProcess.offsetGeomEuclideanCut, geomEuclideanCut)
	}
}

func generateBaseNodesLinks(macroNodes map[gmns.NodeID]*macro.Node, macroLinksProcessed map[gmns.LinkID]*macroLinkProcessing, kind gmns.IDType) (map[gmns.NodeID]*meso.Node, map[gmns.LinkID]*meso.Link, error) {
	nodeIDs, err := gmns.NewIDAllocator(kind)
	if err != nil {
		return nil, nil, err
	}
	linkIDs, err := gmns.NewIDAllocator(kind)
	if err != nil {
		return nil, nil, err
	}
	collectedMesoNodes := make(map[gmns.NodeID]*meso.Node)
	collectedMesoLinks := make(map[gmns.LinkID]*meso.Link)
	// Sort keys for deterministic iteration
	sortedKeys := sortedLinkIDs(macroLinksProcessed)
	for _, macroLinkID := range sortedKeys {
		macroLinkProcess := macroLinksProcessed[macroLinkID]

		// Prepare source mesoscopic node
		var upstreamMesoNode *meso.Node
		sourceMacroNode, ok := macroNodes[macroLinkProcess.sourceMacroNodeID]
		if !ok {
			return nil, nil, errors.Wrapf(macro.ErrNodeNotFound, "Source node ID: %v", macroLinkProcess.sourceMacroNodeID)
		}
		if sourceMacroNode.IsCentroid() {
			// @todo: handle centroids
			return nil, nil, errors.Wrap(ErrNotImplementedYet, "Prepare upstream mesoscopic node from centroid")
		} else {
			upstreamMesoNode = meso.NewNodeFrom(
				nodeIDs.NextNodeID(),
				meso.WithPointGeom(macroLinkProcess.offsetGeomCut[0][0]), // No explicit copy or clone method since Point is not slice, but array
				meso.WithPointEuclideanGeom(macroLinkProcess.offsetGeomEuclideanCut[0][0]),
				meso.WithPointMacroNodeID(macroLinkProcess.sourceMacroNodeID),
				meso.WithPointMacroLinkID(gmns.LinkID{}),
				meso.WithActivityLinkType(sourceMacroNode.ActivityLinkType()),
				meso.WithBoundaryType(types.BOUNDARY_NONE),
			)
			zoneID, hasZone := sourceMacroNode.Zone()
			if hasZone {
				meso.WithMacroZone(zoneID)(upstreamMesoNode)
			}
			collectedMesoNodes[upstreamMesoNode.ID] = upstreamMesoNode
		}

		// Prepare mesoscopic link and target mesoscopic node
		var downstreamMesoNode *meso.Node
		targetMacroNode, ok := macroNodes[macroLinkProcess.targetMacroNodeID]
		if !ok {
			return nil, nil, errors.Wrapf(macro.ErrNodeNotFound, "Target node ID: %v", macroLinkProcess.targetMacroNodeID)
		}
		segmentsToCut := len(macroLinkProcess.lanesInfoCut.LanesList)
		upstreamMesoNodeID := upstreamMesoNode.ID
		for segmentIdx := 0; segmentIdx < segmentsToCut; segmentIdx++ {
			// Prepare mesoscopic node
			if targetMacroNode.IsCentroid() && segmentIdx == segmentsToCut-1 {
				return nil, nil, errors.Wrap(ErrNotImplementedYet, "Prepare downstream mesoscopic node from centroid")
			} else {
				macroNodeID := gmns.NodeID{}
				macroLinkID := macroLinkProcess.id
				var zoneID gmns.NodeID
				hasZone := false
				activityLinkType := types.LINK_UNDEFINED
				if segmentIdx == segmentsToCut-1 {
					macroNodeID = macroLinkProcess.targetMacroNodeID
					macroLinkID = gmns.LinkID{}
					zoneID, hasZone = targetMacroNode.Zone()
					activityLinkType = targetMacroNode.ActivityLinkType()
				}
				downstreamMesoNode = meso.NewNodeFrom(
					nodeIDs.NextNodeID(),
					meso.WithPointGeom(macroLinkProcess.offsetGeomCut[segmentIdx][len(macroLinkProcess.offsetGeomCut[segmentIdx])-1]), // No explicit copy or clone method since Point is not slice, but array
					meso.WithPointEuclideanGeom(macroLinkProcess.offsetGeomEuclideanCut[segmentIdx][len(macroLinkProcess.offsetGeomEuclideanCut[segmentIdx])-1]),
					meso.WithPointMacroNodeID(macroNodeID),
					meso.WithPointMacroLinkID(macroLinkID),
					meso.WithActivityLinkType(activityLinkType),
					meso.WithBoundaryType(types.BOUNDARY_NONE),
				)
				if hasZone {
					meso.WithMacroZone(zoneID)(downstreamMesoNode)
				}
				collectedMesoNodes[downstreamMesoNode.ID] = downstreamMesoNode
			}

			lastMesoLinkID := linkIDs.NextLinkID()
			mesoLink := meso.NewLinkFrom(
				lastMesoLinkID,
				upstreamMesoNodeID,
				downstreamMesoNode.ID,
				meso.WithLanesNum(macroLinkProcess.lanesInfoCut.LanesList[segmentIdx]),
				meso.WithLanesChange(macroLinkProcess.lanesInfoCut.LanesChange[segmentIdx]),
				meso.WithLineGeom(macroLinkProcess.offsetGeomCut[segmentIdx].Clone()),
				meso.WithLineGeomEuclidean(macroLinkProcess.offsetGeomEuclideanCut[segmentIdx].Clone()),
				meso.WithLineMacroLinkID(macroLinkProcess.id),
				meso.WithSegmentIdx(segmentIdx),
				meso.WithMovementID(gmns.MovementID{}),
				meso.WithLineMacroNodeID(gmns.NodeID{}),
				meso.WithLengthMeters(geo.LengthHaversine(macroLinkProcess.offsetGeomCut[segmentIdx])),
			)
			meso.WithOutcomingLinks(lastMesoLinkID)(collectedMesoNodes[upstreamMesoNodeID])
			meso.WithIncomingLinks(lastMesoLinkID)(collectedMesoNodes[downstreamMesoNode.ID])

			// Prepare mesoscopic link
			collectedMesoLinks[mesoLink.ID] = mesoLink
			upstreamMesoNodeID = downstreamMesoNode.ID // This must be done since current upstream node is downstream node for next segment
		}
	}
	return collectedMesoNodes, collectedMesoLinks, nil
}

func connectMesoscopicLinks(
	mesoLinks map[gmns.LinkID]*meso.Link,
	mesoNodes map[gmns.NodeID]*meso.Node,
	macroNodes map[gmns.NodeID]*macro.Node,
	macroLinks map[gmns.LinkID]*macro.Link,
	macroLinksProcessed map[gmns.LinkID]*macroLinkProcessing,
	macroNodesMovements map[gmns.NodeID][]*movement.Movement,
	macroNodesNeedMovement map[gmns.NodeID]bool,
	kind gmns.IDType,
) error {
	linkIDs, err := gmns.NewIDAllocator(kind)
	if err != nil {
		return err
	}
	for linkID := range mesoLinks {
		err := linkIDs.ReserveLinkID(linkID)
		if err != nil {
			return err
		}
	}

	// Collect mesoscopic links for parent macroscopic links
	macroLinkMesoLinks := make(map[gmns.LinkID][]*meso.Link)
	for i := range mesoLinks {
		macroLinkID := mesoLinks[i].MacroLink()
		if _, ok := macroLinkMesoLinks[macroLinkID]; !ok {
			macroLinkMesoLinks[macroLinkID] = make([]*meso.Link, 0, 1)
		}
		macroLinkMesoLinks[macroLinkID] = append(macroLinkMesoLinks[macroLinkID], mesoLinks[i])
	}
	for i := range macroLinkMesoLinks {
		macroLinkData := macroLinkMesoLinks[i]
		/* Sort mesoscopic links by its segment number in parent macroscopic link */
		// @todo: We can achieve better perfomance if does sort during populating data
		sort.Slice(macroLinkData, func(i, j int) bool {
			return macroLinkData[i].SegmentIdx() < macroLinkData[j].SegmentIdx()
		})
	}

	// Start main loop for finding connections between mesoscopic links
	sortedMacroNodes := sortedMacroNodeIDs(macroNodes)
	for _, macroNodeID := range sortedMacroNodes {
		// macroNode := macroNodes[macroNodeID]
		macroNodeMvmts, ok := macroNodesMovements[macroNodeID]
		if !ok {
			continue
		}
		for j := range macroNodeMvmts {
			mvmt := macroNodeMvmts[j]
			incomeMacroLinkID := mvmt.IncomeLink()
			outcomeMacroLinkID := mvmt.OutcomeLink()
			incomingMacroLink, ok := macroLinks[incomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find macro link for further connection: %v", incomeMacroLinkID)
			}
			outcomingMacroLink, ok := macroLinks[outcomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find macro link for further connection: %v", outcomeMacroLinkID)
			}

			incomingMacroLinkProcessed, ok := macroLinksProcessed[incomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find processed macro link for further connection: %v", incomeMacroLinkID)
			}
			outcomingMacroLinkProcessed, ok := macroLinksProcessed[outcomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find processed macro link for further connection: %v", outcomeMacroLinkID)
			}

			incomingMesolinks := macroLinkMesoLinks[incomingMacroLink.ID]
			if len(incomingMesolinks) == 0 {
				panic("No mesoscopic links for incoming macro link")
			}
			outcomingMesolinks := macroLinkMesoLinks[outcomingMacroLink.ID]
			if len(outcomingMesolinks) == 0 {
				panic("No mesoscopic links for outcoming macro link")
			}

			incomingMesoLink := incomingMesolinks[len(incomingMesolinks)-1]
			incomingMesoLinkGeom := incomingMesoLink.Geom()
			incomingMesoLinkGeomEuclidean := incomingMesoLink.GeomEuclidean()
			outcomingMesoLink := outcomingMesolinks[0]
			outcomingMesoLinkGeom := outcomingMesoLink.Geom()
			outcomingMesoLinkGeomEuclidean := outcomingMesoLink.GeomEuclidean()

			geom := orb.LineString{incomingMesoLinkGeom[len(incomingMesoLinkGeom)-1], outcomingMesoLinkGeom[0]}
			geomEuclidean := orb.LineString{incomingMesoLinkGeomEuclidean[len(incomingMesoLinkGeomEuclidean)-1], outcomingMesoLinkGeomEuclidean[0]}
			if macroNodesNeedMovement[macroNodeID] {
				lastMesoLinkID := linkIDs.NextLinkID()
				sourceMesoNodeID := incomingMesoLink.TargetNode()
				targetMesoNodeID := outcomingMesoLink.SourceNode()
				mesoLink := meso.NewLinkFrom(
					lastMesoLinkID,
					sourceMesoNodeID,
					targetMesoNodeID,
					meso.WithLanesNum(mvmt.LanesNum()),
					meso.WithLineGeom(geom),
					meso.WithLineGeomEuclidean(geomEuclidean),
					meso.WithLineMacroLinkID(gmns.LinkID{}),
					meso.WithIsConnection(true),
					meso.WithMovementID(mvmt.ID),
					meso.WithLineMacroNodeID(macroNodeID),
					meso.WithLengthMeters(geo.LengthHaversine(geom)),
					meso.WithMovementCompositeType(mvmt.MvmtTextID()),
					meso.WithMovementMesoLinkIncome(incomingMesoLink.ID),
					meso.WithMovementMesoLinkOutcome(outcomingMesoLink.ID),
					meso.WithMovementIncomeLaneStartSeqID(mvmt.StartIncomeLaneSeqID()),
					meso.WithMovementOutcomeLaneStartSeqID(mvmt.StartOutcomeLaneSeqID()),
				)
				meso.WithOutcomingLinks(lastMesoLinkID)(mesoNodes[sourceMesoNodeID])
				meso.WithIncomingLinks(lastMesoLinkID)(mesoNodes[targetMesoNodeID])
				// Prepare mesoscopic link
				mesoLinks[mesoLink.ID] = mesoLink
			} else {
				if incomingMacroLinkProcessed.downstreamIsTarget && !outcomingMacroLinkProcessed.upstreamIsTarget {
					// remove incoming micro nodes and links of outcomingMesoLink, then connect to incomingMesoLink
					incomingMesoLinkTargetNodeID := incomingMesoLink.TargetNode()
					outcomingMesoLinkSourceNodeID := outcomingMesoLink.SourceNode()

					meso.WithSourceNodeID(incomingMesoLinkTargetNodeID)(outcomingMesoLink)
					meso.WithLineGeom(append(orb.LineString{incomingMesoLinkGeom[len(incomingMesoLinkGeom)-1]}, outcomingMesoLinkGeom[1:]...))(outcomingMesoLink)
					meso.WithLineGeomEuclidean(append(orb.LineString{incomingMesoLinkGeomEuclidean[len(incomingMesoLinkGeomEuclidean)-1]}, outcomingMesoLinkGeomEuclidean[1:]...))(outcomingMesoLink)
					delete(mesoNodes, outcomingMesoLinkSourceNodeID)
				} else if !incomingMacroLinkProcessed.downstreamIsTarget && outcomingMacroLinkProcessed.upstreamIsTarget {
					// remove outgoing micro nodes and links of incomingMesoLink, then connect to outcomingMesoLink
					incomingMesoLinkTargetNodeID := incomingMesoLink.TargetNode()
					outcomingMesoLinkSourceNodeID := outcomingMesoLink.SourceNode()

					meso.WithTargetNodeID(outcomingMesoLinkSourceNodeID)(incomingMesoLink)
					meso.WithLineGeom(append(incomingMesoLinkGeom[:len(incomingMesoLinkGeom)-1], outcomingMesoLinkGeom[0]))(incomingMesoLink)
					meso.WithLineGeomEuclidean(append(incomingMesoLinkGeomEuclidean[:len(incomingMesoLinkGeomEuclidean)-1], outcomingMesoLinkGeomEuclidean[0]))(incomingMesoLink)
					delete(mesoNodes, incomingMesoLinkTargetNodeID)
				}
			}
		}
	}
	return nil
}

func updateBoundaryType(mesoNodes map[gmns.NodeID]*meso.Node, macroNodes map[gmns.NodeID]*macro.Node) error {
	for i := range mesoNodes {
		mesoNode := mesoNodes[i]
		macroNodeID := mesoNode.MacroNode()
		if macroNodeID.IsZero() && mesoNode.MacroLink().IsZero() {
			return errors.Wrapf(ErrBadParentInfo, "Neither macroscopic link nor node for meso node: %v", mesoNode.ID)
		}
		if mesoNode.MacroNode().IsZero() {
			meso.WithBoundaryType(types.BOUNDARY_NONE)(mesoNode)
			continue
		}
		macroNode, ok := macroNodes[macroNodeID]
		if !ok {
			return errors.Wrapf(macro.ErrNodeNotFound, "Can't find macroscopic node with id %v for mesoscopic node %v", macroNodeID, mesoNode.ID)
		}
		meso.WithNodeName(macroNode.Name())(mesoNode)
		meso.WithNodeType(macroNode.NodeType())(mesoNode)
		meso.WithNodeControlType(macroNode.ControlType())(mesoNode)
		altitude, present := macroNode.AltitudeMeters()
		if present {
			meso.WithAltitudeMeters(altitude)(mesoNode)
		}
		macroNodeBoundaryType := macroNode.BoundaryType()
		if macroNodeBoundaryType != types.BOUNDARY_INCOME_OUTCOME {
			meso.WithBoundaryType(macroNodeBoundaryType)(mesoNode)
			continue
		}
		if mesoNode.IncomingLinks().Len() != 0 {
			meso.WithBoundaryType(types.BOUNDARY_INCOME_ONLY)(mesoNode)
			continue
		}
		meso.WithBoundaryType(types.BOUNDARY_OUTCOME_ONLY)(mesoNode)
	}
	return nil
}

func updateLinksProperties(
	mesoNodes map[gmns.NodeID]*meso.Node,
	mesoLinks map[gmns.LinkID]*meso.Link,
	macroNodes map[gmns.NodeID]*macro.Node,
	macroLinks map[gmns.LinkID]*macro.Link,
	movements movement.MovementsStorage,
) error {
	movementMesoLinks := make(map[gmns.LinkID]struct{})
	for i := range mesoLinks {
		mesoLink := mesoLinks[i]
		macroNodeID := mesoLink.MacroNode()
		macroLinkID := mesoLink.MacroLink()

		if macroNodeID.IsZero() && macroLinkID.IsZero() {
			return errors.Wrapf(ErrBadParentInfo, "Neither macroscopic link nor node for mesoscopic link: %v", mesoLink.ID)
		}

		if mesoLink.MacroNode().IsZero() {
			// Inherit macroscopic link properties
			macroLink, ok := macroLinks[macroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find macroscopic link with id %v for mesoscopic link %v", macroLinkID, mesoLink.ID)
			}
			copyMacroRoadAttributes(macroLink, mesoLink)
			meso.WithLinkType(macroLink.LinkType())(mesoLink)
			freeSpeed, hasFreeSpeed := macroLink.FreeSpeed()
			if hasFreeSpeed {
				meso.WithFreeSpeed(freeSpeed)(mesoLink)
			}
			capacity, hasCapacity := macroLink.Capacity()
			if hasCapacity {
				meso.WithCapacity(capacity)(mesoLink)
			}
			meso.WithAllowedAgentTypes(macroLink.AllowedAgentTypes())(mesoLink)
			// Reset contrl type property to default
			meso.WithControlType(types.CONTROL_TYPE_NOT_SIGNAL)(mesoLink)
			continue
		}

		// Collect movement-based links and inherit macroscopic link properties later
		movementMesoLinks[mesoLink.ID] = struct{}{}
		// Inherit macroscopic node properties
		macroNode, ok := macroNodes[macroNodeID]
		if !ok {
			return errors.Wrapf(macro.ErrNodeNotFound, "Can't find macroscopic node with id %v for mesoscopic link %v", macroNodeID, mesoLink.ID)
		}
		meso.WithControlType(macroNode.ControlType())(mesoLink)

		movementID := mesoLink.Movement()
		if movementID.IsZero() {
			return errors.Wrapf(ErrBadParentInfo, "Should have movement ID for mesoscopic link: %v", mesoLink.ID)
		}
		mvmt, ok := movements[movementID]
		if !ok {
			return errors.Wrapf(movement.ErrMvmtNotFound, "Can't find movement with ID %v for mesoscopic link %v", movementID, mesoLink.ID)
		}
		meso.WithMovementCompositeType(mvmt.MvmtTextID())(mesoLink)
	}

	// Inherit macroscopic link properties for movement links
	for mesoLinkID := range movementMesoLinks {
		mesoLink, ok := mesoLinks[mesoLinkID]
		if !ok {
			return errors.Wrapf(meso.ErrLinkNotFound, "Can't find mesoscopic link %v while processing movement links", mesoLinkID)
		}
		sourceMesoNodeID := mesoLink.SourceNode()
		sourceMesoNode, ok := mesoNodes[sourceMesoNodeID]
		if !ok {
			return errors.Wrapf(meso.ErrLinkNotFound, "Can't find source node %v for mesoscopic link %v while processing movement links", sourceMesoNodeID, mesoLinkID)
		}
		incomingMesoLinks := sourceMesoNode.IncomingLinks()
		if incomingMesoLinks.Len() == 0 {
			// @todo: should make warning?
			continue
		}
		upstreamMesoLinkIDRef := incomingMesoLinks.Front()
		if upstreamMesoLinkIDRef == nil {
			// No elements. Should be catched by if-statement above
			continue
		}
		var upstreamMesoLinkID gmns.LinkID
		switch id := upstreamMesoLinkIDRef.Key.(type) {
		case gmns.LinkID:
			upstreamMesoLinkID = id
		default:
			fmt.Println(id)
			return errors.Wrapf(ErrBadInterface, "Can't get correct type for upstream for source mesoscopic node %v of mesoscopic link %v", sourceMesoNode.ID, mesoLinkID)
		}
		upstreamMesoLink, ok := mesoLinks[upstreamMesoLinkID]
		if !ok {
			return errors.Wrapf(meso.ErrLinkNotFound, "Can't find mesoscopic link %v while processing movement links", mesoLinkID)
		}
		// Inherit upstream mesoscopic link properties
		meso.WithLinkType(upstreamMesoLink.LinkType())(mesoLink)
		freeSpeed, hasFreeSpeed := upstreamMesoLink.FreeSpeed()
		if hasFreeSpeed {
			meso.WithFreeSpeed(freeSpeed)(mesoLink)
		}
		capacity, hasCapacity := upstreamMesoLink.Capacity()
		if hasCapacity {
			meso.WithCapacity(capacity)(mesoLink)
		}
		meso.WithAllowedAgentTypes(upstreamMesoLink.AllowedAgentTypes())(mesoLink)
	}
	return nil
}
