package movement

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type movementCodeCase struct {
	name string
	code MovementCompositeType
}

func TestSuppliedMovementCodeAndUses(t *testing.T) {
	cases := []movementCodeCase{
		{
			name: "absent",
		},
		{
			name: "numbered diagonal code",
			code: "NER2",
		},
		{
			name: "custom code retained for later validation",
			code: "custom code",
		},
		{
			name: "explicit legacy undefined",
			code: MOVEMENT_UNDEFINED,
		},
		{
			name: "legacy U-turn code",
			code: MOVEMENT_SBU,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mvmt := NewMovement(gmns.MovementIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.LinkIDFromInt64(3), gmns.LinkIDFromInt64(4), testCase.code, "custom maneuver",
				WithControlType("custom control"),
				WithAllowedUses("delivery, special group"),
				WithAllowedAgentTypes([]types.AgentType{
					types.AGENT_WALK,
				}),
			)
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
			assert.Equal(t, testCase.code, mvmt.MvmtTextID())
			assert.Equal(t, testCase.code.String(), feature.Properties["movement_composite_type"])
			assert.Equal(t, "custom maneuver", feature.Properties["type"])
			assert.Equal(t, "custom control", feature.Properties["control_type"])
			assert.Equal(t, "delivery, special group", feature.Properties["allowed_uses"])
			assert.Equal(t, "delivery, special group", mvmt.AllowedUses())
			WithMvmtTextID("")(mvmt)
			WithType("")(mvmt)
			WithControlType("")(mvmt)
			WithAllowedUses("")(mvmt)
			assert.Empty(t, mvmt.MvmtTextID())
			assert.Empty(t, mvmt.Type())
			assert.Empty(t, mvmt.ControlType())
			assert.Empty(t, mvmt.AllowedUses())
			properties := mvmt.GeoFeature().Properties
			assert.Equal(t, "", properties["movement_composite_type"])
			assert.Equal(t, "", properties["type"])
			assert.Equal(t, "", properties["control_type"])
			assert.Equal(t, "", properties["allowed_uses"])
			assert.Equal(t, []types.AgentType{types.AGENT_WALK}, mvmt.AllowedAgentTypes())
		})
	}
}
