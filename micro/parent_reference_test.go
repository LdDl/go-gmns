package micro

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
				WithParentNodeID(0),
			},
			linkOptions: []func(*Link){
				WithParentLinkID(0),
			},
			hasNode: true,
			hasLink: true,
		},
		{
			name: "negative IDs are supplied values",
			nodeOptions: []func(*Node){
				WithParentNodeID(-1),
			},
			linkOptions: []func(*Link){
				WithParentLinkID(-1),
			},
			parentNode: -1,
			parentLink: -1,
			hasNode:    true,
			hasLink:    true,
		},
		{
			name: "node reference only",
			nodeOptions: []func(*Node){
				WithParentNodeID(41),
			},
			parentNode: 41,
			hasNode:    true,
		},
		{
			name: "link reference only",
			linkOptions: []func(*Link){
				WithParentLinkID(52),
			},
			parentLink: 52,
			hasLink:    true,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			node := NewNodeFrom(1, testCase.nodeOptions...)
			link := NewLinkFrom(2, 1, 3, testCase.linkOptions...)
			WithNodeMesoLinkID(101)(node)
			WithMesoLinkID(201)(link)
			WithMacroLinkID(202)(link)
			WithMacroNodeID(203)(link)
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
				assert.Equal(t, float64(testCase.parentNode), nodeValue)
			} else {
				assert.Nil(t, nodeValue)
			}
			if testCase.hasLink {
				assert.Equal(t, float64(testCase.parentLink), linkValue)
			} else {
				assert.Nil(t, linkValue)
			}
			parentNode, hasNode := node.ParentNode()
			parentLink, hasLink := link.ParentLink()
			assert.Equal(t, testCase.parentNode, parentNode)
			assert.Equal(t, testCase.parentLink, parentLink)
			assert.Equal(t, testCase.hasNode, hasNode)
			assert.Equal(t, testCase.hasLink, hasLink)
			assert.Equal(t, gmns.NodeID(1), node.ID)
			assert.Equal(t, gmns.LinkID(2), link.ID)
			assert.Equal(t, gmns.NodeID(1), link.SourceNode())
			assert.Equal(t, gmns.NodeID(3), link.TargetNode())
			assert.Equal(t, float64(101), nodeFeature.Properties["meso_link_id"])
			assert.Equal(t, float64(201), linkFeature.Properties["meso_link_id"])
			assert.Equal(t, float64(202), linkFeature.Properties["macro_link_id"])
			assert.Equal(t, float64(203), linkFeature.Properties["macro_node_id"])
		})
	}
}
