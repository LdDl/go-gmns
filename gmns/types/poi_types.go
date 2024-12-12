package types

type POIType uint16

const (
	POI_TYPE_UNDEFINED = POIType(iota)
	POI_TYPE_HIGHWAY
	POI_TYPE_RAILWAY
	POI_TYPE_AEROWAY
)
