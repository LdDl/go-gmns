package micro

import (
	"github.com/LdDl/go-gmns/gmns"
)

// Net is representation of a microscopic road network with links and nodes
type Net struct {
	Nodes map[gmns.NodeID]*Node
	Links map[gmns.LinkID]*Link

	maxNodeID gmns.NodeID
	maxLinkID gmns.LinkID
}

// NewNet returns pointer to the new microscopic road network
func NewNet() *Net {
	return &Net{
		Nodes:     make(map[gmns.NodeID]*Node),
		Links:     make(map[gmns.LinkID]*Link),
		maxNodeID: 0,
		maxLinkID: 0,
	}
}

// MaxNodeID returns the current maximum node ID
func (net *Net) MaxNodeID() gmns.NodeID {
	return net.maxNodeID
}

// MaxLinkID returns the current maximum link ID
func (net *Net) MaxLinkID() gmns.LinkID {
	return net.maxLinkID
}

// SetMaxNodeID sets the maximum node ID
func (net *Net) SetMaxNodeID(id gmns.NodeID) {
	net.maxNodeID = id
}

// SetMaxLinkID sets the maximum link ID
func (net *Net) SetMaxLinkID(id gmns.LinkID) {
	net.maxLinkID = id
}

// AddNode adds a node to the network and updates maxNodeID if necessary
func (net *Net) AddNode(node *Node) {
	net.Nodes[node.ID] = node
	if node.ID >= net.maxNodeID {
		net.maxNodeID = node.ID + 1
	}
}

// AddLink adds a link to the network and updates maxLinkID if necessary
func (net *Net) AddLink(link *Link) {
	net.Links[link.ID] = link
	if link.ID >= net.maxLinkID {
		net.maxLinkID = link.ID + 1
	}
}

// DeleteNode removes a node from the network
func (net *Net) DeleteNode(nodeID gmns.NodeID) {
	delete(net.Nodes, nodeID)
}

// DeleteLink removes a link from the network
func (net *Net) DeleteLink(linkID gmns.LinkID) {
	delete(net.Links, linkID)
}
