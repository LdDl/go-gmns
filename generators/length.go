package generators

import (
	"fmt"
	"math"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

func requireMacroLinkLengths(net *macro.Net) error {
	linkIDs := make([]gmns.LinkID, 0, len(net.Links))
	for linkID := range net.Links {
		linkIDs = append(linkIDs, linkID)
	}
	sort.Slice(linkIDs, func(first, second int) bool {
		return linkIDs[first].Less(linkIDs[second])
	})
	for _, linkID := range linkIDs {
		link := net.Links[linkID]
		lengthMeters, hasLength := link.LengthMeters()
		if !hasLength {
			return fmt.Errorf("macro link %v: length is required for meso generation", linkID)
		}
		if lengthMeters <= 0 || math.IsNaN(lengthMeters) || math.IsInf(lengthMeters, 0) {
			return fmt.Errorf("macro link %v: meso generation requires a positive finite length, got %g", linkID, lengthMeters)
		}
	}
	return nil
}

func requiredMesoLength(link *meso.Link) (float64, error) {
	lengthMeters, hasLength := link.LengthMeters()
	if !hasLength {
		return 0, fmt.Errorf("meso link %v: length is required for micro cell generation", link.ID)
	}
	if lengthMeters <= 0 || math.IsNaN(lengthMeters) || math.IsInf(lengthMeters, 0) {
		return 0, fmt.Errorf("meso link %v: micro cell generation requires a positive finite length, got %g", link.ID, lengthMeters)
	}
	return lengthMeters, nil
}

func requireMesoRoadLengths(macroNet *macro.Net, mesoNet *meso.Net) error {
	for _, linkID := range sortedMesoLinkIDs(mesoNet.Links) {
		link := mesoNet.Links[linkID]
		macroLinkID := link.MacroLink()
		if macroLinkID.IsZero() {
			continue
		}
		_, exists := macroNet.Links[macroLinkID]
		if !exists {
			continue
		}
		_, err := requiredMesoLength(link)
		if err != nil {
			return err
		}
	}
	return nil
}
