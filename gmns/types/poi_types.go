package types

// POIType is just type alias for the POI type
type POIType uint16

const (
	POI_TYPE_UNDEFINED = POIType(iota)
	POI_TYPE_HIGHWAY
	POI_TYPE_RAILWAY
	POI_TYPE_AEROWAY
)

var poiTypeStr = []string{"undefined", "highway", "railway", "aeroway"}

func (iotaIdx POIType) String() string {
	return poiTypeStr[iotaIdx]
}
