package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/stretchr/testify/assert"
)

type zoneValue struct {
	id      gmns.NodeID
	present bool
}

type zoneGenerationCase struct {
	name  string
	zones map[gmns.NodeID]zoneValue
}

func TestZonePresenceSurvivesGeneration(t *testing.T) {
	cases := []zoneGenerationCase{
		{
			name: "all absent",
		},
		{
			name: "zero negative and positive zone IDs",
			zones: map[gmns.NodeID]zoneValue{
				gmns.NodeIDFromInt64(1): {
					id:      gmns.NodeIDFromInt64(0),
					present: true,
				},
				gmns.NodeIDFromInt64(2): {
					id:      gmns.NodeIDFromInt64(-1),
					present: true,
				},
				gmns.NodeIDFromInt64(3): {
					id:      gmns.NodeIDFromInt64(42),
					present: true,
				},
			},
		},
		{
			name: "only junction has a zone",
			zones: map[gmns.NodeID]zoneValue{
				gmns.NodeIDFromInt64(2): {
					id:      gmns.NodeIDFromInt64(42),
					present: true,
				},
			},
		},
		{
			name: "junction zone absent between assigned endpoints",
			zones: map[gmns.NodeID]zoneValue{
				gmns.NodeIDFromInt64(1): {
					id:      gmns.NodeIDFromInt64(-1),
					present: true,
				},
				gmns.NodeIDFromInt64(3): {
					id:      gmns.NodeIDFromInt64(0),
					present: true,
				},
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			network := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
			for nodeID, expected := range testCase.zones {
				macro.WithZoneID(expected.id)(network.Nodes[nodeID])
			}
			for _, link := range network.Links {
				macro.WithAllowedAgentTypes([]types.AgentType{
					types.AGENT_AUTO,
					types.AGENT_BIKE,
					types.AGENT_WALK,
				})(link)
			}
			firstLink := network.Links[gmns.LinkIDFromInt64(10)]
			length, hasLength := firstLink.LengthMeters()
			assert.True(t, hasLength)
			macro.WithLanesInfo(macro.LanesInfo{
				LanesList: []int{2, 2},
				LanesChange: [][2]int{
					{0, 0},
					{0, 0},
				},
				LanesChangePoints: []float64{0, length / 2, length},
			})(firstLink)
			movements, err := GenerateMovements(network)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			mesoNet, err := GenerateMesoscopic(network, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			midLinkNodes := 0
			endNodes := 0
			for _, node := range mesoNet.Nodes {
				expected := testCase.zones[node.MacroNode()]
				zoneID, present := node.MacroZone()
				assert.Equal(t, expected.id, zoneID, "meso node %v", node.ID)
				assert.Equal(t, expected.present, present, "meso node %v", node.ID)
				if node.MacroNode().IsZero() {
					midLinkNodes++
					assert.False(t, present, "a mid-link node must not acquire the destination zone")
				} else {
					endNodes++
				}
			}
			assert.Positive(t, midLinkNodes)
			assert.Positive(t, endNodes)

			options := DefaultMicroGenOptions()
			options.SeparateBikeWalk = true
			microNet, err := GenerateMicroscopic(network, mesoNet, movements, options)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			interiorNodes := 0
			bikeEnds := 0
			walkEnds := 0
			roadEnds := 0
			for _, node := range microNet.Nodes {
				expected := zoneValue{}
				if node.IsUpstreamEnd() || node.IsDownstreamEnd() {
					mesoLink := mesoNet.Links[node.MesoLink()]
					if !assert.NotNil(t, mesoLink) {
						continue
					}
					macroLink := network.Links[mesoLink.MacroLink()]
					if !assert.NotNil(t, macroLink) {
						continue
					}
					if node.IsUpstreamEnd() {
						expected = testCase.zones[macroLink.SourceNode()]
					} else {
						expected = testCase.zones[macroLink.TargetNode()]
					}
					switch node.LaneID() {
					case -1:
						bikeEnds++
					case -2:
						walkEnds++
					default:
						roadEnds++
					}
				} else {
					interiorNodes++
				}
				zoneID, present := node.ZoneID()
				assert.Equal(t, expected.id, zoneID, "micro node %v", node.ID)
				assert.Equal(t, expected.present, present, "micro node %v", node.ID)
			}
			assert.Positive(t, interiorNodes)
			assert.Positive(t, roadEnds)
			assert.Positive(t, bikeEnds)
			assert.Positive(t, walkEnds)
			for nodeID, node := range network.Nodes {
				expected := testCase.zones[nodeID]
				zoneID, present := node.Zone()
				assert.Equal(t, expected.id, zoneID)
				assert.Equal(t, expected.present, present)
			}
		})
	}
}
