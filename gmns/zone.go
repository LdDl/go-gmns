package gmns

import (
	"fmt"

	"github.com/paulmach/orb"
)

// Zone uses the same ID type as existing node zone references.
type Zone struct {
	ID        NodeID
	name      string
	boundary  orb.MultiPolygon
	superZone *NodeID
}

func NewZone(id NodeID, options ...func(*Zone)) *Zone {
	zone := &Zone{ID: id}
	for _, option := range options {
		option(zone)
	}
	return zone
}

func (zone *Zone) Name() string {
	return zone.name
}

func (zone *Zone) Boundary() orb.MultiPolygon {
	return zone.boundary.Clone()
}

func (zone *Zone) SuperZone() (NodeID, bool) {
	if zone.superZone == nil {
		return NodeID{}, false
	}
	return *zone.superZone, true
}

func (zone *Zone) Clone() *Zone {
	if zone == nil {
		return nil
	}
	result := *zone
	result.boundary = zone.boundary.Clone()
	if zone.superZone != nil {
		value := *zone.superZone
		result.superZone = &value
	}
	return &result
}

func (zone *Zone) Validate() error {
	if zone == nil {
		return fmt.Errorf("zone is nil")
	}
	err := zone.ID.Validate()
	if err != nil {
		return err
	}
	parent, present := zone.SuperZone()
	if present {
		err = parent.Validate()
		if err != nil {
			return err
		}
	}
	for _, polygon := range zone.boundary {
		if len(polygon) == 0 {
			return fmt.Errorf("zone polygon is empty")
		}
		for _, ring := range polygon {
			if len(ring) < 4 || ring[0] != ring[len(ring)-1] {
				return fmt.Errorf("zone ring must be closed with at least four points")
			}
			for _, point := range ring {
				err = ValidateWGS84(point)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func WithZoneName(name string) func(*Zone) {
	return func(zone *Zone) {
		zone.name = name
	}
}

func WithZoneBoundary(boundary orb.MultiPolygon) func(*Zone) {
	return func(zone *Zone) {
		zone.boundary = boundary.Clone()
	}
}

func WithSuperZone(id NodeID) func(*Zone) {
	return func(zone *Zone) {
		zone.superZone = &id
	}
}

func WithoutSuperZone() func(*Zone) {
	return func(zone *Zone) {
		zone.superZone = nil
	}
}
