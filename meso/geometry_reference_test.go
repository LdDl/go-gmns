package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

type geometryReferenceCase struct {
	name          string
	options       []func(*Link)
	geometryID    gmns.GeometryID
	dirFlag       gmns.DirectionFlag
	hasGeometryID bool
	hasDirFlag    bool
}

func TestGeometryReferencesKeepTravelGeometry(t *testing.T) {
	cases := []geometryReferenceCase{
		{
			name: "absent",
		},
		{
			name: "zero ID and zero flag",
			options: []func(*Link){
				WithGeometryID(gmns.GeometryIDFromInt64(0)),
				WithDirFlag(gmns.DIR_FLAG_NONE),
			},
			geometryID:    gmns.GeometryIDFromInt64(0),
			hasGeometryID: true,
			hasDirFlag:    true,
		},
		{
			name: "reference without orientation",
			options: []func(*Link){
				WithGeometryID(gmns.GeometryIDFromInt64(-1)),
			},
			geometryID:    gmns.GeometryIDFromInt64(-1),
			hasGeometryID: true,
		},
		{
			name: "orientation without reference",
			options: []func(*Link){
				WithDirFlag(gmns.DIR_FLAG_FORWARD),
			},
			dirFlag:    gmns.DIR_FLAG_FORWARD,
			hasDirFlag: true,
		},
		{
			name: "reverse shared geometry",
			options: []func(*Link){
				WithGeometryID(gmns.GeometryIDFromInt64(7)),
				WithDirFlag(gmns.DIR_FLAG_REVERSE),
			},
			geometryID:    gmns.GeometryIDFromInt64(7),
			dirFlag:       gmns.DIR_FLAG_REVERSE,
			hasGeometryID: true,
			hasDirFlag:    true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{
				{37, 55},
				{37.005, 55.002},
				{37.01, 55.01},
			}
			link := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), testCase.options...)
			WithLineGeom(shape.Clone())(link)
			geometryID, hasGeometryID := link.GeometryID()
			dirFlag, hasDirFlag := link.DirFlag()
			assert.Equal(t, testCase.geometryID, geometryID)
			assert.Equal(t, testCase.dirFlag, dirFlag)
			assert.Equal(t, testCase.hasGeometryID, hasGeometryID)
			assert.Equal(t, testCase.hasDirFlag, hasDirFlag)
			assert.Equal(t, shape, link.Geom())
			assert.Equal(t, shape, link.GeoFeature().Geometry, "existing GeoJSON must retain travel-oriented geometry")
			assert.Equal(t, gmns.NodeIDFromInt64(1), link.SourceNode())
			assert.Equal(t, gmns.NodeIDFromInt64(2), link.TargetNode())
		})
	}
}

func TestSharedGeometryReferencesRemainIndependent(t *testing.T) {
	forwardShape := orb.LineString{
		{37, 55},
		{37.005, 55.002},
		{37.01, 55.01},
	}
	backwardShape := forwardShape.Clone()
	backwardShape.Reverse()
	sharedReference := WithGeometryID(gmns.GeometryIDFromInt64(7))
	forward := NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), WithLineGeom(forwardShape), sharedReference, WithDirFlag(gmns.DIR_FLAG_FORWARD))
	backward := NewLinkFrom(gmns.LinkIDFromInt64(11), gmns.NodeIDFromInt64(2), gmns.NodeIDFromInt64(1), WithLineGeom(backwardShape), sharedReference, WithDirFlag(gmns.DIR_FLAG_REVERSE))
	WithGeometryID(gmns.GeometryIDFromInt64(8))(backward)
	WithDirFlag(gmns.DIR_FLAG_NONE)(backward)
	forwardID, hasForwardID := forward.GeometryID()
	backwardID, hasBackwardID := backward.GeometryID()
	forwardFlag, hasForwardFlag := forward.DirFlag()
	backwardFlag, hasBackwardFlag := backward.DirFlag()
	assert.True(t, hasForwardID)
	assert.True(t, hasBackwardID)
	assert.True(t, hasForwardFlag)
	assert.True(t, hasBackwardFlag)
	assert.Equal(t, gmns.GeometryIDFromInt64(7), forwardID)
	assert.Equal(t, gmns.GeometryIDFromInt64(8), backwardID)
	assert.Equal(t, gmns.DIR_FLAG_FORWARD, forwardFlag)
	assert.Equal(t, gmns.DIR_FLAG_NONE, backwardFlag)
	assert.Equal(t, forwardShape, forward.Geom())
	assert.Equal(t, backwardShape, backward.Geom())
}
