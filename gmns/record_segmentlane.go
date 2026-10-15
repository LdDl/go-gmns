package gmns

import "fmt"

// SegmentLane preserves a supplied standard record independently of calculated lane indexes.
type SegmentLane struct {
	ID           SegmentLaneID
	segmentID    SegmentID
	laneNum      int
	parentLaneID *LaneID
	allowedUses  string
	rightBarrier string
	leftBarrier  string
	widthMeters  *float64
}

func NewSegmentLane(id SegmentLaneID, segmentID SegmentID, laneNum int, options ...func(*SegmentLane)) *SegmentLane {
	lane := &SegmentLane{ID: id, segmentID: segmentID, laneNum: laneNum}
	for _, option := range options {
		option(lane)
	}
	return lane
}

func (lane *SegmentLane) Segment() SegmentID {
	return lane.segmentID
}

func (lane *SegmentLane) LaneNum() int {
	return lane.laneNum
}

func (lane *SegmentLane) ParentLane() (LaneID, bool) {
	if lane.parentLaneID == nil {
		return LaneID{}, false
	}
	return *lane.parentLaneID, true
}

func WithSegmentLaneParentLane(value LaneID) func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.parentLaneID = &value
	}
}

func WithoutSegmentLaneParentLane() func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.parentLaneID = nil
	}
}

func (lane *SegmentLane) AllowedUses() string {
	return lane.allowedUses
}

func WithSegmentLaneAllowedUses(value string) func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.allowedUses = value
	}
}

func (lane *SegmentLane) RightBarrier() string {
	return lane.rightBarrier
}

func WithSegmentLaneRightBarrier(value string) func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.rightBarrier = value
	}
}

func (lane *SegmentLane) LeftBarrier() string {
	return lane.leftBarrier
}

func WithSegmentLaneLeftBarrier(value string) func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.leftBarrier = value
	}
}

func (lane *SegmentLane) WidthMeters() (float64, bool) {
	if lane.widthMeters == nil {
		return 0, false
	}
	return *lane.widthMeters, true
}

func WithSegmentLaneWidthMeters(value float64) func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.widthMeters = &value
	}
}

func WithoutSegmentLaneWidthMeters() func(*SegmentLane) {
	return func(lane *SegmentLane) {
		lane.widthMeters = nil
	}
}

func (lane *SegmentLane) Clone() *SegmentLane {
	if lane == nil {
		return nil
	}
	result := *lane
	if lane.parentLaneID != nil {
		value := *lane.parentLaneID
		result.parentLaneID = &value
	}
	if lane.widthMeters != nil {
		value := *lane.widthMeters
		result.widthMeters = &value
	}
	return &result
}

func (lane *SegmentLane) Validate() error {
	if lane == nil {
		return fmt.Errorf("SegmentLane is nil")
	}
	err := lane.ID.Validate()
	if err != nil {
		return err
	}
	err = lane.segmentID.Validate()
	if err != nil {
		return err
	}
	if lane.laneNum < -10 || lane.laneNum > 10 {
		return fmt.Errorf("invalid laneNum")
	}
	if lane.parentLaneID != nil {
		err = lane.parentLaneID.Validate()
		if err != nil {
			return err
		}
	}
	if lane.widthMeters != nil {
		if !Finite(*lane.widthMeters) || *lane.widthMeters < 0 {
			return fmt.Errorf("invalid widthMeters")
		}
	}
	return nil
}
