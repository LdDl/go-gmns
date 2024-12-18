package geomath

import (
	"testing"

	"github.com/paulmach/orb"
)

func BenchmarkFloat64Comparison(b *testing.B) {
	l1 := orb.LineString{
		{37.36039218255323, 55.8467377344096},
		{37.359087239320985, 55.84725055539954},
		{37.35812303126528, 55.84771453046491},
		{37.357245819426225, 55.84814594090787},
		{37.356433854748815, 55.84839420323749},
	}
	l2 := make(orb.LineString, len(l1))
	copy(l2, l1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orb.Equal(l1, l2)
	}
}

func BenchmarkHashComparison(b *testing.B) {
	l1 := orb.LineString{
		{37.36039218255323, 55.8467377344096},
		{37.359087239320985, 55.84725055539954},
		{37.35812303126528, 55.84771453046491},
		{37.357245819426225, 55.84814594090787},
		{37.356433854748815, 55.84839420323749},
	}
	l2 := make(orb.LineString, len(l1))
	copy(l2, l1)
	hash1 := GeometryHash(l1)
	hash2 := GeometryHash(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hash1 == hash2
	}
}
