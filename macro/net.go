package macro

import (
	"github.com/LdDl/go-gmns/gmns"
)

type Net struct {
	Nodes map[gmns.NodeID]*Node
	Links map[gmns.LinkID]*Link
}
