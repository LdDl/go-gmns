package micro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/stretchr/testify/assert"
)

func TestAllocationChecksDirectMapWritesAndRejectsOverwrite(t *testing.T) {
	net := NewNet()
	zero := gmns.NodeIDFromInt64(0)
	existing := NewNodeFrom(zero)
	net.Nodes[zero] = existing
	assert.Equal(t, gmns.NodeIDFromInt64(1), net.MaxNodeID())
	err := net.AddNode(NewNodeFrom(zero, WithNodeName("replacement")))
	assert.ErrorIs(t, err, gmns.ErrDuplicateID)
	assert.Same(t, existing, net.Nodes[zero])
	first := net.MaxNodeID()
	assert.NoError(t, net.AddNode(NewNodeFrom(first)))
	net.DeleteNode(first)
	assert.NotEqual(t, first, net.MaxNodeID(), "deleting a node must not recycle a reserved ID")
	assert.NoError(t, net.SetMaxNodeID(zero))
	assert.Equal(t, gmns.NodeIDFromInt64(2), net.MaxNodeID())

	linkID := gmns.LinkIDFromInt64(0)
	net.Links[linkID] = NewLinkFrom(linkID, zero, zero)
	assert.Equal(t, gmns.LinkIDFromInt64(1), net.MaxLinkID())
	assert.ErrorIs(t, net.AddLink(NewLinkFrom(linkID, zero, zero)), gmns.ErrDuplicateID)
	assert.Error(t, net.AddNode(nil))
	assert.Error(t, net.AddLink(nil))
	assert.Error(t, net.AddNode(NewNodeFrom(gmns.NodeID{})))
}

func TestConfiguredAllocationAndLargeCounter(t *testing.T) {
	net, err := NewNetWithIDType(gmns.IDTypeString)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, gmns.NodeIDFromString("0"), net.MaxNodeID())
	assert.NoError(t, net.AddNode(NewNodeFrom(net.MaxNodeID())))
	assert.Equal(t, gmns.NodeIDFromString("1"), net.MaxNodeID())
	integer, err := gmns.ParseNodeID("9223372036854775808", gmns.IDTypeInteger)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NoError(t, net.SetMaxNodeID(integer))
	assert.Equal(t, gmns.NodeIDFromString("9223372036854775808"), net.MaxNodeID())
	assert.NoError(t, net.AddNode(NewNodeFrom(net.MaxNodeID())))
	assert.Equal(t, gmns.NodeIDFromString("9223372036854775809"), net.MaxNodeID())
	assert.Error(t, net.SetMaxNodeID(gmns.NodeIDFromString("2")))
	assert.Error(t, net.SetMaxLinkID(gmns.LinkIDFromInt64(-1)))
}
