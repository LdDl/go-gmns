package micro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type linkLaneCountCase struct {
	name    string
	options []func(*Link)
	value   int
	present bool
}

func TestLinkLaneCountGeoJSON(t *testing.T) {
	cases := []linkLaneCountCase{
		{
			name: "absent",
		},
		{
			name: "explicit zero",
			options: []func(*Link){
				WithLanesNum(0),
			},
			present: true,
		},
		{
			name: "negative value is not absence",
			options: []func(*Link){
				WithLanesNum(-1),
			},
			value:   -1,
			present: true,
		},
		{
			name: "positive count",
			options: []func(*Link){
				WithLanesNum(2),
			},
			value:   2,
			present: true,
		},
		{
			name: "replace with zero",
			options: []func(*Link){
				WithLanesNum(2),
				WithLanesNum(0),
			},
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), testCase.options...)
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
			value, exists := feature.Properties["lanes"]
			assert.True(t, exists)
			if testCase.present {
				assert.Equal(t, float64(testCase.value), value)
			} else {
				assert.Nil(t, value)
			}
			lanesNum, present := link.LanesNum()
			assert.Equal(t, testCase.value, lanesNum)
			assert.Equal(t, testCase.present, present)
		})
	}
}

func TestLaneCountIsIndependentOfCellIndex(t *testing.T) {
	link := NewLinkFrom(gmns.LinkIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.NodeIDFromInt64(3), WithLaneID(-2))
	_, present := link.LanesNum()
	assert.False(t, present)
	WithLanesNum(3)(link)
	assert.Equal(t, -2, link.LaneID())
	WithoutLanesNum()(link)
	_, present = link.LanesNum()
	assert.False(t, present)
	assert.Equal(t, -2, link.LaneID())
}
