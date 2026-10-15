package movement

import "github.com/paulmach/orb/encoding/wkt"

// WKT encodes the movement's two-dimensional WGS84 geometry without validation.
func (mvmt *Movement) WKT() string {
	return wkt.MarshalString(mvmt.geom)
}
