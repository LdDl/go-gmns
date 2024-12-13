package macro

import (
	"sort"
)

const (
	resolution = 5.0
	// Default lane width
	LANE_WIDTH = 3.5
)

// LanesInfo is information about lanes for the specific link
type LanesInfo struct {
	LanesList         []int
	LanesChange       [][2]int
	LanesChangePoints []float64
}

// NewLanesInfo creates pointer to the new LanesInfo
func NewLanesInfo(link *Link) LanesInfo {
	lanesInfo := LanesInfo{
		LanesList:         make([]int, 0),
		LanesChange:       make([][2]int, 0),
		LanesChangePoints: make([]float64, 0),
	}
	lengthMeters := link.LengthMeters()
	if lengthMeters < resolution {
		lanesInfo.LanesChangePoints = []float64{0.0, lengthMeters}
	} else {
		lanesChangePointsTemp := []float64{0.0, lengthMeters}
		for len(lanesChangePointsTemp) != 0 {
			target := lanesChangePointsTemp[0]
			remove := make(map[int]struct{})
			for idx, point := range lanesChangePointsTemp {
				if target-resolution <= point && point <= target+resolution {
					remove[idx] = struct{}{}
				}
			}
			lanesInfo.LanesChangePoints = append(lanesInfo.LanesChangePoints, target)
			for idx := range remove {
				lanesChangePointsTemp = append(lanesChangePointsTemp[:idx], lanesChangePointsTemp[idx+1:]...)
			}
		}
		sort.Float64s(lanesInfo.LanesChangePoints)
	}
	for i := 0; i < len(lanesInfo.LanesChangePoints)-1; i++ {
		lanesInfo.LanesList = append(lanesInfo.LanesList, link.LanesNum())
		lanesInfo.LanesChange = append(lanesInfo.LanesChange, [2]int{0.0, 0.0})
	}
	return lanesInfo
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
