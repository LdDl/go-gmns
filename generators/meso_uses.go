package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
	"github.com/LdDl/go-gmns/movement"
)

// expandedUses resolves symbolic permissions without interpreting use names as agents.
// A nil set means unspecified; it must not be confused with an empty intersection.
func expandedUses(net *macro.Net, value string) (map[string]bool, error) {
	if value == "" {
		return nil, nil
	}
	result := make(map[string]bool)
	for _, item := range strings.Split(value, ",") {
		name := strings.TrimSpace(item)
		_, definitionExists := net.UseDefinition(name)
		if definitionExists {
			result[name] = true
			continue
		}
		group, groupExists := net.UseGroup(name)
		if !groupExists {
			return nil, fmt.Errorf("unknown allowed use %q", name)
		}
		for _, use := range group.Uses() {
			result[use] = true
		}
	}
	return result, nil
}

func movementAllowedUses(net *macro.Net, record *movement.Movement) (string, error) {
	var intersection map[string]bool
	values := []string{
		net.Links[record.IncomeLink()].AllowedUses(),
		record.AllowedUses(),
		net.Links[record.OutcomeLink()].AllowedUses(),
	}
	for _, value := range values {
		uses, err := expandedUses(net, value)
		if err != nil {
			return "", fmt.Errorf("movement %v allowed_uses: %w", record.ID, err)
		}
		if uses == nil {
			continue
		}
		if intersection == nil {
			intersection = uses
			continue
		}
		for use := range intersection {
			if !uses[use] {
				delete(intersection, use)
			}
		}
	}
	if intersection == nil {
		return "", nil
	}
	if len(intersection) == 0 {
		return "", fmt.Errorf("movement %v allowed_uses: no common use between inbound link %v, movement and outbound link %v; resolve the conflicting permissions explicitly", record.ID, record.IncomeLink(), record.OutcomeLink())
	}
	names := make([]string, 0, len(intersection))
	for name := range intersection {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ","), nil
}

func applyPreparedUses(source *macro.Net, target *meso.Net) error {
	for _, id := range sortedMesoLinkIDs(target.Links) {
		link := target.Links[id]
		if !link.MacroLink().IsZero() {
			meso.WithAllowedUses(source.Links[link.MacroLink()].AllowedUses())(link)
			continue
		}
		record, exists := source.Movement(link.Movement())
		if !exists {
			return fmt.Errorf("meso link %v: source movement %v not found", id, link.Movement())
		}
		uses, err := movementAllowedUses(source, record)
		if err != nil {
			return err
		}
		meso.WithAllowedUses(uses)(link)
	}
	return nil
}
