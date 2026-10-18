package generators

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

func archiveTestSave(t *testing.T, prepared *MesoPreparationResult) []byte {
	t.Helper()
	data, err := SaveMesoUpdateState(context.Background(), prepared.UpdateState())
	assert.NoError(t, err)
	return data
}

func archiveTestRestore(t *testing.T, prepared *MesoPreparationResult, data []byte) *MesoPreparationResult {
	t.Helper()
	restored, err := RestoreMesoUpdateState(context.Background(), prepared.Macro(), prepared.Meso(), data)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	assert.True(t, equalMesoRecords(prepared.Macro(), restored.Macro()))
	assert.True(t, equalMesoRecords(prepared.Meso(), restored.Meso()))
	return restored
}

func TestMesoArchiveContinuation(t *testing.T) {
	for _, kind := range []gmns.IDType{gmns.IDTypeInteger, gmns.IDTypeString} {
		t.Run(string(kind), func(t *testing.T) {
			base := updateTestPreparation(t, kind)
			data := archiveTestSave(t, base)
			restored := archiveTestRestore(t, base, data)
			road, err := gmns.ParseLinkID("13", kind)
			assert.NoError(t, err)
			edit := NewMesoEdit()
			edit.ChangeRightmostLane(road, 1)
			uninterrupted := updateTestApply(t, base, edit)
			resumed := updateTestApply(t, restored, edit)
			assert.Equal(t, uninterrupted, resumed)
			savedAgain := archiveTestSave(t, resumed.Preparation())
			reopened := archiveTestRestore(t, resumed.Preparation(), savedAgain)
			next := NewMesoEdit()
			next.SetRoadFreeSpeed(road, math.Copysign(0, -1))
			assert.Equal(t, updateTestApply(t, uninterrupted.Preparation(), next), updateTestApply(t, reopened, next))
			// An independent preview and Undo to the old snapshot retain their own state.
			assert.Equal(t, resumed, updateTestApply(t, archiveTestRestore(t, base, data), edit))
			assert.Equal(t, savedAgain, archiveTestSave(t, reopened))
		})
	}
}

func TestMesoArchiveDeletedIDs(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	record, _ := base.Macro().Movement(gmns.MovementIDFromInt64(49))
	edit := NewMesoEdit()
	edit.RemoveMovement(record.ID)
	removed := updateTestApply(t, base, edit)
	reopened := archiveTestRestore(t, removed.Preparation(), archiveTestSave(t, removed.Preparation()))
	record.ID = gmns.MovementIDFromInt64(1000)
	edit = NewMesoEdit()
	edit.SetMovement(record)
	uninterrupted := updateTestApply(t, removed.Preparation(), edit)
	resumed := updateTestApply(t, reopened, edit)
	assert.Equal(t, uninterrupted, resumed)
	for _, id := range removed.DeletedLinks() {
		assert.NotContains(t, resumed.CreatedLinks(), id)
	}
}

func TestMesoArchiveRejectsChangedRecords(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	data := archiveTestSave(t, base)
	source := base.Macro().Clone()
	road := gmns.LinkIDFromInt64(13)
	macro.WithFreeSpeed(99)(source.Links[road])
	restored, err := RestoreMesoUpdateState(context.Background(), source, base.Meso(), data)
	assert.Nil(t, restored)
	assert.Error(t, err)
	target := base.Meso().Clone()
	for _, link := range target.Links {
		meso.WithFreeSpeed(99)(link)
		break
	}
	restored, err = RestoreMesoUpdateState(context.Background(), base.Macro(), target, data)
	assert.Nil(t, restored)
	assert.Error(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	restored, err = RestoreMesoUpdateState(ctx, base.Macro(), base.Meso(), data)
	assert.Nil(t, restored)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, data, archiveTestSave(t, base))
	badEdit := NewMesoEdit()
	badEdit.ChangeRightmostLane(road, -1)
	restored = archiveTestRestore(t, base, data)
	result, err := UpdateMesoscopic(context.Background(), restored.Macro(), restored.Meso(), restored.UpdateState(), badEdit)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Equal(t, data, archiveTestSave(t, restored))
}

type archiveMalformedCase struct {
	name   string
	change func(*MesoArchiveSnapshot)
}

func TestMesoArchiveMalformed(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	data := archiveTestSave(t, base)
	cases := []archiveMalformedCase{
		{name: "profile", change: func(state *MesoArchiveSnapshot) {
			state.Metadata.Algorithm = "osm2gmns"
		}},
		{name: "missing section", change: func(state *MesoArchiveSnapshot) {
			state.Metadata.Sections = state.Metadata.Sections[1:]
		}},
		{name: "duplicate section", change: func(state *MesoArchiveSnapshot) {
			state.Metadata.Sections = append(state.Metadata.Sections, state.Metadata.Sections[0])
		}},
		{name: "missing connector", change: func(state *MesoArchiveSnapshot) {
			state.Metadata.Connections = nil
		}},
		{name: "missing offset", change: func(state *MesoArchiveSnapshot) {
			state.Offsets = nil
		}},
		{name: "incorrect offset", change: func(state *MesoArchiveSnapshot) {
			state.Offsets[0].Offset = !state.Offsets[0].Offset
		}},
		{name: "missing reservation", change: func(state *MesoArchiveSnapshot) {
			state.ReservedLinks = nil
		}},
		{name: "kind", change: func(state *MesoArchiveSnapshot) {
			state.IDType = gmns.IDTypeString
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			snapshot := MesoArchiveSnapshot{}
			err := json.Unmarshal(data, &snapshot)
			assert.NoError(t, err)
			test.change(&snapshot)
			snapshot.Integrity = ""
			snapshot.Integrity = archiveDigest(snapshot)
			changed, err := json.Marshal(snapshot)
			assert.NoError(t, err)
			restored, err := RestoreMesoUpdateState(context.Background(), base.Macro(), base.Meso(), changed)
			assert.Nil(t, restored)
			assert.Error(t, err)
		})
	}
	for _, invalid := range []string{`{}`, `{"Version":"a","Version":"b"}`, string(data) + `{}`, `null`, `{"unexpected":1}`, `{"Version":"a","version":"b"}`} {
		restored, err := RestoreMesoUpdateState(context.Background(), base.Macro(), base.Meso(), []byte(invalid))
		assert.Nil(t, restored)
		assert.Error(t, err)
	}
}

func TestMesoArchiveInitialize(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	metadata := base.UpdateState().snapshot.Metadata
	source, target := base.Macro().Clone(), base.Meso().Clone()
	initialized, err := InitializeMesoUpdateState(context.Background(), source, target, &metadata)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.True(t, equalMesoRecords(base.Macro(), source))
	assert.True(t, equalMesoRecords(base.Meso(), target))
	assert.Equal(t, "initialized-current-records", initialized.UpdateState().snapshot.History)
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), 1)
	expected := updateTestApply(t, base, edit)
	actual := updateTestApply(t, initialized, edit)
	assert.Equal(t, expected.Preparation().Meso(), actual.Preparation().Meso())
	assert.Equal(t, expected.UpdatedLinks(), actual.UpdatedLinks())
	archiveTestRestore(t, initialized, archiveTestSave(t, initialized))
	metadata.Algorithm = "ordinary-osm2gmns"
	rejected, err := InitializeMesoUpdateState(context.Background(), source, target, &metadata)
	assert.Nil(t, rejected)
	assert.Error(t, err)
}

func TestMesoArchiveSignedZero(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	macro.WithFreeSpeed(math.Copysign(0, -1))(source.Links[gmns.LinkIDFromInt64(13)])
	base, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	data := archiveTestSave(t, base)
	archiveTestRestore(t, base, data)
	changed := base.Macro().Clone()
	macro.WithFreeSpeed(0)(changed.Links[gmns.LinkIDFromInt64(13)])
	restored, err := RestoreMesoUpdateState(context.Background(), changed, base.Meso(), data)
	assert.Nil(t, restored)
	assert.Error(t, err)
}

// The independent archive runner supplies imported records; this test covers cache-free inputs.
func TestMesoArchiveImportedCaches(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	source, target := base.Macro().Clone(), base.Meso().Clone()
	for _, link := range source.Links {
		macro.WithLineGeomEuclidean(nil)(link)
		macro.WithLanesInfo(macro.LanesInfo{})(link)
	}
	for _, link := range target.Links {
		meso.WithLineGeomEuclidean(nil)(link)
		meso.WithLineMacroLinkID(gmns.LinkID{})(link)
		meso.WithMovementID(gmns.MovementID{})(link)
		meso.WithLineMacroNodeID(gmns.NodeID{})(link)
		meso.WithIsConnection(false)(link)
	}
	for _, record := range source.MovementRecords() {
		movement.WithLanesNum(0)(record)
		err := source.UpdateMovement(record)
		assert.NoError(t, err)
	}
	restored, err := RestoreMesoUpdateState(context.Background(), source, target, archiveTestSave(t, base))
	assert.NoError(t, err)
	if err != nil {
		return
	}
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), 1)
	result, err := UpdateMesoscopic(context.Background(), source, target, restored.UpdateState(), edit)
	assert.NoError(t, err)
	expected := updateTestApply(t, base, edit)
	assert.Equal(t, expected.UpdatedLinks(), result.UpdatedLinks())
	assert.Equal(t, expected.UpdatedNodes(), result.UpdatedNodes())
	assert.Equal(t, archiveRecords(nil, expected.Preparation().Meso()), archiveRecords(nil, result.Preparation().Meso()))
}

func TestMesoArchiveDetailedInitialization(t *testing.T) {
	source := detailedPreparationNetwork(t, false)
	base, err := PrepareMesoscopic(source, WithDetailedLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	restored := archiveTestRestore(t, base, archiveTestSave(t, base))
	initialized, err := InitializeMesoUpdateState(context.Background(), base.Macro(), base.Meso(), &base.UpdateState().snapshot.Metadata)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	edit := NewMesoEdit()
	sections := base.Sections()
	edit.SetRoadFreeSpeed(sections[0].SourceLink(), 33)
	expected := updateTestApply(t, base, edit)
	assert.Equal(t, expected, updateTestApply(t, restored, edit))
	actual := updateTestApply(t, initialized, edit)
	assert.Equal(t, archiveRecords(nil, expected.Preparation().Meso()), archiveRecords(nil, actual.Preparation().Meso()))
}

func TestMesoArchiveSharedGeometry(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	source, target := base.Macro().Clone(), base.Meso().Clone()
	for id, link := range source.Links {
		geometryID, err := gmns.ParseGeometryID(id.String(), id.Type())
		assert.NoError(t, err)
		err = source.AddGeometry(gmns.NewGeometry(geometryID, link.Geom()))
		assert.NoError(t, err)
		macro.WithGeometryID(geometryID)(link)
		macro.WithDirFlag(gmns.DIR_FLAG_FORWARD)(link)
		macro.WithLineGeom(nil)(link)
	}
	for id, link := range target.Links {
		geometryID, err := gmns.ParseGeometryID(id.String(), id.Type())
		assert.NoError(t, err)
		err = target.AddGeometry(gmns.NewGeometry(geometryID, link.Geom()))
		assert.NoError(t, err)
		meso.WithGeometryID(geometryID)(link)
		meso.WithDirFlag(gmns.DIR_FLAG_FORWARD)(link)
		meso.WithLineGeom(nil)(link)
	}
	metadata := base.UpdateState().snapshot.Metadata
	initialized, err := InitializeMesoUpdateState(context.Background(), source, target, &metadata)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, archiveRecords(source, nil), archiveRecords(initialized.Macro(), nil))
	assert.Equal(t, archiveRecords(nil, target), archiveRecords(nil, initialized.Meso()))
	reopened := archiveTestRestore(t, initialized, archiveTestSave(t, initialized))
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), 1)
	actual := updateTestApply(t, reopened, edit)
	assert.Len(t, actual.RecomputedRoads(), 1)
	assert.Len(t, actual.RecomputedConnections(), 1)
	assert.Equal(t, len(target.GeometryRecords()), len(actual.Preparation().Meso().GeometryRecords()))
}

func TestMesoArchiveInitializationRejectsModifiedInheritance(t *testing.T) {
	base := updateTestPreparation(t, gmns.IDTypeInteger)
	source := base.Macro().Clone()
	macro.WithFreeSpeed(77)(source.Links[gmns.LinkIDFromInt64(13)])
	initialized, err := InitializeMesoUpdateState(context.Background(), source, base.Meso(), &base.UpdateState().snapshot.Metadata)
	assert.Nil(t, initialized)
	assert.Error(t, err)
	target := base.Meso().Clone()
	sections := base.Sections()
	meso.WithFreeSpeed(77)(target.Links[sections[0].Link()])
	initialized, err = InitializeMesoUpdateState(context.Background(), base.Macro(), target, &base.UpdateState().snapshot.Metadata)
	assert.Nil(t, initialized)
	assert.Error(t, err)
}

func TestMesoArchiveDefaultAssumptions(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	macro.WithoutLanesNum()(source.Links[gmns.LinkIDFromInt64(13)])
	base, err := PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections(), WithDefaultMesoLanes(3))
	assert.NoError(t, err)
	if err != nil {
		return
	}
	restored := archiveTestRestore(t, base, archiveTestSave(t, base))
	assert.Len(t, restored.LaneAssumptions(), 1)
	edit := NewMesoEdit()
	edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), 1)
	result := updateTestApply(t, restored, edit)
	assert.Empty(t, result.Preparation().LaneAssumptions())
	archiveTestRestore(t, result.Preparation(), archiveTestSave(t, result.Preparation()))
}
