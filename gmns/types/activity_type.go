package types

// ActivityType is just type alias for the activity type
type ActivityType uint16

const (
	ACTIVITY_NONE = ActivityType(iota)
	ACTIVITY_POI
	ACTIVITY_LINK
)

var activityTypeStr = []string{"none", "poi", "link"}

func (iotaIdx ActivityType) String() string {
	return activityTypeStr[iotaIdx]
}
