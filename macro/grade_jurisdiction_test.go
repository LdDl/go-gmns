package macro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type gradeJurisdictionCase struct {
	name         string
	options      []func(*Link)
	grade        float64
	hasGrade     bool
	jurisdiction string
}

func TestGradeAndJurisdictionGeoJSON(t *testing.T) {
	cases := []gradeJurisdictionCase{
		{
			name: "absent",
		},
		{
			name: "explicit level grade",
			options: []func(*Link){
				WithGrade(0),
			},
			hasGrade: true,
		},
		{
			name: "negative one is a downhill grade",
			options: []func(*Link){
				WithGrade(-1),
			},
			grade:    -1,
			hasGrade: true,
		},
		{
			name: "fractional uphill with operator",
			options: []func(*Link){
				WithGrade(7.25),
				WithJurisdiction("City roads department"),
			},
			grade:        7.25,
			hasGrade:     true,
			jurisdiction: "City roads department",
		},
		{
			name: "fractional downhill",
			options: []func(*Link){
				WithGrade(-7.25),
			},
			grade:    -7.25,
			hasGrade: true,
		},
		{
			name: "operator without grade",
			options: []func(*Link){
				WithJurisdiction("  Город; \"Дороги\"  "),
			},
			jurisdiction: "  Город; \"Дороги\"  ",
		},
		{
			name: "clearing operator preserves grade",
			options: []func(*Link){
				WithGrade(2.5),
				WithJurisdiction("City roads department"),
				WithJurisdiction(""),
			},
			grade:    2.5,
			hasGrade: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{
				{37, 55},
				{37.01, 55.01},
			}
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), testCase.options...)
			WithLineGeom(shape.Clone())(link)
			WithDirFlag(gmns.DIR_FLAG_REVERSE)(link)
			data, err := link.GeoFeature().MarshalJSON()
			assert.NoError(t, err)
			if err != nil {
				return
			}
			feature, err := geojson.UnmarshalFeature(data)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			value, exists := feature.Properties["grade"]
			assert.True(t, exists)
			if testCase.hasGrade {
				assert.Equal(t, testCase.grade, value, "grade stays in percent and is not flipped by shape orientation")
			} else {
				assert.Nil(t, value, "grade must not be inferred from the two-dimensional line")
			}
			assert.Equal(t, testCase.jurisdiction, feature.Properties["jurisdiction"])
			grade, hasGrade := link.Grade()
			assert.Equal(t, testCase.grade, grade)
			assert.Equal(t, testCase.hasGrade, hasGrade)
			assert.Equal(t, testCase.jurisdiction, link.Jurisdiction())
			assert.Equal(t, shape, link.Geom())
			assert.Equal(t, shape, feature.Geometry)
			assert.Equal(t, gmns.NodeIDFromInt64(1), link.SourceNode())
			assert.Equal(t, gmns.NodeIDFromInt64(2), link.TargetNode())
		})
	}
}
