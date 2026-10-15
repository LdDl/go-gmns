package gmns

import "strconv"

// MovementID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type MovementID struct {
	value string
	kind  IDType
}

// MovementIDFromString preserves the exact spelling, including leading zeroes.
func MovementIDFromString(value string) MovementID {
	return MovementID{value: value, kind: IDTypeString}
}

func MovementIDFromInt64(value int64) MovementID {
	return MovementID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseMovementID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseMovementID(value string, kind IDType) (MovementID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return MovementID{}, err
	}
	return MovementID{value: parsed, kind: kind}, nil
}

func (identifier MovementID) String() string {
	return identifier.value
}

func (identifier MovementID) Type() IDType {
	return identifier.kind
}

func (identifier MovementID) IsZero() bool {
	return identifier == (MovementID{})
}

func (identifier MovementID) Validate() error {
	return validateID(identifier.value, identifier.kind)
}

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier MovementID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier MovementID) Less(other MovementID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier MovementID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *MovementID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier MovementID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
