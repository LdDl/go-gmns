package movement

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

type MovementsStorage map[gmns.MovementID]*Movement

func NewMovementsStorage() map[gmns.MovementID]*Movement {
	return make(map[gmns.MovementID]*Movement)
}

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
