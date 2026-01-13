package micro

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
	f.Properties["meso_link_id"] = link.MesoLink()
	f.Properties["macro_link_id"] = link.MacroLink()
	f.Properties["macro_node_id"] = link.MacroNode()
	f.Properties["cell_type"] = link.CellType().String()
	f.Properties["lane_id"] = link.LaneID()
	f.Properties["is_first_movement_cell"] = link.IsFirstMovementCell()
	f.Properties["movement_composite_type"] = link.MovementCompositeType().String()
	f.Properties["additional_travel_cost"] = link.AdditionalTravelCost()
	f.Properties["meso_link_type"] = link.MesoLinkType().String()
	f.Properties["control_type"] = link.ControlType().String()
	allowedAgentTypes := link.AllowedAgentTypes()
	allowedAgentTypesStrs := make([]string, len(allowedAgentTypes))
	for i, agentType := range allowedAgentTypes {
		allowedAgentTypesStrs[i] = agentType.String()
	}
	f.Properties["allowed_agent_types"] = strings.Join(allowedAgentTypesStrs, ",")
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
	f.Properties["meso_link_id"] = node.MesoLink()
	f.Properties["lane_id"] = node.LaneID()
	f.Properties["cell_index"] = node.CellIndex()
	f.Properties["is_upstream_end"] = node.IsUpstreamEnd()
	f.Properties["is_downstream_end"] = node.IsDownstreamEnd()
	f.Properties["zone_id"] = node.ZoneID()
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
