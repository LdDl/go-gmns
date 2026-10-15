package meso

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
)

type parentReferenceCase struct {
	name        string
	nodeOptions []func(*Node)
	linkOptions []func(*Link)
	parentNode  gmns.NodeID
	parentLink  gmns.LinkID
	hasNode     bool
	hasLink     bool
}

func TestParentReferencesGeoJSON(t *testing.T) {
	cases := []parentReferenceCase{
		{
			name: "absent despite source provenance",
		},
		{
			name: "explicit zero IDs",
			nodeOptions: []func(*Node){
				WithParentNodeID(gmns.NodeIDFromInt64(0)),
			},
			linkOptions: []func(*Link){
				WithParentLinkID(gmns.LinkIDFromInt64(0)),
			},
			parentNode: gmns.NodeIDFromInt64(0),
			parentLink: gmns.LinkIDFromInt64(0),
			hasNode:    true,
			hasLink:    true,
		},
		{
			name: "negative IDs are supplied values",
			nodeOptions: []func(*Node){
				WithParentNodeID(gmns.NodeIDFromInt64(-1)),
			},
			linkOptions: []func(*Link){
				WithParentLinkID(gmns.LinkIDFromInt64(-1)),
			},
			parentNode: gmns.NodeIDFromInt64(-1),
			parentLink: gmns.LinkIDFromInt64(-1),
			hasNode:    true,
			hasLink:    true,
		},
		{
			name: "node reference only",
			nodeOptions: []func(*Node){
				WithParentNodeID(gmns.NodeIDFromInt64(41)),
			},
			parentNode: gmns.NodeIDFromInt64(41),
			hasNode:    true,
		},
		{
			name: "link reference only",
			linkOptions: []func(*Link){
				WithParentLinkID(gmns.LinkIDFromInt64(52)),
			},
			parentLink: gmns.LinkIDFromInt64(52),
			hasLink:    true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			node := NewNodeFrom(gmns.NodeIDFromInt64(1), testCase.nodeOptions...)
			link := NewLinkFrom(gmns.LinkIDFromInt64(2), gmns.NodeIDFromInt64(1), gmns.NodeIDFromInt64(3), testCase.linkOptions...)
			WithPointMacroNodeID(gmns.NodeIDFromInt64(101))(node)
			WithPointMacroLinkID(gmns.LinkIDFromInt64(102))(node)
			WithLineMacroNodeID(gmns.NodeIDFromInt64(201))(link)
			WithLineMacroLinkID(gmns.LinkIDFromInt64(202))(link)
			WithMovementID(gmns.MovementIDFromInt64(203))(link)
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
			nodeValue, nodePropertyExists := nodeFeature.Properties["parent_node_id"]
			linkValue, linkPropertyExists := linkFeature.Properties["parent_link_id"]
			assert.True(t, nodePropertyExists)
			assert.True(t, linkPropertyExists)
			if testCase.hasNode {
				integer, isInteger := testCase.parentNode.Int64()
				assert.True(t, isInteger)
				assert.Equal(t, float64(integer), nodeValue)
			} else {
				assert.Nil(t, nodeValue)
			}
			if testCase.hasLink {
				integer, isInteger := testCase.parentLink.Int64()
				assert.True(t, isInteger)
				assert.Equal(t, float64(integer), linkValue)
			} else {
				assert.Nil(t, linkValue)
			}
			parentNode, hasNode := node.ParentNode()
			parentLink, hasLink := link.ParentLink()
			assert.Equal(t, testCase.parentNode, parentNode)
			assert.Equal(t, testCase.parentLink, parentLink)
			assert.Equal(t, testCase.hasNode, hasNode)
			assert.Equal(t, testCase.hasLink, hasLink)
			assert.Equal(t, gmns.NodeIDFromInt64(1), node.ID)
			assert.Equal(t, gmns.LinkIDFromInt64(2), link.ID)
			assert.Equal(t, gmns.NodeIDFromInt64(1), link.SourceNode())
			assert.Equal(t, gmns.NodeIDFromInt64(3), link.TargetNode())
			assert.Equal(t, float64(101), nodeFeature.Properties["macro_node_id"])
			assert.Equal(t, float64(102), nodeFeature.Properties["macro_link_id"])
			assert.Equal(t, float64(201), linkFeature.Properties["macro_node_id"])
			assert.Equal(t, float64(202), linkFeature.Properties["macro_link_id"])
			assert.Equal(t, float64(203), linkFeature.Properties["movement_id"])
		})
	}
}
