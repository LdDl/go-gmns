package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type zoneCase struct {
	name    string
	options []func(*Node)
	zoneID  gmns.NodeID
	present bool
}

func TestZoneGeoJSON(t *testing.T) {
	cases := []zoneCase{
		{
			name: "absent",
		},
		{
			name: "zero is a supplied zone ID",
			options: []func(*Node){
				WithMacroZone(gmns.NodeIDFromInt64(0)),
			},
			zoneID:  gmns.NodeIDFromInt64(0),
			present: true,
		},
		{
			name: "negative one is a supplied zone ID",
			options: []func(*Node){
				WithMacroZone(gmns.NodeIDFromInt64(-1)),
			},
			zoneID:  gmns.NodeIDFromInt64(-1),
			present: true,
		},
		{
			name: "positive zone ID",
			options: []func(*Node){
				WithMacroZone(gmns.NodeIDFromInt64(42)),
			},
			zoneID:  gmns.NodeIDFromInt64(42),
			present: true,
		},
		{
			name: "replace a zone with zero",
			options: []func(*Node){
				WithMacroZone(gmns.NodeIDFromInt64(42)),
				WithMacroZone(gmns.NodeIDFromInt64(0)),
			},
			zoneID:  gmns.NodeIDFromInt64(0),
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			point := orb.Point{37, 55}
			node := NewNodeFrom(gmns.NodeIDFromInt64(7), testCase.options...)
			WithPointGeom(point)(node)
			WithParentNodeID(gmns.NodeIDFromInt64(99))(node)
			data, err := node.GeoFeature().MarshalJSON()
			assert.NoError(t, err)
			if err != nil {
				return
			}
			feature, err := geojson.UnmarshalFeature(data)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			value, exists := feature.Properties["zone_id"]
			assert.True(t, exists)
			if testCase.present {
				integer, isInteger := testCase.zoneID.Int64()
				assert.True(t, isInteger)
				assert.Equal(t, float64(integer), value)
			} else {
				assert.Nil(t, value, "node identity and parent reference do not imply a zone")
			}
			zoneID, present := node.MacroZone()
			assert.Equal(t, testCase.zoneID, zoneID)
			assert.Equal(t, testCase.present, present)
			assert.Equal(t, gmns.NodeIDFromInt64(7), node.ID)
			parentID, hasParent := node.ParentNode()
			assert.True(t, hasParent)
			assert.Equal(t, gmns.NodeIDFromInt64(99), parentID)
			assert.Equal(t, point, node.Geom())
			assert.Equal(t, point, feature.Geometry)
		})
	}
}
