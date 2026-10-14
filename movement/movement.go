package movement

import (
	"sync"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

type autoInc struct {
	sync.Mutex
	id gmns.MovementID
}

func (a *autoInc) ID() (id gmns.MovementID) {
	a.Lock()
	defer a.Unlock()

	id = a.id
	a.id++
	return
}

var (
	ai autoInc
)

func GenMovementID() gmns.MovementID {
	return ai.ID()
}

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
	incomeLaneStart, incomeLaneEnd           *int

	outcomeMacroLinkID                         gmns.LinkID
	startOutcomeLaneSeqID, endOutcomeLaneSeqID int
	outcomeLaneStart, outcomeLaneEnd           *int

	osmNodeID       osm.NodeID
	sourceOsmNodeID osm.NodeID
	targetOsmNodeID osm.NodeID

	mType              MovementType
	mTextID            MovementCompositeType
	controlType        types.ControlType
	lanesNum           int
	penaltySeconds     *float64
	capacityPCEPerHour *float64
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
		outcomeMacroLinkID:    outcomeMacroLinkID,
		startOutcomeLaneSeqID: -1,
		endOutcomeLaneSeqID:   -1,
		osmNodeID:             osm.NodeID(-1),
		sourceOsmNodeID:       osm.NodeID(-1),
		targetOsmNodeID:       osm.NodeID(-1),
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
func (mvmt *Movement) Name() string {
	return mvmt.name
}

// Geom returns corresponding geometry [WGS84]
func (mvmt *Movement) Geom() orb.LineString {
	return mvmt.geom
}

// GeomEuclidean returns corresponding geometry [Euclidean]
func (mvmt *Movement) GeomEuclidean() orb.LineString {
	return mvmt.geomEuclidean
}

// AllowedAgentTypes returns set of allowed agent types. Warning: returning object is a slice.
func (mvmt *Movement) AllowedAgentTypes() []types.AgentType {
	return mvmt.allowedAgentTypes
}

// MacroNode returns parent macro node. Should not output "-1" (could lead to the problems during processing mesoscopic graph).
func (mvmt *Movement) MacroNode() gmns.NodeID {
	return mvmt.macroNodeID
}

// IncomeMacroLink returns income macro link. Should not output "-1" (could lead to the problems during processing mesoscopic graph).
func (mvmt *Movement) IncomeMacroLink() gmns.LinkID {
	return mvmt.incomeMacroLinkID
}

// StartIncomeLaneSeqID returns first index of the income lane. Outputs "-1" if it was not set.
func (mvmt *Movement) StartIncomeLaneSeqID() int {
	return mvmt.startIncomeLaneSeqID
}

// EndIncomeLaneSeqID returns last index of the income lane. Outputs "-1" if it was not set.
func (mvmt *Movement) EndIncomeLaneSeqID() int {
	return mvmt.endIncomeLaneSeqID
}

// IncomeLaneStart returns the first lane number of the income link and whether it was supplied.
func (mvmt *Movement) IncomeLaneStart() (int, bool) {
	if mvmt.incomeLaneStart == nil {
		return 0, false
	}
	return *mvmt.incomeLaneStart, true
}

// IncomeLaneEnd returns the last lane number of the income link and whether it was supplied.
func (mvmt *Movement) IncomeLaneEnd() (int, bool) {
	if mvmt.incomeLaneEnd == nil {
		return 0, false
	}
	return *mvmt.incomeLaneEnd, true
}

// OutcomeMacroLink returns outcome macro link. Should not output "-1" (could lead to the problems during processing mesoscopic graph).
func (mvmt *Movement) OutcomeMacroLink() gmns.LinkID {
	return mvmt.outcomeMacroLinkID
}

// StartOutcomeLaneSeqID returns first index of the outcome lane. Outputs "-1" if it was not set.
func (mvmt *Movement) StartOutcomeLaneSeqID() int {
	return mvmt.startOutcomeLaneSeqID
}

// EndOutcomeLaneSeqID returns last index of the outcome lane. Outputs "-1" if it was not set.
func (mvmt *Movement) EndOutcomeLaneSeqID() int {
	return mvmt.endOutcomeLaneSeqID
}

// OutcomeLaneStart returns the first lane number of the outcome link and whether it was supplied.
func (mvmt *Movement) OutcomeLaneStart() (int, bool) {
	if mvmt.outcomeLaneStart == nil {
		return 0, false
	}
	return *mvmt.outcomeLaneStart, true
}

// OutcomeLaneEnd returns the last lane number of the outcome link and whether it was supplied.
func (mvmt *Movement) OutcomeLaneEnd() (int, bool) {
	if mvmt.outcomeLaneEnd == nil {
		return 0, false
	}
	return *mvmt.outcomeLaneEnd, true
}

// OSMNode returns OSM node identifier. Outputs "-1" if it was not set.
func (mvmt *Movement) OSMNode() osm.NodeID {
	return mvmt.osmNodeID
}

// OSMNodeSource returns source OSM node identifier. Outputs "-1" if it was not set.
func (mvmt *Movement) OSMNodeSource() osm.NodeID {
	return mvmt.sourceOsmNodeID
}

// OSMNodeTarget returns target OSM node identifier. Outputs "-1" if it was not set.
func (mvmt *Movement) OSMNodeTarget() osm.NodeID {
	return mvmt.targetOsmNodeID
}

// Type returns movement type
func (mvmt *Movement) Type() MovementType {
	return mvmt.mType
}

// MvmtTextID returns composite type for the given movement
func (mvmt *Movement) MvmtTextID() MovementCompositeType {
	return mvmt.mTextID
}

// ControlType returns control type
func (mvmt *Movement) ControlType() types.ControlType {
	return mvmt.controlType
}

// LanesNum returns number of lanes. Outputs "-1" if it was not set.
func (mvmt *Movement) LanesNum() int {
	return mvmt.lanesNum
}

// Penalty returns the turn penalty in seconds and whether it was supplied.
func (mvmt *Movement) Penalty() (float64, bool) {
	if mvmt.penaltySeconds == nil {
		return 0, false
	}
	return *mvmt.penaltySeconds, true
}

// Capacity returns saturation capacity in passenger car equivalents per hour and whether it was supplied.
func (mvmt *Movement) Capacity() (float64, bool) {
	if mvmt.capacityPCEPerHour == nil {
		return 0, false
	}
	return *mvmt.capacityPCEPerHour, true
}

// WithCapacity sets saturation capacity for the whole movement in passenger car equivalents per hour.
func WithCapacity(capacityPCEPerHour float64) func(*Movement) {
	return func(mvmt *Movement) {
		value := capacityPCEPerHour
		mvmt.capacityPCEPerHour = &value
	}
}

// WithPenalty sets the turn penalty in seconds, including an explicit zero.
func WithPenalty(seconds float64) func(*Movement) {
	return func(mvmt *Movement) {
		value := seconds
		mvmt.penaltySeconds = &value
	}
}

// WithLinkName sets alias for the movement
func WithName(name string) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.name = name
	}
}

// WithGeom sets geometry [WGS84] for the link. Warning: it does not copy the given slice.
func WithGeom(geom orb.LineString) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.geom = geom
	}
}

// WithLineGeomEuclidean sets geometry [Euclidean] for the link. Warning: it does not copy the given slice.
func WithGeomEuclidean(geomEuclidean orb.LineString) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.geomEuclidean = geomEuclidean
	}
}

// WithAllowedAgentTypes sets allowed agent types for the link. Warning: it does copy argument
func WithAllowedAgentTypes(allowedAgentTypes []types.AgentType) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.allowedAgentTypes = make([]types.AgentType, len(allowedAgentTypes))
		copy(mvmt.allowedAgentTypes, allowedAgentTypes)
	}
}

// WithMacroNodeID sets parent macro node identifier. Should not be used since it necessary part of NewMovement(...).
func WithMacroNodeID(nodeID gmns.NodeID) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.macroNodeID = nodeID
	}
}

// WithIncomeMacroLinkID sets income macro link indetifier. Should not be used since it necessary part of NewMovement(...).
func WithIncomeMacroLinkID(linkID gmns.LinkID) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.incomeMacroLinkID = linkID
	}
}

// WithIncomeLaneSequence sets start and end index for the lane's segment of income macro link
// Notice: those are just indexes of incomeLaneStart and incomeLaneEnd
func WithIncomeLaneSequence(startIdx, endIdx int) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.startIncomeLaneSeqID = startIdx
		mvmt.endIncomeLaneSeqID = endIdx
	}
}

// WithIncomeLane sets both lane numbers of the income macro link, including negative numbers.
func WithIncomeLane(start, end int) func(*Movement) {
	return func(mvmt *Movement) {
		startValue, endValue := start, end
		mvmt.incomeLaneStart = &startValue
		mvmt.incomeLaneEnd = &endValue
	}
}

// WithIncomeLaneStart sets the first lane number without changing the last lane number.
func WithIncomeLaneStart(lane int) func(*Movement) {
	return func(mvmt *Movement) {
		value := lane
		mvmt.incomeLaneStart = &value
	}
}

// WithIncomeLaneEnd sets the last lane number without changing the first lane number.
func WithIncomeLaneEnd(lane int) func(*Movement) {
	return func(mvmt *Movement) {
		value := lane
		mvmt.incomeLaneEnd = &value
	}
}

// WithOutcomeMacroLinkID sets outcome macro link indetifier. Should not be used since it necessary part of NewMovement(...).
func WithOutcomeMacroLinkID(linkID gmns.LinkID) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.outcomeMacroLinkID = linkID
	}
}

// WithOutcomeLaneSequence sets start and end index for the lane's segment of outcome macro link
// Notice: those are just indexes of outcomeLaneStart and outcomeLaneEnd
func WithOutcomeLaneSequence(startIdx int, endIdx int) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.startOutcomeLaneSeqID = startIdx
		mvmt.endOutcomeLaneSeqID = endIdx
	}
}

// WithOutcomeLane sets both lane numbers of the outcome macro link, including negative numbers.
func WithOutcomeLane(start, end int) func(*Movement) {
	return func(mvmt *Movement) {
		startValue, endValue := start, end
		mvmt.outcomeLaneStart = &startValue
		mvmt.outcomeLaneEnd = &endValue
	}
}

// WithOutcomeLaneStart sets the first lane number without changing the last lane number.
func WithOutcomeLaneStart(lane int) func(*Movement) {
	return func(mvmt *Movement) {
		value := lane
		mvmt.outcomeLaneStart = &value
	}
}

// WithOutcomeLaneEnd sets the last lane number without changing the first lane number.
func WithOutcomeLaneEnd(lane int) func(*Movement) {
	return func(mvmt *Movement) {
		value := lane
		mvmt.outcomeLaneEnd = &value
	}
}

// WithOSMNodeID sets OSM node identifier
func WithOSMNodeID(nodeID osm.NodeID) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.osmNodeID = nodeID
	}
}

// WithSourceOSMNodeID sets source OSM node identifier
func WithSourceOSMNodeID(nodeID osm.NodeID) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.sourceOsmNodeID = nodeID
	}
}

// WithTargetOSMNodeID sets target OSM node identifier
func WithTargetOSMNodeID(nodeID osm.NodeID) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.targetOsmNodeID = nodeID
	}
}

// WithType sets type for the movement. Should not be used since it necessary part of NewMovement(...).
func WithType(mType MovementType) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.mType = mType
	}
}

// WithMvmtTextID sets composite type for the movement. Should not be used since it necessary part of NewMovement(...).
func WithMvmtTextID(mvmtTxtID MovementCompositeType) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.mTextID = mvmtTxtID
	}
}

// WithControlType sets control type for the movement
func WithControlType(controlType types.ControlType) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.controlType = controlType
	}
}

// WithLanesNum sets number of lanes for the movement
func WithLanesNum(lanesNum int) func(*Movement) {
	return func(mvmt *Movement) {
		mvmt.lanesNum = lanesNum
	}
}
