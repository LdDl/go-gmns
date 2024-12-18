package meso

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

// GeoFeature returns GeoJSON LineString feature for the given link
func (link *Link) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(link.Geom())
	f.ID = link.ID
	f.Properties["id"] = link.ID
	f.Properties["source_node"] = link.SourceNode()
	f.Properties["target_node"] = link.TargetNode()
	f.Properties["macro_node_id"] = link.MacroNode()
	f.Properties["macro_node_id"] = link.MacroLink()
	f.Properties["link_type"] = link.LinkType()
	f.Properties["control_type"] = link.ControlType()
	f.Properties["movement_id"] = link.Movement()
	f.Properties["movement_composite_type"] = link.MvmtTextID().String()
	allowedAgentTypes := link.AllowedAgentTypes()
	allowedAgentTypesStrs := make([]string, len(link.AllowedAgentTypes()))
	for i, agentType := range allowedAgentTypes {
		allowedAgentTypesStrs[i] = agentType.String()
	}
	f.Properties["allowed_agent_types"] = strings.Join(allowedAgentTypesStrs, ",")
	f.Properties["lanes_num"] = link.LanesNum()
	f.Properties["free_speed"] = link.FreeSpeed()
	f.Properties["capacity"] = link.Capacity()
	f.Properties["length_meters"] = link.LengthMeters()
	return f
}

// GeoFeature returns GeoJSON Point feature for the given node
func (node *Node) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(node.Geom())
	f.ID = node.ID
	f.Properties["id"] = node.ID
	f.Properties["zone_id"] = node.MacroZone()
	f.Properties["macro_node_id"] = node.MacroNode()
	f.Properties["macro_link_id"] = node.MacroLink()
	f.Properties["activity_link_type"] = node.ActivityLinkType().String()
	f.Properties["boundary_type"] = node.BoundaryType().String()
	return f
}

// GeoFeatureCollection returns GeoJSON FeatureCollection consisting of Linestring and Point features for the given road network
func (net *Net) GeoFeatureCollection() *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()
	for _, node := range net.Nodes {
		fc.Append(node.GeoFeature())
	}
	for _, link := range net.Links {
		fc.Append(link.GeoFeature())
	}
	return fc
}
