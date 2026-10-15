package meso

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/elliotchance/orderedmap"
	"github.com/paulmach/orb"
)

type Node struct {
	controlType types.ControlType

	altitudeMeters *float64

	parentNodeID *gmns.NodeID

	name     string
	nodeType string

	ID gmns.NodeID

	geom          orb.Point
	geomEuclidean orb.Point

	macroNodeID gmns.NodeID
	macroLinkID gmns.LinkID

	// Should be inherited from the macroscopic node
	macroZoneID      *gmns.NodeID
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

// AltitudeMeters returns the supplied altitude in meters and whether it is set.
func (node *Node) AltitudeMeters() (float64, bool) {
	if node.altitudeMeters == nil {
		return 0, false
	}
	return *node.altitudeMeters, true
}

// WithAltitudeMeters stores altitude independently of the two-dimensional geometry.
// Zero and negative altitudes are supplied values, not absence.
func WithAltitudeMeters(altitudeMeters float64) func(*Node) {
	return func(node *Node) {
		value := altitudeMeters
		node.altitudeMeters = &value
	}
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

// MacroZone returns the supplied zone identifier and whether it is set.
func (node *Node) MacroZone() (gmns.NodeID, bool) {
	if node.macroZoneID == nil {
		return 0, false
	}
	return *node.macroZoneID, true
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
		value := macroZoneID
		node.macroZoneID = &value
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

// NodeType returns what the node represents, or an empty string when not supplied.
func (node *Node) NodeType() string {
	return node.nodeType
}

// WithNodeType sets what the node represents; an empty string clears the value.
func WithNodeType(nodeType string) func(*Node) {
	return func(node *Node) {
		node.nodeType = nodeType
	}
}

// Name returns the supplied node name, or an empty string when not supplied.
func (node *Node) Name() string {
	return node.name
}

// WithNodeName sets the node name; an empty string clears the value.
func WithNodeName(name string) func(*Node) {
	return func(node *Node) {
		node.name = name
	}
}

// ParentNode returns the optional associated node ID within the same network.
// This reference is independent of source provenance across network levels.
func (node *Node) ParentNode() (gmns.NodeID, bool) {
	if node.parentNodeID == nil {
		return 0, false
	}
	return *node.parentNodeID, true
}

// WithParentNodeID sets an associated node reference within the same network.
// Zero and negative IDs are stored as supplied values, not as absence.
func WithParentNodeID(parentNodeID gmns.NodeID) func(*Node) {
	return func(node *Node) {
		value := parentNodeID
		node.parentNodeID = &value
	}
}

// ControlType returns the supplied control category, or empty if absent.
func (node *Node) ControlType() types.ControlType {
	return node.controlType
}

// WithNodeControlType stores or clears a supplied control category.
func WithNodeControlType(controlType types.ControlType) func(*Node) {
	return func(node *Node) {
		node.controlType = controlType
	}
}

// WithoutAltitudeMeters clears the supplied value, restoring absence.
func WithoutAltitudeMeters() func(*Node) {
	return func(node *Node) {
		node.altitudeMeters = nil
	}
}

// WithoutParentNodeID clears the supplied value, restoring absence.
func WithoutParentNodeID() func(*Node) {
	return func(node *Node) {
		node.parentNodeID = nil
	}
}

// WithoutMacroZone clears the supplied value, restoring absence.
func WithoutMacroZone() func(*Node) {
	return func(node *Node) {
		node.macroZoneID = nil
	}
}
