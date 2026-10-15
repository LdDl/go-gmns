package movement

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
)

// ValidateReferences checks macro-owned movements against their owning network.
// It does not reinterpret source movement IDs as references in generated networks.
func (storage MovementsStorage) ValidateReferences(net *macro.Net) error {
	err := net.ValidateReferences()
	if err != nil {
		return err
	}
	for _, identifier := range storage.sortedIDs() {
		mvmt := storage[identifier]
		if mvmt == nil {
			return fmt.Errorf("movement %v: %w", identifier, ErrMvmtNotFound)
		}
		err = identifier.Validate()
		if err != nil {
			return fmt.Errorf("movement key %q: %w", identifier, err)
		}
		if mvmt.ID != identifier {
			return fmt.Errorf("movement key %q does not match record ID %q", identifier, mvmt.ID)
		}
		nodeID := mvmt.MacroNode()
		if net.Nodes[nodeID] == nil {
			return fmt.Errorf("movement %v node_id %q: %w", identifier, nodeID, macro.ErrNodeNotFound)
		}
		income := net.Links[mvmt.IncomeMacroLink()]
		if income == nil {
			return fmt.Errorf("movement %v inbound link %q: %w", identifier, mvmt.IncomeMacroLink(), macro.ErrLinkNotFound)
		}
		outcome := net.Links[mvmt.OutcomeMacroLink()]
		if outcome == nil {
			return fmt.Errorf("movement %v outbound link %q: %w", identifier, mvmt.OutcomeMacroLink(), macro.ErrLinkNotFound)
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

// ValidateIDType checks movement IDs and references against the configured kind.
// Call ValidateReferences separately to resolve the referenced records.
func (storage MovementsStorage) ValidateIDType(kind gmns.IDType) error {
	err := kind.Validate()
	if err != nil {
		return err
	}
	for _, identifier := range storage.sortedIDs() {
		mvmt := storage[identifier]
		if mvmt == nil {
			return fmt.Errorf("movement %v: %w", identifier, ErrMvmtNotFound)
		}
		err = identifier.ValidateType(kind)
		if err != nil {
			return fmt.Errorf("movement %q: %w", identifier, err)
		}
		if mvmt.ID != identifier {
			return fmt.Errorf("movement key %q does not match record ID %q", identifier, mvmt.ID)
		}
		err = mvmt.MacroNode().ValidateType(kind)
		if err != nil {
			return fmt.Errorf("movement %v node_id: %w", identifier, err)
		}
		for _, linkID := range []gmns.LinkID{mvmt.IncomeMacroLink(), mvmt.OutcomeMacroLink()} {
			err = linkID.ValidateType(kind)
			if err != nil {
				return fmt.Errorf("movement %v link reference: %w", identifier, err)
			}
		}
	}
	return nil
}

func (storage MovementsStorage) sortedIDs() []gmns.MovementID {
	ids := make([]gmns.MovementID, 0, len(storage))
	for identifier := range storage {
		ids = append(ids, identifier)
	}
	sort.Slice(ids, func(first, second int) bool { return ids[first].Less(ids[second]) })
	return ids
}
