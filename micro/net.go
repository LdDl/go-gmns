package micro

import (
	"fmt"
	"sort"

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

// ValidateParentReferences checks optional parent references within this network.
// It does not change records or validate topology and source provenance.
func (net *Net) ValidateParentReferences() error {
	if net == nil {
		return fmt.Errorf("network is nil")
	}
	nodeIDs := make([]gmns.NodeID, 0, len(net.Nodes))
	for nodeID := range net.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Slice(nodeIDs, func(first, second int) bool {
		return nodeIDs[first] < nodeIDs[second]
	})
	for _, nodeID := range nodeIDs {
		node := net.Nodes[nodeID]
		if node == nil {
			return fmt.Errorf("node %d: %w", nodeID, ErrNodeNotFound)
		}
		parentID, present := node.ParentNode()
		if !present {
			continue
		}
		parent := net.Nodes[parentID]
		if parent == nil {
			return fmt.Errorf("node %d: parent_node_id %d: %w", node.ID, parentID, ErrNodeNotFound)
		}
		if parent.ID != parentID {
			return fmt.Errorf("node %d: parent_node_id %d resolves to node with ID %d", node.ID, parentID, parent.ID)
		}
	}

	linkIDs := make([]gmns.LinkID, 0, len(net.Links))
	for linkID := range net.Links {
		linkIDs = append(linkIDs, linkID)
	}
	sort.Slice(linkIDs, func(first, second int) bool {
		return linkIDs[first] < linkIDs[second]
	})
	for _, linkID := range linkIDs {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("link %d: %w", linkID, ErrLinkNotFound)
		}
		parentID, present := link.ParentLink()
		if !present {
			continue
		}
		parent := net.Links[parentID]
		if parent == nil {
			return fmt.Errorf("link %d: parent_link_id %d: %w", link.ID, parentID, ErrLinkNotFound)
		}
		if parent.ID != parentID {
			return fmt.Errorf("link %d: parent_link_id %d resolves to link with ID %d", link.ID, parentID, parent.ID)
		}
	}
	return nil
}
