package movement

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type movementCategoryCase struct {
	name  string
	value MovementType
}

func TestMovementCategoryGeoJSON(t *testing.T) {
	cases := []movementCategoryCase{
		{
			name: "",
		},
		{
			name:  "custom maneuver",
			value: MovementType("custom maneuver"),
		},
		{
			name:  "undefined",
			value: MOVEMENT_TYPE_UNDEFINED,
		},
		{
			name:  "thru",
			value: MOVEMENT_TYPE_THRU,
		},
		{
			name:  "right",
			value: MOVEMENT_TYPE_RIGHT,
		},
		{
			name:  "left",
			value: MOVEMENT_TYPE_LEFT,
		},
		{
			name:  "uturn",
			value: MOVEMENT_TYPE_U_TURN,
		},
		{
			name:  "merge",
			value: MOVEMENT_TYPE_MERGE,
		},
		{
			name:  "diverge",
			value: MOVEMENT_TYPE_DIVERGE,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(1, 2, 3, 4, MOVEMENT_UNDEFINED, testCase.value)
			assert.Equal(t, testCase.name, mvmt.Type().String())
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
			assert.Equal(t, testCase.name, feature.Properties["type"])
			assert.Equal(t, testCase.value, mvmt.Type())
			assert.Equal(t, MOVEMENT_UNDEFINED, mvmt.MvmtTextID(), "a supplied type must not fabricate a directional code")
			assert.Equal(t, "", feature.Properties["control_type"], "absence must not become a supplied category")
			WithType(MOVEMENT_TYPE_LEFT)(mvmt)
			assert.Equal(t, MOVEMENT_TYPE_LEFT, mvmt.Type())
			assert.Equal(t, "left", mvmt.GeoFeature().Properties["type"])
		})
	}
}
