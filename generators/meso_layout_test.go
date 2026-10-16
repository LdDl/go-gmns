package generators

import (
	"testing"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/stretchr/testify/assert"
)

func detailedPreparationNetwork(t *testing.T, reverse bool) *macro.Net {
	t.Helper()
	source := preparationTestNetwork(t, gmns.IDTypeString)
	link := source.Links[gmns.LinkIDFromString("13")]
	lane := gmns.NewLane(gmns.LaneIDFromString("001"), link.ID, 3, gmns.WithLaneWidthMeters(3.5))
	err := source.AddLane(lane)
	assert.NoError(t, err)
	reference := link.SourceNode()
	start, end := 40.0, 80.0
	if reverse {
		reference = link.TargetNode()
		start, end = 48, 88
	}
	segment := gmns.NewSegment(gmns.SegmentIDFromString("001"), link.ID, reference, start, end,
		gmns.WithSegmentCapacity(0), gmns.WithSegmentFreeSpeed(25), gmns.WithSegmentLanesNum(2))
	err = source.AddSegment(segment)
	assert.NoError(t, err)
	dropped := gmns.NewSegmentLane(gmns.SegmentLaneIDFromString("001"), segment.ID, 0, gmns.WithSegmentLaneParentLane(lane.ID))
	err = source.AddSegmentLane(dropped)
	assert.NoError(t, err)
	return source
}

func TestPrepareDetailedLocalDrop(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		source := detailedPreparationNetwork(t, reverse)
		expected := detailedPreparationNetwork(t, reverse)
		result, err := PrepareMesoscopic(source, WithDetailedLaneLayout())
		assert.NoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(t, expected, source)
		assert.NoError(t, result.Macro().ValidateDataset())
		assert.NoError(t, result.Meso().ValidateDataset())
		assert.Equal(t, source.LaneRecords(), result.Macro().LaneRecords())
		assert.Equal(t, source.SegmentRecords(), result.Macro().SegmentRecords())
		assert.Equal(t, source.SegmentLaneRecords(), result.Macro().SegmentLaneRecords())
		sections := result.Sections()
		assert.Len(t, sections, 4)
		assert.Equal(t, []int{1, 2, 3}, sections[0].LaneNumbers())
		assert.Equal(t, []int{1, 2}, sections[1].LaneNumbers())
		assert.Equal(t, []int{1, 2, 3}, sections[2].LaneNumbers())
		assert.Equal(t, 40.0, sections[1].StartMeters())
		assert.Equal(t, 80.0, sections[1].EndMeters())
		segment, present := sections[1].Segment()
		assert.True(t, present)
		assert.Equal(t, gmns.SegmentIDFromString("001"), segment)
		assert.Equal(t, []gmns.SegmentLaneID{gmns.SegmentLaneIDFromString("001")}, sections[1].SourceSegmentLanes())
		assert.Empty(t, sections[1].SourceLanes())
		first := result.Meso().Links[sections[0].Link()]
		middle := result.Meso().Links[sections[1].Link()]
		last := result.Meso().Links[sections[2].Link()]
		assert.Equal(t, first.TargetNode(), middle.SourceNode())
		assert.Equal(t, middle.TargetNode(), last.SourceNode())
		assert.Equal(t, first.Geom()[len(first.Geom())-1], middle.Geom()[0])
		speed, present := middle.FreeSpeed()
		assert.True(t, present)
		assert.Equal(t, 25.0, speed)
		_, present = first.FreeSpeed()
		assert.False(t, present)
		count, _ := middle.LanesNum()
		assert.Equal(t, 2, count)
		assert.Equal(t, [2]int{0, -1}, middle.LanesChange())
		rebuilt, err := PrepareMesoscopic(result.Macro(), WithDetailedLaneLayout())
		assert.NoError(t, err)
		assert.Equal(t, result, rebuilt)
		sections[1].lanes[0] = 99
		sections[1].segmentLaneIDs[0] = gmns.SegmentLaneID{}
		assert.Equal(t, []int{1, 2}, result.Sections()[1].lanes)
		assert.False(t, result.Sections()[1].segmentLaneIDs[0].IsZero())
	}
}

func TestPrepareOrderedRangesAndAbsentEnds(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	record := source.MovementRecords()[0]
	movement.WithoutIncomeLaneEnd()(record)
	movement.WithoutOutcomeLaneEnd()(record)
	err := source.UpdateMovement(record)
	assert.NoError(t, err)
	result, err := PrepareMesoscopic(source, WithUniformLaneLayout())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	stored := result.Macro().MovementRecords()[0]
	_, present := stored.IncomeLaneEnd()
	assert.False(t, present)
	_, present = stored.OutcomeLaneEnd()
	assert.False(t, present)
	assert.Equal(t, []int{3}, result.Connections()[0].inbound)
	movement.WithIncomeLane(1, 2)(record)
	movement.WithOutcomeLane(2, 3)(record)
	err = source.UpdateMovement(record)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithUniformLaneLayout())
	assertPreparationError(t, err, "WithOrderedLaneConnections")
	result, err = PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Len(t, result.Connections(), 1)
	connection := result.Connections()[0]
	assert.Equal(t, []int{1, 2}, connection.InboundLanes())
	assert.Equal(t, []int{2, 3}, connection.OutboundLanes())
	count, _ := result.Meso().Links[connection.Link()].LanesNum()
	assert.Equal(t, 2, count)
	movement.WithOutcomeLane(1, 3)(record)
	err = source.UpdateMovement(record)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithUniformLaneLayout(), WithOrderedLaneConnections())
	assertPreparationError(t, err, "equal lane counts (2 inbound, 3 outbound)")
}

func TestPrepareLeftAddedLaneAtJunction(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	link := source.Links[gmns.LinkIDFromInt64(13)]
	segment := gmns.NewSegment(gmns.SegmentIDFromInt64(1), link.ID, link.SourceNode(), 80, 128, gmns.WithSegmentLeftLanesAdded(1))
	err := source.AddSegment(segment)
	assert.NoError(t, err)
	lane := gmns.NewSegmentLane(gmns.SegmentLaneIDFromInt64(1), segment.ID, -1)
	err = source.AddSegmentLane(lane)
	assert.NoError(t, err)
	record := source.MovementRecords()[0]
	movement.WithIncomeLane(-1, 1)(record)
	movement.WithOutcomeLane(1, 2)(record)
	err = source.UpdateMovement(record)
	assert.NoError(t, err)
	result, err := PrepareMesoscopic(source, WithDetailedLaneLayout(), WithOrderedLaneConnections())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, []int{-1, 1, 2, 3}, result.Sections()[1].lanes)
	assert.Equal(t, []int{-1, 1}, result.Connections()[0].inbound)
	assert.Equal(t, []int{1, 2}, result.Connections()[0].outbound)
	assert.Equal(t, 0, result.Meso().Links[result.Connections()[0].linkID].MovementIncomeLaneStartSeqID())
}

type detailedFailureCase struct {
	name    string
	change  func(*macro.Net)
	message string
}

func TestPrepareDetailedDiagnosticsPreserveInput(t *testing.T) {
	cases := []detailedFailureCase{
		{
			name: "overlap",
			change: func(source *macro.Net) {
				link := source.Links[gmns.LinkIDFromString("13")]
				err := source.AddSegment(gmns.NewSegment(gmns.SegmentIDFromString("overlap"), link.ID, link.SourceNode(), 60, 90))
				assert.NoError(t, err)
			},
			message: "segments 001 and overlap overlap",
		},
		{
			name: "endpoint drop",
			change: func(source *macro.Net) {
				record, _ := source.Segment(gmns.SegmentIDFromString("001"))
				replacement := gmns.NewSegment(record.ID, record.Link(), record.RefNode(), 40, 128, gmns.WithSegmentRightLanesAdded(-1))
				err := source.UpdateSegment(replacement)
				assert.NoError(t, err)
			},
			message: "movement 49: inbound lane range 3..3",
		},
		{
			name: "ambiguous count",
			change: func(source *macro.Net) {
				link := source.Links[gmns.LinkIDFromString("13")]
				err := source.AddSegment(gmns.NewSegment(gmns.SegmentIDFromString("ambiguous"), link.ID, link.SourceNode(), 90, 100, gmns.WithSegmentLanesNum(2)))
				assert.NoError(t, err)
			},
			message: "segment ambiguous: lanes 2 disagrees",
		},
		{
			name: "lane width",
			change: func(source *macro.Net) {
				record, _ := source.Lane(gmns.LaneIDFromString("001"))
				gmns.WithLaneWidthMeters(4)(record)
				err := source.UpdateLane(record)
				assert.NoError(t, err)
			},
			message: "lane 001 on link 13: unsupported generation input: variable lane width",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := detailedPreparationNetwork(t, false)
			expected := detailedPreparationNetwork(t, false)
			testCase.change(source)
			testCase.change(expected)
			result, err := PrepareMesoscopic(source, WithDetailedLaneLayout())
			assert.Nil(t, result)
			assertPreparationError(t, err, testCase.message)
			assert.Equal(t, expected, source)
		})
	}
}

func TestPrepareShortTerminalSections(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	link := source.Links[gmns.LinkIDFromInt64(13)]
	segment := gmns.NewSegment(gmns.SegmentIDFromInt64(1), link.ID, link.SourceNode(), 0.001, 127.999)
	err := source.AddSegment(segment)
	assert.NoError(t, err)
	result, err := PrepareMesoscopic(source, WithDetailedLaneLayout())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	for _, section := range result.Sections() {
		length, present := result.Meso().Links[section.Link()].LengthMeters()
		assert.True(t, present)
		assert.Greater(t, length, 0.0)
	}
}

func TestPrepareSegmentOverridesAtConnector(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	for _, name := range []string{"auto", "bike"} {
		err := source.AddUseDefinition(gmns.NewUseDefinition(name, 1, 1))
		assert.NoError(t, err)
	}
	link := source.Links[gmns.LinkIDFromInt64(13)]
	macro.WithAllowedUses("bike")(link)
	segment := gmns.NewSegment(gmns.SegmentIDFromInt64(1), link.ID, link.SourceNode(), 40, 128,
		gmns.WithSegmentAllowedUses("auto"), gmns.WithSegmentFreeSpeed(0), gmns.WithSegmentCapacity(12.5),
		gmns.WithSegmentGrade(-3), gmns.WithSegmentRightOfWayWidthMeters(10), gmns.WithSegmentJurisdiction("city"))
	err := source.AddSegment(segment)
	assert.NoError(t, err)
	movementRecord := source.MovementRecords()[0]
	movement.WithAllowedUses("auto")(movementRecord)
	err = source.UpdateMovement(movementRecord)
	assert.NoError(t, err)
	result, err := PrepareMesoscopic(source, WithDetailedLaneLayout())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	first := result.Meso().Links[result.Sections()[0].linkID]
	last := result.Meso().Links[result.Sections()[1].linkID]
	connector := result.Meso().Links[result.Connections()[0].linkID]
	assert.Equal(t, "bike", first.AllowedUses())
	assert.Equal(t, "auto", last.AllowedUses())
	assert.Equal(t, "auto", connector.AllowedUses())
	speed, present := connector.FreeSpeed()
	assert.True(t, present)
	assert.Zero(t, speed)
	capacity, present := connector.Capacity()
	assert.True(t, present)
	assert.Equal(t, 12.5, capacity)
	grade, present := last.Grade()
	assert.True(t, present)
	assert.Equal(t, -3.0, grade)
	assert.Equal(t, "city", last.Jurisdiction())
	gmns.WithSegmentAllowedUses("bike")(segment)
	err = source.UpdateSegment(segment)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithDetailedLaneLayout())
	assertPreparationError(t, err, "movement 49 allowed_uses: no common use")
}

func TestPrepareDetailedRejectsInteriorDropAndIncompleteAddition(t *testing.T) {
	source := preparationTestNetwork(t, gmns.IDTypeInteger)
	link := source.Links[gmns.LinkIDFromInt64(13)]
	lane := gmns.NewLane(gmns.LaneIDFromInt64(2), link.ID, 2)
	err := source.AddLane(lane)
	assert.NoError(t, err)
	segment := gmns.NewSegment(gmns.SegmentIDFromInt64(1), link.ID, link.SourceNode(), 40, 80)
	err = source.AddSegment(segment)
	assert.NoError(t, err)
	detail := gmns.NewSegmentLane(gmns.SegmentLaneIDFromInt64(1), segment.ID, 0, gmns.WithSegmentLaneParentLane(lane.ID))
	err = source.AddSegmentLane(detail)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithDetailedLaneLayout())
	assertPreparationError(t, err, "interior drops are unsupported")
	detail = gmns.NewSegmentLane(detail.ID, segment.ID, 5)
	err = source.UpdateSegmentLane(detail)
	assert.NoError(t, err)
	_, err = PrepareMesoscopic(source, WithDetailedLaneLayout())
	assertPreparationError(t, err, "added lanes must be consecutive; missing lane 4")
}
