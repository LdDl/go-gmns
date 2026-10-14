package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type altitudeCase struct {
	name    string
	options []func(*Node)
	value   float64
	present bool
}

func TestAltitudeGeoJSON(t *testing.T) {
	cases := []altitudeCase{
		{
			name: "absent",
		},
		{
			name: "explicit zero",
			options: []func(*Node){
				WithAltitudeMeters(0),
			},
			present: true,
		},
		{
			name: "negative one is an altitude",
			options: []func(*Node){
				WithAltitudeMeters(-1),
			},
			value:   -1,
			present: true,
		},
		{
			name: "fractional altitude",
			options: []func(*Node){
				WithAltitudeMeters(123.123456789),
			},
			value:   123.123456789,
			present: true,
		},
		{
			name: "replace altitude with below zero value",
			options: []func(*Node){
				WithAltitudeMeters(100),
				WithAltitudeMeters(-12.75),
			},
			value:   -12.75,
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			point := orb.Point{37, 55}
			node := NewNodeFrom(1, testCase.options...)
			WithPointGeom(point)(node)

			data, err := node.GeoFeature().MarshalJSON()
			assert.NoError(t, err)
			if err != nil {
				return
			}
			feature, err := geojson.UnmarshalFeature(data)
			assert.NoError(t, err)
			if err != nil {
				return
			}

			value, exists := feature.Properties["z_coord"]
			assert.True(t, exists)
			if testCase.present {
				assert.Equal(t, testCase.value, value)
			} else {
				assert.Nil(t, value, "two-dimensional coordinates do not imply an altitude")
			}
			altitude, present := node.AltitudeMeters()
			assert.Equal(t, testCase.value, altitude)
			assert.Equal(t, testCase.present, present)
			assert.Equal(t, point, node.Geom())
			assert.Equal(t, point, feature.Geometry, "altitude stays a property; the point remains two-dimensional")
		})
	}
}

type rightOfWayWidthCase struct {
	name    string
	options []func(*Link)
	value   float64
	present bool
}

func TestRightOfWayWidthGeoJSON(t *testing.T) {
	cases := []rightOfWayWidthCase{
		{
			name: "absent",
		},
		{
			name: "explicit zero",
			options: []func(*Link){
				WithRightOfWayWidthMeters(0),
			},
			present: true,
		},
		{
			name: "fractional whole width",
			options: []func(*Link){
				WithRightOfWayWidthMeters(24.123456789),
			},
			value:   24.123456789,
			present: true,
		},
		{
			name: "replace width with zero",
			options: []func(*Link){
				WithRightOfWayWidthMeters(24.5),
				WithRightOfWayWidthMeters(0),
			},
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{
				{37, 55},
				{37.01, 55.01},
			}
			link := NewLinkFrom(10, 1, 2)
			WithLanesNum(2)(link)
			WithLengthMeters(100)(link)
			WithLineGeom(shape.Clone())(link)
			WithDirFlag(gmns.DIR_FLAG_REVERSE)(link)
			for _, option := range testCase.options {
				option(link)
			}

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

			value, exists := feature.Properties["row_width"]
			assert.True(t, exists)
			if testCase.present {
				assert.Equal(t, testCase.value, value, "whole width is not divided by lanes or negated by orientation")
			} else {
				assert.Nil(t, value, "lane data and geometry do not imply a right-of-way width")
			}
			width, present := link.RightOfWayWidthMeters()
			assert.Equal(t, testCase.value, width)
			assert.Equal(t, testCase.present, present)
			lengthMeters, hasLength := link.LengthMeters()
			assert.True(t, hasLength)
			assert.Equal(t, 100.0, lengthMeters)
			lanesNum, hasLanes := link.LanesNum()
			assert.True(t, hasLanes)
			assert.Equal(t, 2, lanesNum)
			assert.Equal(t, shape, link.Geom())
			assert.Equal(t, shape, feature.Geometry)
			assert.Equal(t, gmns.NodeID(1), link.SourceNode())
			assert.Equal(t, gmns.NodeID(2), link.TargetNode())
		})
	}
}
