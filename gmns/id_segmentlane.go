package gmns

import "strconv"

// SegmentLaneID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type SegmentLaneID struct {
	value string
	kind  IDType
}

// SegmentLaneIDFromString preserves the exact spelling, including leading zeroes.
func SegmentLaneIDFromString(value string) SegmentLaneID {
	return SegmentLaneID{value: value, kind: IDTypeString}
}

func SegmentLaneIDFromInt64(value int64) SegmentLaneID {
	return SegmentLaneID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseSegmentLaneID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseSegmentLaneID(value string, kind IDType) (SegmentLaneID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return SegmentLaneID{}, err
	}
	return SegmentLaneID{value: parsed, kind: kind}, nil
}

func (identifier SegmentLaneID) String() string {
	return identifier.value
}

func (identifier SegmentLaneID) Type() IDType {
	return identifier.kind
}

func (identifier SegmentLaneID) IsZero() bool {
	return identifier == (SegmentLaneID{})
}

func (identifier SegmentLaneID) Validate() error {
	return validateID(identifier.value, identifier.kind)
}

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier SegmentLaneID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier SegmentLaneID) Less(other SegmentLaneID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier SegmentLaneID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *SegmentLaneID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier SegmentLaneID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
