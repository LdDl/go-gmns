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
