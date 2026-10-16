package generators

import (
	"fmt"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb/geo"
)

type mesoPreparationOptions struct {
	uniformLaneLayout bool
}

type MesoPreparationOption func(*mesoPreparationOptions)

// WithUniformLaneLayout accepts constant lane counts numbered 1..lanes along each road.
// This modeling assumption does not create standard lane records.
func WithUniformLaneLayout() MesoPreparationOption {
	return func(options *mesoPreparationOptions) {
		options.uniformLaneLayout = true
	}
}

// PrepareMesoscopic prepares a new network from standard macro records and owned movements.
// It rebuilds derived data independently of source caches and never changes the input.
// The baseline requires explicit single-lane movements and supplied positive lengths/counts.
func PrepareMesoscopic(source *macro.Net, configure ...MesoPreparationOption) (*meso.Net, error) {
	options := mesoPreparationOptions{}
	for _, option := range configure {
		if option == nil {
			return nil, fmt.Errorf("nil meso preparation option")
		}
		option(&options)
	}
	if !options.uniformLaneLayout {
		return nil, fmt.Errorf("meso preparation requires an explicit lane layout: use WithUniformLaneLayout for constant supplied counts")
	}
	err := source.ValidateDataset()
	if err != nil {
		return nil, err
	}
	if source.HasDetailedRecords() {
		return nil, unsupportedPreparationDetail(source)
	}
	working, err := prepareMacroRecords(source)
	if err != nil {
		return nil, err
	}
	prepared, err := generateMesoscopic(working, working.MovementStorage(), true, preparationCutLengths())
	if err != nil {
		return nil, err
	}
	err = prepared.ValidateDataset()
	if err != nil {
		return nil, fmt.Errorf("prepared meso dataset: %w", err)
	}
	return prepared, nil
}

func unsupportedPreparationDetail(source *macro.Net) error {
	lanes := source.LaneRecords()
	if len(lanes) > 0 {
		return fmt.Errorf("%w: lane %v: uniform layout cannot replace supplied detailed records", ErrUnsupportedGenerationInput, lanes[0].ID)
	}
	segments := source.SegmentRecords()
	if len(segments) > 0 {
		return fmt.Errorf("%w: segment %v: uniform layout cannot replace supplied detailed records", ErrUnsupportedGenerationInput, segments[0].ID)
	}
	segmentLanes := source.SegmentLaneRecords()
	return fmt.Errorf("%w: segment_lane %v: uniform layout cannot replace supplied detailed records", ErrUnsupportedGenerationInput, segmentLanes[0].ID)
}

func prepareMacroRecords(source *macro.Net) (*macro.Net, error) {
	working := macro.NewNet()
	err := working.SetConfig(source.Config())
	if err != nil {
		return nil, err
	}
	for _, record := range source.ZoneRecords() {
		err = working.AddZone(record)
		if err != nil {
			return nil, err
		}
	}
	for _, record := range source.UseDefinitionRecords() {
		err = working.AddUseDefinition(record)
		if err != nil {
			return nil, err
		}
	}
	for _, record := range source.UseGroupRecords() {
		err = working.AddUseGroup(record)
		if err != nil {
			return nil, err
		}
	}
	for _, id := range sortedMacroNodeIDs(source.Nodes) {
		node := source.Nodes[id]
		point := node.Geom()
		if point[1] <= -90 || point[1] >= 90 {
			return nil, fmt.Errorf("node %v: Mercator projection is undefined at the poles", id)
		}
		copyNode := macro.NewNodeFrom(id,
			macro.WithNodeName(node.Name()),
			macro.WithNodeType(node.NodeType()),
			macro.WithNodeControlType(node.ControlType()),
			macro.WithPointGeom(point),
			macro.WithPointGeomEuclidean(geomath.PointToEuclidean(point)),
		)
		altitude, present := node.AltitudeMeters()
		if present {
			macro.WithAltitudeMeters(altitude)(copyNode)
		}
		zone, present := node.Zone()
		if present {
			macro.WithZoneID(zone)(copyNode)
		}
		parent, present := node.ParentNode()
		if present {
			macro.WithParentNodeID(parent)(copyNode)
		}
		working.Nodes[id] = copyNode
	}
	for _, link := range macroLinksToSlice(source.Links) {
		copyLink, copyErr := prepareMacroLink(source, link)
		if copyErr != nil {
			return nil, copyErr
		}
		working.Links[link.ID] = copyLink
		macro.WithOutcomingLinks(link.ID)(working.Nodes[link.SourceNode()])
		macro.WithIncomingLinks(link.ID)(working.Nodes[link.TargetNode()])
	}
	for _, record := range source.MovementRecords() {
		err = prepareMovementLanes(working, record)
		if err != nil {
			return nil, err
		}
		err = working.AddMovement(record)
		if err != nil {
			return nil, err
		}
	}
	return working, nil
}

func prepareMacroLink(source *macro.Net, link *macro.Link) (*macro.Link, error) {
	if !link.Directed() {
		return nil, fmt.Errorf("link %v: meso preparation requires directed links", link.ID)
	}
	if link.AllowedUses() != "" {
		return nil, fmt.Errorf("%w: link %v allowed_uses: saved use permissions are not supported by baseline preparation yet", ErrUnsupportedGenerationInput, link.ID)
	}
	toll, hasToll := link.Toll()
	if hasToll && toll != 0 {
		return nil, fmt.Errorf("%w: link %v toll: placement on generated sections is undefined", ErrUnsupportedGenerationInput, link.ID)
	}
	lanes, present := link.LanesNum()
	if !present {
		return nil, fmt.Errorf("link %v lanes: supply a positive count before meso preparation", link.ID)
	}
	if lanes <= 0 || lanes >= len(CUT_LENGTHS) {
		return nil, fmt.Errorf("link %v lanes: baseline preparation supports 1..%d, got %d", link.ID, len(CUT_LENGTHS)-1, lanes)
	}
	length, present := link.LengthMeters()
	if !present || length <= 0 {
		return nil, fmt.Errorf("link %v length: supply a positive length in meters before meso preparation", link.ID)
	}
	shape := link.Geom().Clone()
	geometryID, shared := link.GeometryID()
	if shared && len(shape) == 0 {
		geometry, _ := source.Geometry(geometryID)
		shape = geometry.Geom()
		direction, hasDirection := link.DirFlag()
		if !hasDirection {
			return nil, fmt.Errorf("link %v dir_flag: supply the shared geometry orientation", link.ID)
		}
		if direction == gmns.DIR_FLAG_REVERSE {
			shape.Reverse()
		}
	}
	err := gmns.ValidateLine(shape)
	if err != nil {
		return nil, fmt.Errorf("link %v geometry: %w", link.ID, err)
	}
	for index, point := range shape {
		if point[1] <= -90 || point[1] >= 90 {
			return nil, fmt.Errorf("link %v geometry[%d]: Mercator projection is undefined at the poles", link.ID, index)
		}
		if index > 0 && point == shape[index-1] {
			return nil, fmt.Errorf("link %v geometry[%d]: repeated consecutive point", link.ID, index)
		}
	}
	if geo.LengthHaversine(shape) < 0.01 {
		return nil, fmt.Errorf("link %v geometry: baseline preparation requires at least 0.01 meter", link.ID)
	}
	copyLink := macro.NewLinkFrom(link.ID, link.SourceNode(), link.TargetNode(),
		macro.WithLinkName(link.Name()),
		macro.WithLanesNum(lanes),
		macro.WithLengthMeters(length),
		macro.WithLineGeom(shape),
		macro.WithLineGeomEuclidean(geomath.LineToEuclidean(shape)),
		macro.WithFacilityType(link.FacilityType()),
		macro.WithBikeFacility(link.BikeFacility()),
		macro.WithPedFacility(link.PedFacility()),
		macro.WithParking(link.Parking()),
		macro.WithJurisdiction(link.Jurisdiction()),
	)
	freeSpeed, present := link.FreeSpeed()
	if present {
		macro.WithFreeSpeed(freeSpeed)(copyLink)
	}
	capacity, present := link.Capacity()
	if present {
		macro.WithCapacity(capacity)(copyLink)
	}
	grade, present := link.Grade()
	if present {
		macro.WithGrade(grade)(copyLink)
	}
	width, present := link.RightOfWayWidthMeters()
	if present {
		macro.WithRightOfWayWidthMeters(width)(copyLink)
	}
	if hasToll {
		macro.WithToll(toll)(copyLink)
	}
	parent, present := link.ParentLink()
	if present {
		macro.WithParentLinkID(parent)(copyLink)
	}
	laneInfo, err := macro.NewLanesInfo(copyLink)
	if err != nil {
		return nil, err
	}
	macro.WithLanesInfo(laneInfo)(copyLink)
	return copyLink, nil
}

func prepareMovementLanes(net *macro.Net, record *movement.Movement) error {
	if record.AllowedUses() != "" {
		return fmt.Errorf("%w: movement %v allowed_uses: saved use permissions are not supported by baseline preparation yet", ErrUnsupportedGenerationInput, record.ID)
	}
	inStart, hasInStart := record.IncomeLaneStart()
	inEnd, hasInEnd := record.IncomeLaneEnd()
	outStart, hasOutStart := record.OutcomeLaneStart()
	outEnd, hasOutEnd := record.OutcomeLaneEnd()
	if !hasInStart || !hasInEnd || !hasOutStart || !hasOutEnd {
		return fmt.Errorf("movement %v: supply start_ib_lane, end_ib_lane, start_ob_lane and end_ob_lane before preparation", record.ID)
	}
	inLanes, _ := net.Links[record.IncomeLink()].LanesNum()
	outLanes, _ := net.Links[record.OutcomeLink()].LanesNum()
	if inStart < 1 || inEnd < inStart || inEnd > inLanes {
		return fmt.Errorf("movement %v: inbound lane range %d..%d is outside link %v lanes 1..%d", record.ID, inStart, inEnd, record.IncomeLink(), inLanes)
	}
	if outStart < 1 || outEnd < outStart || outEnd > outLanes {
		return fmt.Errorf("movement %v: outbound lane range %d..%d is outside link %v lanes 1..%d", record.ID, outStart, outEnd, record.OutcomeLink(), outLanes)
	}
	if inStart != inEnd || outStart != outEnd {
		return fmt.Errorf("%w: movement %v: baseline preparation requires exact single-lane scopes; range connection semantics must be resolved explicitly", ErrUnsupportedGenerationInput, record.ID)
	}
	movement.WithIncomeLaneSequence(inStart-1, inEnd-1)(record)
	movement.WithOutcomeLaneSequence(outStart-1, outEnd-1)(record)
	movement.WithLanesNum(1)(record)
	movement.WithAllowedAgentTypes(nil)(record)
	return nil
}

func preparationCutLengths() [100]float64 {
	lengths := [100]float64{2, 8, 12, 14, 16, 18, 20, 22, 24}
	for index := 9; index < len(lengths); index++ {
		lengths[index] = 25
	}
	return lengths
}
