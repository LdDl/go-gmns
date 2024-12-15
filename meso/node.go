package meso

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/elliotchance/orderedmap"
	"github.com/paulmach/orb"
)

type Node struct {
	ID gmns.NodeID

	geom          orb.Point
	geomEuclidean orb.Point

	macroNodeID gmns.NodeID
	macroLinkID gmns.LinkID

	macroZoneID      gmns.NodeID        // Should be inherited from the macroscopic node
	activityLinkType types.LinkType     // Should be inherited from the macroscopic node
	boundaryType     types.BoundaryType // Should be evaluated from macroscopic node and macroscopic link

	incomingLinks  *orderedmap.OrderedMap
	outcomingLinks *orderedmap.OrderedMap
}

// NewLinkFrom creates pointer to the new Node
func NewNodeFrom(id gmns.NodeID, options ...func(*Node)) *Node {
	newNode := &Node{
		ID:               id,
		geom:             orb.Point{},
		geomEuclidean:    orb.Point{},
		macroNodeID:      -1,
		macroLinkID:      -1,
		macroZoneID:      -1,
		activityLinkType: types.LINK_UNDEFINED,
		boundaryType:     types.BOUNDARY_NONE,
		incomingLinks:    orderedmap.NewOrderedMap(),
		outcomingLinks:   orderedmap.NewOrderedMap(),
	}
	for _, option := range options {
		option(newNode)
	}
	return newNode
}

// Geom returns corresponding geometry [WGS84]
func (node *Node) Geom() orb.Point {
	return node.geom
}

// GeomEuclidean returns corresponding geometry [Euclidean]
func (node *Node) GeomEuclidean() orb.Point {
	return node.geomEuclidean
}

// MacroNode returns identifier of parent macroscopic node. Outputs "-1" if it was not set.
func (node *Node) MacroNode() gmns.NodeID {
	return node.macroNodeID
}

// MacroLink returns identifier of parent macroscopic link. Outputs "-1" if it was not set.
func (node *Node) MacroLink() gmns.LinkID {
	return node.macroLinkID
}

// MacroZone returns parent macroscopic node's zone indetifier. Outputs "-1" if it was not set.
func (node *Node) MacroZone() gmns.NodeID {
	return node.macroZoneID
}

// ActivityLinkType returns node activity link type (based on Link information)
func (node *Node) ActivityLinkType() types.LinkType {
	return node.activityLinkType
}

// BoundaryType returns node boundary type
func (node *Node) BoundaryType() types.BoundaryType {
	return node.boundaryType
}

// IncomingLinks returns set of incoming links. Warning: it outputs pointer
func (node *Node) IncomingLinks() *orderedmap.OrderedMap {
	return node.incomingLinks
}

// OutcomingLinks returns set of outcoming links. Warning: it outputs pointer
func (node *Node) OutcomingLinks() *orderedmap.OrderedMap {
	return node.outcomingLinks
}

// WithPointGeom sets geometry [WGS84] for the node
func WithPointGeom(geom orb.Point) func(*Node) {
	return func(node *Node) {
		node.geom = geom
	}
}

// WithPointGeomEuclidean sets geometry [Euclidean] for the node
func WithPointEuclideanGeom(geomEuclidean orb.Point) func(*Node) {
	return func(node *Node) {
		node.geomEuclidean = geomEuclidean
	}
}

// WithPointMacroNodeID sets parent macro node identifier
func WithPointMacroNodeID(macroNodeID gmns.NodeID) func(*Node) {
	return func(node *Node) {
		node.macroNodeID = macroNodeID
	}
}

// WithPointMacroLinkID sets parent macro link identifier
func WithPointMacroLinkID(macroLinkID gmns.LinkID) func(*Node) {
	return func(node *Node) {
		node.macroLinkID = macroLinkID
	}
}

// WithMacroZone sets parent macroscopic node's zone indetifier
func WithMacroZone(macroZoneID gmns.NodeID) func(*Node) {
	return func(node *Node) {
		node.macroZoneID = macroZoneID
	}
}

// WithActivityLinkType sets incident link type
func WithActivityLinkType(activityLinkType types.LinkType) func(*Node) {
	return func(node *Node) {
		node.activityLinkType = activityLinkType
	}
}

// WithBoundaryType sets boundary type for the node
func WithBoundaryType(boundaryType types.BoundaryType) func(*Node) {
	return func(node *Node) {
		node.boundaryType = boundaryType
	}
}

// WithIncomingLinks appends given incoming mesoscopic links to the set of existing incoming links. Warning: it preserves order of insertion
func WithIncomingLinks(linksIDs ...gmns.LinkID) func(*Node) {
	return func(node *Node) {
		for i := range linksIDs {
			node.incomingLinks.Set(linksIDs[i], struct{}{})
		}
	}
}

// WithOutcomingLinks appends given outcoming mesoscopic links to the set of existing outcoming links. Warning: it preserves order of insertion
func WithOutcomingLinks(linksIDs ...gmns.LinkID) func(*Node) {
	return func(node *Node) {
		for i := range linksIDs {
			node.outcomingLinks.Set(linksIDs[i], struct{}{})
		}
	}
}
