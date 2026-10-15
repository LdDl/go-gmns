package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type controlTypeCase struct {
	name       string
	value      ControlType
	hasControl bool
}

func TestControlTypes(t *testing.T) {
	cases := []controlTypeCase{
		{
			name: "",
		},
		{
			name:       "custom control",
			value:      ControlType("custom control"),
			hasControl: true,
		},
		{
			name:  "common",
			value: CONTROL_TYPE_NOT_SIGNAL,
		},
		{
			name:       "signal",
			value:      CONTROL_TYPE_IS_SIGNAL,
			hasControl: true,
		},
		{
			name:  "no_control",
			value: CONTROL_TYPE_NO_CONTROL,
		},
		{
			name:       "yield",
			value:      CONTROL_TYPE_YIELD,
			hasControl: true,
		},
		{
			name:       "stop",
			value:      CONTROL_TYPE_STOP,
			hasControl: true,
		},
		{
			name:       "stop_2_way",
			value:      CONTROL_TYPE_STOP_2_WAY,
			hasControl: true,
		},
		{
			name:       "stop_4_way",
			value:      CONTROL_TYPE_STOP_4_WAY,
			hasControl: true,
		},
		{
			name:       "signal_with_RTOR",
			value:      CONTROL_TYPE_SIGNAL_WITH_RTOR,
			hasControl: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.name, testCase.value.String())
			assert.Equal(t, testCase.hasControl, testCase.value.HasControl())
		})
	}
}
