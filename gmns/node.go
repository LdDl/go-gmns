package gmns

import "strconv"

// NodeID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type NodeID struct {
	value string
	kind  IDType
}

// NodeIDFromString preserves the exact spelling, including leading zeroes.
func NodeIDFromString(value string) NodeID {
	return NodeID{value: value, kind: IDTypeString}
}

func NodeIDFromInt64(value int64) NodeID {
	return NodeID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseNodeID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseNodeID(value string, kind IDType) (NodeID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return NodeID{}, err
	}
	return NodeID{value: parsed, kind: kind}, nil
}

func (identifier NodeID) String() string {
	return identifier.value
}

func (identifier NodeID) Type() IDType {
	return identifier.kind
}

func (identifier NodeID) IsZero() bool {
	return identifier == (NodeID{})
}

func (identifier NodeID) Validate() error {
	return validateID(identifier.value, identifier.kind)
}

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier NodeID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier NodeID) Less(other NodeID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier NodeID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *NodeID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier NodeID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
