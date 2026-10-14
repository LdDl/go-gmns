package generators

import (
	"fmt"

	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

func requiredMacroLaneCount(link *macro.Link) (int, error) {
	lanesNum, hasLanes := link.LanesNum()
	if !hasLanes {
		return 0, fmt.Errorf("macro link %d: lane count is required", link.ID)
	}
	if lanesNum < 0 {
		return 0, fmt.Errorf("macro link %d: lane count must be nonnegative, got %d", link.ID, lanesNum)
	}
	return lanesNum, nil
}

func requiredMesoLaneCount(link *meso.Link) (int, error) {
	lanesNum, hasLanes := link.LanesNum()
	if !hasLanes {
		return 0, fmt.Errorf("meso link %d: lane count is required", link.ID)
	}
	if lanesNum < 0 {
		return 0, fmt.Errorf("meso link %d: lane count must be nonnegative, got %d", link.ID, lanesNum)
	}
	return lanesNum, nil
}

func requireMacroLaneData(net *macro.Net) error {
	for _, link := range macroLinksToSlice(net.Links) {
		lanesNum, err := requiredMacroLaneCount(link)
		if err != nil {
			return err
		}
		if lanesNum == 0 {
			return fmt.Errorf("macro link %d: movement and meso generation require a positive lane count", link.ID)
		}
		laneInfo := link.LanesInfo()
		if len(laneInfo.LanesList) == 0 {
			return fmt.Errorf("macro link %d: prepared lane information is required", link.ID)
		}
		if len(laneInfo.LanesChange) != len(laneInfo.LanesList) {
			return fmt.Errorf("macro link %d: lane-change entries must match the lane segments", link.ID)
		}
		for segment, lanes := range laneInfo.LanesList {
			if lanes <= 0 {
				return fmt.Errorf("macro link %d: lane segment %d requires a positive lane count, got %d", link.ID, segment, lanes)
			}
		}
	}
	return nil
}
