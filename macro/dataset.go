package macro

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/movement"
)

func (net *Net) SetConfig(config *gmns.Config) error {
	err := config.Validate()
	if err != nil {
		return err
	}
	net.config = config.Clone()
	return nil
}

func (net *Net) Config() *gmns.Config {
	return net.config.Clone()
}

func (net *Net) AddGeometry(record *gmns.Geometry) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	if _, exists := net.geometries[key]; exists {
		return fmt.Errorf("geometries %v: %w", key, gmns.ErrDuplicateID)
	}
	if net.geometries == nil {
		net.geometries = make(map[gmns.GeometryID]*gmns.Geometry)
	}
	net.geometries[key] = record.Clone()
	return nil
}

func (net *Net) Geometry(key gmns.GeometryID) (*gmns.Geometry, bool) {
	record, found := net.geometries[key]
	return record.Clone(), found
}

// GeometryRecords returns independent records ordered by key.
func (net *Net) GeometryRecords() []*gmns.Geometry {
	keys := make([]gmns.GeometryID, 0, len(net.geometries))
	for key := range net.geometries {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(first, second int) bool {
		return keys[first].Less(keys[second])
	})
	records := make([]*gmns.Geometry, 0, len(keys))
	for _, key := range keys {
		records = append(records, net.geometries[key].Clone())
	}
	return records
}

func (net *Net) AddZone(record *gmns.Zone) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	if _, exists := net.zones[key]; exists {
		return fmt.Errorf("zones %v: %w", key, gmns.ErrDuplicateID)
	}
	if net.zones == nil {
		net.zones = make(map[gmns.NodeID]*gmns.Zone)
	}
	net.zones[key] = record.Clone()
	return nil
}

func (net *Net) Zone(key gmns.NodeID) (*gmns.Zone, bool) {
	record, found := net.zones[key]
	return record.Clone(), found
}

// ZoneRecords returns independent records ordered by key.
func (net *Net) ZoneRecords() []*gmns.Zone {
	keys := make([]gmns.NodeID, 0, len(net.zones))
	for key := range net.zones {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(first, second int) bool {
		return keys[first].Less(keys[second])
	})
	records := make([]*gmns.Zone, 0, len(keys))
	for _, key := range keys {
		records = append(records, net.zones[key].Clone())
	}
	return records
}

func (net *Net) AddUseDefinition(record *gmns.UseDefinition) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.Use()
	if _, exists := net.useDefinitions[key]; exists {
		return fmt.Errorf("useDefinitions %v: %w", key, gmns.ErrDuplicateID)
	}
	if net.useDefinitions == nil {
		net.useDefinitions = make(map[string]*gmns.UseDefinition)
	}
	net.useDefinitions[key] = record.Clone()
	return nil
}

func (net *Net) UseDefinition(key string) (*gmns.UseDefinition, bool) {
	record, found := net.useDefinitions[key]
	return record.Clone(), found
}

// UseDefinitionRecords returns independent records ordered by key.
func (net *Net) UseDefinitionRecords() []*gmns.UseDefinition {
	keys := make([]string, 0, len(net.useDefinitions))
	for key := range net.useDefinitions {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(first, second int) bool {
		return keys[first] < keys[second]
	})
	records := make([]*gmns.UseDefinition, 0, len(keys))
	for _, key := range keys {
		records = append(records, net.useDefinitions[key].Clone())
	}
	return records
}

func (net *Net) AddUseGroup(record *gmns.UseGroup) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.Name()
	if _, exists := net.useGroups[key]; exists {
		return fmt.Errorf("useGroups %v: %w", key, gmns.ErrDuplicateID)
	}
	if net.useGroups == nil {
		net.useGroups = make(map[string]*gmns.UseGroup)
	}
	net.useGroups[key] = record.Clone()
	return nil
}

func (net *Net) UseGroup(key string) (*gmns.UseGroup, bool) {
	record, found := net.useGroups[key]
	return record.Clone(), found
}

// UseGroupRecords returns independent records ordered by key.
func (net *Net) UseGroupRecords() []*gmns.UseGroup {
	keys := make([]string, 0, len(net.useGroups))
	for key := range net.useGroups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(first, second int) bool {
		return keys[first] < keys[second]
	})
	records := make([]*gmns.UseGroup, 0, len(keys))
	for _, key := range keys {
		records = append(records, net.useGroups[key].Clone())
	}
	return records
}

func (net *Net) AddMovement(record *movement.Movement) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	key := record.ID
	if _, exists := net.movements[key]; exists {
		return fmt.Errorf("movements %v: %w", key, gmns.ErrDuplicateID)
	}
	if net.movements == nil {
		net.movements = make(map[gmns.MovementID]*movement.Movement)
	}
	net.movements[key] = record.Clone()
	return nil
}

func (net *Net) Movement(key gmns.MovementID) (*movement.Movement, bool) {
	record, found := net.movements[key]
	return record.Clone(), found
}

// MovementRecords returns independent records ordered by key.
func (net *Net) MovementRecords() []*movement.Movement {
	keys := make([]gmns.MovementID, 0, len(net.movements))
	for key := range net.movements {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(first, second int) bool {
		return keys[first].Less(keys[second])
	})
	records := make([]*movement.Movement, 0, len(keys))
	for _, key := range keys {
		records = append(records, net.movements[key].Clone())
	}
	return records
}

// MovementStorage returns a detached snapshot for existing generator APIs.
func (net *Net) MovementStorage() movement.MovementsStorage {
	result := movement.NewMovementsStorage()
	for id, record := range net.movements {
		result[id] = record.Clone()
	}
	return result
}
