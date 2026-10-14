package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/stretchr/testify/assert"
)

type mesoOwnershipCase struct {
	name          string
	bidirectional bool
	invalidNode   bool
}

func TestGenerateMesoscopicPreservesSource(t *testing.T) {
	cases := []mesoOwnershipCase{
		{
			name: "single direction",
		},
		{
			name:          "opposite directions sharing lane information",
			bidirectional: true,
		},
		{
			name:          "error after offset preparation",
			bidirectional: true,
			invalidNode:   true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{
				{30, 60},
				{30.002, 60},
				{30.002, 60.002},
			}
			length := geo.LengthHaversine(shape)
			laneInfo := macro.LanesInfo{
				LanesList:         []int{2, 3},
				LanesChange:       [][2]int{{0, 0}, {0, 1}},
				LanesChangePoints: []float64{0, length * 0.4, length},
			}
			expectedLaneInfo := macro.LanesInfo{
				LanesList:         []int{2, 3},
				LanesChange:       [][2]int{{0, 0}, {0, 1}},
				LanesChangePoints: []float64{0, length * 0.4, length},
			}
			network := macro.NewNet()
			network.Nodes[1] = macro.NewNodeFrom(1, macro.WithPointGeom(shape[0]), macro.WithOutcomingLinks(10))
			network.Nodes[2] = macro.NewNodeFrom(2, macro.WithPointGeom(shape[2]), macro.WithIncomingLinks(10))
			network.Links[10] = macro.NewLinkFrom(10, 1, 2,
				macro.WithLineGeom(shape.Clone()),
				macro.WithLineGeomEuclidean(geomath.LineToEuclidean(shape)),
				macro.WithLengthMeters(length),
				macro.WithLanesNum(2),
				macro.WithLanesInfo(laneInfo),
			)
			if testCase.bidirectional {
				reverse := shape.Clone()
				reverse.Reverse()
				network.Links[11] = macro.NewLinkFrom(11, 2, 1,
					macro.WithLineGeom(reverse),
					macro.WithLineGeomEuclidean(geomath.LineToEuclidean(reverse)),
					macro.WithLengthMeters(length),
					macro.WithLanesNum(2),
					macro.WithLanesInfo(laneInfo),
				)
				macro.WithIncomingLinks(11)(network.Nodes[1])
				macro.WithOutcomingLinks(11)(network.Nodes[2])
			}
			movements := movement.NewMovementsStorage()
			if testCase.invalidNode {
				movements[1] = movement.NewMovement(1, 999, 10, 11, movement.MOVEMENT_SBT, movement.MOVEMENT_TYPE_THRU)
			}
			first, err := GenerateMesoscopic(network, movements)
			for _, link := range network.Links {
				assert.Equal(t, expectedLaneInfo, link.LanesInfo(), "source link %d", link.ID)
			}
			assert.Equal(t, shape, network.Links[10].Geom())
			assert.Equal(t, geomath.LineToEuclidean(shape), network.Links[10].GeomEuclidean())
			if testCase.invalidNode {
				assert.ErrorIs(t, err, macro.ErrNodeNotFound)
				assert.Nil(t, first)
				return
			}
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Len(t, first.Links, len(network.Links)*2)
			second, err := GenerateMesoscopic(network, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, first, second, "repeated generation must produce the same graph")
			for _, link := range network.Links {
				assert.Equal(t, expectedLaneInfo, link.LanesInfo(), "source link %d after repeated generation", link.ID)
			}
		})
	}
}
