package generators

import (
	"fmt"
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
)

type preparationSection struct {
	start        float64
	end          float64
	segment      *gmns.Segment
	numbers      []int
	changes      [2]int
	lanes        []*gmns.Lane
	segmentLanes []*gmns.SegmentLane
}

type preparationLayout map[gmns.LinkID][]preparationSection

// WithDetailedLaneLayout accepts standard segments and edge lane additions/drops.
// Unspecified portions use the link's supplied count and lanes numbered 1..lanes.
func WithDetailedLaneLayout() MesoPreparationOption {
	return func(options *mesoPreparationOptions) {
		options.detailedLaneLayout = true
	}
}

// WithOrderedLaneConnections pairs equal-sized movement scopes from left to right.
func WithOrderedLaneConnections() MesoPreparationOption {
	return func(options *mesoPreparationOptions) {
		options.orderedLaneConnections = true
	}
}

func prepareLaneLayout(source *macro.Net, working *macro.Net) (preparationLayout, error) {
	layout := make(preparationLayout)
	for _, link := range macroLinksToSlice(working.Links) {
		count, _ := link.LanesNum()
		length, _ := link.LengthMeters()
		base := preparationSection{start: 0, end: length, numbers: sectionLaneNumbers(count, 0, 0)}
		seen := make(map[int]bool)
		for _, lane := range source.LaneRecords() {
			if lane.Link() != link.ID {
				continue
			}
			number := lane.LaneNum()
			if number < 1 || number > count || seen[number] {
				return nil, fmt.Errorf("lane %v: duplicate or unsupported base lane number %d on link %v; expected unique numbers in 1..%d", lane.ID, number, link.ID, count)
			}
			seen[number] = true
			base.lanes = append(base.lanes, lane)
		}
		segments := make([]preparationSection, 0)
		for _, segment := range source.SegmentRecords() {
			if segment.Link() != link.ID {
				continue
			}
			section := preparationSection{start: segment.StartMeters(), end: segment.EndMeters(), segment: segment, lanes: base.lanes}
			if segment.RefNode() == link.TargetNode() {
				section.start = length - segment.EndMeters()
				section.end = length - segment.StartMeters()
			}
			if section.start >= section.end {
				return nil, fmt.Errorf("segment %v: preparation requires a positive interval", segment.ID)
			}
			toll, present := segment.Toll()
			if present && toll != 0 {
				return nil, fmt.Errorf("%w: segment %v toll: placement on generated sections is undefined", ErrUnsupportedGenerationInput, segment.ID)
			}
			for _, lane := range source.SegmentLaneRecords() {
				if lane.Segment() == segment.ID {
					section.segmentLanes = append(section.segmentLanes, lane)
				}
			}
			err := prepareSectionLanes(source, count, &section)
			if err != nil {
				return nil, err
			}
			segments = append(segments, section)
		}
		sort.Slice(segments, func(first, second int) bool {
			if segments[first].start == segments[second].start {
				return segments[first].segment.ID.Less(segments[second].segment.ID)
			}
			return segments[first].start < segments[second].start
		})
		cursor := 0.0
		for index, section := range segments {
			if section.start < cursor {
				return nil, fmt.Errorf("%w: segments %v and %v overlap on link %v; nested and overlapping intervals are not supported by preparation", ErrUnsupportedGenerationInput, segments[index-1].segment.ID, section.segment.ID, link.ID)
			}
			if section.start > cursor {
				gap := base
				gap.start = cursor
				gap.end = section.start
				layout[link.ID] = append(layout[link.ID], gap)
			}
			layout[link.ID] = append(layout[link.ID], section)
			cursor = section.end
		}
		if cursor < length {
			base.start = cursor
			layout[link.ID] = append(layout[link.ID], base)
		}
		info := macro.LanesInfo{LanesChangePoints: []float64{0}}
		for _, section := range layout[link.ID] {
			err := validateSectionLaneAttributes(working, link, &section)
			if err != nil {
				return nil, err
			}
			info.LanesList = append(info.LanesList, len(section.numbers))
			info.LanesChange = append(info.LanesChange, section.changes)
			info.LanesChangePoints = append(info.LanesChangePoints, section.end)
		}
		macro.WithLanesInfo(info)(link)
	}
	return layout, nil
}

func sectionLaneNumbers(base, left, right int) []int {
	numbers := make([]int, 0, base+left+right)
	for number := -left; number < 0; number++ {
		numbers = append(numbers, number)
	}
	start := 1
	if left < 0 {
		start -= left
	}
	for number := start; number <= base+right; number++ {
		numbers = append(numbers, number)
	}
	return numbers
}

func prepareSectionLanes(source *macro.Net, base int, section *preparationSection) error {
	segment := section.segment
	left, hasLeft := segment.LeftLanesAdded()
	right, hasRight := segment.RightLanesAdded()
	changed := make(map[int]bool)
	parents := make(map[gmns.LaneID]bool)
	added := make(map[int]bool)
	dropped := make(map[int]bool)
	for _, lane := range section.segmentLanes {
		number := lane.LaneNum()
		parentID, hasParent := lane.ParentLane()
		if number != 0 && changed[number] {
			return fmt.Errorf("segment_lane %v: duplicate lane number %d in segment %v", lane.ID, number, segment.ID)
		}
		changed[number] = true
		if hasParent {
			if parents[parentID] {
				return fmt.Errorf("segment_lane %v: duplicate parent lane %v in segment %v", lane.ID, parentID, segment.ID)
			}
			parents[parentID] = true
			parent, _ := source.Lane(parentID)
			if number == 0 {
				dropped[parent.LaneNum()] = true
			} else if number != parent.LaneNum() {
				return fmt.Errorf("%w: segment_lane %v renumbers parent lane %v; only unchanged numbering and edge drops are supported", ErrUnsupportedGenerationInput, lane.ID, parentID)
			}
		} else {
			if number >= 1 && number <= base {
				return fmt.Errorf("segment_lane %v: existing lane %d requires parent_lane_id", lane.ID, number)
			}
			added[number] = true
		}
	}
	if !hasLeft && !hasRight {
		for number := 1; number <= base && dropped[number]; number++ {
			left--
		}
		for number := base; number > -left && dropped[number]; number-- {
			right--
		}
		for number := range added {
			if number < 0 {
				if left < 0 {
					return fmt.Errorf("segment %v: simultaneous addition and drop on the left is unsupported", segment.ID)
				}
				if -number > left {
					left = -number
				}
			} else if number-base > right {
				if right < 0 {
					return fmt.Errorf("segment %v: simultaneous addition and drop on the right is unsupported", segment.ID)
				}
				right = number - base
			}
		}
	}
	if left < -base || right < -base || left > 99 || right > 99 || base+left+right < 1 || base+left+right > 99 {
		return fmt.Errorf("segment %v: edge changes must leave 1..99 lanes", segment.ID)
	}
	numbers := sectionLaneNumbers(base, left, right)
	if len(numbers) != base+left+right {
		return fmt.Errorf("segment %v: inconsistent edge lane changes", segment.ID)
	}
	count, hasCount := segment.LanesNum()
	if hasCount && count != len(numbers) {
		return fmt.Errorf("segment %v: lanes %d disagrees with base count and edge changes (%d); supply explicit sides or segment_lane records", segment.ID, count, len(numbers))
	}
	for number := range dropped {
		if lanePosition(numbers, number) >= 0 {
			return fmt.Errorf("segment %v: dropped lane %d is still in the layout; interior drops are unsupported", segment.ID, number)
		}
	}
	for _, lane := range section.segmentLanes {
		if lane.LaneNum() != 0 && lanePosition(numbers, lane.LaneNum()) < 0 {
			return fmt.Errorf("segment_lane %v: lane %d is outside the segment layout", lane.ID, lane.LaneNum())
		}
	}
	if !hasLeft && !hasRight {
		for _, number := range numbers {
			if (number < 0 || number > base) && !added[number] {
				return fmt.Errorf("segment %v: added lanes must be consecutive; missing lane %d", segment.ID, number)
			}
		}
	}
	section.numbers = numbers
	section.changes = [2]int{left, right}
	return nil
}

func lanePosition(numbers []int, number int) int {
	for index, candidate := range numbers {
		if number == candidate {
			return index
		}
	}
	return -1
}

func sectionUses(link *macro.Link, section *preparationSection) string {
	if section.segment != nil && section.segment.AllowedUses() != "" {
		return section.segment.AllowedUses()
	}
	return link.AllowedUses()
}

func validateSectionLaneAttributes(source *macro.Net, link *macro.Link, section *preparationSection) error {
	for _, lane := range section.lanes {
		if lanePosition(section.numbers, lane.LaneNum()) < 0 {
			continue
		}
		uses := lane.AllowedUses()
		width, hasWidth := lane.WidthMeters()
		left, right := lane.LeftBarrier(), lane.RightBarrier()
		for _, override := range section.segmentLanes {
			parent, present := override.ParentLane()
			if present && parent == lane.ID {
				if override.AllowedUses() != "" {
					uses = override.AllowedUses()
				}
				value, supplied := override.WidthMeters()
				if supplied {
					width, hasWidth = value, true
				}
				if override.LeftBarrier() != "" {
					left = override.LeftBarrier()
				}
				if override.RightBarrier() != "" {
					right = override.RightBarrier()
				}
			}
		}
		err := validatePreparedLaneAttributes(source, uses, width, hasWidth, left, right, sectionUses(link, section))
		if err != nil {
			return fmt.Errorf("lane %v on link %v: %w", lane.ID, link.ID, err)
		}
	}
	for _, lane := range section.segmentLanes {
		_, hasParent := lane.ParentLane()
		if hasParent {
			continue
		}
		width, present := lane.WidthMeters()
		err := validatePreparedLaneAttributes(source, lane.AllowedUses(), width, present, lane.LeftBarrier(), lane.RightBarrier(), sectionUses(link, section))
		if err != nil {
			return fmt.Errorf("segment_lane %v: %w", lane.ID, err)
		}
	}
	return nil
}

func validatePreparedLaneAttributes(source *macro.Net, uses string, width float64, hasWidth bool, left, right, roadUses string) error {
	if hasWidth && width != macro.LANE_WIDTH {
		return fmt.Errorf("%w: variable lane width requires a separate offset policy; supported width is %g meters", ErrUnsupportedGenerationInput, macro.LANE_WIDTH)
	}
	if left != "" || right != "" {
		return fmt.Errorf("%w: lane barriers require lane-level connections", ErrUnsupportedGenerationInput)
	}
	if uses != "" {
		laneUses, err := expandedUses(source, uses)
		if err != nil {
			return err
		}
		sectionUses, err := expandedUses(source, roadUses)
		if err != nil {
			return err
		}
		equal := len(laneUses) == len(sectionUses)
		for name := range laneUses {
			equal = equal && sectionUses[name]
		}
		if !equal {
			return fmt.Errorf("%w: lane-specific allowed_uses differs from its section; aggregate meso connectors cannot represent it", ErrUnsupportedGenerationInput)
		}
	}
	return nil
}
