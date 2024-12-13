package macro

import (
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
