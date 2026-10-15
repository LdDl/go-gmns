package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/stretchr/testify/assert"
)

type requiredLaneCountCase struct {
	name    string
	value   int
	present bool
}

func TestGeneratorsRejectMissingOrNegativeLaneCounts(t *testing.T) {
	cases := []requiredLaneCountCase{
		{
			name: "absent",
		},
		{
			name:    "negative",
			value:   -1,
			present: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			macroNet, mesoNet := laneCountTestNetworks()
			macroLink := macroNet.Links[gmns.LinkIDFromInt64(10)]
			mesoLink := mesoNet.Links[gmns.LinkIDFromInt64(100)]
			if testCase.present {
				macro.WithLanesNum(testCase.value)(macroLink)
				meso.WithLanesNum(testCase.value)(mesoLink)
			}
			movements, err := GenerateMovements(macroNet)
			assert.Error(t, err)
			assert.Nil(t, movements)
			if err != nil {
				assert.Contains(t, err.Error(), "macro link 10: lane count")
			}
			generatedMeso, err := GenerateMesoscopic(macroNet, movement.NewMovementsStorage())
			assert.Error(t, err)
			assert.Nil(t, generatedMeso)
			if err != nil {
				assert.Contains(t, err.Error(), "macro link 10: lane count")
			}
			generatedMicro, err := GenerateMicroscopic(macroNet, mesoNet, movement.NewMovementsStorage())
			assert.Error(t, err)
			assert.Nil(t, generatedMicro)
			if err != nil {
				assert.Contains(t, err.Error(), "macro link 10: lane count")
			}
			value, present := macroLink.LanesNum()
			assert.Equal(t, testCase.value, value)
			assert.Equal(t, testCase.present, present)

			macro.WithLanesNum(2)(macroLink)
			generatedMicro, err = GenerateMicroscopic(macroNet, mesoNet, movement.NewMovementsStorage())
			assert.Error(t, err)
			assert.Nil(t, generatedMicro)
			if err != nil {
				assert.Contains(t, err.Error(), "meso link 100: lane count")
			}
			value, present = mesoLink.LanesNum()
			assert.Equal(t, testCase.value, value)
			assert.Equal(t, testCase.present, present)
			assert.Len(t, macroNet.Nodes, 2)
			assert.Len(t, mesoNet.Nodes, 2)
		})
	}
}

type requiredLaneDataCase struct {
	name    string
	options []func(*macro.Link)
}

func TestMovementAndMesoRequirePositivePreparedLanes(t *testing.T) {
	cases := []requiredLaneDataCase{
		{
			name: "zero motor lanes",
			options: []func(*macro.Link){
				macro.WithLanesNum(0),
			},
		},
		{
			name: "no prepared segments",
			options: []func(*macro.Link){
				macro.WithLanesInfo(macro.LanesInfo{}),
			},
		},
		{
			name: "missing lane changes",
			options: []func(*macro.Link){
				macro.WithLanesInfo(macro.LanesInfo{
					LanesList:         []int{2},
					LanesChangePoints: []float64{0, 50},
				}),
			},
		},
		{
			name: "zero lanes in a segment",
			options: []func(*macro.Link){
				macro.WithLanesInfo(macro.LanesInfo{
					LanesList: []int{0},
					LanesChange: [][2]int{
						{0, 0},
					},
					LanesChangePoints: []float64{0, 50},
				}),
			},
		},
		{
			name: "negative lanes in a segment",
			options: []func(*macro.Link){
				macro.WithLanesInfo(macro.LanesInfo{
					LanesList: []int{-1},
					LanesChange: [][2]int{
						{0, 0},
					},
					LanesChangePoints: []float64{0, 50},
				}),
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			network := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
			link := network.Links[gmns.LinkIDFromInt64(10)]
			for _, option := range testCase.options {
				option(link)
			}
			originalInfo := link.LanesInfo()
			originalShape := link.Geom().Clone()
			movements, err := GenerateMovements(network)
			assert.Error(t, err)
			assert.Nil(t, movements)
			if err != nil {
				assert.Contains(t, err.Error(), "macro link 10")
			}
			mesoNet, err := GenerateMesoscopic(network, movement.NewMovementsStorage())
			assert.Error(t, err)
			assert.Nil(t, mesoNet)
			if err != nil {
				assert.Contains(t, err.Error(), "macro link 10")
			}
			assert.Equal(t, originalInfo, link.LanesInfo())
			assert.Equal(t, originalShape, link.Geom())
		})
	}
}

func TestMicroAllowsZeroMotorLanesWithSeparateBikeWalk(t *testing.T) {
	macroNet, mesoNet := laneCountTestNetworks()
	macro.WithLanesNum(0)(macroNet.Links[gmns.LinkIDFromInt64(10)])
	meso.WithLanesNum(0)(mesoNet.Links[gmns.LinkIDFromInt64(100)])
	options := DefaultMicroGenOptions()
	options.SeparateBikeWalk = true
	microNet, err := GenerateMicroscopic(macroNet, mesoNet, movement.NewMovementsStorage(), options)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NotEmpty(t, microNet.Nodes)
	assert.NotEmpty(t, microNet.Links)
	bikeNodes := 0
	walkNodes := 0
	for _, node := range microNet.Nodes {
		switch node.LaneID() {
		case -1:
			bikeNodes++
		case -2:
			walkNodes++
		default:
			assert.Fail(t, "zero motor lanes must not generate motor-lane nodes")
		}
	}
	assert.Positive(t, bikeNodes)
	assert.Positive(t, walkNodes)
	for _, link := range microNet.Links {
		assert.Contains(t, []int{-1, -2}, link.LaneID())
	}
	macroLanes, macroPresent := macroNet.Links[gmns.LinkIDFromInt64(10)].LanesNum()
	mesoLanes, mesoPresent := mesoNet.Links[gmns.LinkIDFromInt64(100)].LanesNum()
	assert.Zero(t, macroLanes)
	assert.True(t, macroPresent)
	assert.Zero(t, mesoLanes)
	assert.True(t, mesoPresent)
}

func TestMicroRejectsMissingConnectorLaneCount(t *testing.T) {
	macroNet := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	movements, err := GenerateMovements(macroNet)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	mesoNet, err := GenerateMesoscopic(macroNet, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	connectorCount := 0
	for linkID, original := range mesoNet.Links {
		if original.MacroNode() != gmns.NodeIDFromInt64(2) {
			continue
		}
		connectorCount++
		connector := meso.NewLinkFrom(linkID, original.SourceNode(), original.TargetNode(),
			meso.WithLineMacroNodeID(original.MacroNode()),
			meso.WithLineMacroLinkID(gmns.LinkID{}),
			meso.WithMovementID(original.Movement()),
			meso.WithMovementMesoLinkIncome(original.MovementMesoLinkIncome()),
			meso.WithMovementMesoLinkOutcome(original.MovementMesoLinkOutcome()),
			meso.WithMovementIncomeLaneStartSeqID(original.MovementIncomeLaneStartSeqID()),
			meso.WithMovementOutcomeLaneStartSeqID(original.MovementOutcomeLaneStartSeqID()),
		)
		mesoNet.Links[linkID] = connector
	}
	assert.Equal(t, 1, connectorCount)
	microNet, err := GenerateMicroscopic(macroNet, mesoNet, movements)
	assert.Error(t, err)
	assert.Nil(t, microNet)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to connect micro links")
		assert.Contains(t, err.Error(), "lane count is required")
	}
	for _, link := range mesoNet.Links {
		if link.MacroNode() == gmns.NodeIDFromInt64(2) {
			_, present := link.LanesNum()
			assert.False(t, present)
		}
	}
}

func laneCountTestNetworks() (*macro.Net, *meso.Net) {
	shape := orb.LineString{
		{30, 60},
		{30.002, 60},
	}
	length := geo.LengthHaversine(shape)
	macroNet := macro.NewNet()
	macroNet.Nodes[gmns.NodeIDFromInt64(1)] = macro.NewNodeFrom(gmns.NodeIDFromInt64(1), macro.WithOutcomingLinks(gmns.LinkIDFromInt64(10)))
	macroNet.Nodes[gmns.NodeIDFromInt64(2)] = macro.NewNodeFrom(gmns.NodeIDFromInt64(2), macro.WithIncomingLinks(gmns.LinkIDFromInt64(10)))
	macroNet.Links[gmns.LinkIDFromInt64(10)] = macro.NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2),
		macro.WithLengthMeters(length),
		macro.WithLineGeom(shape.Clone()),
		macro.WithLineGeomEuclidean(geomath.LineToEuclidean(shape)),
		macro.WithAllowedAgentTypes([]types.AgentType{
			types.AGENT_BIKE,
			types.AGENT_WALK,
		}),
		macro.WithLanesInfo(macro.LanesInfo{
			LanesList: []int{2},
			LanesChange: [][2]int{
				{0, 0},
			},
			LanesChangePoints: []float64{0, length},
		}),
	)
	mesoNet := meso.NewNet()
	mesoNet.Nodes[gmns.NodeIDFromInt64(1)] = meso.NewNodeFrom(gmns.NodeIDFromInt64(1))
	mesoNet.Nodes[gmns.NodeIDFromInt64(2)] = meso.NewNodeFrom(gmns.NodeIDFromInt64(2))
	mesoNet.Links[gmns.LinkIDFromInt64(100)] = meso.NewLinkFrom(gmns.LinkIDFromInt64(100), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2),
		meso.WithLineMacroLinkID(gmns.LinkIDFromInt64(10)),
		meso.WithLineMacroNodeID(gmns.NodeID{}),
		meso.WithMovementID(gmns.MovementID{}),
		meso.WithLengthMeters(length),
		meso.WithLineGeom(shape.Clone()),
		meso.WithLineGeomEuclidean(geomath.LineToEuclidean(shape)),
	)
	return macroNet, mesoNet
}
