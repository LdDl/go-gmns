package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type controlTypeCase struct {
	name       string
	value      ControlType
	numeric    uint16
	hasControl bool
}

func TestControlTypes(t *testing.T) {
	cases := []controlTypeCase{
		{
			name:    "common",
			value:   CONTROL_TYPE_NOT_SIGNAL,
			numeric: 0,
		},
		{
			name:       "signal",
			value:      CONTROL_TYPE_IS_SIGNAL,
			numeric:    1,
			hasControl: true,
		},
		{
			name:    "no_control",
			value:   CONTROL_TYPE_NO_CONTROL,
			numeric: 2,
		},
		{
			name:       "yield",
			value:      CONTROL_TYPE_YIELD,
			numeric:    3,
			hasControl: true,
		},
		{
			name:       "stop",
			value:      CONTROL_TYPE_STOP,
			numeric:    4,
			hasControl: true,
		},
		{
			name:       "stop_2_way",
			value:      CONTROL_TYPE_STOP_2_WAY,
			numeric:    5,
			hasControl: true,
		},
		{
			name:       "stop_4_way",
			value:      CONTROL_TYPE_STOP_4_WAY,
			numeric:    6,
			hasControl: true,
		},
		{
			name:       "signal_with_RTOR",
			value:      CONTROL_TYPE_SIGNAL_WITH_RTOR,
			numeric:    7,
			hasControl: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.numeric, uint16(testCase.value))
			assert.Equal(t, testCase.name, testCase.value.String())
			assert.Equal(t, testCase.hasControl, testCase.value.HasControl())
		})
	}
}
