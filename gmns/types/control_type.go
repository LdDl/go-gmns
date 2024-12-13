package types

// ControlType is just type alias for the control type
type ControlType uint16

const (
	CONTROL_TYPE_NOT_SIGNAL = ControlType(iota)
	CONTROL_TYPE_IS_SIGNAL
)

var controlTypeStr = []string{"common", "signal"}

func (iotaIdx ControlType) String() string {
	return controlTypeStr[iotaIdx]
}
