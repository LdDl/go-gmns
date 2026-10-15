package macro

import (
	"math"
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/stretchr/testify/assert"
)

type laneLengthCase struct {
	name    string
	value   float64
	present bool
	valid   bool
}

func TestNewLanesInfoRequiresUsableLength(t *testing.T) {
	cases := []laneLengthCase{
		{
			name: "absent",
		},
		{
			name:    "zero",
			present: true,
		},
		{
			name:    "negative",
			value:   -1,
			present: true,
		},
		{
			name:    "NaN",
			value:   math.NaN(),
			present: true,
		},
		{
			name:    "positive infinity",
			value:   math.Inf(1),
			present: true,
		},
		{
			name:    "negative infinity",
			value:   math.Inf(-1),
			present: true,
		},
		{
			name:    "short link",
			value:   4.5,
			present: true,
			valid:   true,
		},
		{
			name:    "exactly five meters",
			value:   5,
			present: true,
			valid:   true,
		},
		{
			name:    "long link",
			value:   123.125,
			present: true,
			valid:   true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), WithLanesNum(2))
			if testCase.present {
				WithLengthMeters(testCase.value)(link)
			}
			laneInfo, err := NewLanesInfo(link)
			if testCase.valid {
				assert.NoError(t, err)
				assert.Equal(t, []float64{0, testCase.value}, laneInfo.LanesChangePoints)
				assert.Equal(t, []int{2}, laneInfo.LanesList)
				assert.Equal(t, [][2]int{{0, 0}}, laneInfo.LanesChange)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), "macro link 10")
					assert.Contains(t, err.Error(), "length")
				}
				assert.Equal(t, LanesInfo{}, laneInfo)
			}
			length, present := link.LengthMeters()
			assert.Equal(t, testCase.present, present)
			if math.IsNaN(testCase.value) {
				assert.True(t, math.IsNaN(length))
			} else {
				assert.Equal(t, testCase.value, length)
			}
			assert.Equal(t, LanesInfo{}, link.LanesInfo(), "preparation must not mutate source lane data")
		})
	}
}

func TestNewLanesInfoRejectsNilLink(t *testing.T) {
	laneInfo, err := NewLanesInfo(nil)
	assert.Error(t, err)
	assert.Equal(t, LanesInfo{}, laneInfo)
}

type preparedLaneCountCase struct {
	name    string
	value   int
	present bool
	valid   bool
}

func TestNewLanesInfoRequiresLaneCount(t *testing.T) {
	cases := []preparedLaneCountCase{
		{
			name: "absent",
		},
		{
			name:    "negative",
			value:   -1,
			present: true,
		},
		{
			name:    "zero motor lanes",
			present: true,
			valid:   true,
		},
		{
			name:    "two motor lanes",
			value:   2,
			present: true,
			valid:   true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), WithLengthMeters(50))
			if testCase.present {
				WithLanesNum(testCase.value)(link)
			}
			laneInfo, err := NewLanesInfo(link)
			if testCase.valid {
				assert.NoError(t, err)
				assert.Equal(t, []int{testCase.value}, laneInfo.LanesList)
				assert.Equal(t, []float64{0, 50}, laneInfo.LanesChangePoints)
				assert.Equal(t, [][2]int{{0, 0}}, laneInfo.LanesChange)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), "macro link 10")
					assert.Contains(t, err.Error(), "lane count")
				}
				assert.Equal(t, LanesInfo{}, laneInfo)
			}
			value, present := link.LanesNum()
			assert.Equal(t, testCase.value, value)
			assert.Equal(t, testCase.present, present)
			assert.Equal(t, LanesInfo{}, link.LanesInfo())
		})
	}
}
