package movement

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type capacityCase struct {
	name    string
	options []func(*Movement)
	want    float64
	present bool
}

func TestCapacityGeoJSON(t *testing.T) {
	cases := []capacityCase{
		{
			name: "absent",
		},
		{
			name: "zero",
			options: []func(*Movement){
				WithCapacity(0),
			},
			present: true,
		},
		{
			name: "fractional",
			options: []func(*Movement){
				WithCapacity(777.25),
			},
			want:    777.25,
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(gmns.MovementIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.LinkIDFromInt64(3), gmns.LinkIDFromInt64(4), MOVEMENT_SBT, MOVEMENT_TYPE_THRU, testCase.options...)
			WithLanesNum(2)(mvmt)
			WithPenalty(12.5)(mvmt)
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
			value, exists := feature.Properties["capacity"]
			assert.True(t, exists, "capacity property is missing")
			if testCase.present {
				assert.Equal(t, testCase.want, value, "movement capacity must not be scaled by lane count")
			} else {
				assert.Nil(t, value, "absent capacity must be null")
			}
			assert.Equal(t, 12.5, feature.Properties["penalty"], "capacity must not replace penalty")
			capacity, present := mvmt.Capacity()
			assert.Equal(t, testCase.want, capacity, "capacity after export")
			assert.Equal(t, testCase.present, present, "capacity presence after export")
		})
	}
}
