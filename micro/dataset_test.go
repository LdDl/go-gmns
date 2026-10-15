package micro

import (
	"math"
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func preparedDataset(t *testing.T) *Net {
	t.Helper()
	net := NewNet()
	config := gmns.NewConfig(gmns.IDTypeString, gmns.WithDatasetName("prepared"), gmns.WithCurrency("USD"))
	assert.NoError(t, net.SetConfig(config))
	first := gmns.NodeIDFromString("001")
	last := gmns.NodeIDFromString("1")
	road := gmns.LinkIDFromString("road")
	zoneID := gmns.NodeIDFromString("zone")
	shapeID := gmns.GeometryIDFromString("shape")
	shape := orb.LineString{{30, 60}, {30.001, 60}}
	net.Nodes[first] = NewNodeFrom(first, WithPointGeom(shape[0]), WithZoneID(zoneID))
	net.Nodes[last] = NewNodeFrom(last, WithPointGeom(shape[1]))
	net.Links[road] = NewLinkFrom(road, first, last, WithLineGeom(shape), WithGeometryID(shapeID), WithDirFlag(gmns.DIR_FLAG_FORWARD), WithLengthMeters(100), WithAllowedUses("motor"))
	assert.NoError(t, net.AddGeometry(gmns.NewGeometry(shapeID, shape)))
	assert.NoError(t, net.AddZone(gmns.NewZone(zoneID)))
	assert.NoError(t, net.AddUseDefinition(gmns.NewUseDefinition("car", 1.2, 1)))
	assert.NoError(t, net.AddUseGroup(gmns.NewUseGroup("motor", []string{"car"})))
	assert.NoError(t, net.ValidateDataset())
	return net
}

func TestDatasetCopiesAndTransactionalEdits(t *testing.T) {
	net := preparedDataset(t)
	road := gmns.LinkIDFromString("road")
	shapeID := gmns.GeometryIDFromString("shape")
	zoneID := gmns.NodeIDFromString("zone")
	config := net.Config()
	gmns.WithDatasetName("changed")(config)
	assert.Equal(t, "prepared", net.Config().DatasetName())
	assert.Equal(t, "meter", net.Config().LongLength())
	assert.Equal(t, "km/h", net.Config().Speed())
	assert.Equal(t, "EPSG:4326", net.Config().CRS())
	geometry, found := net.Geometry(shapeID)
	assert.True(t, found)
	geometry.ID = gmns.GeometryIDFromString("changed")
	shape := geometry.Geom()
	shape[0][0] = 100
	original, found := net.Geometry(shapeID)
	assert.True(t, found)
	assert.Equal(t, 30.0, original.Geom()[0][0])
	assert.ErrorIs(t, net.AddGeometry(original), gmns.ErrDuplicateID)
	assert.Error(t, net.RemoveGeometry(shapeID))
	assert.Error(t, net.RemoveZone(zoneID))
	assert.Error(t, net.RemoveUseDefinition("car"))
	assert.Error(t, net.RemoveUseGroup("motor"))
	badShape := original.Geom()
	badShape[0][0] = 31
	assert.Error(t, net.UpdateGeometry(gmns.NewGeometry(shapeID, badShape)))
	assert.NoError(t, net.ValidateDataset(), "failed update must roll back")
	updated := gmns.NewZone(zoneID, gmns.WithZoneName("renamed"))
	assert.NoError(t, net.UpdateZone(updated))
	gmns.WithZoneName("outside")(updated)
	stored, found := net.Zone(zoneID)
	assert.True(t, found)
	assert.Equal(t, "renamed", stored.Name())
	WithAllowedUses("")(net.Links[road])
	assert.NoError(t, net.RemoveUseGroup("motor"))
	assert.NoError(t, net.RemoveUseDefinition("car"))
	assert.NoError(t, net.ValidateDataset())
}

type invalidDatasetCase struct {
	name   string
	change func(*Net)
}

func TestDatasetValidationRejectsBrokenRecords(t *testing.T) {
	cases := []invalidDatasetCase{
		{
			name: "dangling geometry",
			change: func(net *Net) {
				WithGeometryID(gmns.GeometryIDFromString("missing"))(net.Links[gmns.LinkIDFromString("road")])
			},
		},
		{
			name: "wrong orientation",
			change: func(net *Net) {
				WithDirFlag(gmns.DIR_FLAG_REVERSE)(net.Links[gmns.LinkIDFromString("road")])
			},
		},
		{
			name: "invalid grade",
			change: func(net *Net) {
				WithGrade(101)(net.Links[gmns.LinkIDFromString("road")])
			},
		},
		{
			name: "nonfinite capacity",
			change: func(net *Net) {
				WithCapacity(math.NaN())(net.Links[gmns.LinkIDFromString("road")])
			},
		},
		{
			name: "undefined use",
			change: func(net *Net) {
				WithAllowedUses("missing")(net.Links[gmns.LinkIDFromString("road")])
			},
		},
		{
			name: "coordinate outside WGS84",
			change: func(net *Net) {
				WithPointGeom(orb.Point{200, 60})(net.Nodes[gmns.NodeIDFromString("001")])
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			net := preparedDataset(t)
			testCase.change(net)
			err := net.ValidateDataset()
			assert.Error(t, err)
			assert.EqualError(t, net.ValidateDataset(), err.Error())
		})
	}
}

func TestDatasetDetailedRecordsAndOwnedMovements(t *testing.T) {
	net := preparedDataset(t)
	road := gmns.LinkIDFromString("road")
	first := gmns.NodeIDFromString("001")
	last := gmns.NodeIDFromString("1")
	laneID := gmns.LaneIDFromString("001")
	segmentID := gmns.SegmentIDFromString("segment")
	segmentLaneID := gmns.SegmentLaneIDFromString("drop")
	lane := gmns.NewLane(laneID, road, -1, gmns.WithLaneWidthMeters(3.5), gmns.WithLaneAllowedUses("car"))
	assert.NoError(t, net.AddLane(lane))
	gmns.WithLaneWidthMeters(99)(lane)
	stored, found := net.Lane(laneID)
	assert.True(t, found)
	width, present := stored.WidthMeters()
	assert.True(t, present)
	assert.Equal(t, 3.5, width)
	assert.NoError(t, net.AddSegment(gmns.NewSegment(segmentID, road, first, 10, 50, gmns.WithSegmentGrade(-10))))
	assert.NoError(t, net.AddSegmentLane(gmns.NewSegmentLane(segmentLaneID, segmentID, 0, gmns.WithSegmentLaneParentLane(laneID))))
	assert.True(t, net.HasDetailedRecords())
	assert.NoError(t, net.ValidateDataset())
	assert.Error(t, net.RemoveLane(laneID))
	assert.Error(t, net.RemoveSegment(segmentID))
	assert.Error(t, net.UpdateSegment(gmns.NewSegment(segmentID, road, first, 0, 101)))
	assert.NoError(t, net.ValidateDataset())
	assert.NoError(t, net.RemoveSegmentLane(segmentLaneID))
	assert.NoError(t, net.RemoveSegment(segmentID))
	assert.NoError(t, net.RemoveLane(laneID))
	assert.False(t, net.HasDetailedRecords())
	// The movement refers to this network's own links, irrespective of resolution.
	back := gmns.LinkIDFromString("back")
	net.Links[back] = NewLinkFrom(back, last, first)
	identifier := gmns.MovementIDFromString("turn")
	mvmt := movement.NewMovement(identifier, last, road, back, "", "", movement.WithPenalty(0), movement.WithCapacity(125.5))
	assert.NoError(t, net.AddMovement(mvmt))
	movement.WithPenalty(99)(mvmt)
	storedMovement, found := net.Movement(identifier)
	assert.True(t, found)
	penalty, present := storedMovement.Penalty()
	assert.True(t, present)
	assert.Zero(t, penalty)
	assert.NoError(t, net.ValidateDataset())
	movement.WithNodeID(first)(storedMovement)
	assert.Error(t, net.UpdateMovement(storedMovement))
	assert.NoError(t, net.ValidateDataset())
	snapshot := net.MovementStorage()
	delete(snapshot, identifier)
	assert.Len(t, net.MovementRecords(), 1)
}
