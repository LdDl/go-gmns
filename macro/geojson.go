package macro

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

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

func (node *Node) GeoFeature() geojson.Feature {
	return geojson.Feature{}
}
