package generators

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/movement"
)

// MesoPreparationAlgorithm identifies the geometry and inheritance rules.
const MesoPreparationAlgorithm = "standard-macro-sections-3"

// MesoUpdateAlgorithm identifies the supported local editing profile.
const MesoUpdateAlgorithm = "local-macro-roads-1"

// MesoEdit describes one explicit transaction against a bound preparation state.
// It cannot change source geometry, endpoints, catalogs or detailed lane tables.
type MesoEdit struct {
	roads     map[gmns.LinkID]*mesoRoadEdit
	movements map[gmns.MovementID]*movement.Movement
	removed   map[gmns.MovementID]bool
}

type mesoRoadEdit struct {
	name      *string
	speed     *float64
	speedSet  bool
	laneDelta int
	laneSet   bool
}

func NewMesoEdit() *MesoEdit {
	return &MesoEdit{
		roads:     make(map[gmns.LinkID]*mesoRoadEdit),
		movements: make(map[gmns.MovementID]*movement.Movement),
		removed:   make(map[gmns.MovementID]bool),
	}
}

func (edit *MesoEdit) road(id gmns.LinkID) *mesoRoadEdit {
	if edit.roads == nil {
		edit.roads = make(map[gmns.LinkID]*mesoRoadEdit)
	}
	if edit.roads[id] == nil {
		edit.roads[id] = &mesoRoadEdit{}
	}
	return edit.roads[id]
}

func (edit *MesoEdit) RenameRoad(id gmns.LinkID, name string) {
	edit.road(id).name = &name
}

func (edit *MesoEdit) SetRoadFreeSpeed(id gmns.LinkID, speed float64) {
	road := edit.road(id)
	road.speedSet = true
	road.speed = &speed
}

func (edit *MesoEdit) ClearRoadFreeSpeed(id gmns.LinkID) {
	road := edit.road(id)
	road.speedSet = true
	road.speed = nil
}

// ChangeRightmostLane accepts +1 or -1 over one complete directed road.
// Supplied detailed lane records on that road require a different editing profile.
func (edit *MesoEdit) ChangeRightmostLane(id gmns.LinkID, delta int) {
	road := edit.road(id)
	road.laneDelta = delta
	road.laneSet = true
}

// SetMovement supplies an explicit new movement or revised lane ranges of an existing one.
// Existing movements must retain all other standard and calculated properties.
func (edit *MesoEdit) SetMovement(record *movement.Movement) {
	if edit.movements == nil {
		edit.movements = make(map[gmns.MovementID]*movement.Movement)
	}
	if record == nil {
		edit.movements[gmns.MovementID{}] = nil
		return
	}
	edit.movements[record.ID] = record.Clone()
}

// RemoveMovement removes an explicit connection; absence does not encode a legal prohibition.
func (edit *MesoEdit) RemoveMovement(id gmns.MovementID) {
	if edit.removed == nil {
		edit.removed = make(map[gmns.MovementID]bool)
	}
	edit.removed[id] = true
}

// MesoUpdateConflict identifies an operation that needs caller resolution.
type MesoUpdateConflict struct {
	targetNode gmns.NodeID
	code       string
	source     gmns.LinkID
	movement   gmns.MovementID
	target     gmns.LinkID
	side       string
	message    string
}

func (conflict *MesoUpdateConflict) Code() string {
	return conflict.code
}

func (conflict *MesoUpdateConflict) SourceLink() gmns.LinkID {
	return conflict.source
}

func (conflict *MesoUpdateConflict) Movement() gmns.MovementID {
	return conflict.movement
}

func (conflict *MesoUpdateConflict) TargetLink() gmns.LinkID {
	return conflict.target
}

func (conflict *MesoUpdateConflict) TargetNode() gmns.NodeID {
	return conflict.targetNode
}

// Side returns inbound, outbound, both, or an empty value for a non-lane conflict.
func (conflict *MesoUpdateConflict) Side() string {
	return conflict.side
}

func (conflict *MesoUpdateConflict) Message() string {
	return conflict.message
}

// MesoUpdateError reports structured conflicts without a partially accepted result.
type MesoUpdateError struct {
	conflicts []MesoUpdateConflict
}

func (failure *MesoUpdateError) Error() string {
	if len(failure.conflicts) == 0 {
		return "meso update rejected"
	}
	return failure.conflicts[0].message
}

func (failure *MesoUpdateError) Conflicts() []MesoUpdateConflict {
	return append([]MesoUpdateConflict(nil), failure.conflicts...)
}

func mesoUpdateFailure(code, message string) error {
	return &MesoUpdateError{conflicts: []MesoUpdateConflict{{code: code, message: message}}}
}
