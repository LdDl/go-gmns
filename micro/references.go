package micro

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
)

// ValidateReferences checks record IDs, map keys, link endpoints and same-network parents.
// It checks zone/geometry ID validity, but their targets need ValidateExternalReferences.
// Source provenance and calculated adjacency indexes are separate from this check.
func (net *Net) ValidateReferences() error {
	if net == nil {
		return fmt.Errorf("network is nil")
	}
	nodeIDs := net.sortedNodeIDs()
	for _, nodeID := range nodeIDs {
		node := net.Nodes[nodeID]
		if node == nil {
			return fmt.Errorf("node %v: %w", nodeID, ErrNodeNotFound)
		}
		err := nodeID.Validate()
		if err != nil {
			return fmt.Errorf("node key %q: %w", nodeID, err)
		}
		if node.ID != nodeID {
			return fmt.Errorf("node key %q does not match record ID %q", nodeID, node.ID)
		}
		zoneID, present := node.ZoneID()
		if present {
			err = zoneID.Validate()
			if err != nil {
				return fmt.Errorf("node %v zone_id: %w", nodeID, err)
			}
		}
		parentID, present := node.ParentNode()
		if present {
			err = parentID.Validate()
			if err != nil {
				return fmt.Errorf("node %v parent_node_id: %w", nodeID, err)
			}
		}
	}
	for _, linkID := range net.sortedLinkIDs() {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("link %v: %w", linkID, ErrLinkNotFound)
		}
		err := linkID.Validate()
		if err != nil {
			return fmt.Errorf("link key %q: %w", linkID, err)
		}
		if link.ID != linkID {
			return fmt.Errorf("link key %q does not match record ID %q", linkID, link.ID)
		}
		for _, nodeID := range []gmns.NodeID{link.SourceNode(), link.TargetNode()} {
			err = nodeID.Validate()
			if err != nil {
				return fmt.Errorf("link %v endpoint: %w", linkID, err)
			}
			if net.Nodes[nodeID] == nil {
				return fmt.Errorf("link %v endpoint %q: %w", linkID, nodeID, ErrNodeNotFound)
			}
		}
		geometryID, present := link.GeometryID()
		if present {
			err = geometryID.Validate()
			if err != nil {
				return fmt.Errorf("link %v geometry_id: %w", linkID, err)
			}
		}
		parentID, present := link.ParentLink()
		if present {
			err = parentID.Validate()
			if err != nil {
				return fmt.Errorf("link %v parent_link_id: %w", linkID, err)
			}
		}
	}
	return net.ValidateParentReferences()
}

// ValidateIDType checks every standard ID/reference against one configured kind.
// Generator provenance can refer to a different dataset and is not checked here.
func (net *Net) ValidateIDType(kind gmns.IDType) error {
	err := kind.Validate()
	if err != nil {
		return err
	}
	err = net.ValidateReferences()
	if err != nil {
		return err
	}
	for _, nodeID := range net.sortedNodeIDs() {
		node := net.Nodes[nodeID]
		ids := []gmns.NodeID{node.ID}
		parent, present := node.ParentNode()
		if present {
			ids = append(ids, parent)
		}
		zone, present := node.ZoneID()
		if present {
			ids = append(ids, zone)
		}
		for _, identifier := range ids {
			err = identifier.ValidateType(kind)
			if err != nil {
				return fmt.Errorf("node %v ID/reference %q: %w", nodeID, identifier, err)
			}
		}
	}
	for _, linkID := range net.sortedLinkIDs() {
		link := net.Links[linkID]
		ids := []gmns.LinkID{link.ID}
		parent, present := link.ParentLink()
		if present {
			ids = append(ids, parent)
		}
		for _, identifier := range ids {
			err = identifier.ValidateType(kind)
			if err != nil {
				return fmt.Errorf("link %v ID/reference %q: %w", linkID, identifier, err)
			}
		}
		geometry, present := link.GeometryID()
		if present {
			err = geometry.ValidateType(kind)
			if err != nil {
				return fmt.Errorf("link %v geometry_id: %w", linkID, err)
			}
		}
	}
	return nil
}

// ValidateExternalReferences resolves supplied zone and geometry IDs against explicit catalogs.
// Empty catalogs are valid only when the network has no references to those resources.
func (net *Net) ValidateExternalReferences(zoneIDs []gmns.NodeID, geometryIDs []gmns.GeometryID) error {
	err := net.ValidateReferences()
	if err != nil {
		return err
	}
	zones := make(map[gmns.NodeID]struct{}, len(zoneIDs))
	for _, identifier := range zoneIDs {
		err = identifier.Validate()
		if err != nil {
			return fmt.Errorf("zone: %w", err)
		}
		if _, exists := zones[identifier]; exists {
			return fmt.Errorf("zone %v: %w", identifier, gmns.ErrDuplicateID)
		}
		zones[identifier] = struct{}{}
	}
	geometries := make(map[gmns.GeometryID]struct{}, len(geometryIDs))
	for _, identifier := range geometryIDs {
		err = identifier.Validate()
		if err != nil {
			return fmt.Errorf("geometry: %w", err)
		}
		if _, exists := geometries[identifier]; exists {
			return fmt.Errorf("geometry %v: %w", identifier, gmns.ErrDuplicateID)
		}
		geometries[identifier] = struct{}{}
	}
	for _, nodeID := range net.sortedNodeIDs() {
		zoneID, present := net.Nodes[nodeID].ZoneID()
		if present {
			if _, exists := zones[zoneID]; !exists {
				return fmt.Errorf("node %v: zone_id %q not found", nodeID, zoneID)
			}
		}
	}
	for _, linkID := range net.sortedLinkIDs() {
		geometryID, present := net.Links[linkID].GeometryID()
		if present {
			if _, exists := geometries[geometryID]; !exists {
				return fmt.Errorf("link %v: geometry_id %q not found", linkID, geometryID)
			}
		}
	}
	return nil
}

func (net *Net) sortedNodeIDs() []gmns.NodeID {
	ids := make([]gmns.NodeID, 0, len(net.Nodes))
	for id := range net.Nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	return ids
}

func (net *Net) sortedLinkIDs() []gmns.LinkID {
	ids := make([]gmns.LinkID, 0, len(net.Links))
	for id := range net.Links {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	return ids
}
