package macro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
)

// Clone returns an independent copy, including calculated data.
func (node *Node) Clone() *Node {
	if node == nil {
		return nil
	}
	copyRecord := *node
	if node.altitudeMeters != nil {
		value := *node.altitudeMeters
		copyRecord.altitudeMeters = &value
	}
	if node.parentNodeID != nil {
		value := *node.parentNodeID
		copyRecord.parentNodeID = &value
	}
	if node.incomingLinks != nil {
		copyRecord.incomingLinks = make([]gmns.LinkID, len(node.incomingLinks))
		copy(copyRecord.incomingLinks, node.incomingLinks)
	}
	if node.outcomingLinks != nil {
		copyRecord.outcomingLinks = make([]gmns.LinkID, len(node.outcomingLinks))
		copy(copyRecord.outcomingLinks, node.outcomingLinks)
	}
	if node.zoneID != nil {
		value := *node.zoneID
		copyRecord.zoneID = &value
	}
	return &copyRecord
}

// Clone returns an independent copy, including calculated data.
func (link *Link) Clone() *Link {
	if link == nil {
		return nil
	}
	copyRecord := *link
	if link.toll != nil {
		value := *link.toll
		copyRecord.toll = &value
	}
	if link.rightOfWayWidthMeters != nil {
		value := *link.rightOfWayWidthMeters
		copyRecord.rightOfWayWidthMeters = &value
	}
	if link.gradePercent != nil {
		value := *link.gradePercent
		copyRecord.gradePercent = &value
	}
	if link.parentLinkID != nil {
		value := *link.parentLinkID
		copyRecord.parentLinkID = &value
	}
	if link.lanesInfo.LanesList != nil {
		copyRecord.lanesInfo.LanesList = make([]int, len(link.lanesInfo.LanesList))
		copy(copyRecord.lanesInfo.LanesList, link.lanesInfo.LanesList)
	}
	if link.lanesInfo.LanesChange != nil {
		copyRecord.lanesInfo.LanesChange = make([][2]int, len(link.lanesInfo.LanesChange))
		copy(copyRecord.lanesInfo.LanesChange, link.lanesInfo.LanesChange)
	}
	if link.lanesInfo.LanesChangePoints != nil {
		copyRecord.lanesInfo.LanesChangePoints = make([]float64, len(link.lanesInfo.LanesChangePoints))
		copy(copyRecord.lanesInfo.LanesChangePoints, link.lanesInfo.LanesChangePoints)
	}
	if link.geom != nil {
		copyRecord.geom = make(orb.LineString, len(link.geom))
		copy(copyRecord.geom, link.geom)
	}
	if link.geomEuclidean != nil {
		copyRecord.geomEuclidean = make(orb.LineString, len(link.geomEuclidean))
		copy(copyRecord.geomEuclidean, link.geomEuclidean)
	}
	if link.geometryID != nil {
		value := *link.geometryID
		copyRecord.geometryID = &value
	}
	if link.dirFlag != nil {
		value := *link.dirFlag
		copyRecord.dirFlag = &value
	}
	if link.allowedAgentTypes != nil {
		copyRecord.allowedAgentTypes = make([]types.AgentType, len(link.allowedAgentTypes))
		copy(copyRecord.allowedAgentTypes, link.allowedAgentTypes)
	}
	if link.lengthMeters != nil {
		value := *link.lengthMeters
		copyRecord.lengthMeters = &value
	}
	if link.freeSpeed != nil {
		value := *link.freeSpeed
		copyRecord.freeSpeed = &value
	}
	if link.lanesNum != nil {
		value := *link.lanesNum
		copyRecord.lanesNum = &value
	}
	if link.capacity != nil {
		value := *link.capacity
		copyRecord.capacity = &value
	}
	return &copyRecord
}

// Clone returns independent records and catalogs without recalculating geometry.
func (net *Net) Clone() *Net {
	if net == nil {
		return nil
	}
	copyNet := *net
	copyNet.config = net.config.Clone()
	if net.Nodes != nil {
		copyNet.Nodes = make(map[gmns.NodeID]*Node, len(net.Nodes))
		for id, record := range net.Nodes {
			copyNet.Nodes[id] = record.Clone()
		}
	}
	if net.Links != nil {
		copyNet.Links = make(map[gmns.LinkID]*Link, len(net.Links))
		for id, record := range net.Links {
			copyNet.Links[id] = record.Clone()
		}
	}
	if net.lanes != nil {
		copyNet.lanes = make(map[gmns.LaneID]*gmns.Lane, len(net.lanes))
		for id, record := range net.lanes {
			copyNet.lanes[id] = record.Clone()
		}
	}
	if net.segments != nil {
		copyNet.segments = make(map[gmns.SegmentID]*gmns.Segment, len(net.segments))
		for id, record := range net.segments {
			copyNet.segments[id] = record.Clone()
		}
	}
	if net.segmentLanes != nil {
		copyNet.segmentLanes = make(map[gmns.SegmentLaneID]*gmns.SegmentLane, len(net.segmentLanes))
		for id, record := range net.segmentLanes {
			copyNet.segmentLanes[id] = record.Clone()
		}
	}
	if net.geometries != nil {
		copyNet.geometries = make(map[gmns.GeometryID]*gmns.Geometry, len(net.geometries))
		for id, record := range net.geometries {
			copyNet.geometries[id] = record.Clone()
		}
	}
	if net.zones != nil {
		copyNet.zones = make(map[gmns.NodeID]*gmns.Zone, len(net.zones))
		for id, record := range net.zones {
			copyNet.zones[id] = record.Clone()
		}
	}
	if net.useDefinitions != nil {
		copyNet.useDefinitions = make(map[string]*gmns.UseDefinition, len(net.useDefinitions))
		for id, record := range net.useDefinitions {
			copyNet.useDefinitions[id] = record.Clone()
		}
	}
	if net.useGroups != nil {
		copyNet.useGroups = make(map[string]*gmns.UseGroup, len(net.useGroups))
		for id, record := range net.useGroups {
			copyNet.useGroups[id] = record.Clone()
		}
	}
	if net.movements != nil {
		copyNet.movements = make(map[gmns.MovementID]*movement.Movement, len(net.movements))
		for id, record := range net.movements {
			copyNet.movements[id] = record.Clone()
		}
	}
	return &copyNet
}
