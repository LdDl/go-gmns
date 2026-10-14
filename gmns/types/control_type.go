package types

// ControlType identifies intersection or movement control.
type ControlType uint16

const (
	CONTROL_TYPE_NOT_SIGNAL = ControlType(iota)
	CONTROL_TYPE_IS_SIGNAL
	CONTROL_TYPE_NO_CONTROL
	CONTROL_TYPE_YIELD
	CONTROL_TYPE_STOP
	CONTROL_TYPE_STOP_2_WAY
	CONTROL_TYPE_STOP_4_WAY
	CONTROL_TYPE_SIGNAL_WITH_RTOR
)

var controlTypeStr = []string{
	"common",
	"signal",
	"no_control",
	"yield",
	"stop",
	"stop_2_way",
	"stop_4_way",
	"signal_with_RTOR",
}

func (controlType ControlType) String() string {
	return controlTypeStr[controlType]
}

// HasControl reports whether the category requires preserving an explicit junction.
// The legacy common category and explicit no_control allow topology shortcuts.
func (controlType ControlType) HasControl() bool {
	return controlType != CONTROL_TYPE_NOT_SIGNAL && controlType != CONTROL_TYPE_NO_CONTROL
}
