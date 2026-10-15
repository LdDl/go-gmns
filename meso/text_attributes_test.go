package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type textAttributesCase struct {
	name     string
	nodeName string
	nodeType string
	linkName string
	set      bool
	clear    bool
}

func TestTextAttributesGeoJSON(t *testing.T) {
	cases := []textAttributesCase{
		{
			name: "absent",
		},
		{
			name:     "supplied names and node type",
			nodeName: "Central station",
			nodeType: types.NODE_TYPE_TRANSIT_STATION,
			linkName: "Main Street",
			set:      true,
		},
		{
			name:     "intersection constant",
			nodeType: types.NODE_TYPE_INTERSECTION,
			set:      true,
		},
		{
			name:     "park and ride constant",
			nodeType: types.NODE_TYPE_PARK_AND_RIDE,
			set:      true,
		},
		{
			name:     "custom text is preserved exactly",
			nodeName: "  Площадь; \"Центр\"  ",
			nodeType: "custom:platform/access",
			linkName: "Переход; север\nЮг",
			set:      true,
		},
		{
			name:     "clear supplied values",
			nodeName: "Central station",
			nodeType: types.NODE_TYPE_TRANSIT_STATION,
			linkName: "Main Street",
			set:      true,
			clear:    true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			node := NewNodeFrom(gmns.NodeIDFromInt64(1))
			link := NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3))
			if testCase.set {
				WithNodeName(testCase.nodeName)(node)
				WithNodeType(testCase.nodeType)(node)
				WithLinkName(testCase.linkName)(link)
			}
			wantNodeName := testCase.nodeName
			wantNodeType := testCase.nodeType
			wantLinkName := testCase.linkName
			if testCase.clear {
				WithNodeName("")(node)
				WithNodeType("")(node)
				WithLinkName("")(link)
				wantNodeName = ""
				wantNodeType = ""
				wantLinkName = ""
			}
			nodeData, err := node.GeoFeature().MarshalJSON()
			assert.NoError(t, err)
			if err != nil {
				return
			}
			nodeFeature, err := geojson.UnmarshalFeature(nodeData)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			linkData, err := link.GeoFeature().MarshalJSON()
			assert.NoError(t, err)
			if err != nil {
				return
			}
			linkFeature, err := geojson.UnmarshalFeature(linkData)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, wantNodeName, nodeFeature.Properties["name"])
			assert.Equal(t, wantNodeType, nodeFeature.Properties["node_type"])
			assert.Equal(t, wantLinkName, linkFeature.Properties["name"])
			assert.Equal(t, wantNodeName, node.Name())
			assert.Equal(t, wantNodeType, node.NodeType())
			assert.Equal(t, wantLinkName, link.Name())
			assert.Equal(t, float64(1), nodeFeature.ID)
			assert.Equal(t, float64(1), linkFeature.Properties["source_node"])
			assert.Equal(t, float64(3), linkFeature.Properties["target_node"])
			assert.Equal(t, "undefined", nodeFeature.Properties["activity_link_type"])
		})
	}
}
