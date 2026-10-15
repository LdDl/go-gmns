package micro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/stretchr/testify/assert"
)

type parentValidationCase struct {
	name       string
	prepare    func(*Net)
	wantError  error
	wantDetail string
}

func TestValidateParentReferences(t *testing.T) {
	cases := []parentValidationCase{
		{
			name: "empty network",
		},
		{
			name: "absent parent references",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1))
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3))
			},
		},
		{
			name: "valid zero references",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(0)))
				net.Nodes[gmns.NodeIDFromInt64(0)] = NewNodeFrom(gmns.NodeIDFromInt64(0))
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(0)))
				net.Links[gmns.LinkIDFromInt64(0)] = NewLinkFrom(gmns.LinkIDFromInt64(0), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3))
			},
		},
		{
			name: "valid negative references",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(-1)))
				net.Nodes[gmns.NodeIDFromInt64(-1)] = NewNodeFrom(gmns.NodeIDFromInt64(-1))
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(-1)))
				net.Links[gmns.LinkIDFromInt64(-1)] = NewLinkFrom(gmns.LinkIDFromInt64(-1), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3))
			},
		},
		{
			name: "missing node parent",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(10)))
			},
			wantError:  ErrNodeNotFound,
			wantDetail: "node 1: parent_node_id 10",
		},
		{
			name: "missing link parent",
			prepare: func(net *Net) {
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(20)))
			},
			wantError:  ErrLinkNotFound,
			wantDetail: "link 2: parent_link_id 20",
		},
		{
			name: "zero is not absence",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(0)))
			},
			wantError:  ErrNodeNotFound,
			wantDetail: "node 1: parent_node_id 0",
		},
		{
			name: "negative one is not absence",
			prepare: func(net *Net) {
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(-1)))
			},
			wantError:  ErrLinkNotFound,
			wantDetail: "link 2: parent_link_id -1",
		},
		{
			name: "node parent cannot resolve to link",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(10)))
				net.Links[gmns.LinkIDFromInt64(10)] = NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3))
			},
			wantError:  ErrNodeNotFound,
			wantDetail: "node 1: parent_node_id 10",
		},
		{
			name: "link parent cannot resolve to node",
			prepare: func(net *Net) {
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(20)))
				net.Nodes[gmns.NodeIDFromInt64(20)] = NewNodeFrom(gmns.NodeIDFromInt64(20))
			},
			wantError:  ErrLinkNotFound,
			wantDetail: "link 2: parent_link_id 20",
		},
		{
			name: "nil node target",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(10)))
				net.Nodes[gmns.NodeIDFromInt64(10)] = nil
			},
			wantError:  ErrNodeNotFound,
			wantDetail: "node 1: parent_node_id 10",
		},
		{
			name: "nil link target",
			prepare: func(net *Net) {
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(20)))
				net.Links[gmns.LinkIDFromInt64(20)] = nil
			},
			wantError:  ErrLinkNotFound,
			wantDetail: "link 2: parent_link_id 20",
		},
		{
			name: "node target key does not match record ID",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(10)))
				net.Nodes[gmns.NodeIDFromInt64(10)] = NewNodeFrom(gmns.NodeIDFromInt64(11))
			},
			wantDetail: "node 1: parent_node_id 10 resolves to node with ID 11",
		},
		{
			name: "link target key does not match record ID",
			prepare: func(net *Net) {
				net.Links[gmns.LinkIDFromInt64(2)] = NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), WithParentLinkID(gmns.LinkIDFromInt64(20)))
				net.Links[gmns.LinkIDFromInt64(20)] = NewLinkFrom(gmns.LinkIDFromInt64(21), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3))
			},
			wantDetail: "link 2: parent_link_id 20 resolves to link with ID 21",
		},
		{
			name: "nil node record",
			prepare: func(net *Net) {
				net.Nodes[gmns.NodeIDFromInt64(1)] = nil
			},
			wantError:  ErrNodeNotFound,
			wantDetail: "node 1:",
		},
		{
			name: "nil link record",
			prepare: func(net *Net) {
				net.Links[gmns.LinkIDFromInt64(2)] = nil
			},
			wantError:  ErrLinkNotFound,
			wantDetail: "link 2:",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			net := NewNet()
			if testCase.prepare != nil {
				testCase.prepare(net)
			}
			err := net.ValidateParentReferences()
			if testCase.wantDetail == "" {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			if err == nil {
				return
			}
			assert.Contains(t, err.Error(), testCase.wantDetail)
			if testCase.wantError != nil {
				assert.ErrorIs(t, err, testCase.wantError)
			}
		})
	}
}

func TestParentValidationIsLocalAndDoesNotRepairReferences(t *testing.T) {
	net := NewNet()
	net.Nodes[gmns.NodeIDFromInt64(1)] = NewNodeFrom(gmns.NodeIDFromInt64(1), WithParentNodeID(gmns.NodeIDFromInt64(10)))
	net.Nodes[gmns.NodeIDFromInt64(2)] = NewNodeFrom(gmns.NodeIDFromInt64(2), WithParentNodeID(gmns.NodeIDFromInt64(20)))
	other := NewNet()
	other.Nodes[gmns.NodeIDFromInt64(10)] = NewNodeFrom(gmns.NodeIDFromInt64(10))
	other.Nodes[gmns.NodeIDFromInt64(20)] = NewNodeFrom(gmns.NodeIDFromInt64(20))
	firstErr := net.ValidateParentReferences()
	assert.ErrorIs(t, firstErr, ErrNodeNotFound)
	if firstErr == nil {
		return
	}
	assert.Contains(t, firstErr.Error(), "node 1: parent_node_id 10")
	secondErr := net.ValidateParentReferences()
	assert.EqualError(t, secondErr, firstErr.Error(), "error order must be deterministic")
	parent, present := net.Nodes[gmns.NodeIDFromInt64(1)].ParentNode()
	assert.Equal(t, gmns.NodeIDFromInt64(10), parent)
	assert.True(t, present)
	assert.Len(t, net.Nodes, 2)
	assert.Len(t, other.Nodes, 2)
	net.Nodes[gmns.NodeIDFromInt64(10)] = other.Nodes[gmns.NodeIDFromInt64(10)]
	net.Nodes[gmns.NodeIDFromInt64(20)] = other.Nodes[gmns.NodeIDFromInt64(20)]
	err := net.ValidateParentReferences()
	assert.NoError(t, err)
}

func TestParentValidationHandlesEmptyAndNilNetwork(t *testing.T) {
	net := &Net{}
	err := net.ValidateParentReferences()
	assert.NoError(t, err)
	var missing *Net
	err = missing.ValidateParentReferences()
	assert.EqualError(t, err, "network is nil")
}
