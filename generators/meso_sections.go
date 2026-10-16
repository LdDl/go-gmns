package generators

import (
	"fmt"
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb/geo"
)

func copyPreparationDetails(source *macro.Net, target *macro.Net) error {
	for _, record := range source.LaneRecords() {
		err := target.AddLane(record)
		if err != nil {
			return err
		}
	}
	for _, record := range source.SegmentRecords() {
		err := target.AddSegment(record)
		if err != nil {
			return err
		}
	}
	for _, record := range source.SegmentLaneRecords() {
		err := target.AddSegmentLane(record)
		if err != nil {
			return err
		}
	}
	return target.ValidateDataset()
}

func applyPreparedSections(source *macro.Net, target *meso.Net, layout preparationLayout) error {
	for _, id := range sortedMesoLinkIDs(target.Links) {
		link := target.Links[id]
		if link.MacroLink().IsZero() {
			continue
		}
		section := layout[link.MacroLink()][link.SegmentIdx()]
		meso.WithAllowedUses(sectionUses(source.Links[link.MacroLink()], &section))(link)
		record := section.segment
		if record == nil {
			continue
		}
		grade, present := record.Grade()
		if present {
			meso.WithGrade(grade)(link)
		}
		capacity, present := record.Capacity()
		if present {
			meso.WithCapacity(capacity)(link)
		}
		speed, present := record.FreeSpeed()
		if present {
			meso.WithFreeSpeed(speed)(link)
		}
		width, present := record.RightOfWayWidthMeters()
		if present {
			meso.WithRightOfWayWidthMeters(width)(link)
		}
		toll, present := record.Toll()
		if present {
			meso.WithToll(toll)(link)
		}
		if record.BikeFacility() != "" {
			meso.WithBikeFacility(record.BikeFacility())(link)
		}
		if record.PedFacility() != "" {
			meso.WithPedFacility(record.PedFacility())(link)
		}
		if record.Parking() != "" {
			meso.WithParking(record.Parking())(link)
		}
		if record.Jurisdiction() != "" {
			meso.WithJurisdiction(record.Jurisdiction())(link)
		}
	}
	for _, id := range sortedMesoLinkIDs(target.Links) {
		link := target.Links[id]
		if link.Movement().IsZero() {
			continue
		}
		record, _ := source.Movement(link.Movement())
		inbound := layout[record.IncomeLink()]
		outbound := layout[record.OutcomeLink()]
		inSection := inbound[len(inbound)-1]
		outSection := outbound[0]
		values := []string{sectionUses(source.Links[record.IncomeLink()], &inSection), record.AllowedUses(), sectionUses(source.Links[record.OutcomeLink()], &outSection)}
		uses, err := intersectMovementUses(source, record, values)
		if err != nil {
			return err
		}
		meso.WithAllowedUses(uses)(link)
		// Connector speed and capacity follow the effective upstream section.
		if inSection.segment != nil {
			speed, present := inSection.segment.FreeSpeed()
			if present {
				meso.WithFreeSpeed(speed)(link)
			}
			capacity, present := inSection.segment.Capacity()
			if present {
				meso.WithCapacity(capacity)(link)
			}
		}
	}
	return nil
}

func (processing *macroLinkProcessing) performPreparationCut() error {
	points := append([]float64(nil), processing.lanesInfo.LanesChangePoints...)
	points[0] = processing.upstreamCutLen
	points[len(points)-1] = processing.lengthMetersOffset - processing.downstreamCutLen
	processing.lanesInfoCut.LanesList = append([]int(nil), processing.lanesInfo.LanesList...)
	processing.lanesInfoCut.LanesChange = append([][2]int(nil), processing.lanesInfo.LanesChange...)
	for index := range processing.lanesInfoCut.LanesList {
		start, end := points[index], points[index+1]
		if !gmns.Finite(start) || !gmns.Finite(end) || start < 0 || end > processing.lengthMetersOffset || start >= end {
			return fmt.Errorf("link %v section %d: interval collapses after geometry scaling and cuts", processing.id, index)
		}
		shape := geomath.SubstringHaversine(processing.offsetGeom, start, end)
		err := gmns.ValidateLine(shape)
		if err != nil {
			return fmt.Errorf("link %v section %d: %w", processing.id, index, err)
		}
		if geo.LengthHaversine(shape) <= 0 {
			return fmt.Errorf("link %v section %d: geometry collapses at floating-point precision", processing.id, index)
		}
		processing.offsetGeomCut = append(processing.offsetGeomCut, shape)
		processing.offsetGeomEuclideanCut = append(processing.offsetGeomEuclideanCut, geomath.LineToEuclidean(shape))
	}
	return nil
}
