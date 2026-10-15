package generators

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// ErrUndirectedLinkUnsupported indicates input that the directed-network generators cannot process.
var ErrUndirectedLinkUnsupported = fmt.Errorf("generation requires directed links")

func requireDirectedMacroLinks(net *macro.Net) error {
	if net == nil {
		return fmt.Errorf("macro network is nil")
	}
	linkIDs := make([]gmns.LinkID, 0, len(net.Links))
	for linkID := range net.Links {
		linkIDs = append(linkIDs, linkID)
	}
	sort.Slice(linkIDs, func(first, second int) bool {
		return linkIDs[first].Less(linkIDs[second])
	})
	for _, linkID := range linkIDs {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("macro link %v: %w", linkID, macro.ErrLinkNotFound)
		}
		if !link.Directed() {
			return fmt.Errorf("macro link %v: %w", link.ID, ErrUndirectedLinkUnsupported)
		}
	}
	return nil
}

func requireDirectedMesoLinks(net *meso.Net) error {
	if net == nil {
		return fmt.Errorf("meso network is nil")
	}
	for _, linkID := range sortedMesoLinkIDs(net.Links) {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("meso link %v: %w", linkID, meso.ErrLinkNotFound)
		}
		if !link.Directed() {
			return fmt.Errorf("meso link %v: %w", link.ID, ErrUndirectedLinkUnsupported)
		}
	}
	return nil
}
