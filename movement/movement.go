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

// NewMovement creates pointer to the new Movement
func NewMovement(id gmns.MovementID, macroNodeID gmns.NodeID, incomeMacroLinkID, outcomeMacroLinkID gmns.LinkID, mvmtTxtID MovementCompositeType, mvmtType MovementType, options ...func(*Movement)) *Movement {
	newMovement := &Movement{
		name:                  "",
		geom:                  orb.LineString{},
		geomEuclidean:         orb.LineString{},
		allowedAgentTypes:     []types.AgentType{},
		ID:                    id,
		macroNodeID:           macroNodeID,
		incomeMacroLinkID:     incomeMacroLinkID,
		startIncomeLaneSeqID:  -1,
		endIncomeLaneSeqID:    -1,
		incomeLaneStart:       -1,
		incomeLaneEnd:         -1,
		outcomeMacroLinkID:    outcomeMacroLinkID,
		startOutcomeLaneSeqID: -1,
		endOutcomeLaneSeqID:   -1,
		outcomeLaneStart:      -1,
		outcomeLaneEnd:        -1,
		osmNodeID:             osm.NodeID(-1),
		fromOsmNodeID:         osm.NodeID(-1),
		toOsmNodeID:           osm.NodeID(-1),
		mType:                 mvmtType,
		mTextID:               mvmtTxtID,
		controlType:           types.CONTROL_TYPE_NOT_SIGNAL,
		lanesNum:              -1,
	}
	for _, option := range options {
		option(newMovement)
	}
	return newMovement
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

// MacroNode returns parent macro node. Should not output "-1" (could lead to the problems during processing mesoscopic graph).
func (movement *Movement) MacroNode() gmns.NodeID {
	return movement.macroNodeID
}

// IncomeMacroLink returns income macro link. Should not output "-1" (could lead to the problems during processing mesoscopic graph).
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

// OutcomeMacroLink returns outcome macro link. Should not output "-1" (could lead to the problems during processing mesoscopic graph).
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

// WithLinkName sets alias for the movement
func WithName(name string) func(*Movement) {
	return func(m *Movement) {
		m.name = name
	}
}

// WithGeom sets geometry [WGS84] for the link. Warning: it does not copy the given slice.
func WithGeom(geom orb.LineString) func(*Movement) {
	return func(movement *Movement) {
		movement.geom = geom
	}
}

// WithLineGeomEuclidean sets geometry [Euclidean] for the link. Warning: it does not copy the given slice.
func WithGeomEuclidean(geomEuclidean orb.LineString) func(*Movement) {
	return func(movement *Movement) {
		movement.geomEuclidean = geomEuclidean
	}
}

// WithAllowedAgentTypes sets allowed agent types for the link. Warning: it does copy argument
func WithAllowedAgentTypes(allowedAgentTypes []types.AgentType) func(*Movement) {
	return func(movement *Movement) {
		movement.allowedAgentTypes = make([]types.AgentType, len(allowedAgentTypes))
		copy(movement.allowedAgentTypes, allowedAgentTypes)
	}
}

// WithMacroNodeID sets parent macro node identifier. Should not be used since it necessary part of NewMovement(...).
func WithMacroNodeID(nodeID gmns.NodeID) func(*Movement) {
	return func(movement *Movement) {
		movement.macroNodeID = nodeID
	}
}

// WithIncomeMacroLinkID sets income macro link indetifier. Should not be used since it necessary part of NewMovement(...).
func WithIncomeMacroLinkID(linkID gmns.LinkID) func(*Movement) {
	return func(movement *Movement) {
		movement.incomeMacroLinkID = linkID
	}
}

// WithStartIncomeLaneSeqID sets first index of the income lane
func WithStartIncomeLaneSeqID(seqID int) func(*Movement) {
	return func(movement *Movement) {
		movement.startIncomeLaneSeqID = seqID
	}
}

// WithEndIncomeLaneSeqID sets last index of the income lane
func WithEndIncomeLaneSeqID(seqID int) func(*Movement) {
	return func(movement *Movement) {
		movement.endIncomeLaneSeqID = seqID
	}
}

// WithIncomeLaneStart sets first lane of the income link
func WithIncomeLaneStart(start int) func(*Movement) {
	return func(movement *Movement) {
		movement.incomeLaneStart = start
	}
}

// WithIncomeLaneEnd sets last lane of the income link
func WithIncomeLaneEnd(end int) func(*Movement) {
	return func(movement *Movement) {
		movement.incomeLaneEnd = end
	}
}

// WithOutcomeMacroLinkID sets outcome macro link indetifier. Should not be used since it necessary part of NewMovement(...).
func WithOutcomeMacroLinkID(linkID gmns.LinkID) func(*Movement) {
	return func(movement *Movement) {
		movement.outcomeMacroLinkID = linkID
	}
}

// WithStartOutcomeLaneSeqID sets first index of the outcome lane
func WithStartOutcomeLaneSeqID(seqID int) func(*Movement) {
	return func(movement *Movement) {
		movement.startOutcomeLaneSeqID = seqID
	}
}

// WithEndOutcomeLaneSeqID sets last index of the outcome lane
func WithEndOutcomeLaneSeqID(seqID int) func(*Movement) {
	return func(movement *Movement) {
		movement.endOutcomeLaneSeqID = seqID
	}
}

// WithOutcomeLaneStart sets first lane of the outcome link
func WithOutcomeLaneStart(start int) func(*Movement) {
	return func(movement *Movement) {
		movement.outcomeLaneStart = start
	}
}

// WithOutcomeLaneEnd sets last lane of the outcome link
func WithOutcomeLaneEnd(end int) func(*Movement) {
	return func(movement *Movement) {
		movement.outcomeLaneEnd = end
	}
}

// WithOSMNodeID sets OSM node identifier
func WithOSMNodeID(nodeID osm.NodeID) func(*Movement) {
	return func(movement *Movement) {
		movement.osmNodeID = nodeID
	}
}

// WithFromOSMNodeID sets source OSM node identifier
func WithFromOSMNodeID(nodeID osm.NodeID) func(*Movement) {
	return func(movement *Movement) {
		movement.fromOsmNodeID = nodeID
	}
}

// WithToOSMNodeID sets target OSM node identifier
func WithToOSMNodeID(nodeID osm.NodeID) func(*Movement) {
	return func(movement *Movement) {
		movement.toOsmNodeID = nodeID
	}
}

// WithType sets type for the movement. Should not be used since it necessary part of NewMovement(...).
func WithType(mType MovementType) func(*Movement) {
	return func(movement *Movement) {
		movement.mType = mType
	}
}

// WithMvmtTextID sets composite type for the movement. Should not be used since it necessary part of NewMovement(...).
func WithMvmtTextID(mvmtTxtID MovementCompositeType) func(*Movement) {
	return func(movement *Movement) {
		movement.mTextID = mvmtTxtID
	}
}

// WithControlType sets control type for the movement
func WithControlType(controlType types.ControlType) func(*Movement) {
	return func(movement *Movement) {
		movement.controlType = controlType
	}
}

// WithLanesNum sets number of lanes for the movement
func WithLanesNum(lanesNum int) func(*Movement) {
	return func(movement *Movement) {
		movement.lanesNum = lanesNum
	}
}
