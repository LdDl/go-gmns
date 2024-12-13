package macro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

// Link representation of an edge in a macroscopic graph
type Link struct {
	lanesInfo          LanesInfo
	name               string
	geom               orb.LineString
	geomEuclidean      orb.LineString
	allowedAgentTypes  []types.AgentType
	lengthMeters       float64
	freeSpeed          float64
	maxSpeed           float64
	ID                 gmns.LinkID
	sourceNodeID       gmns.NodeID
	targetNodeID       gmns.NodeID
	osmWayID           osm.WayID
	sourceOsmNodeID    osm.NodeID
	targetOsmNodeID    osm.NodeID
	lanesNum           int
	capacity           int
	linkClass          types.LinkClass
	linkType           types.LinkType
	linkConnectionType types.LinkConnectionType
	controlType        types.ControlType
	wasBidirectional   bool
}

// NewLinkFrom creates pointer to the new Link
func NewLinkFrom(id gmns.LinkID, sourceNodeID, targetNodeID gmns.NodeID, options ...func(*Link)) *Link {
	newLink := &Link{
		lanesInfo:          LanesInfo{},
		name:               "",
		geom:               orb.LineString{},
		geomEuclidean:      orb.LineString{},
		allowedAgentTypes:  []types.AgentType{},
		lengthMeters:       -1,
		freeSpeed:          -1,
		maxSpeed:           -1,
		ID:                 id,
		sourceNodeID:       sourceNodeID,
		targetNodeID:       targetNodeID,
		osmWayID:           osm.WayID(-1),
		sourceOsmNodeID:    osm.NodeID(-1),
		targetOsmNodeID:    osm.NodeID(-1),
		lanesNum:           -1,
		capacity:           -1,
		linkClass:          types.LINK_CLASS_UNDEFINED,
		linkType:           types.LINK_UNDEFINED,
		linkConnectionType: types.NOT_A_LINK,
		controlType:        types.CONTROL_TYPE_NOT_SIGNAL,
		wasBidirectional:   false,
	}
	for _, option := range options {
		option(newLink)
	}
	return newLink
}

// MaxLanes returns max number of lanes on the link. This method is not strictly corresponds to any of the struct field i.e. it is not just getter.
func (link *Link) MaxLanes() int {
	if len(link.lanesInfo.LanesList) == 0 {
		return -1
	}
	max := link.lanesInfo.LanesList[0]
	for _, lane := range link.lanesInfo.LanesList {
		if lane > max {
			max = lane
		}
	}
	return max
}

// GetIncomingLanes returns number of incoming lanes. The first value in lanes list is that number.
func (link *Link) GetIncomingLanes() int {
	if len(link.lanesInfo.LanesList) == 0 {
		return 0
	}
	return link.lanesInfo.LanesList[0]
}

// GetOutcomingLanes returns number of outcoming lanes. The last value in lanes list is that number. Outputs "-1" in case of no lanes at all.
func (link *Link) GetOutcomingLanes() int {
	idx := len(link.lanesInfo.LanesList) - 1
	if idx < 0 {
		return -1
	}
	return link.lanesInfo.LanesList[idx]
}

// GetOutcomingLaneIndices returns slice of the lane indices (for lane changes)
func (link *Link) GetOutcomingLaneIndices() []int {
	lanesInfo := link.lanesInfo
	idx := len(lanesInfo.LanesChange) - 1
	if idx < 0 {
		return make([]int, 0)
	}
	return laneIndices(link.lanesNum, lanesInfo.LanesChange[idx][0], lanesInfo.LanesChange[idx][1])
}

// Name returns link alias.
func (link *Link) Name() string {
	return link.name
}

// LanesInfo just returns copy of the lanes information object.
func (link *Link) LanesInfo() LanesInfo {
	return link.lanesInfo
}

// Geom returns corresponding geometry [WGS84]
func (link *Link) Geom() orb.LineString {
	return link.geom
}

// GeomEuclidean returns corresponding geometry [Euclidean]
func (link *Link) GeomEuclidean() orb.LineString {
	return link.geomEuclidean
}

// AllowedAgentTypes returns set of allowed agent types. Warning: returning object is a slice.
func (link *Link) AllowedAgentTypes() []types.AgentType {
	return link.allowedAgentTypes
}

// LengthMeters returns length of the underlying geometry [WGS84]. Outputs "-1" if it was not set.
func (link *Link) LengthMeters() float64 {
	return link.lengthMeters
}

// FreeSpeed returns free flow speed for the link. Outputs "-1" if it was not set.
func (link *Link) FreeSpeed() float64 {
	return link.freeSpeed
}

// MaxSpeed returns speed restriction the link. Outputs "-1" if it was not set.
func (link *Link) MaxSpeed() float64 {
	return link.maxSpeed
}

// OSMWay returns link identifier corresponding to OSM data. Outputs "-1" if it was not set.
func (link *Link) OSMWay() osm.WayID {
	return link.osmWayID
}

// SourceNode returns identifier of source node. Outputs "-1" if it was not set.
func (link *Link) SourceNode() gmns.NodeID {
	return link.sourceNodeID
}

// TargetNode returns identifier of target node. Outputs "-1" if it was not set.
func (link *Link) TargetNode() gmns.NodeID {
	return link.targetNodeID
}

// SourceOSMNode returns source node identifier corresponding to OSM data
func (link *Link) SourceOSMNode() osm.NodeID {
	return link.sourceOsmNodeID
}

// TargetOSMNode returns target node identifier corresponding to OSM data
func (link *Link) TargetOSMNode() osm.NodeID {
	return link.targetOsmNodeID
}

// LanesNum returns number of lanes. Outputs "-1" if it was not set.
func (link *Link) LanesNum() int {
	return link.lanesNum
}

// Capacity returns max capacity of the link. Outputs "-1" if it was not set.
func (link *Link) Capacity() int {
	return link.capacity
}

// LinkClass returns link class
func (link *Link) LinkClass() types.LinkClass {
	return link.linkClass
}

// LinkType returns link type
func (link *Link) LinkType() types.LinkType {
	return link.linkType
}

// LinkConnectionType returns type of the link connection
func (link *Link) LinkConnectionType() types.LinkConnectionType {
	return link.linkConnectionType
}

// ControlType returns control type
func (link *Link) ControlType() types.ControlType {
	return link.controlType
}

// WasBidirectional returns whether if the link was bidirectional
func (link *Link) WasBidirectional() bool {
	return link.wasBidirectional
}

// WithLinkName sets alias for the link
func WithLinkName(name string) func(*Link) {
	return func(link *Link) {
		link.name = name
	}
}

// WithLanesInfo sets information about lanes for the link
func WithLanesInfo(lanesInfo LanesInfo) func(*Link) {
	return func(link *Link) {
		link.lanesInfo = lanesInfo
	}
}

// WithLineGeom sets geometry [WGS84] for the link. Warning: it does not copy the given slice.
func WithLineGeom(geom orb.LineString) func(*Link) {
	return func(link *Link) {
		link.geom = geom
	}
}

// WithLineGeomEuclidean sets geometry [Euclidean] for the link. Warning: it does not copy the given slice.
func WithLineGeomEuclidean(geomEuclidean orb.LineString) func(*Link) {
	return func(link *Link) {
		link.geomEuclidean = geomEuclidean
	}
}

// WithAllowedAgentTypes sets allowed agent types for the link. Warning: it does copy argument
func WithAllowedAgentTypes(allowedAgentTypes []types.AgentType) func(*Link) {
	return func(link *Link) {
		link.allowedAgentTypes = make([]types.AgentType, len(allowedAgentTypes))
		copy(link.allowedAgentTypes, allowedAgentTypes)
	}
}

// WithLengthMeters sets underlying geometry [WGS84] length in meters. Warning: it should be called explicitly after setting geometry [WGS84]
func WithLengthMeters(lengthMeters float64) func(*Link) {
	return func(link *Link) {
		link.lengthMeters = lengthMeters
	}
}

// WithFreeSpeed sets free flow speed for the link
func WithFreeSpeed(freeSpeed float64) func(*Link) {
	return func(link *Link) {
		link.freeSpeed = freeSpeed
	}
}

// WithMaxSpeed sets speed restriction the link
func WithMaxSpeed(maxSpeed float64) func(*Link) {
	return func(link *Link) {
		link.maxSpeed = maxSpeed
	}
}

// WithOSMWayID sets corresponding OSM node identifier for the link
func WithOSMWayID(osmWayID osm.WayID) func(*Link) {
	return func(link *Link) {
		link.osmWayID = osmWayID
	}
}

// WithSourceNodeID sets source node identifier
func WithSourceNodeID(nodeID gmns.NodeID) func(*Link) {
	return func(link *Link) {
		link.sourceNodeID = nodeID
	}
}

// WithTargetNodeID sets target node identifier
func WithTargetNodeID(nodeID gmns.NodeID) func(*Link) {
	return func(link *Link) {
		link.targetNodeID = nodeID
	}
}

// WithSourceOSMNodeID set corresponding source OSM node identifier for the link
func WithSourceOSMNodeID(osmNodeID osm.NodeID) func(*Link) {
	return func(link *Link) {
		link.sourceOsmNodeID = osmNodeID
	}
}

// WithTargetOSMNodeID sets corresponding target OSM node identifier for the link
func WithTargetOSMNodeID(osmNodeID osm.NodeID) func(*Link) {
	return func(link *Link) {
		link.targetOsmNodeID = osmNodeID
	}
}

// WithLanesNum sets number of lanes for the link
func WithLanesNum(lanesNum int) func(*Link) {
	return func(link *Link) {
		link.lanesNum = lanesNum
	}
}

// WithCapacity sets max capacity for the link
func WithCapacity(capacity int) func(*Link) {
	return func(link *Link) {
		link.capacity = capacity
	}
}

// WithLinkClass sets link class
func WithLinkClass(capacity types.LinkClass) func(*Link) {
	return func(link *Link) {
		link.linkClass = capacity
	}
}

// WithLinkType sets link type
func WithLinkType(capacity types.LinkType) func(*Link) {
	return func(link *Link) {
		link.linkType = capacity
	}
}

// WithLinkConnectionType sets type for the link connection
func WithLinkConnectionType(capacity types.LinkConnectionType) func(*Link) {
	return func(link *Link) {
		link.linkConnectionType = capacity
	}
}

// WithLinkControlType sets control type for the link. Should correspond source vertex
func WithLinkControlType(controlType types.ControlType) func(*Link) {
	return func(link *Link) {
		link.controlType = controlType
	}
}

// WithBidirectionalSource sets information about whether source data (e.g. OSM) has bidirectional (undirected) flag for the link
func WithBidirectionalSource(wasBidirectional bool) func(*Link) {
	return func(link *Link) {
		link.wasBidirectional = wasBidirectional
	}
}
