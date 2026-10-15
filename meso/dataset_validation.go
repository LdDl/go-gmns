package meso

import (
	"fmt"

	"github.com/LdDl/go-gmns/gmns"
)

// ValidateDataset checks the prepared records without constructing, normalizing or mutating them.
func (net *Net) ValidateDataset() error {
	if net == nil {
		return fmt.Errorf("network is nil")
	}
	err := net.config.Validate()
	if err != nil {
		return err
	}
	kind := net.config.IDType()
	err = net.ValidateIDType(kind)
	if err != nil {
		return err
	}
	for _, zone := range net.ZoneRecords() {
		err = zone.Validate()
		if err != nil {
			return err
		}
		err = zone.ID.ValidateType(kind)
		if err != nil {
			return err
		}
		visited := make(map[gmns.NodeID]bool)
		current := zone
		for current != nil {
			if visited[current.ID] {
				return fmt.Errorf("zone %v: cyclic super_zone", zone.ID)
			}
			visited[current.ID] = true
			parent, present := current.SuperZone()
			if !present {
				break
			}
			err = parent.ValidateType(kind)
			if err != nil {
				return err
			}
			current = net.zones[parent]
			if current == nil {
				return fmt.Errorf("zone %v: super_zone %v not found", zone.ID, parent)
			}
		}
	}
	for _, shape := range net.GeometryRecords() {
		err = shape.Validate()
		if err != nil {
			return err
		}
		err = shape.ID.ValidateType(kind)
		if err != nil {
			return err
		}
	}
	for _, definition := range net.UseDefinitionRecords() {
		err = definition.Validate()
		if err != nil {
			return err
		}
		if net.useGroups[definition.Use()] != nil {
			return fmt.Errorf("ambiguous use/group name %q", definition.Use())
		}
	}
	for _, group := range net.UseGroupRecords() {
		err = group.Validate()
		if err != nil {
			return err
		}
		for _, name := range group.Uses() {
			if net.useDefinitions[name] == nil {
				return fmt.Errorf("use group %q: definition %q not found", group.Name(), name)
			}
		}
	}
	for _, id := range net.sortedNodeIDs() {
		node := net.Nodes[id]
		err = gmns.ValidateWGS84(node.Geom())
		if err != nil {
			return fmt.Errorf("node %v: %w", id, err)
		}
		altitude, present := node.AltitudeMeters()
		if present && !gmns.Finite(altitude) {
			return fmt.Errorf("node %v: altitude must be finite", id)
		}
		zone, present := node.MacroZone()
		if present && net.zones[zone] == nil {
			return fmt.Errorf("node %v: zone %v not found", id, zone)
		}
	}
	for _, id := range net.sortedLinkIDs() {
		link := net.Links[id]
		err = validateLinkValues(link)
		if err != nil {
			return fmt.Errorf("link %v: %w", id, err)
		}
		err = gmns.ValidateAllowedUses(link.AllowedUses(), net.useDefinitions, net.useGroups)
		if err != nil {
			return fmt.Errorf("link %v: %w", id, err)
		}
		_, hasToll := link.Toll()
		if hasToll && net.config.Currency() == "" {
			return fmt.Errorf("link %v: toll requires dataset currency", id)
		}
		geometryID, hasGeometry := link.GeometryID()
		direction, hasDirection := link.DirFlag()
		if hasDirection && direction != gmns.DIR_FLAG_FORWARD && direction != gmns.DIR_FLAG_REVERSE {
			return fmt.Errorf("link %v: dir_flag must be 1 or -1", id)
		}
		if hasGeometry {
			geometry := net.geometries[geometryID]
			if geometry == nil {
				return fmt.Errorf("link %v: geometry %v not found", id, geometryID)
			}
			canonical := geometry.Geom()
			inline := link.Geom()
			if len(canonical) > 0 && len(inline) > 0 {
				if len(canonical) != len(inline) {
					return fmt.Errorf("link %v: shared and inline geometries differ", id)
				}
				for index, point := range inline {
					other := index
					if hasDirection && direction == gmns.DIR_FLAG_REVERSE {
						other = len(canonical) - 1 - index
					}
					if point != canonical[other] {
						return fmt.Errorf("link %v: geometry contradicts dir_flag", id)
					}
				}
			}
		}
	}
	err = net.ValidateMovements(net.movements)
	if err != nil {
		return err
	}
	err = net.movements.ValidateIDType(kind)
	if err != nil {
		return err
	}
	for _, mvmt := range net.MovementRecords() {
		err = mvmt.Validate()
		if err != nil {
			return err
		}
		err = gmns.ValidateAllowedUses(mvmt.AllowedUses(), net.useDefinitions, net.useGroups)
		if err != nil {
			return fmt.Errorf("movement %v: %w", mvmt.ID, err)
		}
	}
	return net.validateDetailedRecords()
}

func validateLinkValues(link *Link) error {
	valueLengthMeters, hasLengthMeters := link.LengthMeters()
	if hasLengthMeters && (!gmns.Finite(valueLengthMeters) || valueLengthMeters < 0) {
		return fmt.Errorf("invalid length")
	}
	valueFreeSpeed, hasFreeSpeed := link.FreeSpeed()
	if hasFreeSpeed && (!gmns.Finite(valueFreeSpeed) || valueFreeSpeed < 0 || valueFreeSpeed > 200) {
		return fmt.Errorf("invalid free_speed")
	}
	valueCapacity, hasCapacity := link.Capacity()
	if hasCapacity && (!gmns.Finite(valueCapacity) || valueCapacity < 0) {
		return fmt.Errorf("invalid capacity")
	}
	valueRightOfWayWidthMeters, hasRightOfWayWidthMeters := link.RightOfWayWidthMeters()
	if hasRightOfWayWidthMeters && (!gmns.Finite(valueRightOfWayWidthMeters) || valueRightOfWayWidthMeters < 0) {
		return fmt.Errorf("invalid row_width")
	}
	valueGrade, hasGrade := link.Grade()
	if hasGrade && (!gmns.Finite(valueGrade) || valueGrade < -100 || valueGrade > 100) {
		return fmt.Errorf("invalid grade")
	}
	valueToll, hasToll := link.Toll()
	if hasToll && (!gmns.Finite(valueToll) || false) {
		return fmt.Errorf("invalid toll")
	}
	lanes, hasLanes := link.LanesNum()
	if hasLanes && lanes < 0 {
		return fmt.Errorf("lanes must be nonnegative")
	}
	if len(link.Geom()) > 0 {
		return gmns.ValidateLine(link.Geom())
	}
	return nil
}
