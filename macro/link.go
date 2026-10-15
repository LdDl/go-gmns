package macro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

// Link representation of an edge in a macroscopic graph
type Link struct {
	toll         *float64
	facilityType string
	allowedUses  string

	rightOfWayWidthMeters *float64

	bikeFacility string
	pedFacility  string
	parking      string

	gradePercent *float64
	jurisdiction string

	directed bool

	parentLinkID *gmns.LinkID

	lanesInfo          LanesInfo
	name               string
	geom               orb.LineString
	geomEuclidean      orb.LineString
	geometryID         *gmns.GeometryID
	dirFlag            *gmns.DirectionFlag
	allowedAgentTypes  []types.AgentType
	lengthMeters       *float64
	freeSpeed          *float64
	maxSpeed           float64
	ID                 gmns.LinkID
	sourceNodeID       gmns.NodeID
	targetNodeID       gmns.NodeID
	osmWayID           osm.WayID
	sourceOsmNodeID    osm.NodeID
	targetOsmNodeID    osm.NodeID
	lanesNum           *int
	capacity           *float64
	linkClass          types.LinkClass
	linkType           types.LinkType
	linkConnectionType types.LinkConnectionType
	controlType        types.ControlType
	wasBidirectional   bool
}

// NewLinkFrom creates pointer to the new Link
func NewLinkFrom(id gmns.LinkID, sourceNodeID, targetNodeID gmns.NodeID, options ...func(*Link)) *Link {
	newLink := &Link{
		directed:           true,
		lanesInfo:          LanesInfo{},
		name:               "",
		geom:               orb.LineString{},
		geomEuclidean:      orb.LineString{},
		allowedAgentTypes:  []types.AgentType{},
		maxSpeed:           -1,
		ID:                 id,
		sourceNodeID:       sourceNodeID,
		targetNodeID:       targetNodeID,
		osmWayID:           osm.WayID(-1),
		sourceOsmNodeID:    osm.NodeID(-1),
		targetOsmNodeID:    osm.NodeID(-1),
		linkClass:          types.LINK_CLASS_UNDEFINED,
		linkType:           types.LINK_UNDEFINED,
		linkConnectionType: types.NOT_A_LINK,
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
// Returns at least 1 to avoid negative index calculations in lane connections.
func (link *Link) GetIncomingLanes() int {
	if len(link.lanesInfo.LanesList) == 0 {
		return 1
	}
	lanes := link.lanesInfo.LanesList[0]
	if lanes < 1 {
		return 1
	}
	return lanes
}

// GetOutcomingLanes returns number of outcoming lanes. The last value in lanes list is that number.
// Returns at least 1 to avoid negative index calculations in lane connections.
func (link *Link) GetOutcomingLanes() int {
	idx := len(link.lanesInfo.LanesList) - 1
	if idx < 0 {
		return 1
	}
	lanes := link.lanesInfo.LanesList[idx]
	if lanes < 1 {
		return 1
	}
	return lanes
}

// GetOutcomingLaneIndices returns slice of the lane indices (for lane changes)
func (link *Link) GetOutcomingLaneIndices() []int {
	lanesInfo := link.lanesInfo
	idx := len(lanesInfo.LanesChange) - 1
	if idx < 0 {
		return make([]int, 0)
	}
	lanesNum, hasLanes := link.LanesNum()
	if !hasLanes || lanesNum < 0 {
		return make([]int, 0)
	}
	return laneIndices(lanesNum, lanesInfo.LanesChange[idx][0], lanesInfo.LanesChange[idx][1])
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

// GeometryID returns the optional reference to the GMNS geometry table.
// The boolean distinguishes an absent reference from geometry ID zero.
func (link *Link) GeometryID() (gmns.GeometryID, bool) {
	if link.geometryID == nil {
		return 0, false
	}
	return *link.geometryID, true
}

// DirFlag returns the optional GMNS shape orientation: 1 follows the link's
// endpoints, -1 opposes them, and 0 means undirected or no geometry information.
// An absent flag is distinct from an explicitly supplied zero.
func (link *Link) DirFlag() (gmns.DirectionFlag, bool) {
	if link.dirFlag == nil {
		return gmns.DIR_FLAG_NONE, false
	}
	return *link.dirFlag, true
}

// AllowedAgentTypes returns set of allowed agent types. Warning: returning object is a slice.
func (link *Link) AllowedAgentTypes() []types.AgentType {
	return link.allowedAgentTypes
}

// LengthMeters returns the supplied length in meters and whether it is set.
func (link *Link) LengthMeters() (float64, bool) {
	if link.lengthMeters == nil {
		return 0, false
	}
	return *link.lengthMeters, true
}

// FreeSpeed returns the supplied free-flow speed in km/h and whether it is set.
func (link *Link) FreeSpeed() (float64, bool) {
	if link.freeSpeed == nil {
		return 0, false
	}
	return *link.freeSpeed, true
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

// LanesNum returns the supplied permanent lane count and whether it is set.
func (link *Link) LanesNum() (int, bool) {
	if link.lanesNum == nil {
		return 0, false
	}
	return *link.lanesNum, true
}

// Capacity returns the supplied saturation capacity in PCE/hour/lane and whether it is set.
func (link *Link) Capacity() (float64, bool) {
	if link.capacity == nil {
		return 0, false
	}
	return *link.capacity, true
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

// WithGeometryID sets a reference to a shared shape without changing Geom().
func WithGeometryID(geometryID gmns.GeometryID) func(*Link) {
	return func(link *Link) {
		value := geometryID
		link.geometryID = &value
	}
}

// WithDirFlag sets GMNS shape orientation without reversing Geom(). For a
// shared canonical shape opposite directed links normally use 1 and -1.
// Existing per-link geometries already follow travel order and use 1.
func WithDirFlag(dirFlag gmns.DirectionFlag) func(*Link) {
	return func(link *Link) {
		value := dirFlag
		link.dirFlag = &value
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
		value := lengthMeters
		link.lengthMeters = &value
	}
}

// WithFreeSpeed stores the supplied free-flow speed in km/h.
func WithFreeSpeed(freeSpeed float64) func(*Link) {
	return func(link *Link) {
		value := freeSpeed
		link.freeSpeed = &value
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

// WithSourceNodeID sets source node identifier. Should not be used since it necessary part of NewLinkFrom(...).
func WithSourceNodeID(nodeID gmns.NodeID) func(*Link) {
	return func(link *Link) {
		link.sourceNodeID = nodeID
	}
}

// WithTargetNodeID sets target node identifier. Should not be used since it necessary part of NewLinkFrom(...).
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
		value := lanesNum
		link.lanesNum = &value
	}
}

// WithCapacity stores the supplied saturation capacity in PCE/hour/lane.
func WithCapacity(capacity float64) func(*Link) {
	return func(link *Link) {
		value := capacity
		link.capacity = &value
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

// ParentLink returns the optional associated link ID within the same network.
// This reference is independent of source provenance across network levels.
func (link *Link) ParentLink() (gmns.LinkID, bool) {
	if link.parentLinkID == nil {
		return 0, false
	}
	return *link.parentLinkID, true
}

// WithParentLinkID sets an associated link reference within the same network.
// Zero and negative IDs are stored as supplied values, not as absence.
func WithParentLinkID(parentLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		value := parentLinkID
		link.parentLinkID = &value
	}
}

// Directed reports whether travel is restricted to source-to-target direction.
func (link *Link) Directed() bool {
	return link.directed
}

// WithDirected sets whether the link is directed. NewLinkFrom defaults to true.
// It does not change geometry, endpoints or source-road bidirectionality.
func WithDirected(directed bool) func(*Link) {
	return func(link *Link) {
		link.directed = directed
	}
}

// RightOfWayWidthMeters returns the supplied width of the entire right-of-way in meters.
// The width covers both directions and is independent of individual lane widths.
func (link *Link) RightOfWayWidthMeters() (float64, bool) {
	if link.rightOfWayWidthMeters == nil {
		return 0, false
	}
	return *link.rightOfWayWidthMeters, true
}

// WithRightOfWayWidthMeters stores the entire right-of-way width without deriving it from lanes.
func WithRightOfWayWidthMeters(rightOfWayWidthMeters float64) func(*Link) {
	return func(link *Link) {
		value := rightOfWayWidthMeters
		link.rightOfWayWidthMeters = &value
	}
}

// Grade returns the supplied grade in percent and whether it was supplied.
// Negative values describe downhill travel.
func (link *Link) Grade() (float64, bool) {
	if link.gradePercent == nil {
		return 0, false
	}
	return *link.gradePercent, true
}

// WithGrade stores a grade in percent without calculating it from geometry.
func WithGrade(gradePercent float64) func(*Link) {
	return func(link *Link) {
		value := gradePercent
		link.gradePercent = &value
	}
}

// Jurisdiction returns the supplied owner/operator, or an empty string when absent.
func (link *Link) Jurisdiction() string {
	return link.jurisdiction
}

// WithJurisdiction sets the owner/operator; an empty string clears the value.
func WithJurisdiction(jurisdiction string) func(*Link) {
	return func(link *Link) {
		link.jurisdiction = jurisdiction
	}
}

// BikeFacility returns the supplied bicycle infrastructure, or an empty string when absent.
func (link *Link) BikeFacility() string {
	return link.bikeFacility
}

// WithBikeFacility stores the bicycle infrastructure; an empty string clears the value.
// Suggested GMNS values are available in gmns/types; custom strings are supported.
func WithBikeFacility(bikeFacility string) func(*Link) {
	return func(link *Link) {
		link.bikeFacility = bikeFacility
	}
}

// PedFacility returns the supplied pedestrian infrastructure, or an empty string when absent.
func (link *Link) PedFacility() string {
	return link.pedFacility
}

// WithPedFacility stores the pedestrian infrastructure; an empty string clears the value.
// Suggested GMNS values are available in gmns/types; custom strings are supported.
func WithPedFacility(pedFacility string) func(*Link) {
	return func(link *Link) {
		link.pedFacility = pedFacility
	}
}

// Parking returns the supplied curbside parking type, or an empty string when absent.
func (link *Link) Parking() string {
	return link.parking
}

// WithParking stores the curbside parking type; an empty string clears the value.
// Suggested GMNS values are available in gmns/types; custom strings are supported.
func WithParking(parking string) func(*Link) {
	return func(link *Link) {
		link.parking = parking
	}
}

// Toll returns the supplied amount in the dataset's currency units and its presence.
func (link *Link) Toll() (float64, bool) {
	if link.toll == nil {
		return 0, false
	}
	return *link.toll, true
}

// WithToll stores an amount without currency conversion or automatic calculation.
func WithToll(toll float64) func(*Link) {
	return func(link *Link) {
		value := toll
		link.toll = &value
	}
}

// FacilityType returns the supplied facility classification, or empty if absent.
func (link *Link) FacilityType() string {
	return link.facilityType
}

// WithFacilityType stores or clears a classification independently of OSM link types.
func WithFacilityType(facilityType string) func(*Link) {
	return func(link *Link) {
		link.facilityType = facilityType
	}
}

// AllowedUses returns the supplied comma-separated use references, or empty if absent.
func (link *Link) AllowedUses() string {
	return link.allowedUses
}

// WithAllowedUses stores or clears use references without deriving agent permissions.
// Definition lookup and syntax validation belong to dataset validation.
func WithAllowedUses(allowedUses string) func(*Link) {
	return func(link *Link) {
		link.allowedUses = allowedUses
	}
}

// WithoutToll clears the supplied value, restoring absence.
func WithoutToll() func(*Link) {
	return func(link *Link) {
		link.toll = nil
	}
}

// WithoutLengthMeters clears the supplied value, restoring absence.
func WithoutLengthMeters() func(*Link) {
	return func(link *Link) {
		link.lengthMeters = nil
	}
}

// WithoutFreeSpeed clears the supplied value, restoring absence.
func WithoutFreeSpeed() func(*Link) {
	return func(link *Link) {
		link.freeSpeed = nil
	}
}

// WithoutCapacity clears the supplied value, restoring absence.
func WithoutCapacity() func(*Link) {
	return func(link *Link) {
		link.capacity = nil
	}
}

// WithoutLanesNum clears the supplied value, restoring absence.
func WithoutLanesNum() func(*Link) {
	return func(link *Link) {
		link.lanesNum = nil
	}
}

// WithoutGrade clears the supplied value, restoring absence.
func WithoutGrade() func(*Link) {
	return func(link *Link) {
		link.gradePercent = nil
	}
}

// WithoutRightOfWayWidthMeters clears the supplied value, restoring absence.
func WithoutRightOfWayWidthMeters() func(*Link) {
	return func(link *Link) {
		link.rightOfWayWidthMeters = nil
	}
}

// WithoutGeometryID clears the supplied value, restoring absence.
func WithoutGeometryID() func(*Link) {
	return func(link *Link) {
		link.geometryID = nil
	}
}

// WithoutDirFlag clears the supplied value, restoring absence.
func WithoutDirFlag() func(*Link) {
	return func(link *Link) {
		link.dirFlag = nil
	}
}

// WithoutParentLinkID clears the supplied value, restoring absence.
func WithoutParentLinkID() func(*Link) {
	return func(link *Link) {
		link.parentLinkID = nil
	}
}
