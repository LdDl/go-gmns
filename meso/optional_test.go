package meso

import (
	"testing"

	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

type optionalLinkFieldCase struct {
	name     string
	set      func(*Link)
	clear    func(*Link)
	present  func(*Link) bool
	property string
}

func TestClearOptionalLinkFields(t *testing.T) {
	cases := []optionalLinkFieldCase{
		{
			name:  "Toll",
			set:   WithToll(0),
			clear: WithoutToll(),
			present: func(link *Link) bool {
				_, present := link.Toll()
				return present
			},
			property: "toll",
		},
		{
			name:  "LengthMeters",
			set:   WithLengthMeters(0),
			clear: WithoutLengthMeters(),
			present: func(link *Link) bool {
				_, present := link.LengthMeters()
				return present
			},
			property: "length_meters",
		},
		{
			name:  "FreeSpeed",
			set:   WithFreeSpeed(0),
			clear: WithoutFreeSpeed(),
			present: func(link *Link) bool {
				_, present := link.FreeSpeed()
				return present
			},
			property: "free_speed",
		},
		{
			name:  "Capacity",
			set:   WithCapacity(0),
			clear: WithoutCapacity(),
			present: func(link *Link) bool {
				_, present := link.Capacity()
				return present
			},
			property: "capacity",
		},
		{
			name:  "LanesNum",
			set:   WithLanesNum(0),
			clear: WithoutLanesNum(),
			present: func(link *Link) bool {
				_, present := link.LanesNum()
				return present
			},
			property: "lanes_num",
		},
		{
			name:  "Grade",
			set:   WithGrade(0),
			clear: WithoutGrade(),
			present: func(link *Link) bool {
				_, present := link.Grade()
				return present
			},
			property: "grade",
		},
		{
			name:  "RightOfWayWidthMeters",
			set:   WithRightOfWayWidthMeters(0),
			clear: WithoutRightOfWayWidthMeters(),
			present: func(link *Link) bool {
				_, present := link.RightOfWayWidthMeters()
				return present
			},
			property: "row_width",
		},
		{
			name:  "GeometryID",
			set:   WithGeometryID(0),
			clear: WithoutGeometryID(),
			present: func(link *Link) bool {
				_, present := link.GeometryID()
				return present
			},
			property: "",
		},
		{
			name:  "DirFlag",
			set:   WithDirFlag(0),
			clear: WithoutDirFlag(),
			present: func(link *Link) bool {
				_, present := link.DirFlag()
				return present
			},
			property: "",
		},
		{
			name:  "ParentLinkID",
			set:   WithParentLinkID(0),
			clear: WithoutParentLinkID(),
			present: func(link *Link) bool {
				_, present := link.ParentLink()
				return present
			},
			property: "parent_link_id",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{
				{30, 60},
				{30.002, 60},
			}
			first := NewLinkFrom(1, 2, 3, WithLineGeom(shape), WithLinkName("retained"), testCase.set)
			second := NewLinkFrom(4, 2, 3, testCase.set)
			for _, field := range cases {
				field.set(first)
			}
			assert.True(t, testCase.present(first), "explicit zero is present")
			testCase.clear(first)
			testCase.clear(first)
			assert.False(t, testCase.present(first), "clearing is idempotent")
			assert.True(t, testCase.present(second), "reusing an option must not couple records")
			for _, field := range cases {
				if field.name != testCase.name {
					assert.True(t, field.present(first), "clearing %s must preserve %s", testCase.name, field.name)
				}
			}
			if testCase.property != "" {
				properties := first.GeoFeature().Properties
				assert.Contains(t, properties, testCase.property)
				assert.Nil(t, properties[testCase.property])
			}
			assert.Equal(t, shape, first.Geom())
			assert.Equal(t, "retained", first.Name())
			assert.Equal(t, second.SourceNode(), first.SourceNode())
			assert.Equal(t, second.TargetNode(), first.TargetNode())
			testCase.set(first)
			assert.True(t, testCase.present(first), "a cleared field can be set again")
		})
	}
}

func TestClearOptionalNodeFields(t *testing.T) {
	node := NewNodeFrom(1,
		WithAltitudeMeters(0),
		WithParentNodeID(0),
		WithMacroZone(0),
		WithNodeName("retained"),
	)
	WithoutAltitudeMeters()(node)
	_, hasAltitude := node.AltitudeMeters()
	assert.False(t, hasAltitude)
	_, hasParent := node.ParentNode()
	assert.True(t, hasParent)
	_, hasZone := node.MacroZone()
	assert.True(t, hasZone)
	WithoutParentNodeID()(node)
	WithoutMacroZone()(node)
	_, hasParent = node.ParentNode()
	_, hasZone = node.MacroZone()
	assert.False(t, hasParent)
	assert.False(t, hasZone)
	assert.Equal(t, "retained", node.Name())
	properties := node.GeoFeature().Properties
	assert.Nil(t, properties["z_coord"])
	assert.Nil(t, properties["zone_id"])
	assert.Nil(t, properties["parent_node_id"])
	net := NewNet()
	net.Nodes[node.ID] = node
	assert.NoError(t, net.ValidateParentReferences(), "cleared parent must not remain a dangling reference")
}
