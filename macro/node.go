package macro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

// Node representation of a vertex in a macroscopic graph
type Node struct {
	incomingLinks  []gmns.LinkID
	outcomingLinks []gmns.LinkID

	name             string
	osmHighway       string
	ID               gmns.NodeID
	osmNodeID        osm.NodeID
	intersectionID   int
	zoneID           gmns.NodeID
	poiID            gmns.PoiID
	controlType      types.ControlType
	boundaryType     types.BoundaryType
	activityType     types.ActivityType
	activityLinkType types.LinkType

	geom          orb.Point
	geomEuclidean orb.Point

	/* Not used */
	isCentroid bool
}

// NewLinkFrom creates pointer to the new Node
func NewNodeFrom(id gmns.NodeID, options ...func(*Node)) *Node {
	newNode := &Node{
		incomingLinks:    make([]gmns.LinkID, 0),
		outcomingLinks:   make([]gmns.LinkID, 0),
		name:             "",
		osmHighway:       "",
		ID:               id,
		osmNodeID:        osm.NodeID(-1),
		intersectionID:   -1,
		zoneID:           -1,
		poiID:            -1,
		controlType:      types.CONTROL_TYPE_NOT_SIGNAL,
		boundaryType:     types.BOUNDARY_NONE,
		activityType:     types.ACTIVITY_NONE,
		activityLinkType: types.LINK_UNDEFINED,
		geom:             orb.Point{},
		geomEuclidean:    orb.Point{},
	}
	for _, option := range options {
		option(newNode)
	}

	return newNode
}

// IncomingLinks returns set of incoming links. Warning: it outputs slice
func (node *Node) IncomingLinks() []gmns.LinkID {
	return node.incomingLinks
}

// OutcomingLinks returns set of outcoming links. Warning: it outputs slice
func (node *Node) OutcomingLinks() []gmns.LinkID {
	return node.outcomingLinks
}

// Name returns node alias
func (node *Node) Name() string {
	return node.name
}

// OSMHighway returns highway tag for corresponding node in OSM data
func (node *Node) OSMHighway() string {
	return node.osmHighway
}

// OSMNode returns node identifier corresponding to OSM data. Outputs "-1" if it was not set.
func (node *Node) OSMNode() osm.NodeID {
	return node.osmNodeID
}

// Intersection returns intersection identifier. Outputs "-1" if it was not set.
func (node *Node) Intersection() int {
	return node.intersectionID
}

// Zone returns zone identifier. Outputs "-1" if it was not set.
func (node *Node) Zone() gmns.NodeID {
	return node.zoneID
}

// POI returns POI identifier. Outputs "-1" if it was not set.
func (node *Node) POI() gmns.PoiID {
	return node.poiID
}

// ControlType returns node control type
func (node *Node) ControlType() types.ControlType {
	return node.controlType
}

// BoundaryType returns node boundary type
func (node *Node) BoundaryType() types.BoundaryType {
	return node.boundaryType
}

// ActivityType returns node activity type
func (node *Node) ActivityType() types.ActivityType {
	return node.activityType
}

// ActivityLinkType returns node activity link type (based on Link information)
func (node *Node) ActivityLinkType() types.LinkType {
	return node.activityLinkType
}

// Geom returns corresponding geometry [WGS84]
func (node *Node) Geom() orb.Point {
	return node.geom
}

// GeomEuclidean returns corresponding geometry [Euclidean]
func (node *Node) GeomEuclidean() orb.Point {
	return node.geomEuclidean
}

// IsCentroid returns whether node is centroid or not
func (node *Node) IsCentroid() bool {
	return node.isCentroid
}

// WithIncomingLinks appends given incoming links identifiers to the node
func WithIncomingLinks(linksIDs ...gmns.LinkID) func(*Node) {
	return func(node *Node) {
		node.incomingLinks = append(node.incomingLinks, linksIDs...)
	}
}

// WithOutcomingLinks appends given outcoming links identifiers to the node
func WithOutcomingLinks(linksIDs ...gmns.LinkID) func(*Node) {
	return func(node *Node) {
		node.outcomingLinks = append(node.outcomingLinks, linksIDs...)
	}
}

// WithNodeName sets alias for the node
func WithNodeName(name string) func(*Node) {
	return func(node *Node) {
		node.name = name
	}
}

// WithOSMHighwayTag sets highway information from the OSM data for the node
func WithOSMHighwayTag(osmHighway string) func(*Node) {
	return func(node *Node) {
		node.osmHighway = osmHighway
	}
}

// WithOSMNodeID sets corresponding OSM node identifier for the node
func WithOSMNodeID(osmNodeID osm.NodeID) func(*Node) {
	return func(node *Node) {
		node.osmNodeID = osmNodeID
	}
}

// WithIntersectionID sets intersection identifier for the node
func WithIntersectionID(intersectionID int) func(*Node) {
	return func(node *Node) {
		node.intersectionID = intersectionID
	}
}

// WithZoneID sets zone identifier for the node
func WithZoneID(zoneID gmns.NodeID) func(*Node) {
	return func(node *Node) {
		node.zoneID = zoneID
	}
}

// WithPOI sets POI identifier for the node
func WithPOI(poiID gmns.PoiID) func(*Node) {
	return func(node *Node) {
		node.poiID = poiID
	}
}

// WithNodeControlType sets control type for the node
func WithNodeControlType(controlType types.ControlType) func(*Node) {
	return func(node *Node) {
		node.controlType = controlType
	}
}

// WithBoundaryType sets boundary type for the node
func WithBoundaryType(boundaryType types.BoundaryType) func(*Node) {
	return func(node *Node) {
		node.boundaryType = boundaryType
	}
}

// WithActivityType sets activity type for the node
func WithActivityType(activityType types.ActivityType) func(*Node) {
	return func(node *Node) {
		node.activityType = activityType
	}
}

// WithActivityLinkType sets incident link type
func WithActivityLinkType(activityLinkType types.LinkType) func(*Node) {
	return func(node *Node) {
		node.activityLinkType = activityLinkType
	}
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
