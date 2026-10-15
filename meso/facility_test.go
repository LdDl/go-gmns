package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type facilityCase struct {
	name         string
	options      []func(*Link)
	bikeFacility string
	pedFacility  string
	parking      string
}

func TestFacilityAttributesGeoJSON(t *testing.T) {
	cases := []facilityCase{
		{
			name: "absent",
		},
		{
			name: "suggested categories",
			options: []func(*Link){
				WithBikeFacility(types.BIKE_FACILITY_SEPARATED_LANE),
				WithPedFacility(types.PED_FACILITY_SIDEWALK),
				WithParking(types.PARKING_PARALLEL),
			},
			bikeFacility: "separated bike lane",
			pedFacility:  "sidewalk",
			parking:      "parallel",
		},
		{
			name: "explicit none is different from absent",
			options: []func(*Link){
				WithBikeFacility(types.BIKE_FACILITY_NONE),
				WithPedFacility(types.PED_FACILITY_NONE),
				WithParking(types.PARKING_NONE),
			},
			bikeFacility: "none",
			pedFacility:  "none",
			parking:      "none",
		},
		{
			name: "unknown is preserved",
			options: []func(*Link){
				WithPedFacility(types.PED_FACILITY_UNKNOWN),
				WithParking(types.PARKING_UNKNOWN),
			},
			pedFacility: "unknown",
			parking:     "unknown",
		},
		{
			name: "custom text is preserved",
			options: []func(*Link){
				WithBikeFacility("  custom:велодорожка  "),
				WithPedFacility("custom:covered passage"),
				WithParking("custom:loading; permit-only"),
			},
			bikeFacility: "  custom:велодорожка  ",
			pedFacility:  "custom:covered passage",
			parking:      "custom:loading; permit-only",
		},
		{
			name: "clear bicycle facility independently",
			options: []func(*Link){
				WithBikeFacility(types.BIKE_FACILITY_SHARED_LANE),
				WithPedFacility(types.PED_FACILITY_CROSSWALK),
				WithParking(types.PARKING_ANGLE),
				WithBikeFacility(""),
			},
			pedFacility: "crosswalk",
			parking:     "angle",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			allowed := []types.AgentType{types.AGENT_AUTO}
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), WithAllowedAgentTypes(allowed))
			for _, option := range testCase.options {
				option(link)
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
			assert.Equal(t, testCase.bikeFacility, feature.Properties["bike_facility"])
			assert.Equal(t, testCase.pedFacility, feature.Properties["ped_facility"])
			assert.Equal(t, testCase.parking, feature.Properties["parking"])
			assert.Equal(t, testCase.bikeFacility, link.BikeFacility())
			assert.Equal(t, testCase.pedFacility, link.PedFacility())
			assert.Equal(t, testCase.parking, link.Parking())
			assert.Equal(t, allowed, link.AllowedAgentTypes(), "infrastructure must not grant travel permission")
			assert.Equal(t, "auto", feature.Properties["allowed_agent_types"])
		})
	}
}

func TestAllowedAgentsDoNotInventFacilities(t *testing.T) {
	allowed := []types.AgentType{types.AGENT_BIKE, types.AGENT_WALK}
	link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), WithAllowedAgentTypes(allowed))
	feature := link.GeoFeature()
	assert.Equal(t, "", feature.Properties["bike_facility"])
	assert.Equal(t, "", feature.Properties["ped_facility"])
	assert.Equal(t, "", feature.Properties["parking"])
	assert.Equal(t, allowed, link.AllowedAgentTypes())
	WithBikeFacility(types.BIKE_FACILITY_NONE)(link)
	WithPedFacility(types.PED_FACILITY_NONE)(link)
	assert.Equal(t, allowed, link.AllowedAgentTypes(), "absence of dedicated infrastructure must not revoke travel permission")
}
