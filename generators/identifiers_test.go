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

type generationIDCase struct {
	name  string
	kind  gmns.IDType
	nodes []gmns.NodeID
	links []gmns.LinkID
}

func TestGenerationPreservesExternalIDsAndLocalAllocation(t *testing.T) {
	large, err := gmns.ParseNodeID("922337203685477580812345", gmns.IDTypeInteger)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	cases := []generationIDCase{
		{
			name: "string IDs including distinct leading zeroes",
			kind: gmns.IDTypeString,
			nodes: []gmns.NodeID{
				gmns.NodeIDFromString("001"),
				gmns.NodeIDFromString("1"),
				gmns.NodeIDFromString("node-A"),
			},
			links: []gmns.LinkID{
				gmns.LinkIDFromString("001"),
				gmns.LinkIDFromString("1"),
			},
		},
		{
			name: "negative zero and large integer IDs",
			kind: gmns.IDTypeInteger,
			nodes: []gmns.NodeID{
				gmns.NodeIDFromInt64(-1),
				gmns.NodeIDFromInt64(0),
				large,
			},
			links: []gmns.LinkID{
				gmns.LinkIDFromInt64(-1),
				gmns.LinkIDFromInt64(0),
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			net := identifierGenerationNet(t, testCase.nodes, testCase.links)
			assert.NoError(t, net.ValidateIDType(testCase.kind))
			movements, err := GenerateMovements(net)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			repeated, err := GenerateMovements(net)
			assert.NoError(t, err)
			assert.Equal(t, movements, repeated, "movement IDs must not depend on previous runs")
			assert.NoError(t, net.ValidateMovements(movements))
			assert.NoError(t, movements.ValidateIDType(testCase.kind))
			assert.Len(t, movements, 1)
			mesoNet, err := GenerateMesoscopic(net, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.NoError(t, mesoNet.ValidateIDType(testCase.kind))
			assert.NoError(t, mesoNet.ValidateSourceReferences(net, movements))
			for _, node := range mesoNet.Nodes {
				if !node.MacroNode().IsZero() {
					assert.Contains(t, net.Nodes, node.MacroNode())
				}
				_, hasParent := node.ParentNode()
				assert.False(t, hasParent, "source lineage must not become a same-resource parent")
			}
			microNet, err := GenerateMicroscopic(net, mesoNet, movements)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.NotEmpty(t, microNet.Links)
			assert.NoError(t, microNet.ValidateIDType(testCase.kind))
			assert.NoError(t, microNet.ValidateSourceReferences(net, mesoNet))
			assert.Len(t, net.Nodes, 3)
			assert.Len(t, net.Links, 2)
			for _, nodeID := range testCase.nodes {
				assert.Equal(t, nodeID, net.Nodes[nodeID].ID)
			}
			for _, linkID := range testCase.links {
				assert.Equal(t, linkID, net.Links[linkID].ID)
			}

			for _, link := range mesoNet.Links {
				if !link.MacroLink().IsZero() {
					meso.WithLineMacroLinkID(gmns.LinkIDFromString("missing"))(link)
					break
				}
			}
			assert.Error(t, mesoNet.ValidateSourceReferences(net, movements))
			invalidMicro, err := GenerateMicroscopic(net, mesoNet, movements)
			assert.Error(t, err)
			assert.Nil(t, invalidMicro)
		})
	}
}

func TestMesoNodeAllocationDoesNotOverlapAfterHundredExpansions(t *testing.T) {
	first := gmns.NodeIDFromInt64(1)
	second := gmns.NodeIDFromInt64(2)
	nodes := map[gmns.NodeID]*macro.Node{
		first:  macro.NewNodeFrom(first),
		second: macro.NewNodeFrom(second),
	}
	shape := orb.LineString{
		{30, 60},
		{30.002, 60},
	}
	processed := make(map[gmns.LinkID]*macroLinkProcessing)
	for index := 0; index < 120; index++ {
		linkID := gmns.LinkIDFromInt64(int64(index))
		processed[linkID] = &macroLinkProcessing{
			id:                     linkID,
			sourceMacroNodeID:      first,
			targetMacroNodeID:      second,
			offsetGeomCut:          []orb.LineString{shape},
			offsetGeomEuclideanCut: []orb.LineString{geomath.LineToEuclidean(shape)},
			lanesInfoCut: macro.LanesInfo{
				LanesList: []int{2},
				LanesChange: [][2]int{
					{0, 0},
				},
			},
		}
	}
	mesoNodes, mesoLinks, err := generateBaseNodesLinks(nodes, processed, gmns.IDTypeInteger)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Len(t, mesoNodes, 240)
	assert.Len(t, mesoLinks, 120)
	net := meso.NewNet()
	net.Nodes = mesoNodes
	net.Links = mesoLinks
	assert.NoError(t, net.ValidateReferences())
	for _, link := range mesoLinks {
		assert.Equal(t, first, mesoNodes[link.SourceNode()].MacroNode())
		assert.Equal(t, second, mesoNodes[link.TargetNode()].MacroNode())
	}
}

func TestGeneratorsRejectMismatchedKeysAndMixedIDKinds(t *testing.T) {
	net := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	link := net.Links[gmns.LinkIDFromInt64(10)]
	link.ID = gmns.LinkIDFromString("10")
	movements, err := GenerateMovements(net)
	assert.Error(t, err)
	assert.Nil(t, movements)
	delete(net.Links, gmns.LinkIDFromInt64(10))
	net.Links[link.ID] = link
	movements, err = GenerateMovements(net)
	assert.Error(t, err)
	assert.Nil(t, movements)
	mesoNet, err := GenerateMesoscopic(net, movement.NewMovementsStorage())
	assert.Error(t, err)
	assert.Nil(t, mesoNet)
	microNet, err := GenerateMicroscopic(net, meso.NewNet(), movement.NewMovementsStorage())
	assert.Error(t, err)
	assert.Nil(t, microNet)
	movements, err = GenerateMovements(nil)
	assert.Error(t, err)
	assert.Nil(t, movements)
}

func identifierGenerationNet(t *testing.T, nodes []gmns.NodeID, links []gmns.LinkID) *macro.Net {
	t.Helper()
	points := []orb.Point{
		{30, 60},
		{30.002, 60},
		{30.004, 60},
	}
	net := macro.NewNet()
	net.Nodes[nodes[0]] = macro.NewNodeFrom(nodes[0], macro.WithPointGeom(points[0]), macro.WithOutcomingLinks(links[0]))
	net.Nodes[nodes[1]] = macro.NewNodeFrom(nodes[1], macro.WithPointGeom(points[1]), macro.WithIncomingLinks(links[0]), macro.WithOutcomingLinks(links[1]), macro.WithNodeControlType(types.CONTROL_TYPE_IS_SIGNAL))
	net.Nodes[nodes[2]] = macro.NewNodeFrom(nodes[2], macro.WithPointGeom(points[2]), macro.WithIncomingLinks(links[1]))
	for index, linkID := range links {
		shape := orb.LineString{points[index], points[index+1]}
		link := macro.NewLinkFrom(linkID, nodes[index], nodes[index+1],
			macro.WithLineGeom(shape),
			macro.WithLineGeomEuclidean(geomath.LineToEuclidean(shape)),
			macro.WithLengthMeters(geo.LengthHaversine(shape)),
			macro.WithLanesNum(2),
			macro.WithAllowedAgentTypes([]types.AgentType{types.AGENT_AUTO}),
		)
		laneInfo, err := macro.NewLanesInfo(link)
		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}
		macro.WithLanesInfo(laneInfo)(link)
		net.Links[linkID] = link
	}
	return net
}

func TestMicroSegmentOrderIsIndependentOfStringIDs(t *testing.T) {
	macroID := gmns.LinkIDFromString("road")
	nodeID := gmns.NodeIDFromString("node")
	net := meso.NewNet()
	ids := []gmns.LinkID{
		gmns.LinkIDFromString("2"),
		gmns.LinkIDFromString("10"),
		gmns.LinkIDFromString("1"),
	}
	for index, identifier := range ids {
		net.Links[identifier] = meso.NewLinkFrom(identifier, nodeID, nodeID,
			meso.WithLineMacroLinkID(macroID),
			meso.WithSegmentIdx(index),
		)
	}
	mapping := buildMacroToMesoMapping(net)
	assert.Equal(t, ids, mapping[macroID])
}
