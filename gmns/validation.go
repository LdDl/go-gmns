package gmns

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
)

func Finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func ValidateWGS84(point orb.Point) error {
	if !Finite(point[0]) || !Finite(point[1]) || point[0] < -180 || point[0] > 180 || point[1] < -90 || point[1] > 90 {
		return fmt.Errorf("invalid WGS84 point %v", point)
	}
	return nil
}

func ValidateLine(shape orb.LineString) error {
	if len(shape) < 2 {
		return fmt.Errorf("line requires at least two points")
	}
	for _, point := range shape {
		err := ValidateWGS84(point)
		if err != nil {
			return err
		}
	}
	return nil
}
