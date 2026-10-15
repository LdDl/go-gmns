package meso

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

// GeoFeature returns GeoJSON LineString feature for the given link
func (link *Link) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(link.Geom())
	f.Properties["name"] = link.Name()
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
	f.Properties["facility_type"] = link.FacilityType()
	f.Properties["allowed_uses"] = link.AllowedUses()
	toll, hasToll := link.Toll()
	f.Properties["toll"] = nil
	if hasToll {
		f.Properties["toll"] = toll
	}
	f.Properties["macro_node_id"] = link.MacroNode()
	f.Properties["macro_link_id"] = link.MacroLink()
	f.Properties["link_type"] = link.LinkType().String()
	f.Properties["control_type"] = link.ControlType().String()
	f.Properties["movement_id"] = link.Movement()
	f.Properties["movement_composite_type"] = link.MvmtTextID().String()
	allowedAgentTypes := link.AllowedAgentTypes()
	allowedAgentTypesStrs := make([]string, len(link.AllowedAgentTypes()))
	for i, agentType := range allowedAgentTypes {
		allowedAgentTypesStrs[i] = agentType.String()
	}
	f.Properties["allowed_agent_types"] = strings.Join(allowedAgentTypesStrs, ",")
	lanesNum, hasLanes := link.LanesNum()
	f.Properties["lanes_num"] = nil
	if hasLanes {
		f.Properties["lanes_num"] = lanesNum
	}
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
	lengthMeters, hasLength := link.LengthMeters()
	f.Properties["length_meters"] = nil
	if hasLength {
		f.Properties["length_meters"] = lengthMeters
	}
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
	f.Properties["control_type"] = node.ControlType().String()
	f.Properties["name"] = node.Name()
	f.ID = node.ID
	f.Properties["id"] = node.ID
	parentNodeID, hasParentNode := node.ParentNode()
	f.Properties["parent_node_id"] = nil
	if hasParentNode {
		f.Properties["parent_node_id"] = parentNodeID
	}
	zoneID, hasZone := node.MacroZone()
	f.Properties["zone_id"] = nil
	if hasZone {
		f.Properties["zone_id"] = zoneID
	}
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
