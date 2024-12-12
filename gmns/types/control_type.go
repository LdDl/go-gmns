package types

type ControlType uint16

const (
	CONTROL_TYPE_NOT_SIGNAL = ControlType(iota)
	CONTROL_TYPE_IS_SIGNAL
)

func (iotaIdx ControlType) String() string {
	return [...]string{"common", "signal"}[iotaIdx]
}
