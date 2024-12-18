package geomath

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/paulmach/orb"
)

// GeometryHash encodes the given linestring to the hash. It could be good for comparing geometries
func GeometryHash(geom orb.LineString) string {
	h := md5.New()
	for _, point := range geom {
		h.Write([]byte(fmt.Sprintf("%f,%f;", point[0], point[1])))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func round(x float64) int {
	if x < 0 {
		return int(-math.Floor(-x + 0.5))
	}
	return int(math.Floor(x + 0.5))
}

type Codec struct {
	Dim   int
	Scale float64
}

var defaultCodec = Codec{Dim: 2, Scale: 1e5}

func EncodeUint(buf []byte, u uint) []byte {
	for u >= 32 {
		buf = append(buf, byte((u&31)+95))
		u >>= 5
	}
	buf = append(buf, byte(u+63))
	return buf
}

func EncodeInt(buf []byte, i int) []byte {
	var u uint
	if i < 0 {
		u = uint(^(i << 1))
	} else {
		u = uint(i << 1)
	}
	return EncodeUint(buf, u)
}

func (c Codec) EncodeCoord(buf []byte, coord []float64) []byte {
	for _, x := range coord {
		buf = EncodeInt(buf, round(c.Scale*x))
	}
	return buf
}

func (c Codec) EncodeCoords(buf []byte, coords [][]float64) []byte {
	last := make([]int, c.Dim)
	for _, coord := range coords {
		for i, x := range coord {
			ex := round(c.Scale * x)
			buf = EncodeInt(buf, ex-last[i])
			last[i] = ex
		}
	}
	return buf
}

func EncodeCoord(coord []float64) []byte {
	return defaultCodec.EncodeCoord(nil, coord)
}

func EncodeCoords(coords [][]float64) []byte {
	return defaultCodec.EncodeCoords(nil, coords)
}
