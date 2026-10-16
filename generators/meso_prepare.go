package generators

import (
	"fmt"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb/geo"
)

type mesoPreparationOptions struct {
	uniformLaneLayout      bool
	defaultLanes           *int
	detailedLaneLayout     bool
	orderedLaneConnections bool
}

// WithDefaultMesoLanes fills only absent lane counts using an explicitly accepted assumption.
// Supplied zeroes, negative counts and movement lane references are never replaced.
func WithDefaultMesoLanes(lanes int) MesoPreparationOption {
	return func(options *mesoPreparationOptions) {
		options.defaultLanes = &lanes
	}
}

type MesoPreparationOption func(*mesoPreparationOptions)

// WithUniformLaneLayout accepts constant lane counts numbered 1..lanes along each road.
// This modeling assumption does not create standard lane records.
func WithUniformLaneLayout() MesoPreparationOption {
	return func(options *mesoPreparationOptions) {
		options.uniformLaneLayout = true
	}
}

// PrepareMesoscopic returns completed macro records, a new meso network and applied assumptions.
// It rebuilds derived data independently of source caches and never changes the input.
// Supplied positive lengths and an explicit lane-layout policy are required.
func PrepareMesoscopic(source *macro.Net, configure ...MesoPreparationOption) (*MesoPreparationResult, error) {
	options := mesoPreparationOptions{}
	for _, option := range configure {
		if option == nil {
			return nil, fmt.Errorf("nil meso preparation option")
		}
		option(&options)
	}
	if !options.uniformLaneLayout && !options.detailedLaneLayout {
		return nil, fmt.Errorf("meso preparation requires an explicit lane layout: use WithUniformLaneLayout for constant supplied counts")
	}
	if options.defaultLanes != nil && (*options.defaultLanes <= 0 || *options.defaultLanes >= len(CUT_LENGTHS)) {
		return nil, fmt.Errorf("default meso lanes must be in 1..%d, got %d", len(CUT_LENGTHS)-1, *options.defaultLanes)
	}
	if options.uniformLaneLayout && options.detailedLaneLayout {
		return nil, fmt.Errorf("choose one lane layout: uniform or detailed")
	}
	err := source.ValidateDataset()
	if err != nil {
		return nil, err
	}
	if source.HasDetailedRecords() && !options.detailedLaneLayout {
		return nil, unsupportedPreparationDetail(source)
	}
	working, err := prepareMacroRecords(source, &options)
	if err != nil {
		return nil, err
	}
	layout, err := prepareLaneLayout(source, working)
	if err != nil {
		return nil, err
	}
	for _, record := range source.MovementRecords() {
		err = prepareMovementLanes(working, record, layout, &options)
		if err != nil {
			return nil, err
		}
		err = working.AddMovement(record)
		if err != nil {
			return nil, err
		}
	}
	generationOptions := mesoGenerationOptions{
		preserveConnectors: true,
		standardUses:       true,
		cutLengths:         preparationCutLengths(),
	}
	prepared, err := generateMesoscopic(working, working.MovementStorage(), &generationOptions)
	if err != nil {
		return nil, err
	}
	err = applyPreparedSections(working, prepared, layout)
	if err != nil {
		return nil, err
	}
	err = prepared.ValidateDataset()
	if err != nil {
		return nil, fmt.Errorf("prepared meso dataset: %w", err)
	}
	err = copyPreparationDetails(source, working)
	if err != nil {
		return nil, err
	}
	result := &MesoPreparationResult{macroNet: working, mesoNet: prepared}
	result.sections = preparationLineage(prepared, layout)
	result.connections = preparationConnections(working, prepared, layout)
	for _, link := range macroLinksToSlice(source.Links) {
		_, present := link.LanesNum()
		if !present {
			result.laneAssumptions = append(result.laneAssumptions, LaneCountAssumption{linkID: link.ID, lanes: *options.defaultLanes})
		}
	}
	return result, nil
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

func prepareMacroRecords(source *macro.Net, options *mesoPreparationOptions) (*macro.Net, error) {
	working := macro.NewNet()
	err := working.SetConfig(source.Config())
	if err != nil {
		return nil, err
	}
	for _, record := range source.GeometryRecords() {
		err = working.AddGeometry(record)
		if err != nil {
			return nil, err
		}
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
		copyLink, copyErr := prepareMacroLink(source, link, options)
		if copyErr != nil {
			return nil, copyErr
		}
		working.Links[link.ID] = copyLink
		macro.WithOutcomingLinks(link.ID)(working.Nodes[link.SourceNode()])
		macro.WithIncomingLinks(link.ID)(working.Nodes[link.TargetNode()])
	}
	return working, nil
}

func prepareMacroLink(source *macro.Net, link *macro.Link, options *mesoPreparationOptions) (*macro.Link, error) {
	if !link.Directed() {
		return nil, fmt.Errorf("link %v: meso preparation requires directed links", link.ID)
	}
	toll, hasToll := link.Toll()
	if hasToll && toll != 0 {
		return nil, fmt.Errorf("%w: link %v toll: placement on generated sections is undefined", ErrUnsupportedGenerationInput, link.ID)
	}
	lanes, present := link.LanesNum()
	if !present {
		if options.defaultLanes == nil {
			return nil, fmt.Errorf("link %v lanes: supply a positive count or explicitly accept WithDefaultMesoLanes", link.ID)
		}
		lanes = *options.defaultLanes
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
		macro.WithAllowedUses(link.AllowedUses()),
	)
	if shared {
		macro.WithGeometryID(geometryID)(copyLink)
	}
	direction, present := link.DirFlag()
	if present {
		macro.WithDirFlag(direction)(copyLink)
	}
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

func prepareMovementLanes(net *macro.Net, record *movement.Movement, layout preparationLayout, options *mesoPreparationOptions) error {
	inStart, hasInStart := record.IncomeLaneStart()
	inEnd, hasInEnd := record.IncomeLaneEnd()
	outStart, hasOutStart := record.OutcomeLaneStart()
	outEnd, hasOutEnd := record.OutcomeLaneEnd()
	if !hasInStart || !hasOutStart {
		return fmt.Errorf("movement %v: supply start_ib_lane and start_ob_lane before preparation", record.ID)
	}
	if !hasInEnd {
		inEnd = inStart
	}
	if !hasOutEnd {
		outEnd = outStart
	}
	inbound := layout[record.IncomeLink()]
	outbound := layout[record.OutcomeLink()]
	inNumbers := inbound[len(inbound)-1].numbers
	outNumbers := outbound[0].numbers
	inFirst, inLast := lanePosition(inNumbers, inStart), lanePosition(inNumbers, inEnd)
	outFirst, outLast := lanePosition(outNumbers, outStart), lanePosition(outNumbers, outEnd)
	if inFirst < 0 || inLast < inFirst {
		return fmt.Errorf("movement %v: inbound lane range %d..%d is outside link %v endpoint lanes %v", record.ID, inStart, inEnd, record.IncomeLink(), inNumbers)
	}
	if outFirst < 0 || outLast < outFirst {
		return fmt.Errorf("movement %v: outbound lane range %d..%d is outside link %v endpoint lanes %v", record.ID, outStart, outEnd, record.OutcomeLink(), outNumbers)
	}
	if (inFirst != inLast || outFirst != outLast) && !options.orderedLaneConnections {
		return fmt.Errorf("%w: movement %v: accept WithOrderedLaneConnections for multi-lane ranges", ErrUnsupportedGenerationInput, record.ID)
	}
	if inLast-inFirst != outLast-outFirst {
		return fmt.Errorf("movement %v: ordered connections require equal lane counts (%d inbound, %d outbound); resolve the movement explicitly", record.ID, inLast-inFirst+1, outLast-outFirst+1)
	}
	movement.WithIncomeLaneSequence(inFirst, inLast)(record)
	movement.WithOutcomeLaneSequence(outFirst, outLast)(record)
	movement.WithLanesNum(inLast - inFirst + 1)(record)
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
