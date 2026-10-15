package meso

import (
	"fmt"

	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
)

// ValidateSourceReferences resolves application provenance separately from standard parents.
func (net *Net) ValidateSourceReferences(source *macro.Net, movements movement.MovementsStorage) error {
	err := net.ValidateReferences()
	if err != nil {
		return err
	}
	err = source.ValidateMovements(movements)
	if err != nil {
		return err
	}
	for _, identifier := range net.sortedNodeIDs() {
		node := net.Nodes[identifier]
		macroNodeID := node.MacroNode()
		if !macroNodeID.IsZero() {
			if source.Nodes[macroNodeID] == nil {
				return fmt.Errorf("meso node %v: MacroNode %q: %w", identifier, macroNodeID, macro.ErrNodeNotFound)
			}
		}
		macroLinkID := node.MacroLink()
		if !macroLinkID.IsZero() {
			if source.Links[macroLinkID] == nil {
				return fmt.Errorf("meso node %v: MacroLink %q: %w", identifier, macroLinkID, macro.ErrLinkNotFound)
			}
		}
	}
	for _, identifier := range net.sortedLinkIDs() {
		link := net.Links[identifier]
		macroNodeID := link.MacroNode()
		if !macroNodeID.IsZero() {
			if source.Nodes[macroNodeID] == nil {
				return fmt.Errorf("meso link %v: MacroNode %q: %w", identifier, macroNodeID, macro.ErrNodeNotFound)
			}
		}
		macroLinkID := link.MacroLink()
		if !macroLinkID.IsZero() {
			if source.Links[macroLinkID] == nil {
				return fmt.Errorf("meso link %v: MacroLink %q: %w", identifier, macroLinkID, macro.ErrLinkNotFound)
			}
		}
		movementID := link.Movement()
		if !movementID.IsZero() {
			if movements[movementID] == nil {
				return fmt.Errorf("meso link %v: Movement %q: %w", identifier, movementID, movement.ErrMvmtNotFound)
			}
		}
		movementMesoLinkIncomeID := link.MovementMesoLinkIncome()
		if !movementMesoLinkIncomeID.IsZero() {
			if net.Links[movementMesoLinkIncomeID] == nil {
				return fmt.Errorf("meso link %v: MovementMesoLinkIncome %q: %w", identifier, movementMesoLinkIncomeID, ErrLinkNotFound)
			}
		}
		movementMesoLinkOutcomeID := link.MovementMesoLinkOutcome()
		if !movementMesoLinkOutcomeID.IsZero() {
			if net.Links[movementMesoLinkOutcomeID] == nil {
				return fmt.Errorf("meso link %v: MovementMesoLinkOutcome %q: %w", identifier, movementMesoLinkOutcomeID, ErrLinkNotFound)
			}
		}
	}
	return nil
}
