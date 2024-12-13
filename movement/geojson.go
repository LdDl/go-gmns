package movement

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

// GeoFeature returns GeoJSON LineString feature for the given movement
func (mvmt *Movement) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(nil)
	f.ID = mvmt.ID
	f.Properties["id"] = mvmt.ID
	f.Properties["node_id"] = mvmt.MacroNode()
	f.Properties["osm_node_id"] = mvmt.OSMNode()
	f.Properties["name"] = mvmt.Name()
	f.Properties["in_link_id"] = mvmt.IncomeMacroLink()
	f.Properties["in_lane_start"] = mvmt.IncomeLaneStart()
	f.Properties["in_lane_end"] = mvmt.IncomeLaneEnd()
	f.Properties["out_link_id"] = mvmt.OutcomeMacroLink()
	f.Properties["out_lane_start"] = mvmt.OutcomeLaneStart()
	f.Properties["out_lane_end"] = mvmt.OutcomeLaneEnd()
	f.Properties["lanes_num"] = mvmt.LanesNum()
	f.Properties["from_osm_node_id"] = mvmt.OSMNodeFrom()
	f.Properties["to_osm_node_id"] = mvmt.OSMNodeTo()
	f.Properties["type"] = mvmt.Type()
	f.Properties["penalty"] = -1  // @todo: future works
	f.Properties["capacity"] = -1 // @todo: future works
	f.Properties["control_type"] = mvmt.ControlType().String()
	f.Properties["movement_composite_type"] = mvmt.MvmtTextID().String()
	f.Properties["volume"] = -1     // @todo: future works
	f.Properties["free_speed"] = -1 // @todo: future works
	allowedAgentTypes := mvmt.AllowedAgentTypes()
	allowedAgentTypesStrs := make([]string, len(mvmt.AllowedAgentTypes()))
	for i, agentType := range allowedAgentTypes {
		allowedAgentTypesStrs[i] = agentType.String()
	}
	f.Properties["allowed_agent_types"] = strings.Join(allowedAgentTypesStrs, ",")
	return f
}

// GeoFeatureCollection returns GeoJSON FeatureCollection consisting of Linestring features for the given movements data
func (mvmts MovementsStorage) GeoFeatureCollection() *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()
	for _, mvmt := range mvmts {
		fc.Append(mvmt.GeoFeature())
	}
	return fc
}
