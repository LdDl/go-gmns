package macro

type DirectionType uint16

const (
	DIRECTION_FORWARD = DirectionType(iota + 1)
	DIRECTION_BACKWARD
)
