package movement

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

// MovementsStorage is storage for the movements
type MovementsStorage map[gmns.MovementID]*Movement

// NewMovementsStorage returns new storage for movements. Basically is just shortcut to make(...)
func NewMovementsStorage() MovementsStorage {
	return make(MovementsStorage)
}

// Movement represents maneuver between some road parts
type Movement struct {
	name              string
	geom              orb.LineString
	geomEuclidean     orb.LineString
	allowedAgentTypes []types.AgentType

	ID                                       gmns.MovementID
	macroNodeID                              gmns.NodeID
	incomeMacroLinkID                        gmns.LinkID
	startIncomeLaneSeqID, endIncomeLaneSeqID int
	incomeLaneStart, incomeLaneEnd           int

	outcomeMacroLinkID                         gmns.LinkID
	startOutcomeLaneSeqID, endOutcomeLaneSeqID int
	outcomeLaneStart, outcomeLaneEnd           int

	osmNodeID     osm.NodeID
	fromOsmNodeID osm.NodeID
	toOsmNodeID   osm.NodeID

	mType       MovementType
	mTextID     MovementCompositeType
	controlType types.ControlType
	lanesNum    int
}

// Name returns movement alias.
func (movement *Movement) Name() string {
	return movement.name
}

// Geom returns corresponding geometry [WGS84]
func (movement *Movement) Geom() orb.LineString {
	return movement.geom
}

// GeomEuclidean returns corresponding geometry [Euclidean]
func (movement *Movement) GeomEuclidean() orb.LineString {
	return movement.geomEuclidean
}

// AllowedAgentTypes returns set of allowed agent types. Warning: returning object is a slice.
func (movement *Movement) AllowedAgentTypes() []types.AgentType {
	return movement.allowedAgentTypes
}

// MacroNode returns parent macro node. Outputs "-1" if it was not set.
func (movement *Movement) MacroNode() gmns.NodeID {
	return movement.macroNodeID
}

// IncomeMacroLink returns incoming macro link. Outputs "-1" if it was not set.
func (movement *Movement) IncomeMacroLink() gmns.LinkID {
	return movement.incomeMacroLinkID
}

// StartIncomeLaneSeqID returns first index of the income lane. Outputs "-1" if it was not set.
func (movement *Movement) StartIncomeLaneSeqID() int {
	return movement.startIncomeLaneSeqID
}

// EndIncomeLaneSeqID returns last index of the income lane. Outputs "-1" if it was not set.
func (movement *Movement) EndIncomeLaneSeqID() int {
	return movement.endIncomeLaneSeqID
}

// IncomeLaneStart returns first lane of the income link. Outputs "-1" if it was not set.
func (movement *Movement) IncomeLaneStart() int {
	return movement.incomeLaneStart
}

// IncomeLaneEnd returns last lane of the income link. Outputs "-1" if it was not set.
func (movement *Movement) IncomeLaneEnd() int {
	return movement.incomeLaneEnd
}

// OutcomeMacroLink returns outcoming macro link. Outputs "-1" if it was not set.
func (movement *Movement) OutcomeMacroLink() gmns.LinkID {
	return movement.outcomeMacroLinkID
}

// StartOutcomeLaneSeqID returns first index of the outcome lane. Outputs "-1" if it was not set.
func (movement *Movement) StartOutcomeLaneSeqID() int {
	return movement.startOutcomeLaneSeqID
}

// EndOutcomeLaneSeqID returns last index of the outcome lane. Outputs "-1" if it was not set.
func (movement *Movement) EndOutcomeLaneSeqID() int {
	return movement.endOutcomeLaneSeqID
}

// OutcomeLaneStart returns first lane of the outcome link. Outputs "-1" if it was not set.
func (movement *Movement) OutcomeLaneStart() int {
	return movement.outcomeLaneStart
}

// OutcomeLaneEnd returns last lane of the outcome link. Outputs "-1" if it was not set.
func (movement *Movement) OutcomeLaneEnd() int {
	return movement.outcomeLaneEnd
}

// OSMNode returns OSM node identifier. Outputs "-1" if it was not set.
func (movement *Movement) OSMNode() osm.NodeID {
	return movement.osmNodeID
}

// OSMNodeFrom returns source OSM node identifier. Outputs "-1" if it was not set.
func (movement *Movement) OSMNodeFrom() osm.NodeID {
	return movement.fromOsmNodeID
}

// OSMNodeTo returns target OSM node identifier. Outputs "-1" if it was not set.
func (movement *Movement) OSMNodeTo() osm.NodeID {
	return movement.toOsmNodeID
}

// Type returns movement type
func (movement *Movement) Type() MovementType {
	return movement.mType
}

// MvmtTextID returns composite type for the given movement
func (movement *Movement) MvmtTextID() MovementCompositeType {
	return movement.mTextID
}

// ControlType returns control type
func (movement *Movement) ControlType() types.ControlType {
	return movement.controlType
}

// LanesNum returns number of lanes. Outputs "-1" if it was not set.
func (movement *Movement) LanesNum() int {
	return movement.lanesNum
}
