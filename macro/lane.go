package macro

import (
	"fmt"
	"math"
)

const (
	// Default lane width
	LANE_WIDTH = 3.5
)

// LanesInfo is information about lanes for the specific link
type LanesInfo struct {
	LanesList         []int
	LanesChange       [][2]int
	LanesChangePoints []float64
}

// NewLanesInfo prepares uniform lane information from a supplied positive finite length.
func NewLanesInfo(link *Link) (LanesInfo, error) {
	if link == nil {
		return LanesInfo{}, fmt.Errorf("cannot prepare lane information: link is nil")
	}
	lengthMeters, hasLength := link.LengthMeters()
	if !hasLength {
		return LanesInfo{}, fmt.Errorf("macro link %v: length is required to prepare lane information", link.ID)
	}
	if lengthMeters <= 0 || math.IsNaN(lengthMeters) || math.IsInf(lengthMeters, 0) {
		return LanesInfo{}, fmt.Errorf("macro link %v: lane preparation requires a positive finite length, got %g", link.ID, lengthMeters)
	}
	lanesNum, hasLanes := link.LanesNum()
	if !hasLanes {
		return LanesInfo{}, fmt.Errorf("macro link %v: lane count is required to prepare lane information", link.ID)
	}
	if lanesNum < 0 {
		return LanesInfo{}, fmt.Errorf("macro link %v: lane count must be nonnegative, got %v", link.ID, lanesNum)
	}
	lanesInfo := LanesInfo{
		LanesList: []int{lanesNum},
		LanesChange: [][2]int{
			{0, 0},
		},
		LanesChangePoints: []float64{0, lengthMeters},
	}
	return lanesInfo, nil
}

func laneIndices(lanes int, lanesChangeLeft int, lanesChangeRight int) []int {
	if lanes < lanesChangeLeft || lanes < lanesChangeRight {
		return make([]int, 0)
	}
	laneIndices := make([]int, lanes)
	for i := 1; i <= lanes; i++ {
		laneIndices[i-1] = i
	}
	if lanesChangeLeft < 0 {
		laneIndices = laneIndices[-lanesChangeLeft:]
	} else if lanesChangeLeft > 0 {
		left := make([]int, lanesChangeLeft)
		for i := range left {
			left[i] = -lanesChangeLeft + i
		}
		laneIndices = append(left, laneIndices...)
	}
	if lanesChangeRight < 0 {
		laneIndices = laneIndices[:lanes+lanesChangeRight]
	} else if lanesChangeRight > 0 {
		right := make([]int, lanesChangeRight)
		for i := range right {
			right[i] = lanes + 1 + i
		}
		laneIndices = append(laneIndices, right...)
	}
	return laneIndices
}
