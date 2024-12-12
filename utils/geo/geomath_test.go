package geomath

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/stretchr/testify/assert"
)

func lineAsString(l orb.LineString) string {
	agg := []string{}
	for _, pt := range l {
		agg = append(agg, fmt.Sprintf("[%f, %f]", pt.X(), pt.Y()))
	}
	return "[" + strings.Join(agg, ",") + "]"
}

func TestOffset(t *testing.T) {
	line := orb.LineString{{10.0, 10.0}, {15.0, 10.0}, {18.0, 15.0}, {18.0, 20.0}, {15.0, 24.0}, {12.0, 24.0}, {10.0, 18.0}, {10.0, 15.0}, {13.0, 12.0}, {15.0, 16.0}}
	distance := 1.0

	leftL := lineAsString(OffsetCurve(line, distance))
	rightL := lineAsString(OffsetCurve(line, -distance))

	correctLeft := "[[10.000000, 11.000000],[14.433810, 11.000000],[17.000000, 15.276984],[17.000000, 19.666667],[14.500000, 23.000000],[12.720759, 23.000000],[11.000000, 17.837722],[11.000000, 15.414214],[12.726049, 13.688165],[14.105573, 16.447214]]"
	if leftL != correctLeft {
		t.Errorf("Left offset line should be '%s' but got '%s'", correctLeft, leftL)
	}
	correctRight := "[[10.000000, 9.000000],[15.566190, 9.000000],[19.000000, 14.723016],[19.000000, 20.333333],[15.500000, 25.000000],[11.279241, 25.000000],[9.000000, 18.162278],[9.000000, 14.585786],[13.273951, 10.311835],[15.894427, 15.552786]]"
	if rightL != correctRight {
		t.Errorf("Right offset line should be '%s' but got '%s'", correctRight, rightL)
	}
}

func TestLineSubstring(t *testing.T) {
	lineWKT := "LINESTRING (37.56319128200903 55.78357465483572, 37.565235359279626 55.78497472894253, 37.565822487858156 55.785421030200496, 37.567355545810614 55.784711836767826)"
	line, err := wkt.UnmarshalLineString(lineWKT)
	if err != nil {
		t.Error(err)
		return
	}
	newline := SubstringHaversine(line, 215, 278)
	newLineWKT := wkt.MarshalString(newline)
	correctLine := "LINESTRING(37.56536219999623 55.78507114703719,37.565822487858156 55.785421030200496,37.56600203415945 55.785337974305975)"
	if correctLine != newLineWKT {
		t.Errorf("Correct line should be '%s', but got '%s'", correctLine, newLineWKT)
	}
}

func TestEPSGConverter(t *testing.T) {
	precision := 10e-5

	// Test point
	givenPoint4326 := orb.Point{37.61655751319856, 55.75163877328629}
	expectedPoint3857 := orb.Point{4187456.027182254, 7509131.996742569}
	ansPoint3857 := PointToEuclidean(givenPoint4326)
	assert.InDelta(t, expectedPoint3857[0], ansPoint3857[0], precision, "Wrong X (longitude) in EPSG:3857")
	assert.InDelta(t, expectedPoint3857[1], ansPoint3857[1], precision, "Wrong Y (latitude) in EPSG:3857")
	ansReversedPoint4326 := PointToSpherical(ansPoint3857)
	assert.InDelta(t, givenPoint4326[0], ansReversedPoint4326[0], precision, "Wrong longitude (X) in EPSG:4326")
	assert.InDelta(t, givenPoint4326[1], ansReversedPoint4326[1], precision, "Wrong latitude (Y) in EPSG:4326")

	// Test line
	givenLine4326 := orb.LineString([]orb.Point{{37.61655751319856, 55.75163877328629}, {37.61617406590727, 55.751456041561624}})
	expectedLine3857 := orb.LineString([]orb.Point{{4187456.027182254, 7509131.996742569}, {4187413.342025048, 7509095.852052931}})
	ansLine3857 := LineToEuclidean(givenLine4326)
	assert.Equal(t, len(expectedLine3857), len(ansLine3857), "Incorrect number of points for line in EPSG:3857")
	for i := range ansLine3857 {
		pt := ansLine3857[i]
		assert.InDelta(t, expectedLine3857[i][0], pt[0], precision, fmt.Sprintf("Wrong X (longitude) in EPSG:3857 at pos #%d", i))
		assert.InDelta(t, expectedLine3857[i][1], pt[1], precision, fmt.Sprintf("Wrong Y (latitude) in EPSG:3857 at pos #%d", i))
	}
	ansReversedLine4326 := LineToSpherical(ansLine3857)
	assert.Equal(t, len(givenLine4326), len(ansReversedLine4326), "Incorrect number of points for line in EPSG:4326")
	for i := range ansReversedLine4326 {
		pt := ansReversedLine4326[i]
		assert.InDelta(t, givenLine4326[i][0], pt[0], precision, fmt.Sprintf("Wrong longitude (X) in EPSG:4326 at pos #%d", i))
		assert.InDelta(t, givenLine4326[i][1], pt[1], precision, fmt.Sprintf("Wrong latitude (Y) in EPSG:4326 at pos #%d", i))
	}
}
