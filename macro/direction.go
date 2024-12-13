package macro

// DirectionType is just type alias for the direction of the geometry
type DirectionType uint16

const (
	DIRECTION_FORWARD = DirectionType(iota + 1)
	DIRECTION_BACKWARD
)

var directionTypeStr = []string{"forward", "backward"}

func (iotaIdx DirectionType) String() string {
	return directionTypeStr[iotaIdx-1]
}
