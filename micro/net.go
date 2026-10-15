package micro

import (
	"errors"
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/movement"
)

// Net is representation of a microscopic road network with links and nodes
type Net struct {
	lanes          map[gmns.LaneID]*gmns.Lane
	segments       map[gmns.SegmentID]*gmns.Segment
	segmentLanes   map[gmns.SegmentLaneID]*gmns.SegmentLane
	config         *gmns.Config
	geometries     map[gmns.GeometryID]*gmns.Geometry
	zones          map[gmns.NodeID]*gmns.Zone
	useDefinitions map[string]*gmns.UseDefinition
	useGroups      map[string]*gmns.UseGroup
	movements      movement.MovementsStorage

	Nodes map[gmns.NodeID]*Node
	Links map[gmns.LinkID]*Link

	maxNodeID gmns.IDAllocator
	maxLinkID gmns.IDAllocator
}

// NewNet returns pointer to the new microscopic road network
func NewNet() *Net {
	return &Net{
		Nodes: make(map[gmns.NodeID]*Node),
		Links: make(map[gmns.LinkID]*Link),
	}
}

// NewNetWithIDType configures the kind of newly allocated node and link IDs.
func NewNetWithIDType(kind gmns.IDType) (*Net, error) {
	nodeIDs, err := gmns.NewIDAllocator(kind)
	if err != nil {
		return nil, err
	}
	linkIDs, err := gmns.NewIDAllocator(kind)
	if err != nil {
		return nil, err
	}
	net := NewNet()
	net.maxNodeID = *nodeIDs
	net.maxLinkID = *linkIDs
	return net, nil
}

// MaxNodeID returns the next available node ID in the configured kind, not the greatest existing key.
// Direct map insertions are checked before returning a candidate.
func (net *Net) MaxNodeID() gmns.NodeID {
	for {
		candidate := net.maxNodeID.PeekNodeID()
		if _, exists := net.Nodes[candidate]; !exists {
			return candidate
		}
		net.maxNodeID.NextNodeID()
	}
}

// MaxLinkID returns the next available link ID in the configured kind, not the greatest existing key.
func (net *Net) MaxLinkID() gmns.LinkID {
	for {
		candidate := net.maxLinkID.PeekLinkID()
		if _, exists := net.Links[candidate]; !exists {
			return candidate
		}
		net.maxLinkID.NextLinkID()
	}
}

// SetMaxNodeID sets the integer allocation counter without releasing reserved IDs.
func (net *Net) SetMaxNodeID(id gmns.NodeID) error {
	if id.Type() != gmns.IDTypeInteger {
		return fmt.Errorf("node allocation counter requires an integer ID")
	}
	return net.maxNodeID.SetNextInteger(id.String())
}

// SetMaxLinkID sets the integer allocation counter without releasing reserved IDs.
func (net *Net) SetMaxLinkID(id gmns.LinkID) error {
	if id.Type() != gmns.IDTypeInteger {
		return fmt.Errorf("link allocation counter requires an integer ID")
	}
	return net.maxLinkID.SetNextInteger(id.String())
}

// AddNode adds a valid node without overwriting an existing record.
func (net *Net) AddNode(node *Node) error {
	if node == nil {
		return ErrNodeNotFound
	}
	err := node.ID.Validate()
	if err != nil {
		return err
	}
	if _, exists := net.Nodes[node.ID]; exists {
		return fmt.Errorf("node %v: %w", node.ID, gmns.ErrDuplicateID)
	}
	if net.Nodes == nil {
		net.Nodes = make(map[gmns.NodeID]*Node)
	}
	err = net.maxNodeID.ReserveNodeID(node.ID)
	if err != nil && !errors.Is(err, gmns.ErrDuplicateID) {
		return err
	}
	net.Nodes[node.ID] = node
	return nil
}

// AddLink adds a valid link without overwriting an existing record.
func (net *Net) AddLink(link *Link) error {
	if link == nil {
		return ErrLinkNotFound
	}
	err := link.ID.Validate()
	if err != nil {
		return err
	}
	if _, exists := net.Links[link.ID]; exists {
		return fmt.Errorf("link %v: %w", link.ID, gmns.ErrDuplicateID)
	}
	if net.Links == nil {
		net.Links = make(map[gmns.LinkID]*Link)
	}
	err = net.maxLinkID.ReserveLinkID(link.ID)
	if err != nil && !errors.Is(err, gmns.ErrDuplicateID) {
		return err
	}
	net.Links[link.ID] = link
	return nil
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
		return nodeIDs[first].Less(nodeIDs[second])
	})
	for _, nodeID := range nodeIDs {
		node := net.Nodes[nodeID]
		if node == nil {
			return fmt.Errorf("node %v: %w", nodeID, ErrNodeNotFound)
		}
		parentID, present := node.ParentNode()
		if !present {
			continue
		}
		parent := net.Nodes[parentID]
		if parent == nil {
			return fmt.Errorf("node %v: parent_node_id %v: %w", node.ID, parentID, ErrNodeNotFound)
		}
		if parent.ID != parentID {
			return fmt.Errorf("node %v: parent_node_id %v resolves to node with ID %v", node.ID, parentID, parent.ID)
		}
	}

	linkIDs := make([]gmns.LinkID, 0, len(net.Links))
	for linkID := range net.Links {
		linkIDs = append(linkIDs, linkID)
	}
	sort.Slice(linkIDs, func(first, second int) bool {
		return linkIDs[first].Less(linkIDs[second])
	})
	for _, linkID := range linkIDs {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("link %v: %w", linkID, ErrLinkNotFound)
		}
		parentID, present := link.ParentLink()
		if !present {
			continue
		}
		parent := net.Links[parentID]
		if parent == nil {
			return fmt.Errorf("link %v: parent_link_id %v: %w", link.ID, parentID, ErrLinkNotFound)
		}
		if parent.ID != parentID {
			return fmt.Errorf("link %v: parent_link_id %v resolves to link with ID %v", link.ID, parentID, parent.ID)
		}
	}
	return nil
}
