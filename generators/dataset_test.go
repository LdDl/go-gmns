package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

func TestDatasetCatalogsAndAttributesSurviveGeneration(t *testing.T) {
	net := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	assert.NoError(t, net.SetConfig(gmns.NewConfig(gmns.IDTypeInteger, gmns.WithDatasetName("source"))))
	zoneID := gmns.NodeIDFromInt64(9)
	assert.NoError(t, net.AddZone(gmns.NewZone(zoneID, gmns.WithZoneName("boundary"))))
	first := net.Nodes[gmns.NodeIDFromInt64(1)]
	macro.WithZoneID(zoneID)(first)
	macro.WithNodeName("Start")(first)
	macro.WithNodeType("intersection")(first)
	macro.WithAltitudeMeters(42)(first)
	source := net.Links[gmns.LinkIDFromInt64(10)]
	macro.WithLinkName("Road")(source)
	macro.WithGrade(-5)(source)
	macro.WithJurisdiction("city")(source)
	macro.WithRightOfWayWidthMeters(15)(source)
	assert.NoError(t, net.ValidateDataset())
	movements, err := GenerateMovements(net)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	for _, mvmt := range movements {
		assert.NoError(t, net.AddMovement(mvmt))
	}
	mesoNet, err := GenerateMesoscopic(net, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NoError(t, mesoNet.ValidateDataset())
	assert.Empty(t, mesoNet.MovementRecords(), "source movements are not meso-owned records")
	assert.Equal(t, net.Config(), mesoNet.Config())
	assert.Len(t, mesoNet.ZoneRecords(), 1)
	for _, link := range mesoNet.Links {
		if link.MacroLink() == source.ID {
			assert.Equal(t, "Road", link.Name())
			grade, present := link.Grade()
			assert.True(t, present)
			assert.Equal(t, -5.0, grade)
		}
	}
	microNet, err := GenerateMicroscopic(net, mesoNet, movements)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.NoError(t, microNet.ValidateDataset())
	assert.Empty(t, microNet.MovementRecords())
	assert.Len(t, microNet.ZoneRecords(), 1)
	for _, link := range microNet.Links {
		lanes, present := link.LanesNum()
		assert.True(t, present)
		assert.Equal(t, 1, lanes)
	}
	assert.NoError(t, net.ValidateDataset(), "generation must leave the source valid")
}

func TestGeneratorsDoNotDiscardAuthoredDetail(t *testing.T) {
	net := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	lane := gmns.NewLane(gmns.LaneIDFromInt64(1), gmns.LinkIDFromInt64(10), -1)
	assert.NoError(t, net.AddLane(lane))
	_, err := GenerateMovements(net)
	assert.ErrorIs(t, err, ErrUnsupportedGenerationInput)
	_, err = GenerateMesoscopic(net, movement.NewMovementsStorage())
	assert.ErrorIs(t, err, ErrUnsupportedGenerationInput)
	assert.Len(t, net.LaneRecords(), 1)
}

func TestPreparedMovementsAreNotRegenerated(t *testing.T) {
	net := controlTestNetwork(t, types.CONTROL_TYPE_IS_SIGNAL)
	mvmt := movement.NewMovement(gmns.MovementIDFromInt64(99), gmns.NodeIDFromInt64(2), gmns.LinkIDFromInt64(10), gmns.LinkIDFromInt64(11), "SBT", "thru", movement.WithPenalty(12.5), movement.WithCapacity(777.25))
	assert.NoError(t, net.AddMovement(mvmt))
	result, err := GenerateMovements(net)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Len(t, result, 1)
	actual := result[mvmt.ID]
	if !assert.NotNil(t, actual) {
		return
	}
	penalty, present := actual.Penalty()
	assert.True(t, present)
	assert.Equal(t, 12.5, penalty)
	movement.WithPenalty(0)(actual)
	stored, found := net.Movement(mvmt.ID)
	assert.True(t, found)
	penalty, present = stored.Penalty()
	assert.True(t, present)
	assert.Equal(t, 12.5, penalty)
}
