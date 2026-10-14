package meso

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
)

type Link struct {
	rightOfWayWidthMeters *float64

	bikeFacility string
	pedFacility  string
	parking      string

	gradePercent *float64
	jurisdiction string

	directed bool

	parentLinkID *gmns.LinkID

	name string

	ID gmns.LinkID

	geom          orb.LineString
	geomEuclidean orb.LineString
	geometryID    *gmns.GeometryID
	dirFlag       *gmns.DirectionFlag
	lanesNum      int
	lanesChange   [2]int

	lengthMeters float64

	sourceNodeID gmns.NodeID
	targetNodeID gmns.NodeID

	macroNodeID gmns.NodeID
	macroLinkID gmns.LinkID

	segmentIdx int

	isConnection bool

	/* Movement information */
	movementID                    gmns.MovementID
	movementCompositeType         movement.MovementCompositeType // Inherited from movement
	movementMesoLinkIncome        gmns.LinkID
	movementMesoLinkOutcome       gmns.LinkID
	movementIncomeLaneStartSeqID  int
	movementOutcomeLaneStartSeqID int

	/* Inherited from paret data parameters */
	controlType       types.ControlType // Inherited from macroscopic node
	linkType          types.LinkType    // Inherited either from macroscopic link or from first incoming incident edge in macroscopic node
	freeSpeed         float64           // Inherited either from macroscopic link or from first incoming incident edge in macroscopic node
	capacity          int               // Inherited either from macroscopic link or from first incoming incident edge in macroscopic node
	allowedAgentTypes []types.AgentType // Inherited either from macroscopic link or from first incoming incident edge in macroscopic node
}

func NewLinkFrom(id gmns.LinkID, sourceNodeID, targetNodeID gmns.NodeID, options ...func(*Link)) *Link {
	newLink := &Link{
		directed:      true,
		ID:            id,
		geom:          orb.LineString{},
		geomEuclidean: orb.LineString{},
		lanesNum:      -1,
		lanesChange:   [2]int{0, 0},
		lengthMeters:  -1,
		sourceNodeID:  sourceNodeID,
		targetNodeID:  targetNodeID,
		segmentIdx:    0,
		isConnection:  false,
		// Default movement
		movementID:                    gmns.MovementID(-1),
		movementCompositeType:         movement.MOVEMENT_UNDEFINED,
		movementMesoLinkIncome:        gmns.LinkID(-1),
		movementMesoLinkOutcome:       gmns.LinkID(-1),
		movementIncomeLaneStartSeqID:  -1,
		movementOutcomeLaneStartSeqID: -1,
		// Default parent data
		controlType:       types.CONTROL_TYPE_NOT_SIGNAL,
		linkType:          types.LINK_UNDEFINED,
		freeSpeed:         0.0,
		capacity:          0,
		allowedAgentTypes: []types.AgentType{},
	}
	for _, option := range options {
		option(newLink)
	}
	return newLink
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

// LanesNum returns number of lanes. Outputs "-1" if it was not set.
func (link *Link) LanesNum() int {
	return link.lanesNum
}

func (link *Link) LanesChange() [2]int {
	return link.lanesChange
}

// LengthMeters returns length of the underlying geometry [WGS84]. Outputs "-1" if it was not set.
func (link *Link) LengthMeters() float64 {
	return link.lengthMeters
}

// SourceNode returns identifier of source node. Outputs "-1" if it was not set.
func (link *Link) SourceNode() gmns.NodeID {
	return link.sourceNodeID
}

// TargetNode returns identifier of target node. Outputs "-1" if it was not set.
func (link *Link) TargetNode() gmns.NodeID {
	return link.targetNodeID
}

// MacroNode returns identifier of parent macroscopic node. Outputs "-1" if it was not set.
func (link *Link) MacroNode() gmns.NodeID {
	return link.macroNodeID
}

// MacroLink returns identifier of parent macroscopic link. Outputs "-1" if it was not set.
func (link *Link) MacroLink() gmns.LinkID {
	return link.macroLinkID
}

// SegmentIdx returns segment index in parent macroscopic link.
func (link *Link) SegmentIdx() int {
	return link.segmentIdx
}

// IsConnection returns if the given mesoscopic link is connecting two other mesoscopic links
func (link *Link) IsConnection() bool {
	return link.isConnection
}

// Movement returns parent movement identifier. Outputs "-1" if it was not set.
func (link *Link) Movement() gmns.MovementID {
	return link.movementID
}

// MvmtTextID returns composite type for the parent movement
func (link *Link) MvmtTextID() movement.MovementCompositeType {
	return link.movementCompositeType
}

// MovementMesoLinkIncome returns incoming mesoscopic link identifier in case if the given link is connection two other mesoscopic links. Outputs "-1" if it was not set.
func (link *Link) MovementMesoLinkIncome() gmns.LinkID {
	return link.movementMesoLinkIncome
}

// MovementMesoLinkOutcome returns incoming mesoscopic link identifier in case if the given link is connection two other mesoscopic links. Outputs "-1" if it was not set.
func (link *Link) MovementMesoLinkOutcome() gmns.LinkID {
	return link.movementMesoLinkOutcome
}

// MovementIncomeLaneStartSeqID returns lane identifier of incoming mesoscopic link. Outputs "-1" if it was not set.
func (link *Link) MovementIncomeLaneStartSeqID() int {
	return link.movementIncomeLaneStartSeqID
}

// MovementOutcomeLaneStartSeqID returns lane identifier of outcoming mesoscopic link. Outputs "-1" if it was not set.
func (link *Link) MovementOutcomeLaneStartSeqID() int {
	return link.movementOutcomeLaneStartSeqID
}

// ControlType returns control type
func (link *Link) ControlType() types.ControlType {
	return link.controlType
}

// LinkType returns link type
func (link *Link) LinkType() types.LinkType {
	return link.linkType
}

// FreeSpeed returns free flow speed for the link. Outputs "-1" if it was not set.
func (link *Link) FreeSpeed() float64 {
	return link.freeSpeed
}

// Capacity returns max capacity of the link. Outputs "-1" if it was not set.
func (link *Link) Capacity() int {
	return link.capacity
}

// AllowedAgentTypes returns set of allowed agent types. Warning: returning object is a slice.
func (link *Link) AllowedAgentTypes() []types.AgentType {
	return link.allowedAgentTypes
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

// WithLanesNum sets number of lanes for the link
func WithLanesNum(lanesNum int) func(*Link) {
	return func(link *Link) {
		link.lanesNum = lanesNum
	}
}

// WithLanesChange sets the lanes change array for the Link.
func WithLanesChange(lanesChange [2]int) func(*Link) {
	return func(link *Link) {
		link.lanesChange = lanesChange
	}
}

// WithLengthMeters sets underlying geometry [WGS84] length in meters. Warning: it should be called explicitly after setting geometry [WGS84]
func WithLengthMeters(lengthMeters float64) func(*Link) {
	return func(link *Link) {
		link.lengthMeters = lengthMeters
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

// WithLineMacroNodeID sets parent macro node identifier
func WithLineMacroNodeID(macroNodeID gmns.NodeID) func(*Link) {
	return func(link *Link) {
		link.macroNodeID = macroNodeID
	}
}

// WithLineMacroLinkID sets parent macro link identifier
func WithLineMacroLinkID(macroLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.macroLinkID = macroLinkID
	}
}

// WithSegmentIdx sets the segment index (inside of parent macroscopic link) for the given link.
func WithSegmentIdx(segmentIdx int) func(*Link) {
	return func(link *Link) {
		link.segmentIdx = segmentIdx
	}
}

// WithIsConnection sets whether the given link is a connection between two other mesoscopic links.
func WithIsConnection(isConnection bool) func(*Link) {
	return func(link *Link) {
		link.isConnection = isConnection
	}
}

// WithMovementID sets the movement identifier for the link.
func WithMovementID(movementID gmns.MovementID) func(*Link) {
	return func(link *Link) {
		link.movementID = movementID
	}
}

// WithMovementCompositeType sets the movement composite type for the link.
func WithMovementCompositeType(movementCompositeType movement.MovementCompositeType) func(*Link) {
	return func(link *Link) {
		link.movementCompositeType = movementCompositeType
	}
}

// WithMovementMesoLinkIncome sets the incoming mesoscopic link ID for the link.
func WithMovementMesoLinkIncome(movementMesoLinkIncome gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.movementMesoLinkIncome = movementMesoLinkIncome
	}
}

// WithMovementMesoLinkOutcome sets the outgoing mesoscopic link ID for the link.
func WithMovementMesoLinkOutcome(movementMesoLinkOutcome gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.movementMesoLinkOutcome = movementMesoLinkOutcome
	}
}

// WithMovementIncomeLaneStartSeqID sets the income lane start sequence ID for the link.
func WithMovementIncomeLaneStartSeqID(movementIncomeLaneStartSeqID int) func(*Link) {
	return func(link *Link) {
		link.movementIncomeLaneStartSeqID = movementIncomeLaneStartSeqID
	}
}

// WithMovementOutcomeLaneStartSeqID sets the outcome lane start sequence ID for the link.
func WithMovementOutcomeLaneStartSeqID(movementOutcomeLaneStartSeqID int) func(*Link) {
	return func(link *Link) {
		link.movementOutcomeLaneStartSeqID = movementOutcomeLaneStartSeqID
	}
}

// WithControlType sets the control type for the link.
func WithControlType(controlType types.ControlType) func(*Link) {
	return func(link *Link) {
		link.controlType = controlType
	}
}

// WithLinkType sets the link typk
func WithLinkType(linkType types.LinkType) func(*Link) {
	return func(link *Link) {
		link.linkType = linkType
	}
}

// WithFreeSpeed sets free flow speed for the link
func WithFreeSpeed(freeSpeed float64) func(*Link) {
	return func(link *Link) {
		link.freeSpeed = freeSpeed
	}
}

// WithCapacity sets max capacity for the link
func WithCapacity(capacity int) func(*Link) {
	return func(link *Link) {
		link.capacity = capacity
	}
}

// WithAllowedAgentTypes sets allowed agent types for the link. Warning: it does copy argument
func WithAllowedAgentTypes(allowedAgentTypes []types.AgentType) func(*Link) {
	return func(link *Link) {
		link.allowedAgentTypes = make([]types.AgentType, len(allowedAgentTypes))
		copy(link.allowedAgentTypes, allowedAgentTypes)
	}
}

// Name returns the supplied link name, or an empty string when not supplied.
func (link *Link) Name() string {
	return link.name
}

// WithLinkName sets the link name; an empty string clears the value.
func WithLinkName(name string) func(*Link) {
	return func(link *Link) {
		link.name = name
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
