package gmns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"unicode/utf8"
)

// IDType selects the interpretation of identifiers in a dataset.
type IDType string

const (
	IDTypeString  IDType = "string"
	IDTypeInteger IDType = "integer"
)

func (idType IDType) Validate() error {
	if idType != IDTypeString && idType != IDTypeInteger {
		return fmt.Errorf("unsupported ID type %q", idType)
	}
	return nil
}

func parseID(value string, kind IDType) (string, error) {
	err := kind.Validate()
	if err != nil {
		return "", err
	}
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("ID must be valid UTF-8")
	}
	if value == "" {
		return "", fmt.Errorf("ID must not be empty")
	}
	if kind == IDTypeString {
		return value, nil
	}
	digits := value
	if digits[0] == '-' {
		digits = digits[1:]
	}
	if digits == "" {
		return "", fmt.Errorf("invalid integer ID %q", value)
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return "", fmt.Errorf("invalid integer ID %q", value)
		}
	}
	integer := new(big.Int)
	_, valid := integer.SetString(value, 10)
	if !valid {
		return "", fmt.Errorf("invalid integer ID %q", value)
	}
	return integer.String(), nil
}

func validateID(value string, kind IDType) error {
	canonical, err := parseID(value, kind)
	if err != nil {
		return err
	}
	if canonical != value {
		return fmt.Errorf("integer ID %q is not canonical", value)
	}
	return nil
}

func marshalID(value string, kind IDType) ([]byte, error) {
	if kind == "" && value == "" {
		return []byte("null"), nil
	}
	err := validateID(value, kind)
	if err != nil {
		return nil, err
	}
	if kind == IDTypeString {
		return json.Marshal(value)
	}
	return []byte(value), nil
}

func unmarshalID(data []byte) (string, IDType, error) {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		return "", "", nil
	}
	if !json.Valid(data) {
		return "", "", fmt.Errorf("invalid ID JSON")
	}
	kind := IDTypeInteger
	value := string(data)
	if len(data) > 0 && data[0] == '"' {
		kind = IDTypeString
		err := json.Unmarshal(data, &value)
		if err != nil {
			return "", "", err
		}
	}
	value, err := parseID(value, kind)
	if err != nil {
		return "", "", err
	}
	return value, kind, nil
}

func lessID(value string, kind IDType, otherValue string, otherKind IDType) bool {
	if kind != otherKind {
		return kind < otherKind
	}
	if kind != IDTypeInteger {
		return value < otherValue
	}
	first, firstOK := new(big.Int).SetString(value, 10)
	second, secondOK := new(big.Int).SetString(otherValue, 10)
	if !firstOK || !secondOK {
		return value < otherValue
	}
	return first.Cmp(second) < 0
}

func integerID(value string, kind IDType) (int64, bool) {
	if kind != IDTypeInteger {
		return 0, false
	}
	integer, err := strconv.ParseInt(value, 10, 64)
	return integer, err == nil
}

func validateIDType(actual, expected IDType) error {
	err := expected.Validate()
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("ID type %q does not match dataset type %q", actual, expected)
	}
	return nil
}
