package generators

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func updateTestPreparation(t *testing.T, kind gmns.IDType) *MesoPreparationResult {
	t.Helper()
	source := preparationTestNetwork(t, kind)
	prepared, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	return prepared
}

func updateTestApply(t *testing.T, base *MesoPreparationResult, edit *MesoEdit) *MesoUpdateResult {
	t.Helper()
	source := base.Macro().Clone()
	previous := base.Meso().Clone()
	result, err := UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), base.UpdateState(), edit)
	assert.NoError(t, err)
	assert.Equal(t, source, base.Macro())
	assert.Equal(t, previous, base.Meso())
	if err != nil {
		t.FailNow()
	}
	assert.NoError(t, result.Preparation().Macro().ValidateDataset())
	assert.NoError(t, result.Preparation().Meso().ValidateDataset())
	return result
}

func assertUpdateOracle(t *testing.T, prepared *MesoPreparationResult) {
	t.Helper()
	options := []MesoPreparationOption{WithOrderedLaneConnections()}
	if prepared.options.detailedLaneLayout {
		options = append(options, WithDetailedLaneLayout())
	} else {
		options = append(options, WithUniformLaneLayout())
	}
	oracle, err := PrepareMesoscopic(prepared.Macro(), options...)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, oracle.Meso(), prepared.Meso())
	assert.Equal(t, oracle.Sections(), prepared.Sections())
	assert.Equal(t, oracle.Connections(), prepared.Connections())
}

func TestMesoUpdateAttributesWithoutGeometry(t *testing.T) {
	for _, kind := range []gmns.IDType{gmns.IDTypeInteger, gmns.IDTypeString} {
		t.Run(string(kind), func(t *testing.T) {
			base := updateTestPreparation(t, kind)
			road, err := gmns.ParseLinkID("13", kind)
			assert.NoError(t, err)
			edit := NewMesoEdit()
			edit.RenameRoad(road, "New name")
			edit.SetRoadFreeSpeed(road, 0)
			result := updateTestApply(t, base, edit)
			assert.Empty(t, result.RecomputedRoads())
			assert.Empty(t, result.RecomputedConnections())
			assert.Empty(t, result.UpdatedNodes())
			assertUpdateOracle(t, result.Preparation())
			for id, link := range base.Meso().Links {
				assert.Equal(t, link.Geom(), result.Preparation().Meso().Links[id].Geom())
				assert.Equal(t, link.GeomEuclidean(), result.Preparation().Meso().Links[id].GeomEuclidean())
			}
			edit = NewMesoEdit()
			edit.ClearRoadFreeSpeed(road)
			cleared := updateTestApply(t, result.Preparation(), edit)
			assert.Empty(t, cleared.RecomputedRoads())
			assert.Empty(t, cleared.RecomputedConnections())
			assertUpdateOracle(t, cleared.Preparation())
		})
	}
}

func TestMesoUpdateLaneConflictAndExplicitResolution(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	road := gmns.LinkIDFromInt64(13)
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(road, -1)
	result, err := UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), base.UpdateState(), edit)
	assert.Nil(t, result)
	var failure *MesoUpdateError
	assert.True(t, errors.As(err, &failure))
	if failure == nil {
		return
	}
	conflicts := failure.Conflicts()
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "movement_resolution_required", conflicts[0].Code())
	assert.Equal(t, "inbound", conflicts[0].Side())
	assert.Equal(t, gmns.MovementIDFromInt64(49), conflicts[0].Movement())
	assert.Equal(t, road, conflicts[0].SourceLink())
	record, _ := base.Macro().Movement(gmns.MovementIDFromInt64(49))
	movement.WithIncomeLane(2, 2)(record)
	edit.SetMovement(record)
	accepted := updateTestApply(t, base, edit)
	assert.Equal(t, []gmns.LinkID{road}, accepted.RecomputedRoads())
	assert.Len(t, accepted.RecomputedConnections(), 1)
	assert.Empty(t, accepted.CreatedLinks())
	assert.Empty(t, accepted.DeletedLinks())
	assertUpdateOracle(t, accepted.Preparation())
	repeated := updateTestApply(t, base, edit)
	assert.Equal(t, accepted, repeated)
}

func TestMesoUpdateReverseParallelAndUnrelatedGeometry(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeString)
	road := source.Links[gmns.LinkIDFromString("13")]
	reverseID := gmns.LinkIDFromString("013")
	shape := road.Geom().Clone()
	shape.Reverse()
	reverse := macro.NewLinkFrom(reverseID, road.TargetNode(), road.SourceNode(),
		macro.WithLineGeom(shape), macro.WithLanesNum(2), macro.WithLengthMeters(128), macro.WithFreeSpeed(17),
	)
	source.Links[reverseID] = reverse
	// A parallel edge exercises the existing ordered reverse-hash pairing rule.
	parallel := road.Clone()
	parallel.ID = gmns.LinkIDFromString("14")
	source.Links[parallel.ID] = parallel
	geometryID := gmns.GeometryIDFromString("001")
	err := source.AddGeometry(gmns.NewGeometry(geometryID, road.Geom()))
	assert.NoError(t, err)
	macro.WithGeometryID(geometryID)(road)
	macro.WithDirFlag(gmns.DIR_FLAG_FORWARD)(road)
	macro.WithLineGeom(nil)(road)
	base, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(road.ID, 1)
	edit.SetRoadFreeSpeed(road.ID, 0)
	result := updateTestApply(t, base, edit)
	assert.Equal(t, []gmns.LinkID{road.ID}, result.RecomputedRoads())
	assertUpdateOracle(t, result.Preparation())
	for _, section := range base.Sections() {
		if section.SourceLink() == road.ID {
			continue
		}
		id := section.Link()
		assert.Equal(t, base.Meso().Links[id], result.Preparation().Meso().Links[id])
	}
	edit = NewMesoEdit()
	edit.ChangeRightmostLane(road.ID, 1)
	edit.ChangeRightmostLane(reverseID, 1)
	edit.RenameRoad(road.ID, "Outbound")
	edit.RenameRoad(reverseID, "Inbound")
	both := updateTestApply(t, result.Preparation(), edit)
	assert.Len(t, both.RecomputedRoads(), 2)
	assertUpdateOracle(t, both.Preparation())
	speed, present := both.Preparation().Macro().Links[reverseID].FreeSpeed()
	assert.True(t, present)
	assert.Equal(t, 17.0, speed)
}

func TestMesoUpdateRemovalAndAllocationKeepIdentity(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeString)
	original := base.Connections()[0]
	edit := NewMesoEdit()
	edit.RemoveMovement(original.Movement())
	removed := updateTestApply(t, base, edit)
	assert.Equal(t, []gmns.LinkID{original.Link()}, removed.DeletedLinks())
	assert.Empty(t, removed.RecomputedRoads())
	assert.Empty(t, removed.RecomputedConnections())
	record, _ := base.Macro().Movement(original.Movement())
	record.ID = gmns.MovementIDFromString("001")
	edit = NewMesoEdit()
	edit.SetMovement(record)
	added := updateTestApply(t, removed.Preparation(), edit)
	assert.Len(t, added.CreatedLinks(), 1)
	assert.NotEqual(t, original.Link(), added.CreatedLinks()[0])
	assert.Empty(t, added.RecomputedRoads())
	assert.Len(t, added.RecomputedConnections(), 1)
	for _, section := range base.Sections() {
		assert.Equal(t, base.Meso().Links[section.Link()], added.Preparation().Meso().Links[section.Link()])
	}
	oracle, err := PrepareMesoscopic(added.Preparation().Macro(), WithUniformLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	// Movement identity is the correspondence when allocator history differs from a full build.
	want := oracle.Meso().Links[oracle.Connections()[0].Link()].Clone()
	actual := added.Preparation().Meso().Links[added.CreatedLinks()[0]]
	want.ID = actual.ID
	assert.Equal(t, want, actual)
}

func TestMesoUpdateRejectsStaleManualAndCancelledInputs(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	edit := NewMesoEdit()
	edit.RenameRoad(gmns.LinkIDFromInt64(13), "changed")
	manual := base.Meso().Clone()
	for _, link := range manual.Links {
		meso.WithFreeSpeed(7)(link)
	}
	result, err := UpdateMesoscopic(context.Background(), base.Macro(), manual, base.UpdateState(), edit)
	assert.Nil(t, result)
	var failure *MesoUpdateError
	assert.True(t, errors.As(err, &failure))
	assert.Equal(t, "manual_meso_conflict", failure.Conflicts()[0].Code())
	stale := base.Macro().Clone()
	macro.WithFreeSpeed(5)(stale.Links[gmns.LinkIDFromInt64(13)])
	result, err = UpdateMesoscopic(context.Background(), stale, base.Meso(), base.UpdateState(), edit)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bound preparation state")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err = UpdateMesoscopic(ctx, base.Macro(), base.Meso(), base.UpdateState(), edit)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
	result, err = UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), nil, edit)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "current preparation lineage")
	// Returned records cannot alias the input or its immutable binding.
	accepted := updateTestApply(t, base, edit)
	for _, link := range accepted.Preparation().Meso().Links {
		link.Geom()[0] = orb.Point{0, 0}
	}
	again := updateTestApply(t, base, edit)
	assertUpdateOracle(t, again.Preparation())
}

func TestMesoUpdateDetailedSpeedOverridesAndUnsupportedLaneEdit(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	road := source.Links[gmns.LinkIDFromInt64(13)]
	segment := gmns.NewSegment(gmns.SegmentIDFromInt64(1), road.ID, road.SourceNode(), 64, 128, gmns.WithSegmentFreeSpeed(0))
	err := source.AddSegment(segment)
	assert.NoError(t, err)
	base, err := PrepareMesoscopic(source, WithDetailedLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	edit := NewMesoEdit()
	edit.SetRoadFreeSpeed(road.ID, 42)
	edit.RenameRoad(road.ID, "Detailed road")
	result := updateTestApply(t, base, edit)
	assertUpdateOracle(t, result.Preparation())
	assert.Empty(t, result.RecomputedRoads())
	assert.Empty(t, result.RecomputedConnections())
	assert.Equal(t, base.Macro().SegmentRecords(), result.Preparation().Macro().SegmentRecords())
	connection := result.Preparation().Connections()[0]
	speed, present := result.Preparation().Meso().Links[connection.Link()].FreeSpeed()
	assert.True(t, present)
	assert.Zero(t, speed)
	edit = NewMesoEdit()
	edit.ClearRoadFreeSpeed(road.ID)
	cleared := updateTestApply(t, result.Preparation(), edit)
	assertUpdateOracle(t, cleared.Preparation())
	edit = NewMesoEdit()
	edit.ChangeRightmostLane(road.ID, 1)
	rejected, err := UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), base.UpdateState(), edit)
	assert.Nil(t, rejected)
	assert.Contains(t, err.Error(), "without supplied lane/segment records")
}

func TestMesoUpdateAllLaneSidesAndDisconnectedRoad(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	original, _ := source.Movement(gmns.MovementIDFromInt64(49))
	second := original.Clone()
	second.ID = gmns.MovementIDFromInt64(51)
	movement.WithOutcomeLane(3, 3)(second)
	err := source.AddMovement(second)
	assert.NoError(t, err)
	startID, endID := gmns.NodeIDFromInt64(99), gmns.NodeIDFromInt64(100)
	start, end := orb.Point{38, 55}, orb.Point{38.001, 55}
	source.Nodes[startID] = macro.NewNodeFrom(startID, macro.WithPointGeom(start))
	source.Nodes[endID] = macro.NewNodeFrom(endID, macro.WithPointGeom(end))
	isolated := gmns.LinkIDFromInt64(100)
	source.Links[isolated] = macro.NewLinkFrom(isolated, startID, endID,
		macro.WithLineGeom(orb.LineString{start, end}), macro.WithLanesNum(1), macro.WithLengthMeters(64),
	)
	base, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(original.IncomeLink(), -1)
	edit.ChangeRightmostLane(original.OutcomeLink(), -1)
	rejected, err := UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), base.UpdateState(), edit)
	assert.Nil(t, rejected)
	var failure *MesoUpdateError
	assert.True(t, errors.As(err, &failure))
	assert.Len(t, failure.Conflicts(), 3)
	original, _ = base.Macro().Movement(original.ID)
	second, _ = base.Macro().Movement(second.ID)
	movement.WithIncomeLane(1, 1)(original)
	movement.WithIncomeLane(2, 2)(second)
	movement.WithOutcomeLane(2, 2)(second)
	edit.SetMovement(original)
	edit.SetMovement(second)
	accepted := updateTestApply(t, base, edit)
	assertUpdateOracle(t, accepted.Preparation())
	assert.Len(t, accepted.RecomputedRoads(), 2)
	for _, section := range base.Sections() {
		if section.SourceLink() != isolated {
			continue
		}
		old := base.Meso().Links[section.Link()]
		assert.Equal(t, old, accepted.Preparation().Meso().Links[section.Link()])
		assert.Equal(t, base.Meso().Nodes[old.SourceNode()], accepted.Preparation().Meso().Nodes[old.SourceNode()])
		assert.Equal(t, base.Meso().Nodes[old.TargetNode()], accepted.Preparation().Meso().Nodes[old.TargetNode()])
	}
}

func TestMesoUpdateExtraRecordsAndInvalidLaneDelta(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	previous := base.Meso().Clone()
	extra := previous.Links[gmns.LinkIDFromInt64(0)].Clone()
	extra.ID = gmns.LinkIDFromInt64(9007199254740993)
	previous.Links[extra.ID] = extra
	edit := NewMesoEdit()
	edit.RenameRoad(gmns.LinkIDFromInt64(13), "name")
	result, err := UpdateMesoscopic(context.Background(), base.Macro(), previous, base.UpdateState(), edit)
	assert.Nil(t, result)
	var failure *MesoUpdateError
	assert.True(t, errors.As(err, &failure))
	assert.Equal(t, extra.ID, failure.Conflicts()[0].TargetLink())
	assert.Equal(t, extra, previous.Links[extra.ID])
	edit = NewMesoEdit()
	edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), 0)
	result, err = UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), base.UpdateState(), edit)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "-1 or +1")
}

func TestMesoUpdateConcurrentPreviews(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	done := make(chan error, 2)
	for _, name := range []string{"first", "second"} {
		go func(value string) {
			edit := NewMesoEdit()
			edit.RenameRoad(gmns.LinkIDFromInt64(13), value)
			_, err := UpdateMesoscopic(context.Background(), base.Macro(), base.Meso(), base.UpdateState(), edit)
			done <- err
		}(name)
	}
	assert.NoError(t, <-done)
	assert.NoError(t, <-done)
	assert.Equal(t, "", base.Macro().Links[gmns.LinkIDFromInt64(13)].Name())
}

func TestMesoUpdateBindingPreservesSignedZero(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	edit := NewMesoEdit()
	edit.SetRoadFreeSpeed(gmns.LinkIDFromInt64(13), 0)
	initial := updateTestApply(t, base, edit).Preparation()
	changed := initial.Macro().Clone()
	macro.WithFreeSpeed(math.Copysign(0, -1))(changed.Links[gmns.LinkIDFromInt64(13)])
	result, err := UpdateMesoscopic(context.Background(), changed, initial.Meso(), initial.UpdateState(), NewMesoEdit())
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bound preparation state")
	edit = NewMesoEdit()
	edit.SetRoadFreeSpeed(gmns.LinkIDFromInt64(13), math.Copysign(0, -1))
	accepted := updateTestApply(t, initial, edit)
	assert.Len(t, accepted.UpdatedLinks(), 2)
}
