package gmns

import "strconv"

// SegmentID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type SegmentID struct {
	value string
	kind  IDType
}

// SegmentIDFromString preserves the exact spelling, including leading zeroes.
func SegmentIDFromString(value string) SegmentID {
	return SegmentID{value: value, kind: IDTypeString}
}

func SegmentIDFromInt64(value int64) SegmentID {
	return SegmentID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseSegmentID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseSegmentID(value string, kind IDType) (SegmentID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return SegmentID{}, err
	}
	return SegmentID{value: parsed, kind: kind}, nil
}

func (identifier SegmentID) String() string {
	return identifier.value
}

func (identifier SegmentID) Type() IDType {
	return identifier.kind
}

func (identifier SegmentID) IsZero() bool {
	return identifier == (SegmentID{})
}

func (identifier SegmentID) Validate() error {
	return validateID(identifier.value, identifier.kind)
}

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier SegmentID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier SegmentID) Less(other SegmentID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier SegmentID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *SegmentID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier SegmentID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
