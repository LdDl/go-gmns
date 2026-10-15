package macro

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
)

func (net *Net) AddLane(record *gmns.Lane) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	if _, exists := net.lanes[record.ID]; exists {
		return fmt.Errorf("lanes %v: %w", record.ID, gmns.ErrDuplicateID)
	}
	if net.lanes == nil {
		net.lanes = make(map[gmns.LaneID]*gmns.Lane)
	}
	net.lanes[record.ID] = record.Clone()
	return nil
}

func (net *Net) Lane(id gmns.LaneID) (*gmns.Lane, bool) {
	record, found := net.lanes[id]
	return record.Clone(), found
}

func (net *Net) LaneRecords() []*gmns.Lane {
	ids := make([]gmns.LaneID, 0, len(net.lanes))
	for id := range net.lanes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	records := make([]*gmns.Lane, 0, len(ids))
	for _, id := range ids {
		records = append(records, net.lanes[id].Clone())
	}
	return records
}

func (net *Net) AddSegment(record *gmns.Segment) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	if _, exists := net.segments[record.ID]; exists {
		return fmt.Errorf("segments %v: %w", record.ID, gmns.ErrDuplicateID)
	}
	if net.segments == nil {
		net.segments = make(map[gmns.SegmentID]*gmns.Segment)
	}
	net.segments[record.ID] = record.Clone()
	return nil
}

func (net *Net) Segment(id gmns.SegmentID) (*gmns.Segment, bool) {
	record, found := net.segments[id]
	return record.Clone(), found
}

func (net *Net) SegmentRecords() []*gmns.Segment {
	ids := make([]gmns.SegmentID, 0, len(net.segments))
	for id := range net.segments {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	records := make([]*gmns.Segment, 0, len(ids))
	for _, id := range ids {
		records = append(records, net.segments[id].Clone())
	}
	return records
}

func (net *Net) AddSegmentLane(record *gmns.SegmentLane) error {
	err := record.Validate()
	if err != nil {
		return err
	}
	if _, exists := net.segmentLanes[record.ID]; exists {
		return fmt.Errorf("segmentLanes %v: %w", record.ID, gmns.ErrDuplicateID)
	}
	if net.segmentLanes == nil {
		net.segmentLanes = make(map[gmns.SegmentLaneID]*gmns.SegmentLane)
	}
	net.segmentLanes[record.ID] = record.Clone()
	return nil
}

func (net *Net) SegmentLane(id gmns.SegmentLaneID) (*gmns.SegmentLane, bool) {
	record, found := net.segmentLanes[id]
	return record.Clone(), found
}

func (net *Net) SegmentLaneRecords() []*gmns.SegmentLane {
	ids := make([]gmns.SegmentLaneID, 0, len(net.segmentLanes))
	for id := range net.segmentLanes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(first, second int) bool {
		return ids[first].Less(ids[second])
	})
	records := make([]*gmns.SegmentLane, 0, len(ids))
	for _, id := range ids {
		records = append(records, net.segmentLanes[id].Clone())
	}
	return records
}

// HasDetailedRecords distinguishes supplied standard detail from calculated LanesInfo.
func (net *Net) HasDetailedRecords() bool {
	return len(net.lanes)+len(net.segments)+len(net.segmentLanes) > 0
}

func (net *Net) validateDetailedRecords() error {
	kind := net.config.IDType()
	for _, lane := range net.LaneRecords() {
		err := lane.Validate()
		if err != nil {
			return err
		}
		err = lane.ID.ValidateType(kind)
		if err != nil {
			return err
		}
		if net.Links[lane.Link()] == nil {
			return fmt.Errorf("lane %v: link %v not found", lane.ID, lane.Link())
		}
		err = gmns.ValidateAllowedUses(lane.AllowedUses(), net.useDefinitions, net.useGroups)
		if err != nil {
			return err
		}
	}
	for _, segment := range net.SegmentRecords() {
		err := segment.Validate()
		if err != nil {
			return err
		}
		err = segment.ID.ValidateType(kind)
		if err != nil {
			return err
		}
		link := net.Links[segment.Link()]
		if link == nil {
			return fmt.Errorf("segment %v: link %v not found", segment.ID, segment.Link())
		}
		if segment.RefNode() != link.SourceNode() && segment.RefNode() != link.TargetNode() {
			return fmt.Errorf("segment %v: reference node must be a link endpoint", segment.ID)
		}
		length, hasLength := link.LengthMeters()
		if hasLength && segment.EndMeters() > length {
			return fmt.Errorf("segment %v exceeds its link length", segment.ID)
		}
		_, hasToll := segment.Toll()
		if hasToll && net.config.Currency() == "" {
			return fmt.Errorf("segment %v: toll requires dataset currency", segment.ID)
		}
		err = gmns.ValidateAllowedUses(segment.AllowedUses(), net.useDefinitions, net.useGroups)
		if err != nil {
			return err
		}
	}
	for _, lane := range net.SegmentLaneRecords() {
		err := lane.Validate()
		if err != nil {
			return err
		}
		err = lane.ID.ValidateType(kind)
		if err != nil {
			return err
		}
		segment := net.segments[lane.Segment()]
		if segment == nil {
			return fmt.Errorf("segment lane %v: segment not found", lane.ID)
		}
		parent, hasParent := lane.ParentLane()
		if hasParent {
			record := net.lanes[parent]
			if record == nil || record.Link() != segment.Link() {
				return fmt.Errorf("segment lane %v: parent must belong to the segment link", lane.ID)
			}
		}
		if lane.LaneNum() == 0 && !hasParent {
			return fmt.Errorf("dropped segment lane %v requires parent_lane_id", lane.ID)
		}
		err = gmns.ValidateAllowedUses(lane.AllowedUses(), net.useDefinitions, net.useGroups)
		if err != nil {
			return err
		}
	}
	return nil
}
