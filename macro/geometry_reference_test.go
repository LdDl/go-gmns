package macro

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/paulmach/orb"
)

func TestGeometryAttributesAbsent(t *testing.T) {
	shape := orb.LineString{{37, 55}, {37.01, 55.01}}
	link := NewLinkFrom(10, 1, 2, WithLineGeom(shape))
	_, hasGeometryID := link.GeometryID()
	_, hasDirFlag := link.DirFlag()
	if hasGeometryID || hasDirFlag {
		t.Fatal("existing constructor must not invent geometry attributes")
	}
	if !link.Geom().Equal(shape) {
		t.Fatal("existing geometry changed")
	}
}

type geometryAttributeCase struct {
	name       string
	geometryID gmns.GeometryID
	dirFlag    gmns.DirectionFlag
}

func TestGeometryAttributes(t *testing.T) {
	cases := []geometryAttributeCase{
		{name: "forward", geometryID: 12, dirFlag: gmns.DIR_FLAG_FORWARD},
		{name: "reverse", geometryID: 12, dirFlag: gmns.DIR_FLAG_REVERSE},
		{name: "explicit zero", geometryID: 0, dirFlag: gmns.DIR_FLAG_NONE},
		{name: "negative identifier", geometryID: -1, dirFlag: gmns.DIR_FLAG_FORWARD},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shape := orb.LineString{{37, 55}, {37.01, 55.01}}
			link := NewLinkFrom(10, 1, 2,
				WithLineGeom(shape),
				WithGeometryID(testCase.geometryID),
				WithDirFlag(testCase.dirFlag),
			)
			geometryID, hasGeometryID := link.GeometryID()
			dirFlag, hasDirFlag := link.DirFlag()
			if !hasGeometryID || geometryID != testCase.geometryID {
				t.Fatalf("geometry reference: %d, present: %t", geometryID, hasGeometryID)
			}
			if !hasDirFlag || dirFlag != testCase.dirFlag {
				t.Fatalf("orientation: %d, present: %t", dirFlag, hasDirFlag)
			}
			if !link.Geom().Equal(shape) || link.SourceNode() != 1 || link.TargetNode() != 2 {
				t.Fatal("metadata must not reverse geometry or change topology")
			}
		})
	}
}

func TestSharedGeometryAttributesKeepTravelPaths(t *testing.T) {
	forwardShape := orb.LineString{{37, 55}, {37.005, 55.002}, {37.01, 55.01}}
	backwardShape := forwardShape.Clone()
	backwardShape.Reverse()
	forward := NewLinkFrom(10, 1, 2, WithLineGeom(forwardShape), WithGeometryID(8), WithDirFlag(gmns.DIR_FLAG_FORWARD))
	backward := NewLinkFrom(11, 2, 1, WithLineGeom(backwardShape), WithGeometryID(8), WithDirFlag(gmns.DIR_FLAG_REVERSE))
	forwardID, _ := forward.GeometryID()
	backwardID, _ := backward.GeometryID()
	if forwardID != backwardID || !backward.Geom().Equal(backwardShape) || !forward.Geom().Equal(forwardShape) {
		t.Fatal("shared reference changed existing travel paths")
	}
	WithGeometryID(9)(backward)
	forwardID, _ = forward.GeometryID()
	if forwardID != 8 {
		t.Fatal("updating one link changed another link's reference")
	}
}
