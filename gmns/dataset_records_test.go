package gmns

import (
	"math"
	"testing"

	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestRecordValidationAndCopies(t *testing.T) {
	assert.Error(t, NewConfig(IDType("bad")).Validate())
	assert.Error(t, NewUseDefinition("car", math.NaN(), 1).Validate())
	assert.Error(t, NewUseDefinition("car", 1, -1).Validate())
	assert.NoError(t, NewUseDefinition("parking", 0, 0).Validate())
	assert.Error(t, NewUseGroup("group", []string{"car", "car"}).Validate())
	assert.Error(t, NewUseGroup("group", nil).Validate())
	shape := orb.LineString{{30, 60}, {31, 60}}
	geometry := NewGeometry(GeometryIDFromInt64(0), shape)
	shape[0][0] = 42
	assert.Equal(t, 30.0, geometry.Geom()[0][0])
	assert.NoError(t, geometry.Validate())
	assert.Error(t, NewGeometry(GeometryIDFromInt64(1), orb.LineString{{math.Inf(1), 0}, {0, 0}}).Validate())
	boundary := orb.MultiPolygon{{{{30, 60}, {31, 60}, {31, 61}, {30, 60}}}}
	zone := NewZone(NodeIDFromInt64(-1), WithZoneBoundary(boundary), WithSuperZone(NodeIDFromInt64(0)))
	boundary[0][0][0][0] = 42
	assert.NoError(t, zone.Validate())
	assert.Equal(t, 30.0, zone.Boundary()[0][0][0][0])
	parent, present := zone.SuperZone()
	assert.True(t, present)
	assert.Equal(t, NodeIDFromInt64(0), parent)
	assert.Error(t, NewZone(NodeIDFromInt64(1), WithZoneBoundary(orb.MultiPolygon{{{{0, 0}, {1, 0}, {1, 1}}}})).Validate())
}

func TestDetailedRecordIdentityAndAbsence(t *testing.T) {
	laneID := LaneIDFromString("001")
	assert.NotEqual(t, laneID, LaneIDFromString("1"))
	assert.NotEqual(t, laneID, LaneIDFromInt64(1))
	allocator := IDAllocator{}
	assert.NoError(t, allocator.ReserveLaneID(LaneIDFromInt64(0)))
	assert.Equal(t, LaneIDFromInt64(1), allocator.NextLaneID())
	segment := NewSegment(SegmentIDFromString("segment"), LinkIDFromString("road"), NodeIDFromString("node"), 0, 10, WithSegmentToll(0), WithSegmentLeftLanesAdded(-1))
	assert.NoError(t, segment.Validate())
	toll, present := segment.Toll()
	assert.True(t, present)
	assert.Zero(t, toll)
	_, present = segment.Capacity()
	assert.False(t, present)
	count, present := segment.LeftLanesAdded()
	assert.True(t, present)
	assert.Equal(t, -1, count)
}
