package gmns

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type idCase struct {
	name      string
	value     string
	kind      IDType
	canonical string
	jsonValue string
}

func TestIDKindsPreserveIdentity(t *testing.T) {
	cases := []idCase{
		{
			name:      "leading zero string",
			value:     "001",
			kind:      IDTypeString,
			canonical: "001",
			jsonValue: `"001"`,
		},
		{
			name:      "numeric string",
			value:     "1",
			kind:      IDTypeString,
			canonical: "1",
			jsonValue: `"1"`,
		},
		{
			name:      "text with spaces and Unicode",
			value:     " node-А ",
			kind:      IDTypeString,
			canonical: " node-А ",
			jsonValue: `" node-А "`,
		},
		{
			name:      "integer normalized by configured kind",
			value:     "001",
			kind:      IDTypeInteger,
			canonical: "1",
			jsonValue: "1",
		},
		{
			name:      "zero",
			value:     "0",
			kind:      IDTypeInteger,
			canonical: "0",
			jsonValue: "0",
		},
		{
			name:      "negative one",
			value:     "-1",
			kind:      IDTypeInteger,
			canonical: "-1",
			jsonValue: "-1",
		},
		{
			name:      "beyond int64 and float64 precision",
			value:     "922337203685477580812345678901",
			kind:      IDTypeInteger,
			canonical: "922337203685477580812345678901",
			jsonValue: "922337203685477580812345678901",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			node, err := ParseNodeID(testCase.value, testCase.kind)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, testCase.canonical, node.String())
			assert.Equal(t, testCase.kind, node.Type())
			assert.False(t, node.IsZero())
			assert.NoError(t, node.ValidateType(testCase.kind))
			data, err := json.Marshal(node)
			assert.NoError(t, err)
			assert.Equal(t, testCase.jsonValue, string(data))
			var decoded NodeID
			err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, node, decoded)
			keys := map[NodeID]bool{node: true}
			assert.True(t, keys[decoded])

			link, err := ParseLinkID(testCase.value, testCase.kind)
			assert.NoError(t, err)
			data, err = json.Marshal(link)
			assert.NoError(t, err)
			assert.Equal(t, testCase.jsonValue, string(data))
			var decodedLink LinkID
			err = json.Unmarshal(data, &decodedLink)
			assert.NoError(t, err)
			assert.Equal(t, link, decodedLink)

			mvmt, err := ParseMovementID(testCase.value, testCase.kind)
			assert.NoError(t, err)
			data, err = json.Marshal(mvmt)
			assert.NoError(t, err)
			assert.Equal(t, testCase.jsonValue, string(data))
			var decodedMovement MovementID
			err = json.Unmarshal(data, &decodedMovement)
			assert.NoError(t, err)
			assert.Equal(t, mvmt, decodedMovement)

			geometry, err := ParseGeometryID(testCase.value, testCase.kind)
			assert.NoError(t, err)
			data, err = json.Marshal(geometry)
			assert.NoError(t, err)
			assert.Equal(t, testCase.jsonValue, string(data))
			var decodedGeometry GeometryID
			err = json.Unmarshal(data, &decodedGeometry)
			assert.NoError(t, err)
			assert.Equal(t, geometry, decodedGeometry)
		})
	}
	keys := map[NodeID]bool{
		NodeIDFromString("001"): true,
		NodeIDFromString("1"):   true,
		NodeIDFromInt64(1):      true,
	}
	assert.Len(t, keys, 3)
	assert.True(t, NodeID{}.IsZero())
	assert.False(t, NodeIDFromInt64(0).IsZero())
	assert.False(t, NodeIDFromInt64(-1).IsZero())
	assert.Error(t, NodeID{}.Validate())
	assert.Error(t, NodeIDFromString("").Validate())
	assert.Error(t, NodeIDFromString("1").ValidateType(IDTypeInteger))
	assert.True(t, NodeIDFromInt64(2).Less(NodeIDFromInt64(10)))
	assert.True(t, NodeIDFromString("10").Less(NodeIDFromString("2")))
}

func TestIDRejectsInvalidInputWithoutReplacingValue(t *testing.T) {
	invalid := []string{
		`1.5`,
		`1e3`,
		`true`,
		`[]`,
		`{}`,
		`""`,
		`01`,
	}
	for _, data := range invalid {
		t.Run(data, func(t *testing.T) {
			value := NodeIDFromString("001")
			err := json.Unmarshal([]byte(data), &value)
			assert.Error(t, err)
			assert.Equal(t, NodeIDFromString("001"), value)
		})
	}
	_, err := ParseNodeID("1.0", IDTypeInteger)
	assert.Error(t, err)
	_, err = ParseNodeID("1", IDType("float"))
	assert.Error(t, err)
	value, err := ParseNodeID("9223372036854775808123", IDTypeInteger)
	assert.NoError(t, err)
	_, fits := value.Int64()
	assert.False(t, fits)
	integer, fits := NodeIDFromInt64(-1).Int64()
	assert.True(t, fits)
	assert.Equal(t, int64(-1), integer)
	_, fits = NodeIDFromString("1").Int64()
	assert.False(t, fits)
	err = json.Unmarshal([]byte("null"), &value)
	assert.NoError(t, err)
	assert.True(t, value.IsZero())
}
