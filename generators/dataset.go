package generators

import (
	"fmt"
	"math"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/micro"
)

// ErrUnsupportedGenerationInput indicates records that can be stored but cannot yet be transformed faithfully.
var ErrUnsupportedGenerationInput = fmt.Errorf("unsupported generation input")

func requireSupportedMacroDataset(net *macro.Net) error {
	return requireSupportedMacroDatasetUses(net, false)
}

func requireSupportedMacroDatasetUses(net *macro.Net, standardUses bool) error {
	if net.HasDetailedRecords() {
		return fmt.Errorf("%w: supplied lane/segment records require a dedicated generator", ErrUnsupportedGenerationInput)
	}
	if net.Config() != nil {
		err := net.ValidateDataset()
		if err != nil {
			return err
		}
	}
	for _, link := range macroLinksToSlice(net.Links) {
		if link.AllowedUses() != "" && !standardUses {
			return fmt.Errorf("%w: allowed_uses permission mapping is not implemented", ErrUnsupportedGenerationInput)
		}
		toll, present := link.Toll()
		if present && toll != 0 {
			return fmt.Errorf("%w: toll placement on generated segments is undefined", ErrUnsupportedGenerationInput)
		}
		err := gmns.ValidateLine(link.Geom())
		if err != nil {
			return fmt.Errorf("macro link %v: %w", link.ID, err)
		}
		for _, point := range link.GeomEuclidean() {
			if !gmns.Finite(point[0]) || !gmns.Finite(point[1]) {
				return fmt.Errorf("macro link %v: nonfinite projected coordinate", link.ID)
			}
		}
		if len(link.GeomEuclidean()) != len(link.Geom()) {
			return fmt.Errorf("macro link %v: prepared projected geometry is required", link.ID)
		}
	}
	return nil
}

func requireSupportedMesoDataset(net *meso.Net) error {
	if net.HasDetailedRecords() {
		return fmt.Errorf("%w: supplied lane/segment records require a dedicated generator", ErrUnsupportedGenerationInput)
	}
	if len(net.MovementRecords()) > 0 {
		return fmt.Errorf("%w: transforming meso-owned movements into micro is not implemented", ErrUnsupportedGenerationInput)
	}
	if net.Config() != nil {
		err := net.ValidateDataset()
		if err != nil {
			return err
		}
	}
	for _, id := range sortedMesoLinkIDs(net.Links) {
		link := net.Links[id]
		if !link.MacroLink().IsZero() {
			err := gmns.ValidateLine(link.Geom())
			if err != nil {
				return fmt.Errorf("meso link %v: %w", id, err)
			}
			if len(link.GeomEuclidean()) != len(link.Geom()) {
				return fmt.Errorf("meso link %v: prepared projected geometry is required", id)
			}
			for _, point := range link.GeomEuclidean() {
				if !gmns.Finite(point[0]) || !gmns.Finite(point[1]) {
					return fmt.Errorf("meso link %v: nonfinite projected coordinate", id)
				}
			}
		}
		if link.AllowedUses() != "" {
			return fmt.Errorf("%w: allowed_uses permission mapping is not implemented", ErrUnsupportedGenerationInput)
		}
		toll, present := link.Toll()
		if present && toll != 0 {
			return fmt.Errorf("%w: toll placement on micro cells is undefined", ErrUnsupportedGenerationInput)
		}
	}
	return nil
}

func copyMacroCatalogs(source *macro.Net, target *meso.Net) error {
	config := source.Config()
	var err error
	if config != nil {
		err = target.SetConfig(config)
		if err != nil {
			return err
		}
	}
	for _, zone := range source.ZoneRecords() {
		err = target.AddZone(zone)
		if err != nil {
			return err
		}
	}
	for _, definition := range source.UseDefinitionRecords() {
		err = target.AddUseDefinition(definition)
		if err != nil {
			return err
		}
	}
	for _, group := range source.UseGroupRecords() {
		err = target.AddUseGroup(group)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyMesoCatalogs(source *meso.Net, target *micro.Net) error {
	var err error
	if source.Config() != nil {
		err = target.SetConfig(source.Config())
		if err != nil {
			return err
		}
	}
	for _, zone := range source.ZoneRecords() {
		err = target.AddZone(zone)
		if err != nil {
			return err
		}
	}
	for _, definition := range source.UseDefinitionRecords() {
		err = target.AddUseDefinition(definition)
		if err != nil {
			return err
		}
	}
	for _, group := range source.UseGroupRecords() {
		err = target.AddUseGroup(group)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyMacroRoadAttributes(source *macro.Link, target *meso.Link) {
	meso.WithLinkName(source.Name())(target)
	meso.WithFacilityType(source.FacilityType())(target)
	meso.WithBikeFacility(source.BikeFacility())(target)
	meso.WithPedFacility(source.PedFacility())(target)
	meso.WithParking(source.Parking())(target)
	meso.WithJurisdiction(source.Jurisdiction())(target)
	grade, present := source.Grade()
	if present {
		meso.WithGrade(grade)(target)
	}
	width, present := source.RightOfWayWidthMeters()
	if present {
		meso.WithRightOfWayWidthMeters(width)(target)
	}
	toll, present := source.Toll()
	if present {
		meso.WithToll(toll)(target)
	}
}

func copyMesoRoadAttributes(source *meso.Link, target *micro.Link) {
	micro.WithLinkName(source.Name())(target)
	micro.WithFacilityType(source.FacilityType())(target)
	micro.WithBikeFacility(source.BikeFacility())(target)
	micro.WithPedFacility(source.PedFacility())(target)
	micro.WithParking(source.Parking())(target)
	micro.WithJurisdiction(source.Jurisdiction())(target)
	grade, present := source.Grade()
	if present {
		micro.WithGrade(grade)(target)
	}
	width, present := source.RightOfWayWidthMeters()
	if present {
		micro.WithRightOfWayWidthMeters(width)(target)
	}
	toll, present := source.Toll()
	if present {
		micro.WithToll(toll)(target)
	}
	micro.WithLanesNum(1)(target)
}

func copyMacroEndpoint(source *macro.Node, target *micro.Node) {
	micro.WithNodeName(source.Name())(target)
	micro.WithNodeType(source.NodeType())(target)
	micro.WithNodeControlType(source.ControlType())(target)
	altitude, present := source.AltitudeMeters()
	if present && target.Geom() == source.Geom() {
		micro.WithAltitudeMeters(altitude)(target)
	}
}

// Authored meso heights cannot be projected onto displaced cells without a vertical profile.
func requireSupportedMesoAltitudes(source *macro.Net, target *meso.Net) error {
	for _, node := range target.Nodes {
		altitude, present := node.AltitudeMeters()
		if !present {
			continue
		}
		parent := source.Nodes[node.MacroNode()]
		if parent != nil && parent.Geom() == node.Geom() {
			expected, supplied := parent.AltitudeMeters()
			if supplied && math.Float64bits(expected) == math.Float64bits(altitude) {
				continue
			}
		}
		return fmt.Errorf("%w: meso node %v: micro generation cannot place supplied altitude without a vertical profile", ErrUnsupportedGenerationInput, node.ID)
	}
	return nil
}
