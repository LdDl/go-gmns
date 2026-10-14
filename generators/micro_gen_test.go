package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/micro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

type gapLaneCase struct {
	name       string
	options    []func(*movement.Movement)
	laneOffset int
	joined     bool
}

func TestFixGapsLanePresence(t *testing.T) {
	cases := []gapLaneCase{
		{
			name:       "absent lanes must not select negative lane",
			laneOffset: 1,
		},
		{
			name: "explicit single lanes",
			options: []func(*movement.Movement){
				movement.WithIncomeLane(1, 1),
				movement.WithOutcomeLane(1, 1),
			},
			joined: true,
		},
		{
			name: "absent endpoints mean single lanes",
			options: []func(*movement.Movement){
				movement.WithIncomeLaneStart(1),
				movement.WithOutcomeLaneStart(1),
			},
			joined: true,
		},
		{
			name: "explicit negative lanes",
			options: []func(*movement.Movement){
				movement.WithIncomeLane(-1, -1),
				movement.WithOutcomeLane(-1, -1),
			},
			laneOffset: 1,
			joined:     true,
		},
		{
			name: "missing outbound start",
			options: []func(*movement.Movement){
				movement.WithIncomeLaneStart(1),
				movement.WithOutcomeLaneEnd(1),
			},
		},
		{
			name: "missing inbound start",
			options: []func(*movement.Movement){
				movement.WithIncomeLaneEnd(1),
				movement.WithOutcomeLaneStart(1),
			},
		},
		{
			name: "empty ranges",
			options: []func(*movement.Movement){
				movement.WithIncomeLane(2, 1),
				movement.WithOutcomeLane(2, 1),
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			macroNet := macro.NewNet()
			macroNet.Nodes[2] = macro.NewNodeFrom(2)
			macroNet.Links[3] = macro.NewLinkFrom(3, 1, 2)
			macroNet.Links[4] = macro.NewLinkFrom(4, 2, 3)
			mesoNet := meso.NewNet()
			mesoNet.Links[10] = meso.NewLinkFrom(10, 1, 2, meso.WithLanesNum(1), meso.WithLanesChange([2]int{testCase.laneOffset, 0}))
			mesoNet.Links[20] = meso.NewLinkFrom(20, 3, 4, meso.WithLanesNum(1), meso.WithLanesChange([2]int{testCase.laneOffset, 0}))
			microNet := micro.NewNet()
			microNet.AddNode(micro.NewNodeFrom(100, micro.WithNodeMesoLinkID(10), micro.WithNodeLaneID(1), micro.WithCellIndex(0)))
			microNet.AddNode(micro.NewNodeFrom(200, micro.WithNodeMesoLinkID(20), micro.WithNodeLaneID(1), micro.WithCellIndex(0)))
			microNet.AddNode(micro.NewNodeFrom(201, micro.WithNodeMesoLinkID(20), micro.WithNodeLaneID(1), micro.WithCellIndex(1)))
			microNet.AddLink(micro.NewLinkFrom(300, 200, 201))
			microNet.Nodes[200].AddOutcomingLink(300)
			macroLinkMesoLinks := map[gmns.LinkID][]gmns.LinkID{
				3: {10},
				4: {20},
			}
			flags := &MovementFlags{
				DownstreamIsTarget: map[gmns.LinkID]bool{
					3: true,
				},
			}
			mvmt := movement.NewMovement(1, 2, 3, 4, movement.MOVEMENT_SBT, movement.MOVEMENT_TYPE_THRU, testCase.options...)
			before := mvmt.GeoFeature().Properties
			movements := movement.NewMovementsStorage()
			movements[mvmt.ID] = mvmt
			err := fixGaps(macroNet, mesoNet, microNet, macroLinkMesoLinks, flags, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			if testCase.joined {
				assert.NotContains(t, microNet.Nodes, gmns.NodeID(200))
				assert.Equal(t, gmns.NodeID(100), microNet.Links[300].SourceNode())
			} else {
				assert.Contains(t, microNet.Nodes, gmns.NodeID(200))
				assert.Equal(t, gmns.NodeID(200), microNet.Links[300].SourceNode())
			}
			assert.Equal(t, before, mvmt.GeoFeature().Properties, "generation must not fill absent stored endpoints")
		})
	}
}
