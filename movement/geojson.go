package movement

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

// GeoFeature returns GeoJSON LineString feature for the given movement
func (mvmt *Movement) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(mvmt.Geom())
	f.ID = mvmt.ID
	f.Properties["id"] = mvmt.ID
	f.Properties["node_id"] = mvmt.Node()
	f.Properties["osm_node_id"] = mvmt.OSMNode()
	f.Properties["name"] = mvmt.Name()
	f.Properties["in_link_id"] = mvmt.IncomeLink()
	incomeLaneStart, hasIncomeLaneStart := mvmt.IncomeLaneStart()
	f.Properties["in_lane_start"] = nil
	if hasIncomeLaneStart {
		f.Properties["in_lane_start"] = incomeLaneStart
	}
	incomeLaneEnd, hasIncomeLaneEnd := mvmt.IncomeLaneEnd()
	f.Properties["in_lane_end"] = nil
	if hasIncomeLaneEnd {
		f.Properties["in_lane_end"] = incomeLaneEnd
	}
	f.Properties["out_link_id"] = mvmt.OutcomeLink()
	outcomeLaneStart, hasOutcomeLaneStart := mvmt.OutcomeLaneStart()
	f.Properties["out_lane_start"] = nil
	if hasOutcomeLaneStart {
		f.Properties["out_lane_start"] = outcomeLaneStart
	}
	outcomeLaneEnd, hasOutcomeLaneEnd := mvmt.OutcomeLaneEnd()
	f.Properties["out_lane_end"] = nil
	if hasOutcomeLaneEnd {
		f.Properties["out_lane_end"] = outcomeLaneEnd
	}
	f.Properties["lanes_num"] = mvmt.LanesNum()
	f.Properties["from_osm_node_id"] = mvmt.OSMNodeSource()
	f.Properties["to_osm_node_id"] = mvmt.OSMNodeTarget()
	f.Properties["type"] = mvmt.Type().String()
	f.Properties["allowed_uses"] = mvmt.AllowedUses()
	penalty, hasPenalty := mvmt.Penalty()
	f.Properties["penalty"] = nil
	if hasPenalty {
		f.Properties["penalty"] = penalty
	}
	capacity, hasCapacity := mvmt.Capacity()
	f.Properties["capacity"] = nil
	if hasCapacity {
		f.Properties["capacity"] = capacity
	}
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
