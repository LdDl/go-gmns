package meso

import "github.com/paulmach/orb/encoding/wkt"

// WKT encodes the node's two-dimensional WGS84 geometry without validation.
func (node *Node) WKT() string {
	return wkt.MarshalString(node.geom)
}

// WKT encodes the link's inline geometry without resolving references or applying DirFlag.
func (link *Link) WKT() string {
	return wkt.MarshalString(link.geom)
}
