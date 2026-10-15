package movement

import (
	"fmt"

	"github.com/LdDl/go-gmns/gmns"
)

// Validate checks stored values independently of the owning network.
func (mvmt *Movement) Validate() error {
	if mvmt == nil {
		return fmt.Errorf("movement is nil")
	}
	err := mvmt.ID.Validate()
	if err != nil {
		return err
	}
	err = mvmt.Node().Validate()
	if err != nil {
		return err
	}
	err = mvmt.IncomeLink().Validate()
	if err != nil {
		return err
	}
	err = mvmt.OutcomeLink().Validate()
	if err != nil {
		return err
	}
	penalty, present := mvmt.Penalty()
	if present && !gmns.Finite(penalty) {
		return fmt.Errorf("movement penalty must be finite")
	}
	capacity, present := mvmt.Capacity()
	if present && (!gmns.Finite(capacity) || capacity < 0) {
		return fmt.Errorf("movement capacity must be finite and nonnegative")
	}
	if len(mvmt.geom) > 0 {
		err = gmns.ValidateLine(mvmt.geom)
		if err != nil {
			return err
		}
	}
	return nil
}
