package generators

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
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

func GenerateMesoscopic(macroNet *macro.Net, movements movement.MovementsStorage) (*meso.Net, error) {
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

	// Result map with thread-safe access
	var mu sync.Mutex

	processBatch := func(start, end int, wg *sync.WaitGroup) {
		defer wg.Done()
		localNeedToObserve := make(map[gmns.LinkID]*macroLinkProcessing) // Batch-local map
		for i := start; i < end; i++ {
			macroLink := macroLinks[i]
			macroLinkID := macroLink.ID
			if _, ok := localNeedToObserve[macroLinkID]; ok {
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
					localNeedToObserve[macroLinkID] = &macroLinkProcessing{id: macroLinkID, lanesInfo: macroLink.LanesInfo(), needsOffset: true, sourceMacroNodeID: macroLink.SourceNode(), targetMacroNodeID: macroLink.TargetNode()}
					localNeedToObserve[macroLinkCompareID] = &macroLinkProcessing{id: macroLinkCompareID, lanesInfo: macroLinkCompare.LanesInfo(), needsOffset: true, sourceMacroNodeID: macroLinkCompare.SourceNode(), targetMacroNodeID: macroLinkCompare.TargetNode()}
					break
				}
			}
			if !reversedLinkExists {
				localNeedToObserve[macroLinkID] = &macroLinkProcessing{id: macroLinkID, lanesInfo: macroLink.LanesInfo(), sourceMacroNodeID: macroLink.SourceNode(), targetMacroNodeID: macroLink.TargetNode()}
			}
		}
		// Merge batch-local results into the outside map
		mu.Lock()
		for k, v := range localNeedToObserve {
			needToObserve[k] = v
		}
		mu.Unlock()
	}

	// Spawn worker goroutines
	numWorkers := runtime.NumCPU()
	batchSize := (len(macroLinks) + numWorkers - 1) / numWorkers // Divide equally
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		start := w * batchSize
		end := start + batchSize
		if end > len(macroLinks) {
			end = len(macroLinks)
		}
		wg.Add(1)
		go processBatch(start, end, &wg)
	}
	wg.Wait()

	for macroLinkID := range needToObserve {
		macroLinkProcess := needToObserve[macroLinkID]
		macroLink, ok := macroNet.Links[macroLinkID]
		if !ok {
			return nil, errors.Wrapf(macro.ErrLinkNotFound, "Offset Link ID: %d", macroLinkID)
		}
		if !macroLinkProcess.needsOffset {
			macroLinkProcess.offsetGeomEuclidean = macroLink.GeomEuclidean().Clone()
			macroLinkProcess.offsetGeom = macroLink.Geom().Clone()
			continue
		}
		offsetDistance := 2 * (float64(macroLink.MaxLanes())/2 + 0.5) * macro.LANE_WIDTH
		macroLinkProcess.offsetGeomEuclidean = geomath.OffsetCurve(macroLink.GeomEuclidean(), -offsetDistance)
		macroLinkProcess.offsetGeom = geomath.LineToSpherical(macroLinkProcess.offsetGeomEuclidean)
	}
	// Update breakpoints since geometry has changed
	for macroLinkID := range needToObserve {
		macroLinkProcess := needToObserve[macroLinkID]
		// Re-calcuate length for offset geometry and round to 2 decimal places
		macroLinkProcess.lengthMetersOffset = math.Round(geo.LengthHaversine(macroLinkProcess.offsetGeom)*100.0) / 100.0
		macroLink, ok := macroNet.Links[macroLinkID]
		if !ok {
			return nil, errors.Wrapf(macro.ErrLinkNotFound, "Offset Link ID: %d", macroLinkID)
		}
		for i, item := range macroLinkProcess.lanesInfo.LanesChangePoints {
			macroLinkProcess.lanesInfo.LanesChangePoints[i] = (item / macroLink.LengthMeters()) * macroLinkProcess.lengthMetersOffset
		}
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done preparing geometries offsets. Aggregate movements for nodes")
	}
	st = time.Now()
	macroNodesMovements := make(map[gmns.NodeID][]*movement.Movement, len(macroNet.Nodes))
	for i := range movements {
		mvmt := movements[i]
		macroNodeID := mvmt.MacroNode()
		if _, ok := macroNet.Nodes[macroNodeID]; !ok {
			return nil, errors.Wrapf(macro.ErrNodeNotFound, "Agg movements; Node ID: %d", macroNodeID)
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
	for i := range macroNet.Nodes {
		macroNodesNeedMovement[macroNet.Nodes[i].ID] = true
	}

	for i := range macroNet.Nodes {
		macroNode := macroNet.Nodes[i]
		if macroNode.ControlType() == types.CONTROL_TYPE_IS_SIGNAL {
			continue
		}
		incomingMacroLinks := macroNode.IncomingLinks()
		outcomingMacroLinks := macroNode.OutcomingLinks()
		if len(incomingMacroLinks) == 1 && len(outcomingMacroLinks) >= 1 {
			// Only one incoming link
			incomingMacroLinkID := incomingMacroLinks[0]
			incomingMacroLink, ok := macroNet.Links[incomingMacroLinkID]
			if !ok {
				return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 1: Incoming link ID: %d for macroscopic node %d", incomingMacroLinkID, macroNode.ID)
			}
			badAngle := true
			for j := range outcomingMacroLinks {
				outcomingMacroLinkID := outcomingMacroLinks[j]
				outcomingMacroLink, ok := macroNet.Links[outcomingMacroLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 1: Outcoming link ID: %d for macroscopic node %d", outcomingMacroLinkID, macroNode.ID)
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
					outcomeMacroLinkID := mvmt.OutcomeMacroLink()
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
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Nested outcoming link ID: %d for macroscopic node %d", outcomingMacroLinkID, macroNode.ID)
				}
				outcomingMacroLink.upstreamShortCut = true
			}
		} else if len(incomingMacroLinks) >= 1 && len(outcomingMacroLinks) == 1 {
			// Only one outcoming link
			outcomingMacroLinkID := outcomingMacroLinks[0]
			outcomingMacroLink, ok := macroNet.Links[outcomingMacroLinkID]
			if !ok {
				return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 2: Outcoming link ID: %d for macroscopic node %d", outcomingMacroLinkID, macroNode.ID)
			}
			badAngle := true
			for j := range incomingMacroLinks {
				incomingMacroLinkID := incomingMacroLinks[j]
				incomingMacroLink, ok := macroNet.Links[incomingMacroLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Case 2: Incoming link ID: %d for macroscopic node %d", incomingMacroLinkID, macroNode.ID)
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
					incomeMacroLinkID := mvmt.IncomeMacroLink()
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
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Nested incoming link ID: %d for macroscopic node %d", incomingMacroLinkID, macroNode.ID)
				}
				incomingMacroLink.downstreamShortCut = true
			}
		}
	}

	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done checking necessity of movements. Process movements (calculate cuts' lengths and perform cuts)")
	}
	st = time.Now()
	for macroLinkID := range needToObserve {
		macroLinkProcess := needToObserve[macroLinkID]
		macroLinkProcess.updateCutLength()
		macroLinkProcess.performCut()
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done cuts. Build mesoscopic links")
	}
	st = time.Now()

	mesoNodes, mesoLinks, err := generateBaseNodesLinks(macroNet.Nodes, needToObserve)
	if err != nil {
		return nil, errors.Wrap(err, "Can't generate base mesoscopic nodes and links")
	}
	if VERBOSE {
		log.Info().Str("scope", "gen_meso").Float64("elapsed", time.Since(st).Seconds()).Msg("Done building mesoscopic links. Connect mesoscopic links")
	}
	st = time.Now()

	err = connectMesoscopicLinks(mesoLinks, mesoNodes, macroNet.Nodes, macroNet.Links, needToObserve, macroNodesMovements, macroNodesNeedMovement)
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

	err = updateLinksProperties(mesoNodes, mesoLinks, macroNet.Nodes, macroNet.Links, movements, macroNodesMovements)
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
	return &mesoNet, nil
}

func macroLinksToSlice(links map[gmns.LinkID]*macro.Link) []*macro.Link {
	ans := make([]*macro.Link, 0, len(links))
	for i := range links {
		ans = append(ans, links[i])
	}
	return ans
}

func (macroLinkProcess *macroLinkProcessing) updateCutLength() {
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
			if macroLinkProcess.lengthMetersOffset > math.Min(downstreamMaxCut, CUT_LENGTHS[i])+SHORTCUT_LENGTH+MIN_CUT_LENGTH {
				cutIdx = i
				cutPlaceFound = true
				break
			}
		}
		if cutPlaceFound {
			macroLinkProcess.upstreamCutLen = SHORTCUT_LENGTH
			macroLinkProcess.downstreamCutLen = math.Min(downstreamMaxCut, CUT_LENGTHS[cutIdx])
		} else {
			downStreamCut := math.Min(downstreamMaxCut, CUT_LENGTHS[0])
			totalLen := downStreamCut + SHORTCUT_LENGTH + MIN_CUT_LENGTH
			macroLinkProcess.upstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * SHORTCUT_LENGTH
			macroLinkProcess.downstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * downStreamCut
		}
	} else if macroLinkProcess.downstreamShortCut {
		cutIdx := 0
		cutPlaceFound := false
		for i := macroLinkProcess.lanesInfo.LanesList[len(macroLinkProcess.lanesInfo.LanesList)-1]; i >= 0; i-- {
			if macroLinkProcess.lengthMetersOffset > math.Min(upstreamMaxCut, CUT_LENGTHS[i])+SHORTCUT_LENGTH+MIN_CUT_LENGTH {
				cutIdx = i
				cutPlaceFound = true
				break
			}
		}
		if cutPlaceFound {
			macroLinkProcess.upstreamCutLen = math.Min(upstreamMaxCut, CUT_LENGTHS[cutIdx])
			macroLinkProcess.downstreamCutLen = SHORTCUT_LENGTH
		} else {
			upStreamCut := math.Min(upstreamMaxCut, CUT_LENGTHS[0])
			totalLen := upStreamCut + SHORTCUT_LENGTH + MIN_CUT_LENGTH
			macroLinkProcess.upstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * CUT_LENGTHS[0]
			macroLinkProcess.downstreamCutLen = (macroLinkProcess.lengthMetersOffset / totalLen) * SHORTCUT_LENGTH
		}
	} else {
		cutIdx := 0
		cutPlaceFound := false
		for i := macroLinkProcess.lanesInfo.LanesList[len(macroLinkProcess.lanesInfo.LanesList)-1]; i >= 0; i-- {
			if macroLinkProcess.lengthMetersOffset > math.Min(upstreamMaxCut, CUT_LENGTHS[i])+math.Min(downstreamMaxCut, CUT_LENGTHS[i])+MIN_CUT_LENGTH {
				cutIdx = i
				cutPlaceFound = true
				break
			}
		}
		if cutPlaceFound {
			macroLinkProcess.upstreamCutLen = math.Min(upstreamMaxCut, CUT_LENGTHS[cutIdx])
			macroLinkProcess.downstreamCutLen = math.Min(downstreamMaxCut, CUT_LENGTHS[cutIdx])
		} else {
			upStreamCut := math.Min(upstreamMaxCut, CUT_LENGTHS[0])
			downStreamCut := math.Min(downstreamMaxCut, CUT_LENGTHS[0])
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

func generateBaseNodesLinks(macroNodes map[gmns.NodeID]*macro.Node, macroLinksProcessed map[gmns.LinkID]*macroLinkProcessing) (map[gmns.NodeID]*meso.Node, map[gmns.LinkID]*meso.Link, error) {
	lastMesoLinkID := gmns.LinkID(0)
	expandedMesoNodes := make(map[gmns.NodeID]int)
	collectedMesoNodes := make(map[gmns.NodeID]*meso.Node)
	collectedMesoLinks := make(map[gmns.LinkID]*meso.Link)
	for macroLinkID := range macroLinksProcessed {
		macroLinkProcess := macroLinksProcessed[macroLinkID]

		// Prepare source mesoscopic node
		var upstreamMesoNode *meso.Node
		sourceMacroNode, ok := macroNodes[macroLinkProcess.sourceMacroNodeID]
		if !ok {
			return nil, nil, errors.Wrapf(macro.ErrNodeNotFound, "Source node ID: %d", macroLinkProcess.sourceMacroNodeID)
		}
		if sourceMacroNode.IsCentroid() {
			// @todo: handle centroids
			return nil, nil, errors.Wrap(ErrNotImplementedYet, "Prepare upstream mesoscopic node from centroid")
		} else {
			expNodesNum, ok := expandedMesoNodes[macroLinkProcess.sourceMacroNodeID]
			if !ok {
				expandedMesoNodes[macroLinkProcess.sourceMacroNodeID] = 0
			}
			expandedMesoNodes[macroLinkProcess.sourceMacroNodeID] += 1
			upstreamMesoNode = meso.NewNodeFrom(
				macroLinkProcess.sourceMacroNodeID*100+gmns.NodeID(expNodesNum),
				meso.WithPointGeom(macroLinkProcess.offsetGeomCut[0][0]), // No explicit copy or clone method since Point is not slice, but array
				meso.WithPointEuclideanGeom(macroLinkProcess.offsetGeomEuclideanCut[0][0]),
				meso.WithPointMacroNodeID(macroLinkProcess.sourceMacroNodeID),
				meso.WithPointMacroLinkID(-1),
				meso.WithMacroZone(sourceMacroNode.Zone()),
				meso.WithActivityLinkType(sourceMacroNode.ActivityLinkType()),
				meso.WithBoundaryType(types.BOUNDARY_NONE),
			)
			collectedMesoNodes[upstreamMesoNode.ID] = upstreamMesoNode
		}

		// Prepare mesoscopic link and target mesoscopic node
		var downstreamMesoNode *meso.Node
		targetMacroNode, ok := macroNodes[macroLinkProcess.targetMacroNodeID]
		if !ok {
			return nil, nil, errors.Wrapf(macro.ErrNodeNotFound, "Target node ID: %d", macroLinkProcess.targetMacroNodeID)
		}
		segmentsToCut := len(macroLinkProcess.lanesInfoCut.LanesList)
		upstreamMesoNodeID := upstreamMesoNode.ID
		for segmentIdx := 0; segmentIdx < segmentsToCut; segmentIdx++ {
			// Prepare mesoscopic node
			if targetMacroNode.IsCentroid() && segmentIdx == segmentsToCut-1 {
				return nil, nil, errors.Wrap(ErrNotImplementedYet, "Prepare downstream mesoscopic node from centroid")
			} else {
				expNodesNum, ok := expandedMesoNodes[macroLinkProcess.targetMacroNodeID]
				if !ok {
					expandedMesoNodes[macroLinkProcess.targetMacroNodeID] = 0
				}
				expandedMesoNodes[macroLinkProcess.targetMacroNodeID] += 1
				macroNodeID := gmns.NodeID(-1)
				macroLinkID := macroLinkProcess.id
				zoneID := gmns.NodeID(-1)
				activityLinkType := types.LINK_UNDEFINED
				if segmentIdx == segmentsToCut-1 {
					macroNodeID = macroLinkProcess.targetMacroNodeID
					macroLinkID = gmns.LinkID(-1)
					zoneID = targetMacroNode.Zone()
					activityLinkType = targetMacroNode.ActivityLinkType()
				}
				downstreamMesoNode = meso.NewNodeFrom(
					macroLinkProcess.targetMacroNodeID*100+gmns.NodeID(expNodesNum),
					meso.WithPointGeom(macroLinkProcess.offsetGeomCut[segmentIdx][len(macroLinkProcess.offsetGeomCut[segmentIdx])-1]), // No explicit copy or clone method since Point is not slice, but array
					meso.WithPointEuclideanGeom(macroLinkProcess.offsetGeomEuclideanCut[segmentIdx][len(macroLinkProcess.offsetGeomEuclideanCut[segmentIdx])-1]),
					meso.WithPointMacroNodeID(macroNodeID),
					meso.WithPointMacroLinkID(macroLinkID),
					meso.WithMacroZone(zoneID),
					meso.WithActivityLinkType(activityLinkType),
					meso.WithBoundaryType(types.BOUNDARY_NONE),
				)
				collectedMesoNodes[downstreamMesoNode.ID] = downstreamMesoNode
			}

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
				meso.WithMovementID(-1),
				meso.WithLineMacroNodeID(-1),
				meso.WithLengthMeters(geo.LengthHaversine(macroLinkProcess.offsetGeomCut[segmentIdx])),
			)
			meso.WithOutcomingLinks(lastMesoLinkID)(collectedMesoNodes[upstreamMesoNodeID])
			meso.WithIncomingLinks(lastMesoLinkID)(collectedMesoNodes[downstreamMesoNode.ID])

			// Prepare mesoscopic link
			collectedMesoLinks[mesoLink.ID] = mesoLink
			lastMesoLinkID += 1
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
) error {
	lastMesoLinkID := gmns.LinkID(0)

	// Find max ID (for further links creating)
	for _, mesoLink := range mesoLinks {
		if mesoLink.ID > lastMesoLinkID {
			lastMesoLinkID = mesoLink.ID
		}
	}
	lastMesoLinkID++

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
	for macroNodeID := range macroNodes {
		// macroNode := macroNodes[macroNodeID]
		macroNodeMvmts, ok := macroNodesMovements[macroNodeID]
		if !ok {
			continue
		}
		for j := range macroNodeMvmts {
			mvmt := macroNodeMvmts[j]
			incomeMacroLinkID := mvmt.IncomeMacroLink()
			outcomeMacroLinkID := mvmt.OutcomeMacroLink()
			incomingMacroLink, ok := macroLinks[incomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find macro link for further connection: %d", incomeMacroLinkID)
			}
			outcomingMacroLink, ok := macroLinks[outcomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find macro link for further connection: %d", outcomeMacroLinkID)
			}

			incomingMacroLinkProcessed, ok := macroLinksProcessed[incomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find processed macro link for further connection: %d", incomeMacroLinkID)
			}
			outcomingMacroLinkProcessed, ok := macroLinksProcessed[outcomeMacroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find processed macro link for further connection: %d", outcomeMacroLinkID)
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
				sourceMesoNodeID := incomingMesoLink.TargetNode()
				targetMesoNodeID := outcomingMesoLink.SourceNode()
				mesoLink := meso.NewLinkFrom(
					lastMesoLinkID,
					sourceMesoNodeID,
					targetMesoNodeID,
					meso.WithLanesNum(mvmt.LanesNum()),
					meso.WithLineGeom(geom),
					meso.WithLineGeomEuclidean(geomEuclidean),
					meso.WithLineMacroLinkID(-1),
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
				lastMesoLinkID += 1
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
		if macroNodeID < 0 && mesoNode.MacroLink() < 0 {
			return errors.Wrapf(ErrBadParentInfo, "Neither macroscopic link nor node for meso node: %d", mesoNode.ID)
		}
		if mesoNode.MacroNode() < 0 {
			meso.WithBoundaryType(types.BOUNDARY_NONE)(mesoNode)
			continue
		}
		macroNode, ok := macroNodes[macroNodeID]
		if !ok {
			return errors.Wrapf(macro.ErrNodeNotFound, "Can't find macroscopic node with id %d for mesoscopic node %d", macroNodeID, mesoNode.ID)
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
	macroNodesMovements map[gmns.NodeID][]*movement.Movement,
) error {
	movementMesoLinks := make(map[gmns.LinkID]struct{})
	for i := range mesoLinks {
		mesoLink := mesoLinks[i]
		macroNodeID := mesoLink.MacroNode()
		macroLinkID := mesoLink.MacroLink()

		if macroNodeID < 0 && macroLinkID < 0 {
			return errors.Wrapf(ErrBadParentInfo, "Neither macroscopic link nor node for mesoscopic link: %d", mesoLink.ID)
		}

		if mesoLink.MacroNode() < 0 {
			// Inherit macroscopic link properties
			macroLink, ok := macroLinks[macroLinkID]
			if !ok {
				return errors.Wrapf(macro.ErrLinkNotFound, "Can't find macroscopic link with id %d for mesoscopic link %d", macroLinkID, mesoLink.ID)
			}
			meso.WithLinkType(macroLink.LinkType())(mesoLink)
			meso.WithFreeSpeed(macroLink.FreeSpeed())(mesoLink)
			meso.WithCapacity(macroLink.Capacity())(mesoLink)
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
			return errors.Wrapf(macro.ErrNodeNotFound, "Can't find macroscopic node with id %d for mesoscopic link %d", macroNodeID, mesoLink.ID)
		}
		meso.WithControlType(macroNode.ControlType())(mesoLink)

		movementID := mesoLink.Movement()
		if movementID < 0 {
			return errors.Wrapf(ErrBadParentInfo, "Should have movement ID for mesoscopic link: %d", mesoLink.ID)
		}
		mvmt, ok := movements[movementID]
		if !ok {
			return errors.Wrapf(movement.ErrMvmtNotFound, "Can't find movement with ID %d for mesoscopic link %d", movementID, mesoLink.ID)
		}
		meso.WithMovementCompositeType(mvmt.MvmtTextID())(mesoLink)
	}

	// Inherit macroscopic link properties for movement links
	for mesoLinkID := range movementMesoLinks {
		mesoLink, ok := mesoLinks[mesoLinkID]
		if !ok {
			return errors.Wrapf(meso.ErrLinkNotFound, "Can't find mesoscopic link %d while processing movement links", mesoLinkID)
		}
		sourceMesoNodeID := mesoLink.SourceNode()
		sourceMesoNode, ok := mesoNodes[sourceMesoNodeID]
		if !ok {
			return errors.Wrapf(meso.ErrLinkNotFound, "Can't find source node %d for mesoscopic link %d while processing movement links", sourceMesoNodeID, mesoLinkID)
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
			return errors.Wrapf(ErrBadInterface, "Can't get correct type for upstream for source mesoscopic node %d of mesoscopic link %d", sourceMesoNode.ID, mesoLinkID)
		}
		upstreamMesoLink, ok := mesoLinks[upstreamMesoLinkID]
		if !ok {
			return errors.Wrapf(meso.ErrLinkNotFound, "Can't find mesoscopic link %d while processing movement links", mesoLinkID)
		}
		// Inherit upstream mesoscopic link properties
		meso.WithLinkType(upstreamMesoLink.LinkType())(mesoLink)
		meso.WithFreeSpeed(upstreamMesoLink.FreeSpeed())(mesoLink)
		meso.WithCapacity(upstreamMesoLink.Capacity())(mesoLink)
		meso.WithAllowedAgentTypes(upstreamMesoLink.AllowedAgentTypes())(mesoLink)
	}
	return nil
}
