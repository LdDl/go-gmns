package types

type ActivityType uint16

const (
	ACTIVITY_NONE = ActivityType(iota)
	ACTIVITY_POI
	ACTIVITY_LINK
)

func (iotaIdx ActivityType) String() string {
	return [...]string{"none", "poi", "link"}[iotaIdx]
}
