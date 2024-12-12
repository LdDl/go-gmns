package macro

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
	f.Properties["osm_way_id"] = link.OSMWay()
	f.Properties["source_osm_node_id"] = link.SourceOSMNode()
	f.Properties["target_osm_node_id"] = link.TargetOSMNode()
	f.Properties["link_class"] = link.LinkClass().String()
	f.Properties["is_link"] = link.LinkConnectionType().String()
	f.Properties["link_type"] = link.LinkType().String()
	f.Properties["control_type"] = link.ControlType().String()
	allowedAgentTypes := link.AllowedAgentTypes()
	allowedAgentTypesStrs := make([]string, len(link.AllowedAgentTypes()))
	for i, agentType := range allowedAgentTypes {
		allowedAgentTypesStrs[i] = agentType.String()
	}
	f.Properties["allowed_agent_types"] = strings.Join(allowedAgentTypesStrs, ",")
	f.Properties["was_bidirectional"] = link.WasBidirectional()
	f.Properties["lanes"] = link.LanesNum()
	f.Properties["max_speed"] = link.MaxSpeed()
	f.Properties["free_speed"] = link.FreeSpeed()
	f.Properties["capacity"] = link.Capacity()
	f.Properties["length_meters"] = link.LengthMeters()
	f.Properties["name"] = link.Name()
	return f
}

// GeoFeature returns GeoJSON Point feature for the given node
func (node *Node) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(node.Geom())
	f.ID = node.ID
	f.Properties["id"] = node.ID
	f.Properties["osm_node_id"] = node.OSMNode()
	f.Properties["control_type"] = node.ControlType().String()
	f.Properties["boundary_type"] = node.BoundaryType().String()
	f.Properties["activity_type"] = node.ActivityType().String()
	f.Properties["activity_link_type"] = node.ActivityLinkType().String()
	f.Properties["zone_id"] = node.Zone()
	f.Properties["intersection_id"] = node.Intersection()
	f.Properties["poi_id"] = node.POI()
	f.Properties["osm_highway"] = node.OSMHighway()
	f.Properties["name"] = node.Name()
	return f
}
