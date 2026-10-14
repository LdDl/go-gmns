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
	linkIDs := make([]gmns.LinkID, 0, len(net.Links))
	for linkID := range net.Links {
		linkIDs = append(linkIDs, linkID)
	}
	sort.Slice(linkIDs, func(first, second int) bool {
		return linkIDs[first] < linkIDs[second]
	})
	for _, linkID := range linkIDs {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("macro link %d: %w", linkID, macro.ErrLinkNotFound)
		}
		if !link.Directed() {
			return fmt.Errorf("macro link %d: %w", link.ID, ErrUndirectedLinkUnsupported)
		}
	}
	return nil
}

func requireDirectedMesoLinks(net *meso.Net) error {
	for _, linkID := range sortedMesoLinkIDs(net.Links) {
		link := net.Links[linkID]
		if link == nil {
			return fmt.Errorf("meso link %d: %w", linkID, meso.ErrLinkNotFound)
		}
		if !link.Directed() {
			return fmt.Errorf("meso link %d: %w", link.ID, ErrUndirectedLinkUnsupported)
		}
	}
	return nil
}
