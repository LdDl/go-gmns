package movement

import (
	"encoding/json"
	"github.com/LdDl/go-gmns/gmns"
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type penaltyCase struct {
	name    string
	options []func(*Movement)
	want    float64
	present bool
}

func TestPenaltyGeoJSON(t *testing.T) {
	cases := []penaltyCase{
		{
			name: "absent",
		},
		{
			name: "zero",
			options: []func(*Movement){
				WithPenalty(0),
			},
			present: true,
		},
		{
			name: "fractional",
			options: []func(*Movement){
				WithPenalty(12.5),
			},
			want:    12.5,
			present: true,
		},
		{
			name: "negative is supplied",
			options: []func(*Movement){
				WithPenalty(-1),
			},
			want:    -1,
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(gmns.MovementIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.LinkIDFromInt64(3), gmns.LinkIDFromInt64(4), MOVEMENT_SBT, MOVEMENT_TYPE_THRU, testCase.options...)
			data, err := json.Marshal(mvmt.GeoFeature())
			assert.NoError(t, err)
			if err != nil {
				return
			}
			feature, err := geojson.UnmarshalFeature(data)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			value, exists := feature.Properties["penalty"]
			assert.True(t, exists, "penalty property is missing")
			if testCase.present {
				assert.Equal(t, testCase.want, value, "exported penalty")
			} else {
				assert.Nil(t, value, "absent penalty must be null")
			}
			penalty, present := mvmt.Penalty()
			assert.Equal(t, testCase.want, penalty, "penalty after export")
			assert.Equal(t, testCase.present, present, "penalty presence after export")
		})
	}
}
