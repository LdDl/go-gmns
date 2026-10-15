package meso

import (
	"fmt"
	"strings"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/movement"
)

// UpdateGeometry replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateGeometry(record *gmns.Geometry) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	old, found := net.geometries[key]
	if !found {
		return fmt.Errorf("geometries %v not found", key)
	}
	net.geometries[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.geometries[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveGeometry(key gmns.GeometryID) error {
	if _, found := net.geometries[key]; !found {
		return fmt.Errorf("geometries %v not found", key)
	}
	for _, link := range net.Links {
		if link == nil {
			continue
		}
		id, present := link.GeometryID()
		if present && id == key {
			return fmt.Errorf("geometry %v is referenced by link %v", key, link.ID)
		}
	}
	delete(net.geometries, key)
	return nil
}

// UpdateZone replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateZone(record *gmns.Zone) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	old, found := net.zones[key]
	if !found {
		return fmt.Errorf("zones %v not found", key)
	}
	net.zones[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.zones[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveZone(key gmns.NodeID) error {
	if _, found := net.zones[key]; !found {
		return fmt.Errorf("zones %v not found", key)
	}
	for _, node := range net.Nodes {
		if node == nil {
			continue
		}
		id, present := node.MacroZone()
		if present && id == key {
			return fmt.Errorf("zone %v is referenced by node %v", key, node.ID)
		}
	}
	for _, zone := range net.zones {
		id, present := zone.SuperZone()
		if present && id == key {
			return fmt.Errorf("zone %v is referenced by zone %v", key, zone.ID)
		}
	}
	delete(net.zones, key)
	return nil
}

// UpdateUseDefinition replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateUseDefinition(record *gmns.UseDefinition) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.Use()
	old, found := net.useDefinitions[key]
	if !found {
		return fmt.Errorf("useDefinitions %v not found", key)
	}
	net.useDefinitions[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.useDefinitions[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveUseDefinition(key string) error {
	if _, found := net.useDefinitions[key]; !found {
		return fmt.Errorf("useDefinitions %v not found", key)
	}
	if net.referencesUse(key) {
		return fmt.Errorf("use %q is referenced by a record", key)
	}
	for _, group := range net.useGroups {
		for _, use := range group.Uses() {
			if use == key {
				return fmt.Errorf("use %q is referenced by group %q", key, group.Name())
			}
		}
	}
	delete(net.useDefinitions, key)
	return nil
}

// UpdateUseGroup replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateUseGroup(record *gmns.UseGroup) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.Name()
	old, found := net.useGroups[key]
	if !found {
		return fmt.Errorf("useGroups %v not found", key)
	}
	net.useGroups[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.useGroups[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveUseGroup(key string) error {
	if _, found := net.useGroups[key]; !found {
		return fmt.Errorf("useGroups %v not found", key)
	}
	if net.referencesUse(key) {
		return fmt.Errorf("use %q is referenced by a record", key)
	}
	delete(net.useGroups, key)
	return nil
}

// UpdateMovement replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateMovement(record *movement.Movement) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	old, found := net.movements[key]
	if !found {
		return fmt.Errorf("movements %v not found", key)
	}
	net.movements[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.movements[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveMovement(key gmns.MovementID) error {
	if _, found := net.movements[key]; !found {
		return fmt.Errorf("movements %v not found", key)
	}
	delete(net.movements, key)
	return nil
}

// UpdateLane replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateLane(record *gmns.Lane) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	old, found := net.lanes[key]
	if !found {
		return fmt.Errorf("lanes %v not found", key)
	}
	net.lanes[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.lanes[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveLane(key gmns.LaneID) error {
	if _, found := net.lanes[key]; !found {
		return fmt.Errorf("lanes %v not found", key)
	}
	for _, lane := range net.segmentLanes {
		parent, present := lane.ParentLane()
		if present && parent == key {
			return fmt.Errorf("lane %v is referenced by segment lane %v", key, lane.ID)
		}
	}
	delete(net.lanes, key)
	return nil
}

// UpdateSegment replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateSegment(record *gmns.Segment) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	old, found := net.segments[key]
	if !found {
		return fmt.Errorf("segments %v not found", key)
	}
	net.segments[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.segments[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveSegment(key gmns.SegmentID) error {
	if _, found := net.segments[key]; !found {
		return fmt.Errorf("segments %v not found", key)
	}
	for _, lane := range net.segmentLanes {
		if lane.Segment() == key {
			return fmt.Errorf("segment %v is referenced by segment lane %v", key, lane.ID)
		}
	}
	delete(net.segments, key)
	return nil
}

// UpdateSegmentLane replaces an existing record only if the resulting dataset is valid.
func (net *Net) UpdateSegmentLane(record *gmns.SegmentLane) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	old, found := net.segmentLanes[key]
	if !found {
		return fmt.Errorf("segmentLanes %v not found", key)
	}
	net.segmentLanes[key] = record.Clone()
	err = net.ValidateDataset()
	if err != nil {
		net.segmentLanes[key] = old
		return err
	}
	return nil
}

func (net *Net) RemoveSegmentLane(key gmns.SegmentLaneID) error {
	if _, found := net.segmentLanes[key]; !found {
		return fmt.Errorf("segmentLanes %v not found", key)
	}
	delete(net.segmentLanes, key)
	return nil
}

func (net *Net) referencesUse(key string) bool {
	for _, link := range net.Links {
		if link != nil && containsUse(link.AllowedUses(), key) {
			return true
		}
	}
	for _, mvmt := range net.movements {
		if containsUse(mvmt.AllowedUses(), key) {
			return true
		}
	}
	for _, lane := range net.lanes {
		if containsUse(lane.AllowedUses(), key) {
			return true
		}
	}
	for _, segment := range net.segments {
		if containsUse(segment.AllowedUses(), key) {
			return true
		}
	}
	for _, lane := range net.segmentLanes {
		if containsUse(lane.AllowedUses(), key) {
			return true
		}
	}
	return false
}

func containsUse(text, key string) bool {
	for _, value := range strings.Split(text, ",") {
		if strings.TrimSpace(value) == key {
			return true
		}
	}
	return false
}
