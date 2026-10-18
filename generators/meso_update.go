package generators

import (
	"context"
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
)

// UpdateMesoscopic applies an explicit edit using a current immutable preparation binding.
// Inputs are never changed. Cancellation or any conflict returns no partial result.
// Geometry work is limited to changed roads and their incident explicit connectors.
func UpdateMesoscopic(ctx context.Context, source *macro.Net, previous *meso.Net, state *MesoUpdateState, edit *MesoEdit) (*MesoUpdateResult, error) {
	err := state.validateBase(ctx, source, previous)
	if err != nil {
		return nil, err
	}
	if edit == nil {
		return nil, mesoUpdateFailure("invalid_request", "meso edit is nil")
	}
	prepared := cloneMesoPreparation(state.prepared)
	result := &MesoUpdateResult{prepared: prepared}
	layout := make(preparationLayout, len(state.layout))
	for id, sections := range state.layout {
		layout[id] = sections
	}
	roads := make([]gmns.LinkID, 0, len(edit.roads))
	for id := range edit.roads {
		roads = append(roads, id)
	}
	sort.Slice(roads, func(first, second int) bool {
		return roads[first].Less(roads[second])
	})
	for _, id := range roads {
		err = ctx.Err()
		if err != nil {
			return nil, err
		}
		road := edit.roads[id]
		link := prepared.macroNet.Links[id]
		if link == nil {
			return nil, mesoUpdateFailure("unknown_road", fmt.Sprintf("macro link %v does not exist", id))
		}
		if road.name != nil {
			macro.WithLinkName(*road.name)(link)
		}
		if road.speedSet {
			macro.WithoutFreeSpeed()(link)
			if road.speed != nil {
				macro.WithFreeSpeed(*road.speed)(link)
			}
		}
		if road.laneSet {
			if road.laneDelta != -1 && road.laneDelta != 1 {
				return nil, mesoUpdateFailure("unsupported_edit", "rightmost lane change must be -1 or +1")
			}
			sections := layout[id]
			if len(sections) != 1 || sections[0].segment != nil || len(sections[0].lanes) != 0 || len(sections[0].segmentLanes) != 0 {
				return nil, mesoUpdateFailure("unsupported_edit", fmt.Sprintf("link %v: lane count edits require a whole road without supplied lane/segment records", id))
			}
			count, _ := link.LanesNum()
			count += road.laneDelta
			if count < 1 || count >= len(CUT_LENGTHS) {
				return nil, mesoUpdateFailure("unsupported_edit", fmt.Sprintf("link %v: resulting lane count must be in 1..99", id))
			}
			macro.WithLanesNum(count)(link)
			info, laneErr := macro.NewLanesInfo(link)
			if laneErr != nil {
				return nil, laneErr
			}
			macro.WithLanesInfo(info)(link)
			section := sections[0]
			section.numbers = sectionLaneNumbers(count, 0, 0)
			layout[id] = []preparationSection{section}
		}
		result.affectedRoads = append(result.affectedRoads, id)
	}
	err = applyMesoMovementEdits(prepared.macroNet, edit)
	if err != nil {
		return nil, err
	}
	err = prepared.macroNet.ValidateDataset()
	if err != nil {
		return nil, mesoUpdateFailure("invalid_source", err.Error())
	}
	conflicts := make([]MesoUpdateConflict, 0)
	for _, record := range prepared.macroNet.MovementRecords() {
		_, revised := edit.movements[record.ID]
		incoming := edit.roads[record.IncomeLink()]
		outgoing := edit.roads[record.OutcomeLink()]
		if !revised && incoming == nil && outgoing == nil {
			continue
		}
		result.affectedMovements = append(result.affectedMovements, record.ID)
		issues := mesoMovementConflicts(record, layout, &prepared.options)
		conflicts = append(conflicts, issues...)
		if len(issues) != 0 {
			continue
		}
		err = prepared.macroNet.UpdateMovement(record)
		if err != nil {
			return nil, mesoUpdateFailure("invalid_movement", err.Error())
		}
	}
	if len(conflicts) != 0 {
		return nil, &MesoUpdateError{conflicts: conflicts}
	}
	for _, id := range roads {
		err = ctx.Err()
		if err != nil {
			return nil, err
		}
		road := edit.roads[id]
		if road.laneDelta != 0 {
			err = recomputeMesoRoad(prepared, id, state.offsets[id])
			if err != nil {
				return nil, err
			}
			result.recomputedRoads = append(result.recomputedRoads, id)
		}
		updateMesoRoadAttributes(prepared, id, road, layout)
	}
	err = updateMesoConnections(ctx, state, result, edit, layout)
	if err != nil {
		return nil, err
	}
	assumptions := make([]LaneCountAssumption, 0)
	for _, assumption := range prepared.laneAssumptions {
		road := edit.roads[assumption.linkID]
		if road == nil || !road.laneSet {
			assumptions = append(assumptions, assumption)
		}
	}
	prepared.laneAssumptions = assumptions
	prepared.sections = preparationLineage(prepared.mesoNet, layout)
	prepared.connections = preparationConnections(prepared.macroNet, prepared.mesoNet, layout)
	err = prepared.mesoNet.ValidateDataset()
	if err != nil {
		return nil, mesoUpdateFailure("invalid_result", err.Error())
	}
	collectMesoChanges(state.prepared.mesoNet, result)
	bindMesoPreparation(prepared, layout, state)
	err = ctx.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func applyMesoMovementEdits(source *macro.Net, edit *MesoEdit) error {
	ids := make(map[gmns.MovementID]bool)
	for id := range edit.removed {
		ids[id] = true
	}
	for id := range edit.movements {
		ids[id] = true
	}
	sorted := sortedMesoEditMovements(ids)
	for _, id := range sorted {
		record, supplied := edit.movements[id]
		old, exists := source.Movement(id)
		if edit.removed[id] {
			if supplied || !exists {
				return mesoUpdateFailure("invalid_request", fmt.Sprintf("movement %v: removal is ambiguous or the record is absent", id))
			}
			err := source.RemoveMovement(id)
			if err != nil {
				return err
			}
			continue
		}
		if record == nil {
			return mesoUpdateFailure("invalid_request", "supplied movement is nil")
		}
		if exists {
			normalized := record.Clone()
			movement.WithoutIncomeLaneStart()(normalized)
			movement.WithoutIncomeLaneEnd()(normalized)
			movement.WithoutOutcomeLaneStart()(normalized)
			movement.WithoutOutcomeLaneEnd()(normalized)
			movement.WithoutIncomeLaneStart()(old)
			movement.WithoutIncomeLaneEnd()(old)
			movement.WithoutOutcomeLaneStart()(old)
			movement.WithoutOutcomeLaneEnd()(old)
			if !equalMesoRecords(normalized, old) {
				return mesoUpdateFailure("unsupported_edit", fmt.Sprintf("movement %v: only explicit lane ranges may change", id))
			}
			err := source.RemoveMovement(id)
			if err != nil {
				return err
			}
		}
		err := source.AddMovement(record)
		if err != nil {
			return mesoUpdateFailure("invalid_movement", err.Error())
		}
	}
	return nil
}

func mesoMovementConflicts(record *movement.Movement, layout preparationLayout, options *mesoPreparationOptions) []MesoUpdateConflict {
	conflicts := make([]MesoUpdateConflict, 0)
	inbound, inPresent := layout[record.IncomeLink()]
	outbound, outPresent := layout[record.OutcomeLink()]
	if !inPresent || !outPresent {
		return []MesoUpdateConflict{{code: "invalid_movement", movement: record.ID, message: "movement references an unknown road"}}
	}
	inStart, inSet := record.IncomeLaneStart()
	inEnd, endSet := record.IncomeLaneEnd()
	if !endSet {
		inEnd = inStart
	}
	outStart, outSet := record.OutcomeLaneStart()
	outEnd, endSet := record.OutcomeLaneEnd()
	if !endSet {
		outEnd = outStart
	}
	inNumbers := inbound[len(inbound)-1].numbers
	outNumbers := outbound[0].numbers
	inFirst, inLast := lanePosition(inNumbers, inStart), lanePosition(inNumbers, inEnd)
	outFirst, outLast := lanePosition(outNumbers, outStart), lanePosition(outNumbers, outEnd)
	if !inSet || inFirst < 0 || inLast < inFirst {
		conflicts = append(conflicts, MesoUpdateConflict{code: "movement_resolution_required", source: record.IncomeLink(), movement: record.ID, side: "inbound", message: fmt.Sprintf("movement %v: inbound range %d..%d is outside link %v lanes %v", record.ID, inStart, inEnd, record.IncomeLink(), inNumbers)})
	}
	if !outSet || outFirst < 0 || outLast < outFirst {
		conflicts = append(conflicts, MesoUpdateConflict{code: "movement_resolution_required", source: record.OutcomeLink(), movement: record.ID, side: "outbound", message: fmt.Sprintf("movement %v: outbound range %d..%d is outside link %v lanes %v", record.ID, outStart, outEnd, record.OutcomeLink(), outNumbers)})
	}
	if len(conflicts) != 0 {
		return conflicts
	}
	err := prepareMovementLanes(record, layout, options)
	if err != nil {
		conflicts = append(conflicts, MesoUpdateConflict{code: "movement_resolution_required", movement: record.ID, side: "both", message: err.Error()})
	}
	return conflicts
}

func sortedMesoEditMovements(ids map[gmns.MovementID]bool) []gmns.MovementID {
	sorted := make([]gmns.MovementID, 0, len(ids))
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Slice(sorted, func(first, second int) bool {
		return sorted[first].Less(sorted[second])
	})
	return sorted
}

func collectMesoChanges(previous *meso.Net, result *MesoUpdateResult) {
	next := result.prepared.mesoNet
	for _, id := range sortedMesoLinkIDs(previous.Links) {
		if next.Links[id] == nil {
			result.deletedLinks = append(result.deletedLinks, id)
		}
	}
	for _, id := range sortedMesoLinkIDs(next.Links) {
		if previous.Links[id] == nil {
			result.createdLinks = append(result.createdLinks, id)
		} else if !equalMesoRecords(previous.Links[id], next.Links[id]) {
			result.updatedLinks = append(result.updatedLinks, id)
		}
	}
	for _, id := range sortedMesoNodeIDs(next.Nodes) {
		if !equalMesoRecords(previous.Nodes[id], next.Nodes[id]) {
			result.updatedNodes = append(result.updatedNodes, id)
		}
	}
}

func sortMesoUpdateLinkIDs(ids []gmns.LinkID) {
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
}
