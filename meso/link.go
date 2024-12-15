package meso

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
)

type Link struct {
	ID gmns.LinkID

	geom          orb.LineString
	geomEuclidean orb.LineString
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
