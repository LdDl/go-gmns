package macro

import (
	"github.com/LdDl/go-gmns/gmns"
)

type Link struct {
	lengthMeters float64
	ID           gmns.LinkID
	lanesNum     int
}

// LengthMeters returns length of the underlying geometry [WGS84]
func (link *Link) LengthMeters() float64 {
	return link.lengthMeters
}

// LanesNum returns number of lanes
func (link *Link) LanesNum() int {
	return link.lanesNum
}
