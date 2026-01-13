package micro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/elliotchance/orderedmap"
	"github.com/paulmach/orb"
)

// Node represents a microscopic network node (cell vertex)
type Node struct {
	ID gmns.NodeID

	geom          orb.Point
	geomEuclidean orb.Point

	mesoLinkID gmns.LinkID
	laneID     int // Lane number (1-indexed for regular lanes, -1 for bike, -2 for walk)
	cellIndex  int // Position index within the lane (0-indexed)

	isUpstreamEnd   bool // Is this node at the upstream end of a macroscopic link
	isDownstreamEnd bool // Is this node at the downstream end of a macroscopic link

	zoneID       gmns.NodeID        // Inherited from macroscopic node
	boundaryType types.BoundaryType // Inherited from macroscopic node

	incomingLinks  *orderedmap.OrderedMap
	outcomingLinks *orderedmap.OrderedMap
}

// NewNodeFrom creates pointer to the new microscopic Node
func NewNodeFrom(id gmns.NodeID, options ...func(*Node)) *Node {
	newNode := &Node{
		ID:              id,
		geom:            orb.Point{},
		geomEuclidean:   orb.Point{},
		mesoLinkID:      -1,
		laneID:          0,
		cellIndex:       -1,
		isUpstreamEnd:   false,
		isDownstreamEnd: false,
		zoneID:          -1,
		boundaryType:    types.BOUNDARY_NONE,
		incomingLinks:   orderedmap.NewOrderedMap(),
		outcomingLinks:  orderedmap.NewOrderedMap(),
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

// MesoLink returns identifier of parent mesoscopic link. Outputs "-1" if not set.
func (node *Node) MesoLink() gmns.LinkID {
	return node.mesoLinkID
}

// LaneID returns the lane number. Regular lanes are 1-indexed, bike=-1, walk=-2.
func (node *Node) LaneID() int {
	return node.laneID
}

// CellIndex returns the position index within the lane. Outputs "-1" if not set.
func (node *Node) CellIndex() int {
	return node.cellIndex
}

// IsUpstreamEnd returns true if this node is at upstream end of macroscopic link
func (node *Node) IsUpstreamEnd() bool {
	return node.isUpstreamEnd
}

// IsDownstreamEnd returns true if this node is at downstream end of macroscopic link
func (node *Node) IsDownstreamEnd() bool {
	return node.isDownstreamEnd
}

// ZoneID returns the zone identifier. Outputs "-1" if not set.
func (node *Node) ZoneID() gmns.NodeID {
	return node.zoneID
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
func WithPointGeomEuclidean(geomEuclidean orb.Point) func(*Node) {
	return func(node *Node) {
		node.geomEuclidean = geomEuclidean
	}
}

// WithNodeMesoLinkID sets parent mesoscopic link identifier
func WithNodeMesoLinkID(mesoLinkID gmns.LinkID) func(*Node) {
	return func(node *Node) {
		node.mesoLinkID = mesoLinkID
	}
}

// WithNodeLaneID sets the lane number
func WithNodeLaneID(laneID int) func(*Node) {
	return func(node *Node) {
		node.laneID = laneID
	}
}

// WithCellIndex sets the position index within the lane
func WithCellIndex(cellIndex int) func(*Node) {
	return func(node *Node) {
		node.cellIndex = cellIndex
	}
}

// WithIsUpstreamEnd marks the node as upstream end
func WithIsUpstreamEnd(isUpstreamEnd bool) func(*Node) {
	return func(node *Node) {
		node.isUpstreamEnd = isUpstreamEnd
	}
}

// WithIsDownstreamEnd marks the node as downstream end
func WithIsDownstreamEnd(isDownstreamEnd bool) func(*Node) {
	return func(node *Node) {
		node.isDownstreamEnd = isDownstreamEnd
	}
}

// WithZoneID sets the zone identifier
func WithZoneID(zoneID gmns.NodeID) func(*Node) {
	return func(node *Node) {
		node.zoneID = zoneID
	}
}

// WithBoundaryType sets boundary type for the node
func WithBoundaryType(boundaryType types.BoundaryType) func(*Node) {
	return func(node *Node) {
		node.boundaryType = boundaryType
	}
}

// WithIncomingLinks appends given incoming microscopic links to the set
func WithIncomingLinks(linksIDs ...gmns.LinkID) func(*Node) {
	return func(node *Node) {
		for i := range linksIDs {
			node.incomingLinks.Set(linksIDs[i], struct{}{})
		}
	}
}

// WithOutcomingLinks appends given outcoming microscopic links to the set
func WithOutcomingLinks(linksIDs ...gmns.LinkID) func(*Node) {
	return func(node *Node) {
		for i := range linksIDs {
			node.outcomingLinks.Set(linksIDs[i], struct{}{})
		}
	}
}

// AddIncomingLink adds an incoming link to the node
func (node *Node) AddIncomingLink(linkID gmns.LinkID) {
	node.incomingLinks.Set(linkID, struct{}{})
}

// AddOutcomingLink adds an outcoming link to the node
func (node *Node) AddOutcomingLink(linkID gmns.LinkID) {
	node.outcomingLinks.Set(linkID, struct{}{})
}

// SetUpstreamEnd sets the upstream end flag
func (node *Node) SetUpstreamEnd(isUpstreamEnd bool) {
	node.isUpstreamEnd = isUpstreamEnd
}

// SetDownstreamEnd sets the downstream end flag
func (node *Node) SetDownstreamEnd(isDownstreamEnd bool) {
	node.isDownstreamEnd = isDownstreamEnd
}

// SetZoneID sets the zone identifier
func (node *Node) SetZoneID(zoneID gmns.NodeID) {
	node.zoneID = zoneID
}
