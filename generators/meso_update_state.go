package generators

import (
	"context"
	"fmt"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// MesoUpdateState binds preparation lineage, options and exact records in memory.
// It is immutable and safe to reuse for independent previews. It is application metadata.
// This first profile rejects manual changes to the bound datasets, including extra rows.
type MesoUpdateState struct {
	prepared      *MesoPreparationResult
	layout        preparationLayout
	offsets       map[gmns.LinkID]bool
	reservedLinks map[gmns.LinkID]bool
}

// UpdateState returns the immutable binding created by preparation or an accepted update.
// Imported records without this binding cannot be used for local editing.
func (result *MesoPreparationResult) UpdateState() *MesoUpdateState {
	return result.updateState
}

func (state *MesoUpdateState) Algorithm() string {
	return MesoUpdateAlgorithm
}

// MesoUpdateResult owns a completed transaction and its next source binding.
// Diagnostic IDs are sorted using the dataset's exact identifier kind.
type MesoUpdateResult struct {
	prepared              *MesoPreparationResult
	affectedRoads         []gmns.LinkID
	affectedMovements     []gmns.MovementID
	recomputedRoads       []gmns.LinkID
	recomputedConnections []gmns.LinkID
	createdLinks          []gmns.LinkID
	updatedLinks          []gmns.LinkID
	deletedLinks          []gmns.LinkID
	updatedNodes          []gmns.NodeID
}

func (result *MesoUpdateResult) Preparation() *MesoPreparationResult {
	return result.prepared
}

func (result *MesoUpdateResult) AffectedRoads() []gmns.LinkID {
	return append([]gmns.LinkID(nil), result.affectedRoads...)
}

func (result *MesoUpdateResult) AffectedMovements() []gmns.MovementID {
	return append([]gmns.MovementID(nil), result.affectedMovements...)
}

// RecomputedRoads lists actual offset/cut calculations, excluding copying and validation.
func (result *MesoUpdateResult) RecomputedRoads() []gmns.LinkID {
	return append([]gmns.LinkID(nil), result.recomputedRoads...)
}

// RecomputedConnections lists actual connector geometry calculations.
func (result *MesoUpdateResult) RecomputedConnections() []gmns.LinkID {
	return append([]gmns.LinkID(nil), result.recomputedConnections...)
}

func (result *MesoUpdateResult) CreatedLinks() []gmns.LinkID {
	return append([]gmns.LinkID(nil), result.createdLinks...)
}

func (result *MesoUpdateResult) UpdatedLinks() []gmns.LinkID {
	return append([]gmns.LinkID(nil), result.updatedLinks...)
}

func (result *MesoUpdateResult) DeletedLinks() []gmns.LinkID {
	return append([]gmns.LinkID(nil), result.deletedLinks...)
}

func (result *MesoUpdateResult) UpdatedNodes() []gmns.NodeID {
	return append([]gmns.NodeID(nil), result.updatedNodes...)
}

func cloneMesoPreparation(prepared *MesoPreparationResult) *MesoPreparationResult {
	return &MesoPreparationResult{
		macroNet:        prepared.macroNet.Clone(),
		mesoNet:         prepared.mesoNet.Clone(),
		options:         prepared.options,
		algorithm:       prepared.algorithm,
		sections:        prepared.Sections(),
		connections:     prepared.Connections(),
		laneAssumptions: prepared.LaneAssumptions(),
	}
}

func bindMesoPreparation(prepared *MesoPreparationResult, layout preparationLayout, previous *MesoUpdateState) {
	state := &MesoUpdateState{
		prepared:      cloneMesoPreparation(prepared),
		layout:        layout,
		offsets:       make(map[gmns.LinkID]bool),
		reservedLinks: make(map[gmns.LinkID]bool),
	}
	if previous == nil {
		for id, processing := range mesoRoadProcessing(prepared.macroNet) {
			state.offsets[id] = processing.needsOffset
		}
	} else {
		for id, offset := range previous.offsets {
			state.offsets[id] = offset
		}
		for id := range previous.reservedLinks {
			state.reservedLinks[id] = true
		}
	}
	for id := range prepared.mesoNet.Links {
		state.reservedLinks[id] = true
	}
	prepared.updateState = state
}

func (state *MesoUpdateState) validateBase(ctx context.Context, source *macro.Net, previous *meso.Net) error {
	if ctx == nil {
		return mesoUpdateFailure("invalid_request", "meso update requires a context")
	}
	err := ctx.Err()
	if err != nil {
		return err
	}
	if state == nil || state.prepared == nil || state.prepared.algorithm != MesoPreparationAlgorithm {
		return mesoUpdateFailure("incompatible_lineage", "meso update requires current preparation lineage")
	}
	if !equalMesoRecords(source, state.prepared.macroNet) {
		return mesoUpdateFailure("stale_source", "macro source differs from the bound preparation state")
	}
	if !equalMesoRecords(previous, state.prepared.mesoNet) {
		return manualMesoConflicts(state.prepared.mesoNet, previous)
	}
	return nil
}

func manualMesoConflicts(expected, actual *meso.Net) error {
	conflicts := make([]MesoUpdateConflict, 0)
	if actual != nil {
		links := make(map[gmns.LinkID]*meso.Link)
		nodes := make(map[gmns.NodeID]*meso.Node)
		for id, record := range expected.Links {
			links[id] = record
		}
		for id, record := range actual.Links {
			links[id] = record
		}
		for id, record := range expected.Nodes {
			nodes[id] = record
		}
		for id, record := range actual.Nodes {
			nodes[id] = record
		}
		for _, id := range sortedMesoLinkIDs(links) {
			if !equalMesoRecords(expected.Links[id], actual.Links[id]) {
				conflicts = append(conflicts, MesoUpdateConflict{code: "manual_meso_conflict", target: id, message: fmt.Sprintf("meso link %v differs from the bound preparation state", id)})
			}
		}
		for _, id := range sortedMesoNodeIDs(nodes) {
			if !equalMesoRecords(expected.Nodes[id], actual.Nodes[id]) {
				conflicts = append(conflicts, MesoUpdateConflict{code: "manual_meso_conflict", targetNode: id, message: fmt.Sprintf("meso node %v differs from the bound preparation state", id)})
			}
		}
	}
	if len(conflicts) == 0 {
		conflicts = append(conflicts, MesoUpdateConflict{code: "manual_meso_conflict", message: "meso catalogs or records differ from the bound preparation state"})
	}
	return &MesoUpdateError{conflicts: conflicts}
}
