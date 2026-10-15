package movement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type optionalMovementFieldCase struct {
	name     string
	set      func(*Movement)
	clear    func(*Movement)
	present  func(*Movement) bool
	property string
}

func TestClearOptionalMovementFields(t *testing.T) {
	cases := []optionalMovementFieldCase{
		{
			name:  "Penalty",
			set:   WithPenalty(0),
			clear: WithoutPenalty(),
			present: func(mvmt *Movement) bool {
				_, present := mvmt.Penalty()
				return present
			},
			property: "penalty",
		},
		{
			name:  "Capacity",
			set:   WithCapacity(0),
			clear: WithoutCapacity(),
			present: func(mvmt *Movement) bool {
				_, present := mvmt.Capacity()
				return present
			},
			property: "capacity",
		},
		{
			name:  "IncomeLaneStart",
			set:   WithIncomeLaneStart(0),
			clear: WithoutIncomeLaneStart(),
			present: func(mvmt *Movement) bool {
				_, present := mvmt.IncomeLaneStart()
				return present
			},
			property: "in_lane_start",
		},
		{
			name:  "IncomeLaneEnd",
			set:   WithIncomeLaneEnd(0),
			clear: WithoutIncomeLaneEnd(),
			present: func(mvmt *Movement) bool {
				_, present := mvmt.IncomeLaneEnd()
				return present
			},
			property: "in_lane_end",
		},
		{
			name:  "OutcomeLaneStart",
			set:   WithOutcomeLaneStart(0),
			clear: WithoutOutcomeLaneStart(),
			present: func(mvmt *Movement) bool {
				_, present := mvmt.OutcomeLaneStart()
				return present
			},
			property: "out_lane_start",
		},
		{
			name:  "OutcomeLaneEnd",
			set:   WithOutcomeLaneEnd(0),
			clear: WithoutOutcomeLaneEnd(),
			present: func(mvmt *Movement) bool {
				_, present := mvmt.OutcomeLaneEnd()
				return present
			},
			property: "out_lane_end",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			first := NewMovement(1, 2, 3, 4, "NER2", "custom maneuver", testCase.set)
			second := NewMovement(5, 2, 3, 4, "", "", testCase.set)
			for _, field := range cases {
				field.set(first)
			}
			assert.True(t, testCase.present(first))
			testCase.clear(first)
			testCase.clear(first)
			assert.False(t, testCase.present(first))
			assert.True(t, testCase.present(second))
			for _, field := range cases {
				if field.name != testCase.name {
					assert.True(t, field.present(first), "clearing %s must preserve %s", testCase.name, field.name)
				}
			}
			properties := first.GeoFeature().Properties
			assert.Contains(t, properties, testCase.property)
			assert.Nil(t, properties[testCase.property])
			assert.Equal(t, MovementCompositeType("NER2"), first.MvmtTextID())
			assert.Equal(t, MovementType("custom maneuver"), first.Type())
			assert.Equal(t, second.MacroNode(), first.MacroNode())
			testCase.set(first)
			assert.True(t, testCase.present(first))
		})
	}
}

func TestClearingLaneEndPreservesSignedStartAndSequence(t *testing.T) {
	mvmt := NewMovement(1, 2, 3, 4, "", "",
		WithIncomeLane(-2, -1),
		WithIncomeLaneSequence(0, 1),
		WithOutcomeLane(-3, -2),
		WithOutcomeLaneSequence(2, 3),
	)
	WithoutIncomeLaneEnd()(mvmt)
	WithoutOutcomeLaneEnd()(mvmt)
	start, present := mvmt.IncomeLaneStart()
	assert.Equal(t, -2, start)
	assert.True(t, present)
	start, present = mvmt.OutcomeLaneStart()
	assert.Equal(t, -3, start)
	assert.True(t, present)
	assert.Equal(t, 0, mvmt.StartIncomeLaneSeqID())
	assert.Equal(t, 1, mvmt.EndIncomeLaneSeqID())
	assert.Equal(t, 2, mvmt.StartOutcomeLaneSeqID())
	assert.Equal(t, 3, mvmt.EndOutcomeLaneSeqID())
}
