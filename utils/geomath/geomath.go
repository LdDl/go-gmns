package geomath

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

const (
	d2r    = math.Pi / 180.0
	earthR = 20037508.34
)

func deg2rad(deg float64) float64 {
	return deg * d2r
}

func epsg3857To4326(lat, lng float64) (float64, float64) {
	newLat := lat * 180 / earthR
	newLong := math.Atan(math.Exp(lng*math.Pi/earthR))*360/math.Pi - 90
	return newLat, newLong
}

func epsg4326To3857(lon, lat float64) (float64, float64) {
	x := lon * earthR / 180
	y := math.Log(math.Tan((90+lat)*math.Pi/360)) / (math.Pi / 180)
	y = y * earthR / 180
	return x, y
}

func PointToEuclidean(pt orb.Point) orb.Point {
	euclideanX, euclideanY := epsg4326To3857(pt.Lon(), pt.Lat())
	return orb.Point{euclideanX, euclideanY}
}

func LineToEuclidean(line orb.LineString) orb.LineString {
	newLine := make(orb.LineString, len(line))
	for i, pt := range line {
		newLine[i] = PointToEuclidean(pt)
	}
	return newLine
}

func PointToSpherical(pt orb.Point) orb.Point {
	sphericalX, sphericalY := epsg3857To4326(pt.X(), pt.Y())
	return orb.Point{sphericalX, sphericalY}
}

func LineToSpherical(line orb.LineString) orb.LineString {
	newLine := make(orb.LineString, len(line))
	for i, pt := range line {
		newLine[i] = PointToSpherical(pt)
	}
	return newLine
}

// AngleBetweenLines returs angle between two lines
//
// Note: panics if number of points in any line is less than 2
func AngleBetweenLines(l1 orb.LineString, l2 orb.LineString) float64 {
	angle1 := math.Atan2(l1[len(l1)-1].Y()-l1[0].Y(), l1[len(l1)-1].X()-l1[0].X())
	angle2 := math.Atan2(l2[len(l2)-1].Y()-l2[0].Y(), l2[len(l2)-1].X()-l2[0].X())
	angle := angle2 - angle1
	if angle < -1*math.Pi {
		angle += 2 * math.Pi
	}
	if angle > math.Pi {
		angle -= 2 * math.Pi
	}
	return angle
}

// Returns a line segment between specified distances along the given line
// using DistanceHaversine for more accurate results
// @TODO: Handle edge-cases such as:
// 1. negative values for distances
// 2. startDist > endDist
// 3. startDist > totalLengthMeters
// 4. endDist > totalLengthMeters
func SubstringHaversine(line orb.LineString, startDist float64, endDist float64) orb.LineString {
	var substring orb.LineString
	totalLengthMeters := 0.0
	for i := 1; i < len(line); i++ {
		segmentStart := line[i-1]
		segmentEnd := line[i]
		segmentLengthMeters := geo.DistanceHaversine(segmentStart, segmentEnd)
		totalLengthMeters += segmentLengthMeters
		if totalLengthMeters >= startDist {
			substring = append(substring, segmentStart)
			if totalLengthMeters >= endDist {
				substring = append(substring, segmentEnd)
				break
			}
		}
	}
	startCut, _ := geo.PointAtDistanceAlongLine(line, startDist)
	endCut, _ := geo.PointAtDistanceAlongLine(line, endDist)
	substring[0] = startCut
	substring[len(substring)-1] = endCut
	return substring
}

// Returns a line segment between specified distances along the given line
// using simple Euclidean distance function
// @TODO: Handle edge-cases such as:
// 1. negative values for distances
// 2. startDist > endDist
// 3. startDist > totalLengthMeters
// 4. endDist > totalLengthMeters
func Substring(line orb.LineString, startDist float64, endDist float64) orb.LineString {
	var substring orb.LineString
	totalLengthMeters := 0.0
	for i := 1; i < len(line); i++ {
		segmentStart := line[i-1]
		segmentEnd := line[i]
		segmentLengthMeters := geo.Distance(segmentStart, segmentEnd)
		totalLengthMeters += segmentLengthMeters
		if totalLengthMeters >= startDist {
			substring = append(substring, segmentStart)
			if totalLengthMeters >= endDist {
				substring = append(substring, segmentEnd)
				break
			}
		}
	}
	startCut, _ := geo.PointAtDistanceAlongLine(line, startDist)
	endCut, _ := geo.PointAtDistanceAlongLine(line, endDist)
	substring[0] = startCut
	substring[len(substring)-1] = endCut
	return substring
}

// Checks if two segments intersects and returns intersections Point
// p1, p2 - first segment
// p3, p4 - second segment
// Note: Euclidean space
func intersection(p1, p2, p3, p4 orb.Point) (orb.Point, error) {
	// Calculate the coefficients of the linear equations
	a1 := p2[1] - p1[1]
	b1 := p1[0] - p2[0]
	c1 := a1*p1[0] + b1*p1[1]
	a2 := p4[1] - p3[1]
	b2 := p3[0] - p4[0]
	c2 := a2*p3[0] + b2*p3[1]

	// Calculate the determinant
	det := a1*b2 - a2*b1
	if det == 0 {
		return orb.Point{}, fmt.Errorf("The lines are parallel")
	}

	// Calculate the intersection point
	x := (b2*c1 - b1*c2) / det
	y := (a1*c2 - a2*c1) / det
	return orb.Point{x, y}, nil
}

func OffsetCurve(line orb.LineString, distance float64) orb.LineString {
	// Initialize result list and segment list
	var result orb.LineString
	var segments [][2]orb.Point

	// Iterate over line segments and calculate offset segments
	for i := 1; i < len(line); i++ {
		// Get current and previous points
		p1 := line[i-1]
		p2 := line[i]

		// Calculate the vector between the points
		vec := [2]float64{p2[0] - p1[0], p2[1] - p1[1]}

		// Normalize the vector
		vecLen := math.Sqrt(vec[0]*vec[0] + vec[1]*vec[1])
		vec = [2]float64{vec[0] / vecLen, vec[1] / vecLen}

		// Rotate the vector by 90 degrees
		rotated := [2]float64{-vec[1], vec[0]}

		// Scale the rotated vector by the distance
		offset := [2]float64{rotated[0] * distance, rotated[1] * distance}

		// Calculate the offset points
		op1 := [2]float64{p1[0] + offset[0], p1[1] + offset[1]}
		op2 := [2]float64{p2[0] + offset[0], p2[1] + offset[1]}

		// Add the offset segment to the list of segments
		segments = append(segments, [2]orb.Point{op1, op2})
	}

	result = append(result, segments[0][0])
	// Iterate over the segments and calculate the intersections
	for i := 1; i < len(segments); i++ {
		// Get the current and previous segments
		seg1 := segments[i-1]
		seg2 := segments[i]
		// Calculate the intersection point
		intersection, err := intersection(seg1[0], seg1[1], seg2[0], seg2[1])
		if err != nil {
			continue
		}
		// If there is an intersection, add the intersection and the current segment to the result
		result = append(result, intersection)
	}
	result = append(result, segments[len(segments)-1][1])
	return result
}
