package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/stretchr/testify/assert"
)

func TestPrepareMesoscopicCompletesOnlyMissingLanes(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeString)
	expectedSource := preparationTestNetwork(t, gmns.IDTypeString)
	identifier := gmns.LinkIDFromString("13")
	macro.WithoutLanesNum()(source.Links[identifier])
	macro.WithoutLanesNum()(expectedSource.Links[identifier])
	result, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithDefaultMesoLanes(4))
	assert.NoError(t, err)
	assert.Equal(t, expectedSource, source)
	if err != nil {
		return
	}
	assert.NoError(t, result.Macro().ValidateDataset())
	assert.NoError(t, result.Meso().ValidateDataset())
	lanes, present := result.Macro().Links[identifier].LanesNum()
	assert.True(t, present)
	assert.Equal(t, 4, lanes)
	other := gmns.LinkIDFromString("900719925474099312345")
	lanes, present = result.Macro().Links[other].LanesNum()
	assert.True(t, present)
	assert.Equal(t, 3, lanes)
	assumptions := result.LaneAssumptions()
	assert.Len(t, assumptions, 1)
	if len(assumptions) != 1 {
		return
	}
	assert.Equal(t, identifier, assumptions[0].Link())
	assert.Equal(t, 4, assumptions[0].Lanes())
	assumptions[0].lanes = 999
	assert.Equal(t, 4, result.LaneAssumptions()[0].lanes)
	for _, link := range result.Meso().Links {
		if !link.Movement().IsZero() {
			assert.Equal(t, 2, link.MovementIncomeLaneStartSeqID())
		}
	}
	// Saving the completed macro records makes a later rebuild independent of the default.
	rebuilt, err := PrepareMesoscopic(result.Macro(), WithUniformLaneLayout())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, result.Meso(), rebuilt.Meso())
	assert.Empty(t, rebuilt.LaneAssumptions())
	result.Macro().Links[identifier].Geom()[0][0] = 0
	assert.Equal(t, expectedSource, source)
}

func TestDefaultLanesDoNotRepairInvalidValuesOrReferences(t *testing.T) {
	for _, lanes := range []int{0, -1} {
		source := preparationTestNetwork(t, gmns.IDTypeInteger)
		identifier := gmns.LinkIDFromInt64(13)
		macro.WithLanesNum(lanes)(source.Links[identifier])
		result, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithDefaultMesoLanes(4))
		assert.Nil(t, result)
		assertPreparationError(t, err, "link 13")
		stored, present := source.Links[identifier].LanesNum()
		assert.True(t, present)
		assert.Equal(t, lanes, stored)
	}
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	identifier := gmns.LinkIDFromInt64(13)
	macro.WithoutLanesNum()(source.Links[identifier])
	result, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithDefaultMesoLanes(2))
	assert.Nil(t, result)
	assertPreparationError(t, err, "movement 49: inbound lane range 3..3 is outside link 13 lanes 1..2")
	_, present := source.Links[identifier].LanesNum()
	assert.False(t, present)
	for _, count := range []int{-1, 0, 100} {
		result, err = PrepareMesoscopic(source, WithUniformLaneLayout(), WithDefaultMesoLanes(count))
		assert.Nil(t, result)
		assertPreparationError(t, err, "default meso lanes must be in")
	}
}
