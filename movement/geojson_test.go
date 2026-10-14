package movement

import (
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

func TestGeoFeaturePreservesGeometry(t *testing.T) {
	shape := orb.LineString{
		{37, 55},
		{37.005, 55.002},
		{37.01, 55.01},
	}
	before := shape.Clone()
	mvmt := NewMovement(1, 2, 3, 4, MOVEMENT_SBT, MOVEMENT_TYPE_THRU,
		WithGeom(shape),
	)
	data, err := mvmt.GeoFeature().MarshalJSON()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	feature, err := geojson.UnmarshalFeature(data)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, before, feature.Geometry, "exported movement geometry")
	assert.Equal(t, before, mvmt.Geom(), "export changed the movement geometry")
}
