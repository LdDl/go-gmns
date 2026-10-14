package macro

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
)

// Net is representation of any road network with links and nodes in GMNS specification
type Net struct {
	Nodes map[gmns.NodeID]*Node
	Links map[gmns.LinkID]*Link
}

// NewNet returns pointer to the new macroscopic road network data
func NewNet() *Net {
	return &Net{
		Nodes: make(map[gmns.NodeID]*Node),
		Links: make(map[gmns.LinkID]*Link),
	}
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
