package generators

import (
	"context"
	"math"
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/micro"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestGenerationDoesNotMoveEndpointAltitude(t *testing.T) {
	net := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	for _, node := range net.Nodes {
		macro.WithAltitudeMeters(-12.5)(node)
	}
	before := net.Clone()
	movements, err := GenerateMovements(net)
	if !assert.NoError(t, err) {
		return
	}
	mesoNet, err := GenerateMesoscopic(net, movements)
	if !assert.NoError(t, err) {
		return
	}
	displaced := 0
	for _, node := range mesoNet.Nodes {
		parent := net.Nodes[node.MacroNode()]
		value, present := node.AltitudeMeters()
		if parent != nil && parent.Geom() == node.Geom() {
			assert.True(t, present)
			assert.Equal(t, -12.5, value)
		} else {
			displaced++
			assert.False(t, present)
		}
	}
	assert.Greater(t, displaced, 0)
	microNet, err := GenerateMicroscopic(net, mesoNet, movements)
	if !assert.NoError(t, err) {
		return
	}
	for _, node := range microNet.Nodes {
		value, present := node.AltitudeMeters()
		if present {
			matched := false
			for _, parent := range net.Nodes {
				matched = matched || parent.Geom() == node.Geom()
			}
			assert.True(t, matched)
			assert.Equal(t, -12.5, value)
		}
	}
	assert.Equal(t, before, net)
}

func TestExactMicroEndpointRetainsSignedZero(t *testing.T) {
	point := orb.Point{37, 55}
	source := macro.NewNodeFrom(gmns.NodeIDFromInt64(1), macro.WithPointGeom(point), macro.WithAltitudeMeters(math.Copysign(0, -1)))
	target := micro.NewNodeFrom(gmns.NodeIDFromInt64(2), micro.WithPointGeom(point))
	copyMacroEndpoint(source, target)
	value, present := target.AltitudeMeters()
	assert.True(t, present)
	assert.True(t, math.Signbit(value))
}

func TestHistoricalAltitudeSurvivesRestoreButCannotBeMoved(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	base, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	if !assert.NoError(t, err) {
		return
	}
	for _, node := range base.Meso().Nodes {
		_, present := node.AltitudeMeters()
		assert.False(t, present, "cut endpoints have no measured height")
	}
	archiveTestRestore(t, base, archiveTestSave(t, base))
	// Previous releases copied endpoint altitude to displaced nodes. Preserve saved values.
	target := base.Meso().Clone()
	for _, node := range target.Nodes {
		parent := source.Nodes[node.MacroNode()]
		if parent != nil {
			value, present := parent.AltitudeMeters()
			if present {
				meso.WithAltitudeMeters(value)(node)
			}
		}
	}
	metadata := base.UpdateState().snapshot.Metadata
	historical, err := InitializeMesoUpdateState(context.Background(), source, target, &metadata)
	if !assert.NoError(t, err) {
		return
	}
	restored := archiveTestRestore(t, historical, archiveTestSave(t, historical))
	before := restored.Meso().Clone()
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), 1)
	result, err := UpdateMesoscopic(context.Background(), restored.Macro(), restored.Meso(), restored.UpdateState(), edit)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, before, restored.Meso())
}

func TestMicroRejectsAuthoredMesoAltitude(t *testing.T) {
	net := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	movements, err := GenerateMovements(net)
	if !assert.NoError(t, err) {
		return
	}
	mesoNet, err := GenerateMesoscopic(net, movements)
	if !assert.NoError(t, err) {
		return
	}
	for _, node := range mesoNet.Nodes {
		meso.WithAltitudeMeters(12.5)(node)
		break
	}
	before := mesoNet.Clone()
	result, err := GenerateMicroscopic(net, mesoNet, movements)
	assert.ErrorIs(t, err, ErrUnsupportedGenerationInput)
	assert.Nil(t, result)
	assert.Equal(t, before, mesoNet)
}
