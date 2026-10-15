package gmns

import "fmt"

// Segment preserves a supplied standard record independently of calculated lane indexes.
type Segment struct {
	ID                    SegmentID
	linkID                LinkID
	refNodeID             NodeID
	startMeters           float64
	endMeters             float64
	grade                 *float64
	capacity              *float64
	freeSpeed             *float64
	lanes                 *int
	leftLanesAdded        *int
	rightLanesAdded       *int
	bikeFacility          string
	pedFacility           string
	parking               string
	allowedUses           string
	toll                  *float64
	jurisdiction          string
	rightOfWayWidthMeters *float64
}

func NewSegment(id SegmentID, linkID LinkID, refNodeID NodeID, startMeters float64, endMeters float64, options ...func(*Segment)) *Segment {
	segment := &Segment{ID: id, linkID: linkID, refNodeID: refNodeID, startMeters: startMeters, endMeters: endMeters}
	for _, option := range options {
		option(segment)
	}
	return segment
}

func (segment *Segment) Link() LinkID {
	return segment.linkID
}

func (segment *Segment) RefNode() NodeID {
	return segment.refNodeID
}

func (segment *Segment) StartMeters() float64 {
	return segment.startMeters
}

func (segment *Segment) EndMeters() float64 {
	return segment.endMeters
}

func (segment *Segment) Grade() (float64, bool) {
	if segment.grade == nil {
		return 0, false
	}
	return *segment.grade, true
}

func WithSegmentGrade(value float64) func(*Segment) {
	return func(segment *Segment) {
		segment.grade = &value
	}
}

func WithoutSegmentGrade() func(*Segment) {
	return func(segment *Segment) {
		segment.grade = nil
	}
}

func (segment *Segment) Capacity() (float64, bool) {
	if segment.capacity == nil {
		return 0, false
	}
	return *segment.capacity, true
}

func WithSegmentCapacity(value float64) func(*Segment) {
	return func(segment *Segment) {
		segment.capacity = &value
	}
}

func WithoutSegmentCapacity() func(*Segment) {
	return func(segment *Segment) {
		segment.capacity = nil
	}
}

func (segment *Segment) FreeSpeed() (float64, bool) {
	if segment.freeSpeed == nil {
		return 0, false
	}
	return *segment.freeSpeed, true
}

func WithSegmentFreeSpeed(value float64) func(*Segment) {
	return func(segment *Segment) {
		segment.freeSpeed = &value
	}
}

func WithoutSegmentFreeSpeed() func(*Segment) {
	return func(segment *Segment) {
		segment.freeSpeed = nil
	}
}

func (segment *Segment) LanesNum() (int, bool) {
	if segment.lanes == nil {
		return 0, false
	}
	return *segment.lanes, true
}

func WithSegmentLanesNum(value int) func(*Segment) {
	return func(segment *Segment) {
		segment.lanes = &value
	}
}

func WithoutSegmentLanesNum() func(*Segment) {
	return func(segment *Segment) {
		segment.lanes = nil
	}
}

func (segment *Segment) LeftLanesAdded() (int, bool) {
	if segment.leftLanesAdded == nil {
		return 0, false
	}
	return *segment.leftLanesAdded, true
}

func WithSegmentLeftLanesAdded(value int) func(*Segment) {
	return func(segment *Segment) {
		segment.leftLanesAdded = &value
	}
}

func WithoutSegmentLeftLanesAdded() func(*Segment) {
	return func(segment *Segment) {
		segment.leftLanesAdded = nil
	}
}

func (segment *Segment) RightLanesAdded() (int, bool) {
	if segment.rightLanesAdded == nil {
		return 0, false
	}
	return *segment.rightLanesAdded, true
}

func WithSegmentRightLanesAdded(value int) func(*Segment) {
	return func(segment *Segment) {
		segment.rightLanesAdded = &value
	}
}

func WithoutSegmentRightLanesAdded() func(*Segment) {
	return func(segment *Segment) {
		segment.rightLanesAdded = nil
	}
}

func (segment *Segment) BikeFacility() string {
	return segment.bikeFacility
}

func WithSegmentBikeFacility(value string) func(*Segment) {
	return func(segment *Segment) {
		segment.bikeFacility = value
	}
}

func (segment *Segment) PedFacility() string {
	return segment.pedFacility
}

func WithSegmentPedFacility(value string) func(*Segment) {
	return func(segment *Segment) {
		segment.pedFacility = value
	}
}

func (segment *Segment) Parking() string {
	return segment.parking
}

func WithSegmentParking(value string) func(*Segment) {
	return func(segment *Segment) {
		segment.parking = value
	}
}

func (segment *Segment) AllowedUses() string {
	return segment.allowedUses
}

func WithSegmentAllowedUses(value string) func(*Segment) {
	return func(segment *Segment) {
		segment.allowedUses = value
	}
}

func (segment *Segment) Toll() (float64, bool) {
	if segment.toll == nil {
		return 0, false
	}
	return *segment.toll, true
}

func WithSegmentToll(value float64) func(*Segment) {
	return func(segment *Segment) {
		segment.toll = &value
	}
}

func WithoutSegmentToll() func(*Segment) {
	return func(segment *Segment) {
		segment.toll = nil
	}
}

func (segment *Segment) Jurisdiction() string {
	return segment.jurisdiction
}

func WithSegmentJurisdiction(value string) func(*Segment) {
	return func(segment *Segment) {
		segment.jurisdiction = value
	}
}

func (segment *Segment) RightOfWayWidthMeters() (float64, bool) {
	if segment.rightOfWayWidthMeters == nil {
		return 0, false
	}
	return *segment.rightOfWayWidthMeters, true
}

func WithSegmentRightOfWayWidthMeters(value float64) func(*Segment) {
	return func(segment *Segment) {
		segment.rightOfWayWidthMeters = &value
	}
}

func WithoutSegmentRightOfWayWidthMeters() func(*Segment) {
	return func(segment *Segment) {
		segment.rightOfWayWidthMeters = nil
	}
}

func (segment *Segment) Clone() *Segment {
	if segment == nil {
		return nil
	}
	result := *segment
	if segment.grade != nil {
		value := *segment.grade
		result.grade = &value
	}
	if segment.capacity != nil {
		value := *segment.capacity
		result.capacity = &value
	}
	if segment.freeSpeed != nil {
		value := *segment.freeSpeed
		result.freeSpeed = &value
	}
	if segment.lanes != nil {
		value := *segment.lanes
		result.lanes = &value
	}
	if segment.leftLanesAdded != nil {
		value := *segment.leftLanesAdded
		result.leftLanesAdded = &value
	}
	if segment.rightLanesAdded != nil {
		value := *segment.rightLanesAdded
		result.rightLanesAdded = &value
	}
	if segment.toll != nil {
		value := *segment.toll
		result.toll = &value
	}
	if segment.rightOfWayWidthMeters != nil {
		value := *segment.rightOfWayWidthMeters
		result.rightOfWayWidthMeters = &value
	}
	return &result
}

func (segment *Segment) Validate() error {
	if segment == nil {
		return fmt.Errorf("Segment is nil")
	}
	err := segment.ID.Validate()
	if err != nil {
		return err
	}
	err = segment.linkID.Validate()
	if err != nil {
		return err
	}
	err = segment.refNodeID.Validate()
	if err != nil {
		return err
	}
	if !Finite(segment.startMeters) || segment.startMeters < 0 {
		return fmt.Errorf("invalid startMeters")
	}
	if !Finite(segment.endMeters) || segment.endMeters < 0 {
		return fmt.Errorf("invalid endMeters")
	}
	if segment.grade != nil {
		if !Finite(*segment.grade) || *segment.grade < -100 || *segment.grade > 100 {
			return fmt.Errorf("invalid grade")
		}
	}
	if segment.capacity != nil {
		if !Finite(*segment.capacity) || *segment.capacity < 0 {
			return fmt.Errorf("invalid capacity")
		}
	}
	if segment.freeSpeed != nil {
		if !Finite(*segment.freeSpeed) || *segment.freeSpeed < 0 || *segment.freeSpeed > 200 {
			return fmt.Errorf("invalid freeSpeed")
		}
	}
	if segment.lanes != nil {
	}
	if segment.leftLanesAdded != nil {
	}
	if segment.rightLanesAdded != nil {
	}
	if segment.toll != nil {
		if !Finite(*segment.toll) {
			return fmt.Errorf("invalid toll")
		}
	}
	if segment.rightOfWayWidthMeters != nil {
		if !Finite(*segment.rightOfWayWidthMeters) || *segment.rightOfWayWidthMeters < 0 {
			return fmt.Errorf("invalid rightOfWayWidthMeters")
		}
	}
	if segment.startMeters > segment.endMeters {
		return fmt.Errorf("segment start exceeds end")
	}
	return nil
}
