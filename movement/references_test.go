package movement_test

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

func TestMovementReferencesBelongToOwningNetwork(t *testing.T) {
	first := gmns.NodeIDFromString("001")
	junction := gmns.NodeIDFromString("1")
	last := gmns.NodeIDFromString("end")
	income := gmns.LinkIDFromString("in")
	outcome := gmns.LinkIDFromString("out")
	identifier := gmns.MovementIDFromString("001")
	net := macro.NewNet()
	for _, nodeID := range []gmns.NodeID{first, junction, last} {
		net.Nodes[nodeID] = macro.NewNodeFrom(nodeID)
	}
	net.Links[income] = macro.NewLinkFrom(income, first, junction)
	net.Links[outcome] = macro.NewLinkFrom(outcome, junction, last)
	mvmt := movement.NewMovement(identifier, junction, income, outcome, "", "")
	storage := movement.MovementsStorage{identifier: mvmt}
	assert.NoError(t, net.ValidateMovements(storage))
	assert.NoError(t, storage.ValidateIDType(gmns.IDTypeString))
	assert.Error(t, storage.ValidateIDType(gmns.IDTypeInteger))
	movement.WithNodeID(first)(mvmt)
	assert.Error(t, net.ValidateMovements(storage), "existing but wrong junction must not resolve by membership alone")
	movement.WithNodeID(junction)(mvmt)
	movement.WithIncomeLinkID(gmns.LinkIDFromString("missing"))(mvmt)
	assert.ErrorIs(t, net.ValidateMovements(storage), macro.ErrLinkNotFound)
	movement.WithIncomeLinkID(income)(mvmt)
	mvmt.ID = gmns.MovementIDFromString("1")
	assert.Error(t, net.ValidateMovements(storage))
	storage[identifier] = nil
	assert.ErrorIs(t, net.ValidateMovements(storage), movement.ErrMvmtNotFound)
	assert.Error(t, (*macro.Net)(nil).ValidateMovements(storage))
}
