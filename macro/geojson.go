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
	parentLinkID, hasParentLink := link.ParentLink()
	f.Properties["parent_link_id"] = nil
	if hasParentLink {
		f.Properties["parent_link_id"] = parentLinkID
	}
	f.Properties["source_node"] = link.SourceNode()
	f.Properties["target_node"] = link.TargetNode()
	f.Properties["directed"] = link.Directed()
	rightOfWayWidthMeters, hasRightOfWayWidth := link.RightOfWayWidthMeters()
	f.Properties["row_width"] = nil
	if hasRightOfWayWidth {
		f.Properties["row_width"] = rightOfWayWidthMeters
	}
	grade, hasGrade := link.Grade()
	f.Properties["grade"] = nil
	if hasGrade {
		f.Properties["grade"] = grade
	}
	f.Properties["jurisdiction"] = link.Jurisdiction()
	f.Properties["bike_facility"] = link.BikeFacility()
	f.Properties["ped_facility"] = link.PedFacility()
	f.Properties["parking"] = link.Parking()
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
	freeSpeed, hasFreeSpeed := link.FreeSpeed()
	f.Properties["free_speed"] = nil
	if hasFreeSpeed {
		f.Properties["free_speed"] = freeSpeed
	}
	capacity, hasCapacity := link.Capacity()
	f.Properties["capacity"] = nil
	if hasCapacity {
		f.Properties["capacity"] = capacity
	}
	f.Properties["length_meters"] = link.LengthMeters()
	f.Properties["name"] = link.Name()
	return f
}

// GeoFeature returns GeoJSON Point feature for the given node
func (node *Node) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(node.Geom())
	altitudeMeters, hasAltitude := node.AltitudeMeters()
	f.Properties["z_coord"] = nil
	if hasAltitude {
		f.Properties["z_coord"] = altitudeMeters
	}
	f.Properties["node_type"] = node.NodeType()
	f.ID = node.ID
	f.Properties["id"] = node.ID
	parentNodeID, hasParentNode := node.ParentNode()
	f.Properties["parent_node_id"] = nil
	if hasParentNode {
		f.Properties["parent_node_id"] = parentNodeID
	}
	f.Properties["osm_node_id"] = node.OSMNode()
	f.Properties["control_type"] = node.ControlType().String()
	f.Properties["boundary_type"] = node.BoundaryType().String()
	f.Properties["activity_type"] = node.ActivityType().String()
	f.Properties["activity_link_type"] = node.ActivityLinkType().String()
	zoneID, hasZone := node.Zone()
	f.Properties["zone_id"] = nil
	if hasZone {
		f.Properties["zone_id"] = zoneID
	}
	f.Properties["intersection_id"] = node.Intersection()
	f.Properties["poi_id"] = node.POI()
	f.Properties["osm_highway"] = node.OSMHighway()
	f.Properties["name"] = node.Name()
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
