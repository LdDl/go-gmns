package types

// LinkClass is just type alias for the link class
type LinkClass uint16

const (
	LINK_CLASS_UNDEFINED = LinkClass(iota)
	LINK_CLASS_HIGHWAY
	LINK_CLASS_RAILWAY
	LINK_CLASS_AEROWAY
)

var linkClassStr = []string{"undefined", "highway", "railway", "aeroway"}

func (iotaIdx LinkClass) String() string {
	return linkClassStr[iotaIdx]
}
