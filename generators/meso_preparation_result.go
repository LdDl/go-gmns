package generators

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// MesoPreparationResult owns independent completed source and generated datasets.
// Neither dataset shares mutable geometry with the original input.
type MesoPreparationResult struct {
	macroNet        *macro.Net
	mesoNet         *meso.Net
	laneAssumptions []LaneCountAssumption
	sections        []PreparedSection
	connections     []PreparedConnection
}

// Macro returns the completed source dataset owned by this result.
func (result *MesoPreparationResult) Macro() *macro.Net {
	return result.macroNet
}

// Meso returns the generated dataset owned by this result.
func (result *MesoPreparationResult) Meso() *meso.Net {
	return result.mesoNet
}

// LaneAssumptions returns a copy of the applied defaults, ordered by source link ID.
func (result *MesoPreparationResult) LaneAssumptions() []LaneCountAssumption {
	return append([]LaneCountAssumption(nil), result.laneAssumptions...)
}

// LaneCountAssumption records an explicitly accepted replacement for an absent count.
// It is preparation metadata, not a standard GMNS record.
type LaneCountAssumption struct {
	linkID gmns.LinkID
	lanes  int
}

func (assumption *LaneCountAssumption) Link() gmns.LinkID {
	return assumption.linkID
}

func (assumption *LaneCountAssumption) Lanes() int {
	return assumption.lanes
}
