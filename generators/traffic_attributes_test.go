package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

type trafficValues struct {
	freeSpeed   float64
	hasSpeed    bool
	capacity    float64
	hasCapacity bool
}

type trafficGenerationCase struct {
	name     string
	incoming trafficValues
	outgoing trafficValues
}

func TestTrafficAttributesSurviveGeneration(t *testing.T) {
	cases := []trafficGenerationCase{
		{
			name: "absent values remain absent",
		},
		{
			name: "zero incoming values with fractional outgoing values",
			incoming: trafficValues{
				hasSpeed:    true,
				hasCapacity: true,
			},
			outgoing: trafficValues{
				freeSpeed:   60.125,
				hasSpeed:    true,
				capacity:    1800.75,
				hasCapacity: true,
			},
		},
		{
			name: "different fractional values at each road",
			incoming: trafficValues{
				freeSpeed:   40.125,
				hasSpeed:    true,
				capacity:    1234.125,
				hasCapacity: true,
			},
			outgoing: trafficValues{
				freeSpeed:   60.625,
				hasSpeed:    true,
				capacity:    1800.75,
				hasCapacity: true,
			},
		},
		{
			name: "incoming capacity and outgoing speed are independent",
			incoming: trafficValues{
				capacity:    1234.125,
				hasCapacity: true,
			},
			outgoing: trafficValues{
				freeSpeed: 60.625,
				hasSpeed:  true,
			},
		},
		{
			name: "incoming speed and outgoing capacity are independent",
			incoming: trafficValues{
				freeSpeed: 40.125,
				hasSpeed:  true,
			},
			outgoing: trafficValues{
				capacity:    1800.75,
				hasCapacity: true,
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			network := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
			expectedByMacroLink := map[gmns.LinkID]trafficValues{
				gmns.LinkIDFromInt64(10): testCase.incoming,
				gmns.LinkIDFromInt64(11): testCase.outgoing,
			}
			for linkID, expected := range expectedByMacroLink {
				original := network.Links[linkID]
				lanesNum, hasLanes := original.LanesNum()
				assert.True(t, hasLanes)
				lengthMeters, hasLength := original.LengthMeters()
				assert.True(t, hasLength)
				link := macro.NewLinkFrom(linkID, original.SourceNode(), original.TargetNode(),
					macro.WithLineGeom(original.Geom().Clone()),
					macro.WithLineGeomEuclidean(original.GeomEuclidean().Clone()),
					macro.WithLengthMeters(lengthMeters),
					macro.WithLanesNum(lanesNum),
					macro.WithLanesInfo(original.LanesInfo()),
					macro.WithAllowedAgentTypes(original.AllowedAgentTypes()),
				)
				if expected.hasSpeed {
					macro.WithFreeSpeed(expected.freeSpeed)(link)
				}
				if expected.hasCapacity {
					macro.WithCapacity(expected.capacity)(link)
				}
				network.Links[linkID] = link
			}
			movements, err := GenerateMovements(network)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Len(t, movements, 1)
			for _, mvmt := range movements {
				movement.WithCapacity(777.25)(mvmt)
			}

			mesoNet, err := GenerateMesoscopic(network, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			expectedByMesoLink := make(map[gmns.LinkID]trafficValues)
			connectorCount := 0
			roadCount := 0
			for _, link := range mesoNet.Links {
				expected, exists := expectedByMacroLink[link.MacroLink()]
				if link.MacroNode() == gmns.NodeIDFromInt64(2) {
					expected = testCase.incoming
					exists = true
					connectorCount++
				} else {
					roadCount++
				}
				assert.True(t, exists, "meso link %v must retain its source", link.ID)
				expectedByMesoLink[link.ID] = expected
				freeSpeed, hasSpeed := link.FreeSpeed()
				capacity, hasCapacity := link.Capacity()
				assert.Equal(t, expected.freeSpeed, freeSpeed)
				assert.Equal(t, expected.hasSpeed, hasSpeed)
				assert.Equal(t, expected.capacity, capacity, "capacity remains per lane and is not replaced by movement capacity")
				assert.Equal(t, expected.hasCapacity, hasCapacity)
			}
			assert.Equal(t, 1, connectorCount)
			assert.Equal(t, 2, roadCount)

			microNet, err := GenerateMicroscopic(network, mesoNet, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			movementCells := 0
			firstMovementCells := 0
			laneChanges := 0
			for _, link := range microNet.Links {
				expected, exists := expectedByMesoLink[link.MesoLink()]
				assert.True(t, exists, "micro link %v must retain its meso source", link.ID)
				freeSpeed, hasSpeed := link.FreeSpeed()
				capacity, hasCapacity := link.Capacity()
				assert.Equal(t, expected.freeSpeed, freeSpeed)
				assert.Equal(t, expected.hasSpeed, hasSpeed)
				assert.Equal(t, expected.capacity, capacity)
				assert.Equal(t, expected.hasCapacity, hasCapacity)
				if link.MacroNode() == gmns.NodeIDFromInt64(2) {
					movementCells++
				}
				if link.IsFirstMovementCell() {
					firstMovementCells++
				}
				if link.CellType() != types.CELL_FORWARD {
					laneChanges++
				}
			}
			assert.Equal(t, 2, firstMovementCells, "both lanes have a movement connection")
			assert.Greater(t, movementCells, firstMovementCells, "both intermediate and final movement cells are exercised")
			assert.Positive(t, laneChanges, "lane-change cells are exercised")

			for linkID, expected := range expectedByMacroLink {
				link := network.Links[linkID]
				freeSpeed, hasSpeed := link.FreeSpeed()
				capacity, hasCapacity := link.Capacity()
				assert.Equal(t, expected.freeSpeed, freeSpeed)
				assert.Equal(t, expected.hasSpeed, hasSpeed)
				assert.Equal(t, expected.capacity, capacity)
				assert.Equal(t, expected.hasCapacity, hasCapacity)
			}
			for _, mvmt := range movements {
				capacity, present := mvmt.Capacity()
				assert.True(t, present)
				assert.Equal(t, 777.25, capacity)
			}
		})
	}
}
