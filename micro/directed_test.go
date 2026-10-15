package micro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type directedCase struct {
	name    string
	options []func(*Link)
	want    bool
}

func TestDirectedGeoJSON(t *testing.T) {
	cases := []directedCase{
		{
			name: "constructor defaults to directed",
			want: true,
		},
		{
			name: "explicit directed",
			options: []func(*Link){
				WithDirected(true),
			},
			want: true,
		},
		{
			name: "explicit undirected",
			options: []func(*Link){
				WithDirected(false),
			},
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
			WithDirFlag(gmns.DIR_FLAG_NONE)(link)

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
			assert.Equal(t, testCase.want, feature.Properties["directed"])
			assert.Equal(t, testCase.want, link.Directed())
			assert.Equal(t, shape, feature.Geometry)
			assert.Equal(t, shape, link.Geom())
			assert.Equal(t, gmns.NodeIDFromInt64(1), link.SourceNode())
			assert.Equal(t, gmns.NodeIDFromInt64(2), link.TargetNode())

		})
	}
}
