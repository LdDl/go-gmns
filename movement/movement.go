package movement

import (
	"github.com/LdDl/go-gmns/gmns"
)

type MovementsStorage map[gmns.MovementID]*Movement

func NewMovementsStorage() map[gmns.MovementID]*Movement {
	return make(map[gmns.MovementID]*Movement)
}

type Movement struct {
	ID gmns.MovementID
}
