package gmns_test

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/micro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/stretchr/testify/assert"
)

type wktEncoder interface {
	WKT() string
}

type wktRoundTripCase struct {
	name     string
	record   wktEncoder
	expected orb.Geometry
}

func TestWKTPreservesWGS84Geometry(t *testing.T) {
	point := orb.Point{37.36039218255323, 55.8467377344096}
	shape := orb.LineString{
		point,
		{37.359087239320985, 55.84725055539954},
	}
	originalShape := shape.Clone()
	projectedPoint := orb.Point{4158973, 7518424}
	projectedShape := orb.LineString{
		projectedPoint,
		{4158827, 7518526},
	}
	nodeID := gmns.NodeIDFromString("001")
	targetID := gmns.NodeIDFromString("002")
	linkID := gmns.LinkIDFromString("road")

	macroNode := macro.NewNodeFrom(nodeID,
		macro.WithPointGeom(point),
		macro.WithPointGeomEuclidean(projectedPoint),
		macro.WithAltitudeMeters(123),
	)
	mesoNode := meso.NewNodeFrom(nodeID,
		meso.WithPointGeom(point),
		meso.WithPointEuclideanGeom(projectedPoint),
	)
	microNode := micro.NewNodeFrom(nodeID,
		micro.WithPointGeom(point),
		micro.WithPointGeomEuclidean(projectedPoint),
	)
	macroLink := macro.NewLinkFrom(linkID, nodeID, targetID,
		macro.WithLineGeom(shape),
		macro.WithLineGeomEuclidean(projectedShape),
		macro.WithDirFlag(gmns.DIR_FLAG_REVERSE),
	)
	mesoLink := meso.NewLinkFrom(linkID, nodeID, targetID,
		meso.WithLineGeom(shape),
		meso.WithLineGeomEuclidean(projectedShape),
		meso.WithDirFlag(gmns.DIR_FLAG_REVERSE),
	)
	microLink := micro.NewLinkFrom(linkID, nodeID, targetID,
		micro.WithLineGeom(shape),
		micro.WithLineGeomEuclidean(projectedShape),
		micro.WithDirFlag(gmns.DIR_FLAG_REVERSE),
	)
	maneuver := movement.NewMovement(gmns.MovementIDFromString("turn"), nodeID, linkID, linkID, "", "",
		movement.WithGeom(shape),
		movement.WithGeomEuclidean(projectedShape),
	)
	geometry := gmns.NewGeometry(gmns.GeometryIDFromString("shape"), shape)

	cases := []wktRoundTripCase{
		{
			name:     "macro node",
			record:   macroNode,
			expected: point,
		},
		{
			name:     "meso node",
			record:   mesoNode,
			expected: point,
		},
		{
			name:     "micro node",
			record:   microNode,
			expected: point,
		},
		{
			name:     "macro link",
			record:   macroLink,
			expected: shape,
		},
		{
			name:     "meso link",
			record:   mesoLink,
			expected: shape,
		},
		{
			name:     "micro link",
			record:   microLink,
			expected: shape,
		},
		{
			name:     "movement",
			record:   maneuver,
			expected: shape,
		},
		{
			name:     "canonical geometry",
			record:   geometry,
			expected: shape,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			decoded, err := wkt.Unmarshal(testCase.record.WKT())
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, decoded)
		})
	}
	assert.Equal(t, originalShape, shape)
}

func TestZoneWKTPreservesPolygonsAndHoles(t *testing.T) {
	boundary := orb.MultiPolygon{
		{
			{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}},
			{{1, 1}, {1, 2}, {2, 2}, {2, 1}, {1, 1}},
		},
		{
			{{10, 10}, {12, 10}, {12, 12}, {10, 12}, {10, 10}},
		},
	}
	zone := gmns.NewZone(gmns.NodeIDFromInt64(1), gmns.WithZoneBoundary(boundary))
	decoded, err := wkt.UnmarshalMultiPolygon(zone.WKT())
	assert.NoError(t, err)
	assert.Equal(t, boundary, decoded)
	assert.Equal(t, boundary, zone.Boundary())
}

func TestWKTEmptyGeometry(t *testing.T) {
	nodeID := gmns.NodeIDFromInt64(1)
	linkID := gmns.LinkIDFromInt64(1)
	macroLink := macro.NewLinkFrom(linkID, nodeID, nodeID)
	mesoLink := meso.NewLinkFrom(linkID, nodeID, nodeID)
	microLink := micro.NewLinkFrom(linkID, nodeID, nodeID)
	maneuver := movement.NewMovement(gmns.MovementIDFromInt64(1), nodeID, linkID, linkID, "", "")
	geometry := gmns.NewGeometry(gmns.GeometryIDFromInt64(1), nil)
	zone := gmns.NewZone(nodeID)

	assert.Equal(t, "LINESTRING EMPTY", macroLink.WKT())
	assert.Equal(t, "LINESTRING EMPTY", mesoLink.WKT())
	assert.Equal(t, "LINESTRING EMPTY", microLink.WKT())
	assert.Equal(t, "LINESTRING EMPTY", maneuver.WKT())
	assert.Equal(t, "LINESTRING EMPTY", geometry.WKT())
	assert.Equal(t, "MULTIPOLYGON EMPTY", zone.WKT())
}
