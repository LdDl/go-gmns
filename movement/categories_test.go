package movement

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type movementCategoryCase struct {
	name    string
	value   MovementType
	numeric uint16
}

func TestMovementCategoryGeoJSON(t *testing.T) {
	cases := []movementCategoryCase{
		{
			name:    "undefined",
			value:   MOVEMENT_TYPE_UNDEFINED,
			numeric: 0,
		},
		{
			name:    "thru",
			value:   MOVEMENT_TYPE_THRU,
			numeric: 1,
		},
		{
			name:    "right",
			value:   MOVEMENT_TYPE_RIGHT,
			numeric: 2,
		},
		{
			name:    "left",
			value:   MOVEMENT_TYPE_LEFT,
			numeric: 3,
		},
		{
			name:    "uturn",
			value:   MOVEMENT_TYPE_U_TURN,
			numeric: 4,
		},
		{
			name:    "merge",
			value:   MOVEMENT_TYPE_MERGE,
			numeric: 5,
		},
		{
			name:    "diverge",
			value:   MOVEMENT_TYPE_DIVERGE,
			numeric: 6,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(1, 2, 3, 4, MOVEMENT_UNDEFINED, testCase.value)
			assert.Equal(t, testCase.numeric, uint16(mvmt.Type()))
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
			assert.Equal(t, "common", feature.Properties["control_type"], "the existing default is unchanged")
			WithType(MOVEMENT_TYPE_LEFT)(mvmt)
			assert.Equal(t, MOVEMENT_TYPE_LEFT, mvmt.Type())
			assert.Equal(t, "left", mvmt.GeoFeature().Properties["type"])
		})
	}
}
