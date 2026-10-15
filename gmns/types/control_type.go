package types

// ControlType stores a supplied control category. An empty value means absent.
type ControlType string

const (
	CONTROL_TYPE_NOT_SIGNAL       ControlType = "common"
	CONTROL_TYPE_IS_SIGNAL        ControlType = "signal"
	CONTROL_TYPE_NO_CONTROL       ControlType = "no_control"
	CONTROL_TYPE_YIELD            ControlType = "yield"
	CONTROL_TYPE_STOP             ControlType = "stop"
	CONTROL_TYPE_STOP_2_WAY       ControlType = "stop_2_way"
	CONTROL_TYPE_STOP_4_WAY       ControlType = "stop_4_way"
	CONTROL_TYPE_SIGNAL_WITH_RTOR ControlType = "signal_with_RTOR"
)

func (controlType ControlType) String() string {
	return string(controlType)
}

// HasControl preserves explicit junctions for supplied control categories.
// Absent, legacy common and explicit no_control values allow topology shortcuts.
// Custom categories retain a junction without interpreting their control rules.
func (controlType ControlType) HasControl() bool {
	return controlType != "" && controlType != CONTROL_TYPE_NOT_SIGNAL && controlType != CONTROL_TYPE_NO_CONTROL
}
