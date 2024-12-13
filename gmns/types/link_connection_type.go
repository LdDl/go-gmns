package types

// LinkConnectionType is just type alias for the link connection type
type LinkConnectionType uint16

const (
	// Plain way
	NOT_A_LINK = LinkConnectionType(iota)
	// Connection between two roads
	IS_LINK
)

var linkConnectionTypeStr = []string{"no", "yes"}

func (iotaIdx LinkConnectionType) String() string {
	return linkConnectionTypeStr[iotaIdx]
}
