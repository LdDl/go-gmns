package micro

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/movement"
)

// ValidateMovements resolves movement references in this network, independently of source provenance.
func (net *Net) ValidateMovements(storage movement.MovementsStorage) error {
	err := net.ValidateReferences()
	if err != nil {
		return err
	}
	for _, identifier := range sortedMovementIDs(storage) {
		mvmt := storage[identifier]
		if mvmt == nil {
			return fmt.Errorf("movement %v: %w", identifier, movement.ErrMvmtNotFound)
		}
		err = identifier.Validate()
		if err != nil {
			return fmt.Errorf("movement key %q: %w", identifier, err)
		}
		if mvmt.ID != identifier {
			return fmt.Errorf("movement key %q does not match record ID %q", identifier, mvmt.ID)
		}
		nodeID := mvmt.Node()
		if net.Nodes[nodeID] == nil {
			return fmt.Errorf("movement %v node_id %q: %w", identifier, nodeID, ErrNodeNotFound)
		}
		income := net.Links[mvmt.IncomeLink()]
		if income == nil {
			return fmt.Errorf("movement %v inbound link %q: %w", identifier, mvmt.IncomeLink(), ErrLinkNotFound)
		}
		outcome := net.Links[mvmt.OutcomeLink()]
		if outcome == nil {
			return fmt.Errorf("movement %v outbound link %q: %w", identifier, mvmt.OutcomeLink(), ErrLinkNotFound)
		}
		if income.TargetNode() != nodeID && (income.Directed() || income.SourceNode() != nodeID) {
			return fmt.Errorf("movement %v: inbound link does not enter node %q", identifier, nodeID)
		}
		if outcome.SourceNode() != nodeID && (outcome.Directed() || outcome.TargetNode() != nodeID) {
			return fmt.Errorf("movement %v: outbound link does not leave node %q", identifier, nodeID)
		}
	}
	return nil
}

func sortedMovementIDs(storage movement.MovementsStorage) []gmns.MovementID {
	ids := make([]gmns.MovementID, 0, len(storage))
	for id := range storage {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	return ids
}
