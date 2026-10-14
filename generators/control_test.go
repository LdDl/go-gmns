package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/utils/geomath"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/stretchr/testify/assert"
)

type junctionControlCase struct {
	name          string
	controlType   types.ControlType
	needsMovement bool
}

func TestControlSurvivesGeneration(t *testing.T) {
	cases := []junctionControlCase{
		{
			name:        "common",
			controlType: types.CONTROL_TYPE_NOT_SIGNAL,
		},
		{
			name:        "no_control",
			controlType: types.CONTROL_TYPE_NO_CONTROL,
		},
		{
			name:          "signal",
			controlType:   types.CONTROL_TYPE_IS_SIGNAL,
			needsMovement: true,
		},
		{
			name:          "yield",
			controlType:   types.CONTROL_TYPE_YIELD,
			needsMovement: true,
		},
		{
			name:          "stop",
			controlType:   types.CONTROL_TYPE_STOP,
			needsMovement: true,
		},
		{
			name:          "stop_2_way",
			controlType:   types.CONTROL_TYPE_STOP_2_WAY,
			needsMovement: true,
		},
		{
			name:          "stop_4_way",
			controlType:   types.CONTROL_TYPE_STOP_4_WAY,
			needsMovement: true,
		},
		{
			name:          "signal_with_RTOR",
			controlType:   types.CONTROL_TYPE_SIGNAL_WITH_RTOR,
			needsMovement: true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			network := controlTestNetwork(t, testCase.controlType)
			assert.Equal(t, testCase.name, network.Nodes[2].GeoFeature().Properties["control_type"])
			movements, err := GenerateMovements(network)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Len(t, movements, 1)
			for _, mvmt := range movements {
				assert.Equal(t, testCase.controlType, mvmt.ControlType())
				assert.Equal(t, testCase.name, mvmt.GeoFeature().Properties["control_type"])
			}

			flags := ComputeMovementFlags(network, movements)
			assert.Equal(t, testCase.needsMovement, flags.NodesNeedMovement[2])
			assert.Equal(t, !testCase.needsMovement, flags.DownstreamIsTarget[10])
			assert.False(t, flags.UpstreamIsTarget[11])

			mesoNet, err := GenerateMesoscopic(network, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			connectorCount := 0
			for _, link := range mesoNet.Links {
				if link.MacroNode() != 2 {
					continue
				}
				connectorCount++
				assert.Equal(t, testCase.controlType, link.ControlType())
				assert.Equal(t, testCase.name, link.GeoFeature().Properties["control_type"])
			}
			if testCase.needsMovement {
				assert.Equal(t, 1, connectorCount, "controlled junction keeps its movement connector")
			} else {
				assert.Zero(t, connectorCount, "uncontrolled pass-through still permits the shortcut")
			}

			microNet, err := GenerateMicroscopic(network, mesoNet, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			movementCells := 0
			for _, link := range microNet.Links {
				if link.MacroNode() != 2 {
					continue
				}
				movementCells++
				assert.Equal(t, testCase.controlType, link.ControlType())
				assert.Equal(t, testCase.name, link.GeoFeature().Properties["control_type"])
			}
			if testCase.needsMovement {
				assert.Positive(t, movementCells, "controlled junction keeps movement cells")
			} else {
				assert.Zero(t, movementCells)
			}
			assert.Equal(t, testCase.controlType, network.Nodes[2].ControlType())
		})
	}
}

func controlTestNetwork(t *testing.T, controlType types.ControlType) *macro.Net {
	t.Helper()
	points := []orb.Point{
		{30, 60},
		{30.002, 60},
		{30.004, 60},
	}
	network := macro.NewNet()
	network.Nodes[1] = macro.NewNodeFrom(1,
		macro.WithPointGeom(points[0]),
		macro.WithOutcomingLinks(10),
	)
	network.Nodes[2] = macro.NewNodeFrom(2,
		macro.WithPointGeom(points[1]),
		macro.WithIncomingLinks(10),
		macro.WithOutcomingLinks(11),
		macro.WithNodeControlType(controlType),
	)
	network.Nodes[3] = macro.NewNodeFrom(3,
		macro.WithPointGeom(points[2]),
		macro.WithIncomingLinks(11),
	)
	links := []*macro.Link{
		macro.NewLinkFrom(10, 1, 2, macro.WithLineGeom(orb.LineString{points[0], points[1]})),
		macro.NewLinkFrom(11, 2, 3, macro.WithLineGeom(orb.LineString{points[1], points[2]})),
	}
	for _, link := range links {
		macro.WithLineGeomEuclidean(geomath.LineToEuclidean(link.Geom()))(link)
		macro.WithLengthMeters(geo.LengthHaversine(link.Geom()))(link)
		macro.WithLanesNum(2)(link)
		laneInfo, err := macro.NewLanesInfo(link)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		macro.WithLanesInfo(laneInfo)(link)
		macro.WithFreeSpeed(40)(link)
		macro.WithCapacity(1800)(link)
		macro.WithAllowedAgentTypes([]types.AgentType{types.AGENT_AUTO})(link)
		network.Links[link.ID] = link
	}
	return network
}
