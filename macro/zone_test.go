package macro

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
				WithZoneID(0),
			},
			present: true,
		},
		{
			name: "negative one is a supplied zone ID",
			options: []func(*Node){
				WithZoneID(-1),
			},
			zoneID:  -1,
			present: true,
		},
		{
			name: "positive zone ID",
			options: []func(*Node){
				WithZoneID(42),
			},
			zoneID:  42,
			present: true,
		},
		{
			name: "replace a zone with zero",
			options: []func(*Node){
				WithZoneID(42),
				WithZoneID(0),
			},
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			point := orb.Point{37, 55}
			node := NewNodeFrom(7, testCase.options...)
			WithPointGeom(point)(node)
			WithParentNodeID(99)(node)
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
				assert.Equal(t, float64(testCase.zoneID), value)
			} else {
				assert.Nil(t, value, "node identity and parent reference do not imply a zone")
			}
			zoneID, present := node.Zone()
			assert.Equal(t, testCase.zoneID, zoneID)
			assert.Equal(t, testCase.present, present)
			assert.Equal(t, gmns.NodeID(7), node.ID)
			parentID, hasParent := node.ParentNode()
			assert.True(t, hasParent)
			assert.Equal(t, gmns.NodeID(99), parentID)
			assert.Equal(t, point, node.Geom())
			assert.Equal(t, point, feature.Geometry)
		})
	}
}
