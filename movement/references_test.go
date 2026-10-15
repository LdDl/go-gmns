package movement

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
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
	mvmt := NewMovement(identifier, junction, income, outcome, "", "")
	storage := MovementsStorage{identifier: mvmt}
	assert.NoError(t, storage.ValidateReferences(net))
	assert.NoError(t, storage.ValidateIDType(gmns.IDTypeString))
	assert.Error(t, storage.ValidateIDType(gmns.IDTypeInteger))
	WithMacroNodeID(first)(mvmt)
	assert.Error(t, storage.ValidateReferences(net), "existing but wrong junction must not resolve by membership alone")
	WithMacroNodeID(junction)(mvmt)
	WithIncomeMacroLinkID(gmns.LinkIDFromString("missing"))(mvmt)
	assert.ErrorIs(t, storage.ValidateReferences(net), macro.ErrLinkNotFound)
	WithIncomeMacroLinkID(income)(mvmt)
	mvmt.ID = gmns.MovementIDFromString("1")
	assert.Error(t, storage.ValidateReferences(net))
	storage[identifier] = nil
	assert.ErrorIs(t, storage.ValidateReferences(net), ErrMvmtNotFound)
	assert.Error(t, storage.ValidateReferences(nil))
}
