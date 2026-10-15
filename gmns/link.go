package gmns

import "strconv"

// LinkID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type LinkID struct {
	value string
	kind  IDType
}

// LinkIDFromString preserves the exact spelling, including leading zeroes.
func LinkIDFromString(value string) LinkID {
	return LinkID{value: value, kind: IDTypeString}
}

func LinkIDFromInt64(value int64) LinkID {
	return LinkID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseLinkID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseLinkID(value string, kind IDType) (LinkID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return LinkID{}, err
	}
	return LinkID{value: parsed, kind: kind}, nil
}

func (identifier LinkID) String() string {
	return identifier.value
}

func (identifier LinkID) Type() IDType {
	return identifier.kind
}

func (identifier LinkID) IsZero() bool {
	return identifier == (LinkID{})
}

func (identifier LinkID) Validate() error {
	return validateID(identifier.value, identifier.kind)
}

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier LinkID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier LinkID) Less(other LinkID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier LinkID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *LinkID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier LinkID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
