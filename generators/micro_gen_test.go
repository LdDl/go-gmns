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
			macroNet.Nodes[gmns.NodeIDFromInt64(2)] = macro.NewNodeFrom(gmns.NodeIDFromInt64(2))
			macroNet.Links[gmns.LinkIDFromInt64(3)] = macro.NewLinkFrom(gmns.LinkIDFromInt64(3), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2))
			macroNet.Links[gmns.LinkIDFromInt64(4)] = macro.NewLinkFrom(gmns.LinkIDFromInt64(4), gmns.NodeIDFromInt64(2), gmns.NodeIDFromInt64(3))
			mesoNet := meso.NewNet()
			mesoNet.Links[gmns.LinkIDFromInt64(10)] = meso.NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), meso.WithLanesNum(1), meso.WithLanesChange([2]int{testCase.laneOffset, 0}))
			mesoNet.Links[gmns.LinkIDFromInt64(20)] = meso.NewLinkFrom(gmns.LinkIDFromInt64(20), gmns.NodeIDFromInt64(3), gmns.NodeIDFromInt64(4), meso.WithLanesNum(1), meso.WithLanesChange([2]int{testCase.laneOffset, 0}))
			microNet := micro.NewNet()
			microNet.AddNode(micro.NewNodeFrom(gmns.NodeIDFromInt64(100), micro.WithNodeMesoLinkID(gmns.LinkIDFromInt64(10)), micro.WithNodeLaneID(1), micro.WithCellIndex(0)))
			microNet.AddNode(micro.NewNodeFrom(gmns.NodeIDFromInt64(200), micro.WithNodeMesoLinkID(gmns.LinkIDFromInt64(20)), micro.WithNodeLaneID(1), micro.WithCellIndex(0)))
			microNet.AddNode(micro.NewNodeFrom(gmns.NodeIDFromInt64(201), micro.WithNodeMesoLinkID(gmns.LinkIDFromInt64(20)), micro.WithNodeLaneID(1), micro.WithCellIndex(1)))
			microNet.AddLink(micro.NewLinkFrom(gmns.LinkIDFromInt64(300), gmns.NodeIDFromInt64(200), gmns.NodeIDFromInt64(201)))
			microNet.Nodes[gmns.NodeIDFromInt64(200)].AddOutcomingLink(gmns.LinkIDFromInt64(300))
			macroLinkMesoLinks := map[gmns.LinkID][]gmns.LinkID{
				gmns.LinkIDFromInt64(3): {gmns.LinkIDFromInt64(10)},
				gmns.LinkIDFromInt64(4): {gmns.LinkIDFromInt64(20)},
			}
			flags := &MovementFlags{
				DownstreamIsTarget: map[gmns.LinkID]bool{
					gmns.LinkIDFromInt64(3): true,
				},
			}
			mvmt := movement.NewMovement(gmns.MovementIDFromInt64(1), gmns.NodeIDFromInt64(2), gmns.LinkIDFromInt64(3), gmns.LinkIDFromInt64(4), movement.MOVEMENT_SBT, movement.MOVEMENT_TYPE_THRU, testCase.options...)
			before := mvmt.GeoFeature().Properties
			movements := movement.NewMovementsStorage()
			movements[mvmt.ID] = mvmt
			err := fixGaps(macroNet, mesoNet, microNet, macroLinkMesoLinks, flags, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			if testCase.joined {
				assert.NotContains(t, microNet.Nodes, gmns.NodeIDFromInt64(200))
				assert.Equal(t, gmns.NodeIDFromInt64(100), microNet.Links[gmns.LinkIDFromInt64(300)].SourceNode())
			} else {
				assert.Contains(t, microNet.Nodes, gmns.NodeIDFromInt64(200))
				assert.Equal(t, gmns.NodeIDFromInt64(200), microNet.Links[gmns.LinkIDFromInt64(300)].SourceNode())
			}
			assert.Equal(t, before, mvmt.GeoFeature().Properties, "generation must not fill absent stored endpoints")
		})
	}
}
