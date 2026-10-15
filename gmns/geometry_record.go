package gmns

import (
	"fmt"

	"github.com/paulmach/orb"
)

// Geometry stores the canonical shape referenced by links. Link direction is separate.
type Geometry struct {
	ID   GeometryID
	geom orb.LineString
}

func NewGeometry(id GeometryID, shape orb.LineString) *Geometry {
	return &Geometry{ID: id, geom: shape.Clone()}
}

func (geometry *Geometry) Geom() orb.LineString {
	return geometry.geom.Clone()
}

func (geometry *Geometry) Clone() *Geometry {
	if geometry == nil {
		return nil
	}
	return NewGeometry(geometry.ID, geometry.geom)
}

func (geometry *Geometry) Validate() error {
	if geometry == nil {
		return fmt.Errorf("geometry is nil")
	}
	err := geometry.ID.Validate()
	if err != nil {
		return err
	}
	// Geometry is optional in the standard record, but required when resolving a shape.
	if len(geometry.geom) == 0 {
		return nil
	}
	return ValidateLine(geometry.geom)
}
