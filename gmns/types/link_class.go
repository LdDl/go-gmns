package types

type LinkClass uint16

const (
	LINK_CLASS_UNDEFINED = LinkClass(iota)
	LINK_CLASS_HIGHWAY
	LINK_CLASS_RAILWAY
	LINK_CLASS_AEROWAY
)

func (iotaIdx LinkClass) String() string {
	return [...]string{"undefined", "highway", "railway", "aeroway"}[iotaIdx]
}
