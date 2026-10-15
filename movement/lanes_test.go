package movement

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type laneCase struct {
	name    string
	options []func(*Movement)
	values  [4]int
	present [4]bool
}

func TestLanePresenceGeoJSON(t *testing.T) {
	cases := []laneCase{
		{
			name: "absent",
		},
		{
			name: "explicit negative numbers",
			options: []func(*Movement){
				WithIncomeLane(-2, -1),
				WithOutcomeLane(-1, -1),
			},
			values:  [4]int{-2, -1, -1, -1},
			present: [4]bool{true, true, true, true},
		},
		{
			name: "explicit zeros",
			options: []func(*Movement){
				WithIncomeLane(0, 0),
				WithOutcomeLane(0, 0),
			},
			present: [4]bool{true, true, true, true},
		},
		{
			name: "single lanes with absent endpoints",
			options: []func(*Movement){
				WithIncomeLaneStart(-1),
				WithOutcomeLaneStart(2),
			},
			values:  [4]int{-1, 0, 2, 0},
			present: [4]bool{true, false, true, false},
		},
		{
			name: "endpoints supplied independently",
			options: []func(*Movement){
				WithIncomeLaneEnd(-1),
				WithOutcomeLaneEnd(3),
			},
			values:  [4]int{0, -1, 0, 3},
			present: [4]bool{false, true, false, true},
		},
		{
			name: "individual options preserve other bound",
			options: []func(*Movement){
				WithIncomeLane(1, 2),
				WithIncomeLaneEnd(3),
				WithOutcomeLane(1, 4),
				WithOutcomeLaneStart(2),
			},
			values:  [4]int{1, 3, 2, 4},
			present: [4]bool{true, true, true, true},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(gmns.MovementIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.LinkIDFromInt64(3), gmns.LinkIDFromInt64(4), MOVEMENT_SBT, MOVEMENT_TYPE_THRU, testCase.options...)
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
			properties := []string{"in_lane_start", "in_lane_end", "out_lane_start", "out_lane_end"}
			for index, property := range properties {
				value, exists := feature.Properties[property]
				assert.True(t, exists, property)
				if testCase.present[index] {
					assert.Equal(t, float64(testCase.values[index]), value, property)
				} else {
					assert.Nil(t, value, property)
				}
			}
			incomeStart, hasIncomeStart := mvmt.IncomeLaneStart()
			incomeEnd, hasIncomeEnd := mvmt.IncomeLaneEnd()
			outcomeStart, hasOutcomeStart := mvmt.OutcomeLaneStart()
			outcomeEnd, hasOutcomeEnd := mvmt.OutcomeLaneEnd()
			assert.Equal(t, testCase.values, [4]int{incomeStart, incomeEnd, outcomeStart, outcomeEnd})
			assert.Equal(t, testCase.present, [4]bool{hasIncomeStart, hasIncomeEnd, hasOutcomeStart, hasOutcomeEnd})
			assert.Equal(t, -1, mvmt.StartIncomeLaneSeqID(), "exported lane numbers must not set calculated sequence indices")
			assert.Equal(t, -1, mvmt.EndIncomeLaneSeqID())
			assert.Equal(t, -1, mvmt.StartOutcomeLaneSeqID())
			assert.Equal(t, -1, mvmt.EndOutcomeLaneSeqID())
		})
	}
}
