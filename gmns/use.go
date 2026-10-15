package gmns

import (
	"fmt"
	"strings"
)

// UseDefinition describes a travel mode or a non-travel use such as parking.
type UseDefinition struct {
	use               string
	personsPerVehicle float64
	pce               float64
	specialConditions string
	description       string
}

func NewUseDefinition(use string, personsPerVehicle, pce float64, options ...func(*UseDefinition)) *UseDefinition {
	definition := &UseDefinition{use: use, personsPerVehicle: personsPerVehicle, pce: pce}
	for _, option := range options {
		option(definition)
	}
	return definition
}

func (definition *UseDefinition) Use() string {
	return definition.use
}

func (definition *UseDefinition) PersonsPerVehicle() float64 {
	return definition.personsPerVehicle
}

func (definition *UseDefinition) PCE() float64 {
	return definition.pce
}

func (definition *UseDefinition) SpecialConditions() string {
	return definition.specialConditions
}

func (definition *UseDefinition) Description() string {
	return definition.description
}

func (definition *UseDefinition) Clone() *UseDefinition {
	if definition == nil {
		return nil
	}
	result := *definition
	return &result
}

func (definition *UseDefinition) Validate() error {
	if definition == nil {
		return fmt.Errorf("use definition is nil")
	}
	err := validateUseName(definition.use)
	if err != nil {
		return err
	}
	if !Finite(definition.personsPerVehicle) || definition.personsPerVehicle < 0 {
		return fmt.Errorf("persons_per_vehicle must be finite and nonnegative")
	}
	if !Finite(definition.pce) || definition.pce < 0 {
		return fmt.Errorf("pce must be finite and nonnegative")
	}
	return nil
}

func WithUseDescription(text string) func(*UseDefinition) {
	return func(definition *UseDefinition) {
		definition.description = text
	}
}

func WithUseSpecialConditions(text string) func(*UseDefinition) {
	return func(definition *UseDefinition) {
		definition.specialConditions = text
	}
}

// UseGroup names a list of concrete use definitions; groups do not recursively contain groups.
type UseGroup struct {
	name        string
	uses        []string
	description string
}

func NewUseGroup(name string, uses []string, options ...func(*UseGroup)) *UseGroup {
	group := &UseGroup{name: name, uses: append([]string(nil), uses...)}
	for _, option := range options {
		option(group)
	}
	return group
}

func (group *UseGroup) Name() string {
	return group.name
}

func (group *UseGroup) Uses() []string {
	return append([]string(nil), group.uses...)
}

func (group *UseGroup) Description() string {
	return group.description
}

func (group *UseGroup) Clone() *UseGroup {
	if group == nil {
		return nil
	}
	result := *group
	result.uses = group.Uses()
	return &result
}

func (group *UseGroup) Validate() error {
	if group == nil {
		return fmt.Errorf("use group is nil")
	}
	err := validateUseName(group.name)
	if err != nil {
		return err
	}
	if len(group.uses) == 0 {
		return fmt.Errorf("use group must contain uses")
	}
	seen := make(map[string]bool)
	for _, name := range group.uses {
		err = validateUseName(name)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("duplicate use %q in group", name)
		}
		seen[name] = true
	}
	return nil
}

func WithUseGroupDescription(text string) func(*UseGroup) {
	return func(group *UseGroup) {
		group.description = text
	}
}

func validateUseName(name string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(name) != name || strings.Contains(name, ",") {
		return fmt.Errorf("invalid use name %q", name)
	}
	return nil
}

// ValidateAllowedUses resolves the stored comma-separated references without changing their spelling.
func ValidateAllowedUses(value string, definitions map[string]*UseDefinition, groups map[string]*UseGroup) error {
	if value == "" {
		return nil
	}
	for _, item := range strings.Split(value, ",") {
		name := strings.TrimSpace(item)
		if name == "" || (definitions[name] == nil && groups[name] == nil) {
			return fmt.Errorf("unknown allowed use %q", name)
		}
	}
	return nil
}
