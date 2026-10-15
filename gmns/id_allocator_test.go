package gmns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDAllocatorAvoidsReservedAndReturnedIDs(t *testing.T) {
	allocator, err := NewIDAllocator(IDTypeString)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NoError(t, allocator.ReserveNodeID(NodeIDFromString("0")))
	assert.NoError(t, allocator.ReserveNodeID(NodeIDFromString("001")))
	first := allocator.NextNodeID()
	second := allocator.NextNodeID()
	assert.Equal(t, NodeIDFromString("1"), first)
	assert.Equal(t, NodeIDFromString("2"), second)
	assert.ErrorIs(t, allocator.ReserveNodeID(first), ErrDuplicateID)
	assert.NoError(t, allocator.SetNextInteger("0"))
	assert.Equal(t, NodeIDFromString("3"), allocator.NextNodeID())
	assert.Error(t, allocator.SetNextInteger("-1"))
	assert.Error(t, allocator.SetNextInteger("1.2"))
	assert.Equal(t, NodeIDFromString("4"), allocator.NextNodeID())
}

func TestIDAllocatorHasIndependentNamespacesAndNoInt64Overflow(t *testing.T) {
	nodes := &IDAllocator{}
	links := &IDAllocator{}
	assert.Equal(t, NodeIDFromInt64(0), nodes.NextNodeID())
	assert.Equal(t, LinkIDFromInt64(0), links.NextLinkID())
	assert.NoError(t, nodes.ReserveNodeID(NodeIDFromInt64(-1)))
	assert.NoError(t, nodes.SetNextInteger("9223372036854775807"))
	assert.Equal(t, "9223372036854775807", nodes.NextNodeID().String())
	assert.Equal(t, "9223372036854775808", nodes.NextNodeID().String())
	assert.Equal(t, "9223372036854775809", nodes.NextNodeID().String())
	assert.Error(t, nodes.ReserveNodeID(NodeID{}))
	_, err := NewIDAllocator(IDType("float"))
	assert.Error(t, err)
}
