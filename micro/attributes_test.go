package micro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type suppliedAttributeCase struct {
	name     string
	toll     float64
	hasToll  bool
	facility string
	uses     string
	control  types.ControlType
}

func TestSuppliedAttributesGeoJSON(t *testing.T) {
	cases := []suppliedAttributeCase{
		{
			name: "absent",
		},
		{
			name:     "zero and custom categories",
			hasToll:  true,
			facility: "custom arterial / дорога",
			uses:     "delivery, special group",
			control:  "custom control",
		},
		{
			name:    "fractional toll and explicit common",
			toll:    2.123456789,
			hasToll: true,
			control: types.CONTROL_TYPE_NOT_SIGNAL,
		},
		{
			name:    "negative supplied toll is not absence",
			toll:    -1,
			hasToll: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2),
				WithFacilityType(testCase.facility),
				WithAllowedUses(testCase.uses),
				WithControlType(testCase.control),
				WithMesoLinkType(types.LINK_PRIMARY),
				WithAllowedAgentTypes([]types.AgentType{
					types.AGENT_WALK,
				}),
			)
			if testCase.hasToll {
				WithToll(testCase.toll)(link)
			}
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
			toll, present := link.Toll()
			assert.Equal(t, testCase.toll, toll)
			assert.Equal(t, testCase.hasToll, present)
			if testCase.hasToll {
				assert.Equal(t, testCase.toll, feature.Properties["toll"])
			} else {
				assert.Contains(t, feature.Properties, "toll")
				assert.Nil(t, feature.Properties["toll"])
			}
			assert.Equal(t, testCase.facility, link.FacilityType())
			assert.Equal(t, testCase.uses, link.AllowedUses())
			assert.Equal(t, testCase.facility, feature.Properties["facility_type"])
			assert.Equal(t, testCase.uses, feature.Properties["allowed_uses"])
			assert.Equal(t, testCase.control.String(), feature.Properties["control_type"])
			WithFacilityType("")(link)
			WithAllowedUses("")(link)
			WithControlType("")(link)
			assert.Empty(t, link.FacilityType())
			assert.Empty(t, link.AllowedUses())
			assert.Empty(t, link.ControlType())
			assert.Equal(t, "", link.GeoFeature().Properties["facility_type"])
			assert.Equal(t, "", link.GeoFeature().Properties["allowed_uses"])
			assert.Equal(t, "", link.GeoFeature().Properties["control_type"])
			assert.Equal(t, types.LINK_PRIMARY, link.MesoLinkType())
			assert.Equal(t, []types.AgentType{types.AGENT_WALK}, link.AllowedAgentTypes())
		})
	}
}

func TestNodeControlPresence(t *testing.T) {
	node := NewNodeFrom(gmns.NodeIDFromInt64(1))
	assert.Empty(t, node.ControlType())
	assert.Equal(t, "", node.GeoFeature().Properties["control_type"])
	WithNodeControlType("custom control")(node)
	assert.Equal(t, types.ControlType("custom control"), node.ControlType())
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
	assert.Equal(t, "custom control", feature.Properties["control_type"])
	WithNodeControlType(types.CONTROL_TYPE_NO_CONTROL)(node)
	assert.Equal(t, "no_control", node.GeoFeature().Properties["control_type"])
	WithNodeControlType("")(node)
	assert.Empty(t, node.ControlType())
	assert.Equal(t, "", node.GeoFeature().Properties["control_type"])
}

func TestLinkControlDefaultsToAbsent(t *testing.T) {
	link := NewLinkFrom(gmns.LinkIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.NodeIDFromInt64(3))
	assert.Empty(t, link.ControlType())
	assert.Equal(t, "", link.GeoFeature().Properties["control_type"])
}

func TestSuppliedConnectorCode(t *testing.T) {
	link := NewLinkFrom(gmns.LinkIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.NodeIDFromInt64(3))
	assert.Empty(t, link.MovementCompositeType())
	assert.Equal(t, "", link.GeoFeature().Properties["movement_composite_type"])
	WithMovementCompositeType("NER2")(link)
	assert.Equal(t, movement.MovementCompositeType("NER2"), link.MovementCompositeType())
	assert.Equal(t, "NER2", link.GeoFeature().Properties["movement_composite_type"])
	WithMovementCompositeType("")(link)
	assert.Empty(t, link.MovementCompositeType())
	assert.Equal(t, "", link.GeoFeature().Properties["movement_composite_type"])
}
