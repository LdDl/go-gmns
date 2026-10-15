package generators

import (
	"fmt"

	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

func requiredMacroLaneCount(link *macro.Link) (int, error) {
	lanesNum, hasLanes := link.LanesNum()
	if !hasLanes {
		return 0, fmt.Errorf("macro link %v: lane count is required", link.ID)
	}
	if lanesNum < 0 {
		return 0, fmt.Errorf("macro link %v: lane count must be nonnegative, got %v", link.ID, lanesNum)
	}
	return lanesNum, nil
}

func requiredMesoLaneCount(link *meso.Link) (int, error) {
	lanesNum, hasLanes := link.LanesNum()
	if !hasLanes {
		return 0, fmt.Errorf("meso link %v: lane count is required", link.ID)
	}
	if lanesNum < 0 {
		return 0, fmt.Errorf("meso link %v: lane count must be nonnegative, got %v", link.ID, lanesNum)
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
			return fmt.Errorf("macro link %v: movement and meso generation require a positive lane count", link.ID)
		}
		laneInfo := link.LanesInfo()
		if len(laneInfo.LanesList) == 0 {
			return fmt.Errorf("macro link %v: prepared lane information is required", link.ID)
		}
		if len(laneInfo.LanesChange) != len(laneInfo.LanesList) {
			return fmt.Errorf("macro link %v: lane-change entries must match the lane segments", link.ID)
		}
		for segment, lanes := range laneInfo.LanesList {
			if lanes <= 0 {
				return fmt.Errorf("macro link %v: lane segment %v requires a positive lane count, got %v", link.ID, segment, lanes)
			}
		}
	}
	return nil
}
