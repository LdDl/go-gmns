package gmns

import "strconv"

// LaneID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type LaneID struct {
	value string
	kind  IDType
}

// LaneIDFromString preserves the exact spelling, including leading zeroes.
func LaneIDFromString(value string) LaneID {
	return LaneID{value: value, kind: IDTypeString}
}

func LaneIDFromInt64(value int64) LaneID {
	return LaneID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseLaneID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseLaneID(value string, kind IDType) (LaneID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return LaneID{}, err
	}
	return LaneID{value: parsed, kind: kind}, nil
}

func (identifier LaneID) String() string {
	return identifier.value
}

func (identifier LaneID) Type() IDType {
	return identifier.kind
}

func (identifier LaneID) IsZero() bool {
	return identifier == (LaneID{})
}

func (identifier LaneID) Validate() error {
	return validateID(identifier.value, identifier.kind)
}

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier LaneID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier LaneID) Less(other LaneID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier LaneID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *LaneID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier LaneID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
