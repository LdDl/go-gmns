package movement

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
)

// ValidateIDType checks movement IDs and references against the configured kind.
// Call the owning network's ValidateMovements method to resolve references.
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
		err = mvmt.Node().ValidateType(kind)
		if err != nil {
			return fmt.Errorf("movement %v node_id: %w", identifier, err)
		}
		for _, linkID := range []gmns.LinkID{mvmt.IncomeLink(), mvmt.OutcomeLink()} {
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
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	return ids
}
