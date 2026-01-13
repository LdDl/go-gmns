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
