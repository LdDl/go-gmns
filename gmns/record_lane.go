package gmns

import "fmt"

// Lane preserves a supplied standard record independently of calculated lane indexes.
type Lane struct {
	ID           LaneID
	linkID       LinkID
	laneNum      int
	allowedUses  string
	rightBarrier string
	leftBarrier  string
	widthMeters  *float64
}

func NewLane(id LaneID, linkID LinkID, laneNum int, options ...func(*Lane)) *Lane {
	lane := &Lane{ID: id, linkID: linkID, laneNum: laneNum}
	for _, option := range options {
		option(lane)
	}
	return lane
}

func (lane *Lane) Link() LinkID {
	return lane.linkID
}

func (lane *Lane) LaneNum() int {
	return lane.laneNum
}

func (lane *Lane) AllowedUses() string {
	return lane.allowedUses
}

func WithLaneAllowedUses(value string) func(*Lane) {
	return func(lane *Lane) {
		lane.allowedUses = value
	}
}

func (lane *Lane) RightBarrier() string {
	return lane.rightBarrier
}

func WithLaneRightBarrier(value string) func(*Lane) {
	return func(lane *Lane) {
		lane.rightBarrier = value
	}
}

func (lane *Lane) LeftBarrier() string {
	return lane.leftBarrier
}

func WithLaneLeftBarrier(value string) func(*Lane) {
	return func(lane *Lane) {
		lane.leftBarrier = value
	}
}

func (lane *Lane) WidthMeters() (float64, bool) {
	if lane.widthMeters == nil {
		return 0, false
	}
	return *lane.widthMeters, true
}

func WithLaneWidthMeters(value float64) func(*Lane) {
	return func(lane *Lane) {
		lane.widthMeters = &value
	}
}

func WithoutLaneWidthMeters() func(*Lane) {
	return func(lane *Lane) {
		lane.widthMeters = nil
	}
}

func (lane *Lane) Clone() *Lane {
	if lane == nil {
		return nil
	}
	result := *lane
	if lane.widthMeters != nil {
		value := *lane.widthMeters
		result.widthMeters = &value
	}
	return &result
}

func (lane *Lane) Validate() error {
	if lane == nil {
		return fmt.Errorf("Lane is nil")
	}
	err := lane.ID.Validate()
	if err != nil {
		return err
	}
	err = lane.linkID.Validate()
	if err != nil {
		return err
	}
	if lane.laneNum < -10 || lane.laneNum > 10 {
		return fmt.Errorf("invalid laneNum")
	}
	if lane.widthMeters != nil {
		if !Finite(*lane.widthMeters) || *lane.widthMeters < 0 {
			return fmt.Errorf("invalid widthMeters")
		}
	}
	return nil
}
