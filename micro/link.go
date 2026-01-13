package micro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
)

// Link represents a microscopic network link (cell)
type Link struct {
	ID gmns.LinkID

	geom          orb.LineString
	geomEuclidean orb.LineString

	lengthMeters float64

	sourceNodeID gmns.NodeID
	targetNodeID gmns.NodeID

	// Parent references
	mesoLinkID  gmns.LinkID
	macroLinkID gmns.LinkID
	macroNodeID gmns.NodeID

	// Cell properties
	cellType types.CellType
	laneID   int // Lane number this cell belongs to

	// Movement properties
	isFirstMovementCell   bool
	movementCompositeType movement.MovementCompositeType
	additionalTravelCost  float64

	// Inherited from parent data
	mesoLinkType      types.LinkType
	controlType       types.ControlType
	freeSpeed         float64
	capacity          int
	allowedAgentTypes []types.AgentType
}

// NewLinkFrom creates pointer to the new microscopic Link
func NewLinkFrom(id gmns.LinkID, sourceNodeID, targetNodeID gmns.NodeID, options ...func(*Link)) *Link {
	newLink := &Link{
		ID:                    id,
		geom:                  orb.LineString{},
		geomEuclidean:         orb.LineString{},
		lengthMeters:          -1,
		sourceNodeID:          sourceNodeID,
		targetNodeID:          targetNodeID,
		mesoLinkID:            -1,
		macroLinkID:           -1,
		macroNodeID:           -1,
		cellType:              types.CELL_FORWARD,
		laneID:                0,
		isFirstMovementCell:   false,
		movementCompositeType: movement.MOVEMENT_UNDEFINED,
		additionalTravelCost:  0.0,
		mesoLinkType:          types.LINK_UNDEFINED,
		controlType:           types.CONTROL_TYPE_NOT_SIGNAL,
		freeSpeed:             0.0,
		capacity:              0,
		allowedAgentTypes:     []types.AgentType{},
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

// LengthMeters returns length of the underlying geometry. Outputs "-1" if not set.
func (link *Link) LengthMeters() float64 {
	return link.lengthMeters
}

// SourceNode returns identifier of source node
func (link *Link) SourceNode() gmns.NodeID {
	return link.sourceNodeID
}

// TargetNode returns identifier of target node
func (link *Link) TargetNode() gmns.NodeID {
	return link.targetNodeID
}

// MesoLink returns identifier of parent mesoscopic link. Outputs "-1" if not set.
func (link *Link) MesoLink() gmns.LinkID {
	return link.mesoLinkID
}

// MacroLink returns identifier of parent macroscopic link. Outputs "-1" if not set.
func (link *Link) MacroLink() gmns.LinkID {
	return link.macroLinkID
}

// MacroNode returns identifier of parent macroscopic node. Outputs "-1" if not set.
func (link *Link) MacroNode() gmns.NodeID {
	return link.macroNodeID
}

// CellType returns the cell type (forward or lane_change)
func (link *Link) CellType() types.CellType {
	return link.cellType
}

// LaneID returns the lane number this cell belongs to
func (link *Link) LaneID() int {
	return link.laneID
}

// IsFirstMovementCell returns true if this is the first cell in a movement
func (link *Link) IsFirstMovementCell() bool {
	return link.isFirstMovementCell
}

// MovementCompositeType returns the movement composite type
func (link *Link) MovementCompositeType() movement.MovementCompositeType {
	return link.movementCompositeType
}

// AdditionalTravelCost returns additional travel cost for this cell
func (link *Link) AdditionalTravelCost() float64 {
	return link.additionalTravelCost
}

// MesoLinkType returns the parent mesoscopic link type
func (link *Link) MesoLinkType() types.LinkType {
	return link.mesoLinkType
}

// ControlType returns control type
func (link *Link) ControlType() types.ControlType {
	return link.controlType
}

// FreeSpeed returns free flow speed for the link
func (link *Link) FreeSpeed() float64 {
	return link.freeSpeed
}

// Capacity returns max capacity of the link
func (link *Link) Capacity() int {
	return link.capacity
}

// AllowedAgentTypes returns set of allowed agent types
func (link *Link) AllowedAgentTypes() []types.AgentType {
	return link.allowedAgentTypes
}

// SetSourceNode updates the source node ID
func (link *Link) SetSourceNode(nodeID gmns.NodeID) {
	link.sourceNodeID = nodeID
}

// SetTargetNode updates the target node ID
func (link *Link) SetTargetNode(nodeID gmns.NodeID) {
	link.targetNodeID = nodeID
}

// WithLineGeom sets geometry [WGS84] for the link
func WithLineGeom(geom orb.LineString) func(*Link) {
	return func(link *Link) {
		link.geom = geom
	}
}

// WithLineGeomEuclidean sets geometry [Euclidean] for the link
func WithLineGeomEuclidean(geomEuclidean orb.LineString) func(*Link) {
	return func(link *Link) {
		link.geomEuclidean = geomEuclidean
	}
}

// WithLengthMeters sets length in meters
func WithLengthMeters(lengthMeters float64) func(*Link) {
	return func(link *Link) {
		link.lengthMeters = lengthMeters
	}
}

// WithMesoLinkID sets parent mesoscopic link identifier
func WithMesoLinkID(mesoLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.mesoLinkID = mesoLinkID
	}
}

// WithMacroLinkID sets parent macroscopic link identifier
func WithMacroLinkID(macroLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.macroLinkID = macroLinkID
	}
}

// WithMacroNodeID sets parent macroscopic node identifier
func WithMacroNodeID(macroNodeID gmns.NodeID) func(*Link) {
	return func(link *Link) {
		link.macroNodeID = macroNodeID
	}
}

// WithCellType sets the cell type
func WithCellType(cellType types.CellType) func(*Link) {
	return func(link *Link) {
		link.cellType = cellType
	}
}

// WithLaneID sets the lane number
func WithLaneID(laneID int) func(*Link) {
	return func(link *Link) {
		link.laneID = laneID
	}
}

// WithIsFirstMovementCell marks the link as first movement cell
func WithIsFirstMovementCell(isFirst bool) func(*Link) {
	return func(link *Link) {
		link.isFirstMovementCell = isFirst
	}
}

// WithMovementCompositeType sets the movement composite type
func WithMovementCompositeType(mct movement.MovementCompositeType) func(*Link) {
	return func(link *Link) {
		link.movementCompositeType = mct
	}
}

// WithAdditionalTravelCost sets additional travel cost
func WithAdditionalTravelCost(cost float64) func(*Link) {
	return func(link *Link) {
		link.additionalTravelCost = cost
	}
}

// WithMesoLinkType sets the parent mesoscopic link type
func WithMesoLinkType(linkType types.LinkType) func(*Link) {
	return func(link *Link) {
		link.mesoLinkType = linkType
	}
}

// WithControlType sets the control type
func WithControlType(controlType types.ControlType) func(*Link) {
	return func(link *Link) {
		link.controlType = controlType
	}
}

// WithFreeSpeed sets free flow speed
func WithFreeSpeed(freeSpeed float64) func(*Link) {
	return func(link *Link) {
		link.freeSpeed = freeSpeed
	}
}

// WithCapacity sets max capacity
func WithCapacity(capacity int) func(*Link) {
	return func(link *Link) {
		link.capacity = capacity
	}
}

// WithAllowedAgentTypes sets allowed agent types (copies the slice)
func WithAllowedAgentTypes(allowedAgentTypes []types.AgentType) func(*Link) {
	return func(link *Link) {
		link.allowedAgentTypes = make([]types.AgentType, len(allowedAgentTypes))
		copy(link.allowedAgentTypes, allowedAgentTypes)
	}
}
