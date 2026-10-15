package micro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type linkLengthCase struct {
	name    string
	options []func(*Link)
	value   float64
	present bool
}

func TestLinkLengthGeoJSON(t *testing.T) {
	cases := []linkLengthCase{
		{
			name: "absent despite supplied geometry",
		},
		{
			name: "explicit zero",
			options: []func(*Link){
				WithLengthMeters(0),
			},
			present: true,
		},
		{
			name: "supplied negative value is not absence",
			options: []func(*Link){
				WithLengthMeters(-1),
			},
			value:   -1,
			present: true,
		},
		{
			name: "fractional supplied length",
			options: []func(*Link){
				WithLengthMeters(123.123456789),
			},
			value:   123.123456789,
			present: true,
		},
		{
			name: "replace a length with zero",
			options: []func(*Link){
				WithLengthMeters(12.5),
				WithLengthMeters(0),
			},
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{
				{30, 60},
				{30.002, 60},
			}
			link := NewLinkFrom(gmns.LinkIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.NodeIDFromInt64(3), testCase.options...)
			WithLineGeom(shape.Clone())(link)
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
			value, exists := feature.Properties["length_meters"]
			assert.True(t, exists)
			if testCase.present {
				assert.Equal(t, testCase.value, value)
			} else {
				assert.Nil(t, value, "serialization must not calculate a missing length")
			}
			length, present := link.LengthMeters()
			assert.Equal(t, testCase.value, length)
			assert.Equal(t, testCase.present, present)
			assert.Equal(t, shape, link.Geom())
			assert.Equal(t, shape, feature.Geometry)
		})
	}
}
