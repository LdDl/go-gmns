package micro

import (
	"fmt"

	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// ValidateSourceReferences resolves application provenance separately from standard parents.
func (net *Net) ValidateSourceReferences(macroNet *macro.Net, mesoNet *meso.Net) error {
	err := net.ValidateReferences()
	if err != nil {
		return err
	}
	err = macroNet.ValidateReferences()
	if err != nil {
		return err
	}
	err = mesoNet.ValidateReferences()
	if err != nil {
		return err
	}
	for _, identifier := range net.sortedNodeIDs() {
		node := net.Nodes[identifier]
		mesoLinkID := node.MesoLink()
		if !mesoLinkID.IsZero() {
			if mesoNet.Links[mesoLinkID] == nil {
				return fmt.Errorf("micro node %v: MesoLink %q: %w", identifier, mesoLinkID, meso.ErrLinkNotFound)
			}
		}
	}
	for _, identifier := range net.sortedLinkIDs() {
		link := net.Links[identifier]
		mesoLinkID := link.MesoLink()
		if !mesoLinkID.IsZero() {
			if mesoNet.Links[mesoLinkID] == nil {
				return fmt.Errorf("micro link %v: MesoLink %q: %w", identifier, mesoLinkID, meso.ErrLinkNotFound)
			}
		}
		macroNodeID := link.MacroNode()
		if !macroNodeID.IsZero() {
			if macroNet.Nodes[macroNodeID] == nil {
				return fmt.Errorf("micro link %v: MacroNode %q: %w", identifier, macroNodeID, macro.ErrNodeNotFound)
			}
		}
		macroLinkID := link.MacroLink()
		if !macroLinkID.IsZero() {
			if macroNet.Links[macroLinkID] == nil {
				return fmt.Errorf("micro link %v: MacroLink %q: %w", identifier, macroLinkID, macro.ErrLinkNotFound)
			}
		}
	}
	return nil
}
