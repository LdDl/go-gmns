package gmns

import "github.com/paulmach/orb/encoding/wkt"

// WKT encodes the stored canonical shape without validation or direction changes.
func (geometry *Geometry) WKT() string {
	return wkt.MarshalString(geometry.geom)
}

// WKT encodes the zone boundary as a two-dimensional MultiPolygon without validation.
func (zone *Zone) WKT() string {
	return wkt.MarshalString(zone.boundary)
}
