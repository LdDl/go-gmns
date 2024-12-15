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
