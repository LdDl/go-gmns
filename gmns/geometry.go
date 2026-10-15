package gmns

import "strconv"

// GeometryID is an immutable, comparable identifier with an explicit string or integer kind.
// Its zero value denotes absence; integer zero and integer -1 are actual identifiers.
type GeometryID struct {
	value string
	kind  IDType
}

// GeometryIDFromString preserves the exact spelling, including leading zeroes.
func GeometryIDFromString(value string) GeometryID {
	return GeometryID{value: value, kind: IDTypeString}
}

func GeometryIDFromInt64(value int64) GeometryID {
	return GeometryID{value: strconv.FormatInt(value, 10), kind: IDTypeInteger}
}

// ParseGeometryID applies the configured kind without floating-point conversion.
// Integer spellings are canonicalized; string spellings are preserved exactly.
func ParseGeometryID(value string, kind IDType) (GeometryID, error) {
	parsed, err := parseID(value, kind)
	if err != nil {
		return GeometryID{}, err
	}
	return GeometryID{value: parsed, kind: kind}, nil
}

func (identifier GeometryID) String() string  { return identifier.value }
func (identifier GeometryID) Type() IDType    { return identifier.kind }
func (identifier GeometryID) IsZero() bool    { return identifier == (GeometryID{}) }
func (identifier GeometryID) Validate() error { return validateID(identifier.value, identifier.kind) }

// Int64 returns a numeric value only for integer IDs within the int64 range.
func (identifier GeometryID) Int64() (int64, bool) {
	return integerID(identifier.value, identifier.kind)
}

// Less provides deterministic ordering by kind, then by integer value or exact text.
func (identifier GeometryID) Less(other GeometryID) bool {
	return lessID(identifier.value, identifier.kind, other.value, other.kind)
}

func (identifier GeometryID) MarshalJSON() ([]byte, error) {
	return marshalID(identifier.value, identifier.kind)
}

func (identifier *GeometryID) UnmarshalJSON(data []byte) error {
	value, kind, err := unmarshalID(data)
	if err != nil {
		return err
	}
	identifier.value = value
	identifier.kind = kind
	return nil
}

// DirectionFlag describes shape orientation relative to a link's endpoints.
// It is independent of macro.DirectionType used when constructing OSM links.
type DirectionFlag int8

const (
	DIR_FLAG_REVERSE DirectionFlag = -1
	DIR_FLAG_NONE    DirectionFlag = 0
	DIR_FLAG_FORWARD DirectionFlag = 1
)

// ValidateType checks validity and agreement with the dataset identifier kind.
func (identifier GeometryID) ValidateType(kind IDType) error {
	err := identifier.Validate()
	if err != nil {
		return err
	}
	return validateIDType(identifier.kind, kind)
}
