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

func TestGeneratorsRejectUndirectedMacroLinks(t *testing.T) {
	net := macro.NewNet()
	net.Links[gmns.LinkIDFromInt64(20)] = macro.NewLinkFrom(gmns.LinkIDFromInt64(20), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), macro.WithDirected(false))
	net.Links[gmns.LinkIDFromInt64(10)] = macro.NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), macro.WithDirected(false))
	movements, err := GenerateMovements(net)
	assert.ErrorIs(t, err, ErrUndirectedLinkUnsupported)
	assert.Nil(t, movements)
	if err == nil {
		return
	}
	assert.Contains(t, err.Error(), "macro link 10")
	mesoNet, err := GenerateMesoscopic(net, movement.NewMovementsStorage())
	assert.ErrorIs(t, err, ErrUndirectedLinkUnsupported)
	assert.Nil(t, mesoNet)
	if err == nil {
		return
	}
	assert.Contains(t, err.Error(), "macro link 10")
	microNet, err := GenerateMicroscopic(net, meso.NewNet(), movement.NewMovementsStorage())
	assert.ErrorIs(t, err, ErrUndirectedLinkUnsupported)
	assert.Nil(t, microNet)
	if err == nil {
		return
	}
	assert.Contains(t, err.Error(), "macro link 10")
	assert.False(t, net.Links[gmns.LinkIDFromInt64(10)].Directed())
	assert.False(t, net.Links[gmns.LinkIDFromInt64(20)].Directed())
}

func TestMicroGeneratorRejectsUndirectedMesoLinks(t *testing.T) {
	net := meso.NewNet()
	net.Links[gmns.LinkIDFromInt64(20)] = meso.NewLinkFrom(gmns.LinkIDFromInt64(20), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), meso.WithDirected(false))
	net.Links[gmns.LinkIDFromInt64(10)] = meso.NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2), meso.WithDirected(false))
	result, err := GenerateMicroscopic(macro.NewNet(), net, movement.NewMovementsStorage())
	assert.ErrorIs(t, err, ErrUndirectedLinkUnsupported)
	assert.Nil(t, result)
	if err == nil {
		return
	}
	assert.Contains(t, err.Error(), "meso link 10")
	assert.False(t, net.Links[gmns.LinkIDFromInt64(10)].Directed())
	assert.False(t, net.Links[gmns.LinkIDFromInt64(20)].Directed())
}

func TestDirectedGenerationPipeline(t *testing.T) {
	shape := orb.LineString{
		{30, 60},
		{30.002, 60},
	}
	net := macro.NewNet()
	net.Nodes[gmns.NodeIDFromInt64(1)] = macro.NewNodeFrom(gmns.NodeIDFromInt64(1), macro.WithPointGeom(shape[0]), macro.WithOutcomingLinks(gmns.LinkIDFromInt64(10)))
	net.Nodes[gmns.NodeIDFromInt64(2)] = macro.NewNodeFrom(gmns.NodeIDFromInt64(2), macro.WithPointGeom(shape[1]), macro.WithIncomingLinks(gmns.LinkIDFromInt64(10)))
	link := macro.NewLinkFrom(gmns.LinkIDFromInt64(10), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(2),
		macro.WithLineGeom(shape),
		macro.WithLineGeomEuclidean(geomath.LineToEuclidean(shape)),
		macro.WithLengthMeters(geo.LengthHaversine(shape)),
		macro.WithLanesNum(2),
		macro.WithFreeSpeed(40),
		macro.WithCapacity(1800),
		macro.WithAllowedAgentTypes([]types.AgentType{types.AGENT_AUTO}),
	)
	laneInfo, err := macro.NewLanesInfo(link)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	macro.WithLanesInfo(laneInfo)(link)
	net.Links[gmns.LinkIDFromInt64(10)] = link
	movements, err := GenerateMovements(net)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	mesoNet, err := GenerateMesoscopic(net, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NotEmpty(t, mesoNet.Links)
	for _, generated := range mesoNet.Links {
		assert.True(t, generated.Directed())
		assert.Equal(t, true, generated.GeoFeature().Properties["directed"])
	}
	microNet, err := GenerateMicroscopic(net, mesoNet, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NotEmpty(t, microNet.Links)
	for _, generated := range microNet.Links {
		assert.True(t, generated.Directed())
		assert.Equal(t, true, generated.GeoFeature().Properties["directed"])
	}
	assert.True(t, link.Directed())
	assert.Equal(t, shape, link.Geom())
}
