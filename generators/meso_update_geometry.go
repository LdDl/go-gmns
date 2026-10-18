package generators

import (
	"context"
	"fmt"
	"math"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func recomputeMesoRoad(prepared *MesoPreparationResult, id gmns.LinkID, offset bool) error {
	source := prepared.macroNet.Links[id]
	processing := &macroLinkProcessing{id: id, lanesInfo: source.LanesInfo()}
	processing.offsetGeom = source.Geom().Clone()
	if len(processing.offsetGeom) == 0 {
		geometryID, _ := source.GeometryID()
		geometry, exists := prepared.macroNet.Geometry(geometryID)
		if !exists {
			return mesoUpdateFailure("invalid_source", "source road has no geometry")
		}
		processing.offsetGeom = geometry.Geom()
		direction, _ := source.DirFlag()
		if direction == gmns.DIR_FLAG_REVERSE {
			processing.offsetGeom.Reverse()
		}
	}
	processing.offsetGeomEuclidean = source.GeomEuclidean().Clone()
	if offset {
		distance := 2 * (float64(source.MaxLanes())/2 + 0.5) * macro.LANE_WIDTH
		processing.offsetGeomEuclidean = geomath.OffsetCurve(source.GeomEuclidean(), -distance)
		processing.offsetGeom = geomath.LineToSpherical(processing.offsetGeomEuclidean)
	}
	processing.lengthMetersOffset = geo.LengthHaversine(processing.offsetGeom)
	if !gmns.Finite(processing.lengthMetersOffset) || processing.lengthMetersOffset <= 0 {
		return fmt.Errorf("link %v: degenerate offset geometry", id)
	}
	length, _ := source.LengthMeters()
	points := make([]float64, len(processing.lanesInfo.LanesChangePoints))
	for index, point := range processing.lanesInfo.LanesChangePoints {
		points[index] = (point / length) * processing.lengthMetersOffset
	}
	processing.lanesInfo.LanesChangePoints = points
	processing.updateCutLength(preparationCutLengths())
	processing.upstreamCutLen = math.Min(processing.upstreamCutLen, (points[1]-points[0])/2)
	processing.downstreamCutLen = math.Min(processing.downstreamCutLen, (points[len(points)-1]-points[len(points)-2])/2)
	err := processing.performPreparationCut()
	if err != nil {
		return err
	}
	for _, section := range prepared.sections {
		if section.sourceLinkID != id {
			continue
		}
		link := prepared.mesoNet.Links[section.linkID]
		shape := processing.offsetGeomCut[0]
		projected := processing.offsetGeomEuclideanCut[0]
		count, _ := source.LanesNum()
		_, wasShared := link.GeometryID()
		if wasShared {
			meso.WithoutGeometryID()(link)
			meso.WithDirFlag(gmns.DIR_FLAG_FORWARD)(link)
		}
		meso.WithLineGeom(shape.Clone())(link)
		meso.WithLineGeomEuclidean(projected.Clone())(link)
		meso.WithLanesNum(count)(link)
		meso.WithLengthMeters(geo.LengthHaversine(shape))(link)
		upstream := prepared.mesoNet.Nodes[link.SourceNode()]
		downstream := prepared.mesoNet.Nodes[link.TargetNode()]
		meso.WithPointGeom(shape[0])(upstream)
		meso.WithPointEuclideanGeom(projected[0])(upstream)
		meso.WithPointGeom(shape[len(shape)-1])(downstream)
		meso.WithPointEuclideanGeom(projected[len(projected)-1])(downstream)
	}
	return nil
}

func updateMesoRoadAttributes(prepared *MesoPreparationResult, id gmns.LinkID, edit *mesoRoadEdit, layout preparationLayout) {
	source := prepared.macroNet.Links[id]
	for _, section := range prepared.sections {
		if section.sourceLinkID != id {
			continue
		}
		target := prepared.mesoNet.Links[section.linkID]
		if edit.name != nil {
			meso.WithLinkName(source.Name())(target)
		}
		if edit.speedSet {
			speed, supplied := source.FreeSpeed()
			effective := layout[id][target.SegmentIdx()].segment
			if effective != nil {
				override, present := effective.FreeSpeed()
				if present {
					speed, supplied = override, true
				}
			}
			meso.WithoutFreeSpeed()(target)
			if supplied {
				meso.WithFreeSpeed(speed)(target)
			}
		}
	}
}

func updateMesoConnections(ctx context.Context, state *MesoUpdateState, result *MesoUpdateResult, edit *MesoEdit, layout preparationLayout) error {
	prepared := result.prepared
	target := prepared.mesoNet
	roadSections := make(map[gmns.LinkID][]*meso.Link)
	for _, id := range sortedMesoLinkIDs(target.Links) {
		link := target.Links[id]
		if link.MacroLink().IsZero() {
			continue
		}
		sections := roadSections[link.MacroLink()]
		if sections == nil {
			sections = make([]*meso.Link, len(layout[link.MacroLink()]))
		}
		sections[link.SegmentIdx()] = link
		roadSections[link.MacroLink()] = sections
	}
	connectionIDs := make(map[gmns.MovementID]gmns.LinkID)
	affected := make(map[gmns.MovementID]bool)
	for _, id := range result.affectedMovements {
		affected[id] = true
	}
	for _, connection := range prepared.connections {
		connectionIDs[connection.movementID] = connection.linkID
		if edit.removed[connection.movementID] {
			link := target.Links[connection.linkID]
			target.Nodes[link.SourceNode()].OutcomingLinks().Delete(link.ID)
			target.Nodes[link.TargetNode()].IncomingLinks().Delete(link.ID)
			delete(target.Links, link.ID)
			affected[connection.movementID] = true
		}
	}
	kind, err := macroIdentifierType(prepared.macroNet)
	if err != nil {
		return err
	}
	allocator, err := gmns.NewIDAllocator(kind)
	if err != nil {
		return err
	}
	for id := range state.reservedLinks {
		err = allocator.ReserveLinkID(id)
		if err != nil {
			return err
		}
	}
	result.affectedMovements = sortedMesoEditMovements(affected)
	for _, id := range result.affectedMovements {
		err = ctx.Err()
		if err != nil {
			return err
		}
		record, exists := prepared.macroNet.Movement(id)
		if !exists {
			continue
		}
		incoming := roadSections[record.IncomeLink()]
		outgoing := roadSections[record.OutcomeLink()]
		upstream := incoming[len(incoming)-1]
		downstream := outgoing[0]
		linkID, existing := connectionIDs[id]
		link := target.Links[linkID]
		if !existing {
			linkID = allocator.NextLinkID()
			link = meso.NewLinkFrom(linkID, upstream.TargetNode(), downstream.SourceNode(),
				meso.WithIsConnection(true),
				meso.WithMovementID(id),
				meso.WithLineMacroNodeID(record.Node()),
				meso.WithControlType(prepared.macroNet.Nodes[record.Node()].ControlType()),
				meso.WithLinkType(upstream.LinkType()),
				meso.WithAllowedAgentTypes(upstream.AllowedAgentTypes()),
			)
			capacity, present := upstream.Capacity()
			if present {
				meso.WithCapacity(capacity)(link)
			}
			target.Links[linkID] = link
			meso.WithOutcomingLinks(linkID)(target.Nodes[link.SourceNode()])
			meso.WithIncomingLinks(linkID)(target.Nodes[link.TargetNode()])
		}
		inEdit := edit.roads[record.IncomeLink()]
		outEdit := edit.roads[record.OutcomeLink()]
		geometryChanged := !existing || inEdit != nil && inEdit.laneDelta != 0 || outEdit != nil && outEdit.laneDelta != 0
		if geometryChanged {
			inShape, shapeErr := archiveMesoShape(target, upstream)
			if shapeErr != nil {
				return shapeErr
			}
			outShape, shapeErr := archiveMesoShape(target, downstream)
			if shapeErr != nil {
				return shapeErr
			}
			incomingGeometry := upstream.Clone()
			outgoingGeometry := downstream.Clone()
			meso.WithLineGeom(inShape)(incomingGeometry)
			meso.WithLineGeom(outShape)(outgoingGeometry)
			updateMesoConnectionGeometry(link, incomingGeometry, outgoingGeometry)
			result.recomputedConnections = append(result.recomputedConnections, linkID)
		}
		_, rangesChanged := edit.movements[id]
		if !existing || rangesChanged || geometryChanged {
			meso.WithLanesNum(record.LanesNum())(link)
			meso.WithMovementCompositeType(record.MvmtTextID())(link)
			meso.WithMovementMesoLinkIncome(upstream.ID)(link)
			meso.WithMovementMesoLinkOutcome(downstream.ID)(link)
			meso.WithMovementIncomeLaneStartSeqID(record.StartIncomeLaneSeqID())(link)
			meso.WithMovementOutcomeLaneStartSeqID(record.StartOutcomeLaneSeqID())(link)
			err = updateMesoConnectionUses(prepared, link, record, layout)
			if err != nil {
				return err
			}
		}
		if !existing || inEdit != nil && inEdit.speedSet {
			meso.WithoutFreeSpeed()(link)
			speed, present := upstream.FreeSpeed()
			if present {
				meso.WithFreeSpeed(speed)(link)
			}
		}
	}
	sortMesoUpdateLinkIDs(result.recomputedConnections)
	return nil
}

func updateMesoConnectionGeometry(link, upstream, downstream *meso.Link) {
	inShape, outShape := upstream.Geom(), downstream.Geom()
	inProjected, outProjected := upstream.GeomEuclidean(), downstream.GeomEuclidean()
	shape := orb.LineString{inShape[len(inShape)-1], outShape[0]}
	projected := orb.LineString{inProjected[len(inProjected)-1], outProjected[0]}
	_, wasShared := link.GeometryID()
	if wasShared {
		meso.WithoutGeometryID()(link)
		meso.WithDirFlag(gmns.DIR_FLAG_FORWARD)(link)
	}
	meso.WithLineGeom(shape)(link)
	meso.WithLineGeomEuclidean(projected)(link)
	meso.WithLengthMeters(geo.LengthHaversine(shape))(link)
}

func updateMesoConnectionUses(prepared *MesoPreparationResult, link *meso.Link, record *movement.Movement, layout preparationLayout) error {
	inbound := layout[record.IncomeLink()]
	outbound := layout[record.OutcomeLink()]
	inSection := inbound[len(inbound)-1]
	outSection := outbound[0]
	values := []string{sectionUses(prepared.macroNet.Links[record.IncomeLink()], &inSection), record.AllowedUses(), sectionUses(prepared.macroNet.Links[record.OutcomeLink()], &outSection)}
	uses, err := intersectMovementUses(prepared.macroNet, record, values)
	if err != nil {
		return err
	}
	meso.WithAllowedUses(uses)(link)
	return nil
}
