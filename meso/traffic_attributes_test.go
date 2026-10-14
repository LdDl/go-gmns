package meso

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type trafficAttributesCase struct {
	name        string
	options     []func(*Link)
	freeSpeed   float64
	hasSpeed    bool
	capacity    float64
	hasCapacity bool
}

func TestTrafficAttributesGeoJSON(t *testing.T) {
	cases := []trafficAttributesCase{
		{
			name: "both absent",
		},
		{
			name: "zero capacity with absent speed",
			options: []func(*Link){
				WithCapacity(0),
			},
			hasCapacity: true,
		},
		{
			name: "zero speed with absent capacity",
			options: []func(*Link){
				WithFreeSpeed(0),
			},
			hasSpeed: true,
		},
		{
			name: "fractional supplied values",
			options: []func(*Link){
				WithCapacity(1234.123456789),
				WithFreeSpeed(43.123456789),
			},
			capacity:    1234.123456789,
			hasCapacity: true,
			freeSpeed:   43.123456789,
			hasSpeed:    true,
		},
		{
			name: "supplied negative values are not treated as absence",
			options: []func(*Link){
				WithCapacity(-1),
				WithFreeSpeed(-1),
			},
			capacity:    -1,
			hasCapacity: true,
			freeSpeed:   -1,
			hasSpeed:    true,
		},
		{
			name: "replace supplied values with zeros",
			options: []func(*Link){
				WithCapacity(1800.5),
				WithFreeSpeed(40.5),
				WithCapacity(0),
				WithFreeSpeed(0),
			},
			hasCapacity: true,
			hasSpeed:    true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			link := NewLinkFrom(1, 2, 3, testCase.options...)
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

			capacity, hasCapacity := link.Capacity()
			assert.Equal(t, testCase.capacity, capacity)
			assert.Equal(t, testCase.hasCapacity, hasCapacity)
			value, exists := feature.Properties["capacity"]
			assert.True(t, exists)
			if testCase.hasCapacity {
				assert.Equal(t, testCase.capacity, value)
			} else {
				assert.Nil(t, value)
			}

			freeSpeed, hasSpeed := link.FreeSpeed()
			assert.Equal(t, testCase.freeSpeed, freeSpeed)
			assert.Equal(t, testCase.hasSpeed, hasSpeed)
			value, exists = feature.Properties["free_speed"]
			assert.True(t, exists)
			if testCase.hasSpeed {
				assert.Equal(t, testCase.freeSpeed, value)
			} else {
				assert.Nil(t, value)
			}
		})
	}
}
