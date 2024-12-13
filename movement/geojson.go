package movement

import (
	"github.com/paulmach/orb/geojson"
)

// GeoFeature returns GeoJSON LineString feature for the given movement
func (mvmt *Movement) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(nil)
	f.ID = mvmt.ID
	panic("@tbd")
	return f
}

// GeoFeatureCollection returns GeoJSON FeatureCollection consisting of Linestring features for the given movements data
func (mvmts MovementsStorage) GeoFeatureCollection() *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()
	for _, mvmt := range mvmts {
		fc.Append(mvmt.GeoFeature())
	}
	return fc
}
