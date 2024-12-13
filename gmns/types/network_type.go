package types

// NetworkType is just type alias for the network type
type NetworkType uint16

const (
	NETWORK_UNDEFINED = NetworkType(iota)
	NETWORK_AUTO
	NETWORK_BIKE
	NETWORK_WALK
	NETWORK_RAILWAY
	NETWORK_AEROWAY
)

var networkTypeStr = []string{"undefined", "auto", "bike", "walk", "railway", "aeroway"}

func (iotaIdx NetworkType) String() string {
	return networkTypeStr[iotaIdx]
}

var (
	networkTypesAll = map[NetworkType]struct{}{
		NETWORK_AUTO:      {},
		NETWORK_BIKE:      {},
		NETWORK_WALK:      {},
		NETWORK_RAILWAY:   {},
		NETWORK_UNDEFINED: {},
	}
	networkTypesDefault = map[NetworkType]struct{}{
		NETWORK_AUTO: {},
	}
)
