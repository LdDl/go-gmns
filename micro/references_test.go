package micro

import (
	"encoding/json"
	"github.com/LdDl/go-gmns/gmns"
	"github.com/stretchr/testify/assert"
	"testing"
)

func identifierTestNet() *Net {
	first := gmns.NodeIDFromString("001")
	second := gmns.NodeIDFromString("1")
	linkID := gmns.LinkIDFromString("road-A")
	net := NewNet()
	net.Nodes[first] = NewNodeFrom(first)
	net.Nodes[second] = NewNodeFrom(second, WithParentNodeID(first))
	net.Links[linkID] = NewLinkFrom(linkID, first, second)
	return net
}

type referenceIDCase struct {
	name    string
	prepare func(*Net)
}

func TestValidateReferencesWithExactIDs(t *testing.T) {
	cases := []referenceIDCase{
		{
			name:    "missing endpoint",
			prepare: func(net *Net) { delete(net.Nodes, gmns.NodeIDFromString("001")) },
		},
		{
			name:    "numeric reference cannot resolve a string key",
			prepare: func(net *Net) { net.Links[gmns.LinkIDFromString("road-A")].SetSourceNode(gmns.NodeIDFromInt64(1)) },
		},
		{
			name:    "mismatched node key",
			prepare: func(net *Net) { net.Nodes[gmns.NodeIDFromString("001")].ID = gmns.NodeIDFromString("other") },
		},
		{
			name:    "mismatched link key",
			prepare: func(net *Net) { net.Links[gmns.LinkIDFromString("road-A")].ID = gmns.LinkIDFromString("other") },
		},
		{
			name:    "nil record",
			prepare: func(net *Net) { net.Links[gmns.LinkIDFromString("road-A")] = nil },
		},
		{
			name: "missing parent link",
			prepare: func(net *Net) {
				WithParentLinkID(gmns.LinkIDFromString("missing"))(net.Links[gmns.LinkIDFromString("road-A")])
			},
		},
		{
			name:    "explicit absent ID is not a valid supplied zone reference",
			prepare: func(net *Net) { WithZoneID(gmns.NodeID{})(net.Nodes[gmns.NodeIDFromString("001")]) },
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			net := identifierTestNet()
			testCase.prepare(net)
			err := net.ValidateReferences()
			assert.Error(t, err)
			repeated := net.ValidateReferences()
			assert.Equal(t, err, repeated, "validation is deterministic and does not repair data")
		})
	}
	net := identifierTestNet()
	assert.NoError(t, net.ValidateReferences())
	assert.NoError(t, net.ValidateIDType(gmns.IDTypeString))
	assert.Error(t, net.ValidateIDType(gmns.IDTypeInteger))
	assert.Len(t, net.Nodes, 2)
	var missing *Net
	assert.Error(t, missing.ValidateReferences())
}

func TestExternalReferencesUseExactResourceIDs(t *testing.T) {
	net := identifierTestNet()
	zoneID := gmns.NodeIDFromString("001")
	geometryID := gmns.GeometryIDFromString("shape-A")
	WithZoneID(zoneID)(net.Nodes[zoneID])
	WithGeometryID(geometryID)(net.Links[gmns.LinkIDFromString("road-A")])
	assert.NoError(t, net.ValidateReferences())
	assert.Error(t, net.ValidateExternalReferences(nil, nil))
	assert.Error(t, net.ValidateExternalReferences([]gmns.NodeID{gmns.NodeIDFromString("1")}, []gmns.GeometryID{geometryID}))
	assert.Error(t, net.ValidateExternalReferences([]gmns.NodeID{zoneID}, nil))
	assert.NoError(t, net.ValidateExternalReferences([]gmns.NodeID{zoneID}, []gmns.GeometryID{geometryID}))
	assert.ErrorIs(t, net.ValidateExternalReferences([]gmns.NodeID{zoneID, zoneID}, nil), gmns.ErrDuplicateID)
}

func TestGeoJSONPreservesIDKindAndPrecision(t *testing.T) {
	integer, err := gmns.ParseNodeID("922337203685477580812345", gmns.IDTypeInteger)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	ids := []gmns.NodeID{gmns.NodeIDFromString("001"), integer}
	for _, identifier := range ids {
		node := NewNodeFrom(identifier)
		data, err := node.GeoFeature().MarshalJSON()
		assert.NoError(t, err)
		if err != nil {
			return
		}
		feature := make(map[string]json.RawMessage)
		err = json.Unmarshal(data, &feature)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		var decoded gmns.NodeID
		err = json.Unmarshal(feature["id"], &decoded)
		assert.NoError(t, err)
		assert.Equal(t, identifier, decoded)
	}
}
