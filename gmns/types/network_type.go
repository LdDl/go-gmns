package types

type NetworkType uint16

const (
	NETWORK_UNDEFINED = NetworkType(iota)
	NETWORK_AUTO
	NETWORK_BIKE
	NETWORK_WALK
	NETWORK_RAILWAY
	NETWORK_AEROWAY
)

func (iotaIdx NetworkType) String() string {
	return [...]string{"undefined", "auto", "bike", "walk", "railway", "aeroway"}[iotaIdx]
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
