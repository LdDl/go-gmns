package generators

import (
	"context"
	"fmt"
	"reflect"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
)

func validateArchiveDatasets(source *macro.Net, target *meso.Net) error {
	err := source.ValidateDataset()
	if err != nil {
		return mesoUpdateFailure("invalid_source", err.Error())
	}
	err = target.ValidateDataset()
	if err != nil {
		return mesoUpdateFailure("invalid_target", err.Error())
	}
	if source.Config().IDType() != target.Config().IDType() {
		return mesoUpdateFailure("invalid_target", "macro and meso identifier kinds differ")
	}
	if target.HasDetailedRecords() || len(target.MovementRecords()) != 0 {
		return mesoUpdateFailure("unsupported_profile", "meso-owned detailed tables or movements are outside the preparation profile")
	}
	return nil
}

func restoreArchiveRecords(ctx context.Context, source *macro.Net, target *meso.Net, metadata *MesoArchiveMetadata) (*MesoPreparationResult, preparationLayout, error) {
	if ctx == nil || metadata == nil {
		return nil, nil, mesoUpdateFailure("invalid_request", "initialization requires context and preparation metadata")
	}
	err := ctx.Err()
	if err != nil {
		return nil, nil, err
	}
	if metadata.Algorithm != MesoPreparationAlgorithm {
		return nil, nil, mesoUpdateFailure("unsupported_profile", "requires standard-macro-sections-3; ordinary OSM generator ancestry does not identify this profile")
	}
	options := mesoPreparationOptions{}
	switch metadata.Options.LaneLayout {
	case "uniform":
		options.uniformLaneLayout = true
	case "detailed":
		options.detailedLaneLayout = true
	default:
		return nil, nil, mesoUpdateFailure("unsupported_profile", "missing or unsupported lane layout")
	}
	switch metadata.Options.MovementPolicy {
	case "single":
	case "ordered":
		options.orderedLaneConnections = true
	default:
		return nil, nil, mesoUpdateFailure("unsupported_profile", "missing or unsupported movement policy")
	}
	if metadata.Options.DefaultLanes != nil {
		count := *metadata.Options.DefaultLanes
		if count < 1 || count > 99 {
			return nil, nil, mesoUpdateFailure("invalid_state", "invalid default lane count")
		}
		options.defaultLanes = &count
	}
	err = validateArchiveDatasets(source, target)
	if err != nil {
		return nil, nil, err
	}
	if source.HasDetailedRecords() && !options.detailedLaneLayout {
		return nil, nil, mesoUpdateFailure("unsupported_profile", "uniform policy cannot own supplied detailed tables")
	}
	prepared := &MesoPreparationResult{macroNet: source.Clone(), mesoNet: target.Clone(), algorithm: MesoPreparationAlgorithm, options: options}
	for _, node := range prepared.macroNet.Nodes {
		macro.WithPointGeomEuclidean(geomath.PointToEuclidean(node.Geom()))(node)
	}
	for _, link := range macroLinksToSlice(prepared.macroNet.Links) {
		count, present := link.LanesNum()
		if !present || count < 1 || count > 99 {
			return nil, nil, mesoUpdateFailure("invalid_source", fmt.Sprintf("link %v: completed positive lane count is required", link.ID))
		}
		// Reconstruct only disposable source caches; retain canonical records exactly.
		validated, validateErr := prepareMacroLink(source, link, &options)
		if validateErr != nil {
			return nil, nil, mesoUpdateFailure("invalid_source", validateErr.Error())
		}
		macro.WithLineGeomEuclidean(validated.GeomEuclidean())(link)
	}
	layout, err := prepareLaneLayout(prepared.macroNet, prepared.macroNet)
	if err != nil {
		return nil, nil, mesoUpdateFailure("invalid_source", err.Error())
	}
	for _, record := range prepared.macroNet.MovementRecords() {
		err = prepareMovementLanes(record, layout, &options)
		if err != nil {
			return nil, nil, mesoUpdateFailure("invalid_movement", err.Error())
		}
		err = prepared.macroNet.RemoveMovement(record.ID)
		if err != nil {
			return nil, nil, err
		}
		err = prepared.macroNet.AddMovement(record)
		if err != nil {
			return nil, nil, err
		}
	}
	err = prepared.macroNet.ValidateDataset()
	if err != nil {
		return nil, nil, err
	}
	err = restoreArchiveSections(prepared, layout, metadata)
	if err != nil {
		return nil, nil, err
	}
	err = restoreArchiveConnections(prepared, layout, metadata)
	if err != nil {
		return nil, nil, err
	}
	err = validateArchiveInheritance(prepared, layout)
	if err != nil {
		return nil, nil, err
	}
	assumed := make(map[gmns.LinkID]bool)
	for _, entry := range metadata.Assumptions {
		id, parseErr := archiveLinkID(entry.Link, source.Config().IDType())
		link := source.Links[id]
		if parseErr != nil || link == nil || assumed[id] || options.defaultLanes == nil || entry.Lanes != *options.defaultLanes {
			return nil, nil, mesoUpdateFailure("invalid_lineage", "invalid lane assumption")
		}
		count, _ := link.LanesNum()
		if count != entry.Lanes {
			return nil, nil, mesoUpdateFailure("invalid_lineage", "lane assumption differs from completed source")
		}
		assumed[id] = true
		prepared.laneAssumptions = append(prepared.laneAssumptions, LaneCountAssumption{linkID: id, lanes: count})
	}
	for _, id := range sortedMesoLinkIDs(prepared.mesoNet.Links) {
		link := prepared.mesoNet.Links[id]
		shape, shapeErr := archiveMesoShape(prepared.mesoNet, link)
		if shapeErr != nil {
			return nil, nil, shapeErr
		}
		meso.WithLineGeomEuclidean(geomath.LineToEuclidean(shape))(link)
		from, to := prepared.mesoNet.Nodes[link.SourceNode()], prepared.mesoNet.Nodes[link.TargetNode()]
		meso.WithPointEuclideanGeom(geomath.PointToEuclidean(from.Geom()))(from)
		meso.WithPointEuclideanGeom(geomath.PointToEuclidean(to.Geom()))(to)
		if !equalMesoRecords(shape[0], from.Geom()) || !equalMesoRecords(shape[len(shape)-1], to.Geom()) {
			return nil, nil, archiveLinkFailure(id, "geometry endpoints differ from their meso nodes")
		}
		_, hasOutgoing := from.OutcomingLinks().Get(id)
		if !hasOutgoing {
			meso.WithOutcomingLinks(id)(from)
		}
		_, hasIncoming := to.IncomingLinks().Get(id)
		if !hasIncoming {
			meso.WithIncomingLinks(id)(to)
		}
	}
	err = ctx.Err()
	if err != nil {
		return nil, nil, err
	}
	return prepared, layout, nil
}

func restoreArchiveSections(prepared *MesoPreparationResult, layout preparationLayout, metadata *MesoArchiveMetadata) error {
	kind := prepared.macroNet.Config().IDType()
	seen := make(map[gmns.LinkID]bool)
	positions := make(map[gmns.LinkID]map[int]*meso.Link)
	ownedNodes := make(map[gmns.NodeID]gmns.LinkID)
	for _, entry := range metadata.Sections {
		id, err := archiveLinkID(entry.Link, kind)
		if err != nil || prepared.mesoNet.Links[id] == nil || seen[id] {
			return mesoUpdateFailure("invalid_lineage", "missing, invalid or duplicate section link")
		}
		sourceID, err := archiveLinkID(entry.Source, kind)
		source := prepared.macroNet.Links[sourceID]
		if err != nil || source == nil {
			return archiveLinkFailure(id, "section source link is missing")
		}
		start, err := archiveParseFloat(entry.Start)
		if err != nil {
			return archiveLinkFailure(id, err.Error())
		}
		end, err := archiveParseFloat(entry.End)
		if err != nil {
			return archiveLinkFailure(id, err.Error())
		}
		position := -1
		for index, section := range layout[sourceID] {
			if archiveFloat(section.start) == entry.Start && archiveFloat(section.end) == entry.End {
				position = index
			}
		}
		if position < 0 {
			return archiveLinkFailure(id, "section interval differs from current source layout")
		}
		if positions[sourceID] == nil {
			positions[sourceID] = make(map[int]*meso.Link)
		}
		if positions[sourceID][position] != nil {
			return archiveLinkFailure(id, "duplicate section interval")
		}
		link := prepared.mesoNet.Links[id]
		if link.IsConnection() || !link.Movement().IsZero() || !link.MacroLink().IsZero() && link.MacroLink() != sourceID {
			return archiveLinkFailure(id, "section metadata contradicts existing ancestry")
		}
		section := layout[sourceID][position]
		count, present := link.LanesNum()
		if !present || count != len(section.numbers) || !reflect.DeepEqual(entry.Lanes, section.numbers) {
			return archiveLinkFailure(id, "section lane numbers differ from source layout")
		}
		meso.WithLineMacroLinkID(sourceID)(link)
		meso.WithSegmentIdx(position)(link)
		meso.WithLanesChange(section.changes)(link)
		expected := PreparedSection{linkID: id, sourceLinkID: sourceID, start: start, end: end, lanes: section.numbers}
		if section.segment != nil {
			expected.segmentID = section.segment.ID
		}
		if expected.segmentID.String() != entry.Segment {
			return archiveLinkFailure(id, "section segment differs from source layout")
		}
		laneIDs := []string{}
		for _, lane := range section.lanes {
			if lanePosition(section.numbers, lane.LaneNum()) >= 0 {
				laneIDs = append(laneIDs, lane.ID.String())
			}
		}
		segmentLaneIDs := []string{}
		for _, lane := range section.segmentLanes {
			segmentLaneIDs = append(segmentLaneIDs, lane.ID.String())
		}
		if !equalArchiveStrings(laneIDs, entry.LaneIDs) || !equalArchiveStrings(segmentLaneIDs, entry.SegmentLaneIDs) {
			return archiveLinkFailure(id, "section lane ancestry differs from current detailed records")
		}
		for _, nodeID := range []gmns.NodeID{link.SourceNode(), link.TargetNode()} {
			owner, exists := ownedNodes[nodeID]
			if exists && owner != sourceID {
				return archiveLinkFailure(id, "section node is shared by unrelated source roads")
			}
			ownedNodes[nodeID] = sourceID
		}
		positions[sourceID][position] = link
		seen[id] = true
	}
	for _, source := range macroLinksToSlice(prepared.macroNet.Links) {
		sections := positions[source.ID]
		if len(sections) != len(layout[source.ID]) {
			return mesoUpdateFailure("invalid_lineage", fmt.Sprintf("source link %v has missing sections", source.ID))
		}
		for index := 0; index < len(sections); index++ {
			link := sections[index]
			if index > 0 && sections[index-1].TargetNode() != link.SourceNode() {
				return archiveLinkFailure(link.ID, "adjacent sections are disconnected")
			}
			if index == 0 {
				err := restoreArchiveNode(prepared, link.SourceNode(), source.SourceNode(), gmns.LinkID{})
				if err != nil {
					return err
				}
			}
			parentNode, parentLink := gmns.NodeID{}, source.ID
			if index == len(sections)-1 {
				parentNode, parentLink = source.TargetNode(), gmns.LinkID{}
			}
			err := restoreArchiveNode(prepared, link.TargetNode(), parentNode, parentLink)
			if err != nil {
				return err
			}
		}
	}
	if len(ownedNodes) != len(prepared.mesoNet.Nodes) {
		return mesoUpdateFailure("invalid_lineage", "extra meso nodes have no section ownership")
	}
	prepared.sections = preparationLineageSections(prepared.mesoNet, layout, seen)
	return nil
}

func preparationLineageSections(target *meso.Net, layout preparationLayout, seen map[gmns.LinkID]bool) []PreparedSection {
	roads := meso.NewNet()
	for id := range seen {
		roads.Links[id] = target.Links[id]
	}
	return preparationLineage(roads, layout)
}

func restoreArchiveNode(prepared *MesoPreparationResult, id, parent gmns.NodeID, road gmns.LinkID) error {
	node := prepared.mesoNet.Nodes[id]
	if !node.MacroNode().IsZero() && node.MacroNode() != parent || !node.MacroLink().IsZero() && node.MacroLink() != road {
		return mesoUpdateFailure("invalid_lineage", fmt.Sprintf("node %v has contradictory ancestry", id))
	}
	expected := meso.NewNodeFrom(id, meso.WithPointGeom(node.Geom()))
	if !parent.IsZero() {
		source := prepared.macroNet.Nodes[parent]
		meso.WithNodeName(source.Name())(expected)
		meso.WithNodeType(source.NodeType())(expected)
		meso.WithNodeControlType(source.ControlType())(expected)
		altitude, present := source.AltitudeMeters()
		_, supplied := node.AltitudeMeters()
		if present && (supplied || node.Geom() == source.Geom()) {
			meso.WithAltitudeMeters(altitude)(expected)
		}
		zone, present := source.Zone()
		if present {
			meso.WithMacroZone(zone)(expected)
		}
	}
	if archiveRecord("node", node, "meso").Hash != archiveRecord("node", expected, "meso").Hash {
		return mesoUpdateFailure("invalid_lineage", fmt.Sprintf("node %v inherited attributes differ from source", id))
	}
	meso.WithPointMacroNodeID(parent)(node)
	meso.WithPointMacroLinkID(road)(node)
	return nil
}

func restoreArchiveConnections(prepared *MesoPreparationResult, layout preparationLayout, metadata *MesoArchiveMetadata) error {
	kind := prepared.macroNet.Config().IDType()
	seen := make(map[gmns.LinkID]bool)
	roads := make(map[gmns.LinkID][]*meso.Link)
	for _, section := range prepared.sections {
		link := prepared.mesoNet.Links[section.linkID]
		if roads[section.sourceLinkID] == nil {
			roads[section.sourceLinkID] = make([]*meso.Link, len(layout[section.sourceLinkID]))
		}
		roads[section.sourceLinkID][link.SegmentIdx()] = link
		seen[link.ID] = true
	}
	movements := make(map[gmns.MovementID]bool)
	for _, entry := range metadata.Connections {
		id, err := archiveLinkID(entry.Link, kind)
		if err != nil || prepared.mesoNet.Links[id] == nil || seen[id] {
			return mesoUpdateFailure("invalid_lineage", "missing, invalid or duplicate connector link")
		}
		movementID, err := gmns.ParseMovementID(entry.Movement, kind)
		record, exists := prepared.macroNet.Movement(movementID)
		if err != nil || movementID.String() != entry.Movement || !exists || movements[movementID] {
			return archiveLinkFailure(id, "missing or duplicate connector movement")
		}
		inbound, outbound := roads[record.IncomeLink()], roads[record.OutcomeLink()]
		upstream, downstream := inbound[len(inbound)-1], outbound[0]
		link := prepared.mesoNet.Links[id]
		if link.SourceNode() != upstream.TargetNode() || link.TargetNode() != downstream.SourceNode() {
			return archiveLinkFailure(id, "connector endpoints differ from source movement ownership")
		}
		if !link.MacroLink().IsZero() || !link.Movement().IsZero() && link.Movement() != movementID || !link.MacroNode().IsZero() && link.MacroNode() != record.Node() {
			return archiveLinkFailure(id, "connector metadata contradicts existing ancestry")
		}
		count, present := link.LanesNum()
		inNumbers := layout[record.IncomeLink()][len(inbound)-1].numbers
		outNumbers := layout[record.OutcomeLink()][0].numbers
		firstIn, firstOut := record.StartIncomeLaneSeqID(), record.StartOutcomeLaneSeqID()
		if !present || count != record.LanesNum() || !reflect.DeepEqual(entry.Inbound, inNumbers[firstIn:firstIn+record.LanesNum()]) || !reflect.DeepEqual(entry.Outbound, outNumbers[firstOut:firstOut+record.LanesNum()]) {
			return archiveLinkFailure(id, "connector lane scopes differ from current movement")
		}
		meso.WithIsConnection(true)(link)
		meso.WithMovementID(movementID)(link)
		meso.WithLineMacroNodeID(record.Node())(link)
		meso.WithMovementMesoLinkIncome(upstream.ID)(link)
		meso.WithMovementMesoLinkOutcome(downstream.ID)(link)
		meso.WithMovementIncomeLaneStartSeqID(firstIn)(link)
		meso.WithMovementOutcomeLaneStartSeqID(firstOut)(link)
		meso.WithMovementCompositeType(record.MvmtTextID())(link)
		seen[id], movements[movementID] = true, true
	}
	if len(seen) != len(prepared.mesoNet.Links) || len(movements) != len(prepared.macroNet.MovementRecords()) {
		return mesoUpdateFailure("invalid_lineage", "missing connector associations or unowned meso links")
	}
	prepared.connections = preparationConnections(prepared.macroNet, prepared.mesoNet, layout)
	return nil
}

func validateArchiveInheritance(prepared *MesoPreparationResult, layout preparationLayout) error {
	expected := prepared.mesoNet.Clone()
	for _, section := range prepared.sections {
		source := prepared.macroNet.Links[section.sourceLinkID]
		target := expected.Links[section.linkID]
		meso.WithoutGrade()(target)
		meso.WithoutRightOfWayWidthMeters()(target)
		meso.WithoutToll()(target)
		copyMacroRoadAttributes(source, target)
		meso.WithoutFreeSpeed()(target)
		speed, present := source.FreeSpeed()
		if present {
			meso.WithFreeSpeed(speed)(target)
		}
		meso.WithoutCapacity()(target)
		capacity, present := source.Capacity()
		if present {
			meso.WithCapacity(capacity)(target)
		}
	}
	for _, connection := range prepared.connections {
		target := expected.Links[connection.linkID]
		source := expected.Links[target.MovementMesoLinkIncome()]
		meso.WithoutFreeSpeed()(target)
		speed, present := source.FreeSpeed()
		if present {
			meso.WithFreeSpeed(speed)(target)
		}
		meso.WithoutCapacity()(target)
		capacity, present := source.Capacity()
		if present {
			meso.WithCapacity(capacity)(target)
		}
	}
	err := applyPreparedSections(prepared.macroNet, expected, layout)
	if err != nil {
		return mesoUpdateFailure("invalid_lineage", err.Error())
	}
	for _, id := range sortedMesoLinkIDs(expected.Links) {
		if archiveRecord("link", expected.Links[id], "meso").Hash != archiveRecord("link", prepared.mesoNet.Links[id], "meso").Hash {
			return archiveLinkFailure(id, "inherited attributes differ from current source or detailed overrides")
		}
	}
	return nil
}

func archiveLinkFailure(id gmns.LinkID, message string) error {
	return &MesoUpdateError{conflicts: []MesoUpdateConflict{{code: "invalid_lineage", target: id, message: fmt.Sprintf("meso link %v: %s", id, message)}}}
}

func equalArchiveStrings(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func archiveMesoShape(net *meso.Net, link *meso.Link) (orb.LineString, error) {
	shape := link.Geom()
	id, shared := link.GeometryID()
	if len(shape) == 0 && shared {
		record, _ := net.Geometry(id)
		shape = record.Geom()
		direction, present := link.DirFlag()
		if !present {
			return nil, archiveLinkFailure(link.ID, "shared geometry orientation is absent")
		}
		if direction == gmns.DIR_FLAG_REVERSE {
			shape.Reverse()
		}
	}
	err := gmns.ValidateLine(shape)
	if err != nil {
		return nil, archiveLinkFailure(link.ID, err.Error())
	}
	return shape, nil
}
