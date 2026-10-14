package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

func TestGeoFeaturePreservesLinkProperties(t *testing.T) {
	link := NewLinkFrom(1, 2, 3,
		WithLineMacroNodeID(17),
		WithLineMacroLinkID(29),
		WithLinkType(types.LINK_SECONDARY),
		WithControlType(types.CONTROL_TYPE_IS_SIGNAL),
	)
	data, err := link.GeoFeature().MarshalJSON()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	feature, err := geojson.UnmarshalFeature(data)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, float64(17), feature.Properties["macro_node_id"])
	assert.Equal(t, float64(29), feature.Properties["macro_link_id"])
	assert.Equal(t, "secondary", feature.Properties["link_type"])
	assert.Equal(t, "signal", feature.Properties["control_type"])
}
