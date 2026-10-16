package generators

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// PreparedSection relates a generated road section to its source interval before junction cuts.
// Lane numbers identify continuity across adjacent sections, not generated standard IDs.
type PreparedSection struct {
	linkID         gmns.LinkID
	sourceLinkID   gmns.LinkID
	segmentID      gmns.SegmentID
	start          float64
	end            float64
	lanes          []int
	laneIDs        []gmns.LaneID
	segmentLaneIDs []gmns.SegmentLaneID
}

func (section *PreparedSection) Link() gmns.LinkID {
	return section.linkID
}

func (section *PreparedSection) SourceLink() gmns.LinkID {
	return section.sourceLinkID
}

func (section *PreparedSection) Segment() (gmns.SegmentID, bool) {
	return section.segmentID, !section.segmentID.IsZero()
}

func (section *PreparedSection) StartMeters() float64 {
	return section.start
}

func (section *PreparedSection) EndMeters() float64 {
	return section.end
}

func (section *PreparedSection) LaneNumbers() []int {
	return append([]int(nil), section.lanes...)
}

func (section *PreparedSection) SourceLanes() []gmns.LaneID {
	return append([]gmns.LaneID(nil), section.laneIDs...)
}

func (section *PreparedSection) SourceSegmentLanes() []gmns.SegmentLaneID {
	return append([]gmns.SegmentLaneID(nil), section.segmentLaneIDs...)
}

// PreparedConnection records the ordered lane pairs represented by one meso connector.
type PreparedConnection struct {
	linkID     gmns.LinkID
	movementID gmns.MovementID
	inbound    []int
	outbound   []int
}

func (connection *PreparedConnection) Link() gmns.LinkID {
	return connection.linkID
}

func (connection *PreparedConnection) Movement() gmns.MovementID {
	return connection.movementID
}

func (connection *PreparedConnection) InboundLanes() []int {
	return append([]int(nil), connection.inbound...)
}

func (connection *PreparedConnection) OutboundLanes() []int {
	return append([]int(nil), connection.outbound...)
}

func (result *MesoPreparationResult) Sections() []PreparedSection {
	sections := append([]PreparedSection(nil), result.sections...)
	for index := range sections {
		sections[index].lanes = result.sections[index].LaneNumbers()
		sections[index].laneIDs = result.sections[index].SourceLanes()
		sections[index].segmentLaneIDs = result.sections[index].SourceSegmentLanes()
	}
	return sections
}

func (result *MesoPreparationResult) Connections() []PreparedConnection {
	connections := append([]PreparedConnection(nil), result.connections...)
	for index := range connections {
		connections[index].inbound = result.connections[index].InboundLanes()
		connections[index].outbound = result.connections[index].OutboundLanes()
	}
	return connections
}

func preparationLineage(target *meso.Net, layout preparationLayout) []PreparedSection {
	sections := make([]PreparedSection, 0)
	for _, id := range sortedMesoLinkIDs(target.Links) {
		link := target.Links[id]
		if link.MacroLink().IsZero() {
			continue
		}
		source := layout[link.MacroLink()][link.SegmentIdx()]
		section := PreparedSection{linkID: id, sourceLinkID: link.MacroLink(), start: source.start, end: source.end, lanes: append([]int(nil), source.numbers...)}
		if source.segment != nil {
			section.segmentID = source.segment.ID
		}
		for _, lane := range source.lanes {
			if lanePosition(source.numbers, lane.LaneNum()) >= 0 {
				section.laneIDs = append(section.laneIDs, lane.ID)
			}
		}
		for _, lane := range source.segmentLanes {
			section.segmentLaneIDs = append(section.segmentLaneIDs, lane.ID)
		}
		sections = append(sections, section)
	}
	return sections
}

func preparationConnections(source *macro.Net, target *meso.Net, layout preparationLayout) []PreparedConnection {
	connections := make([]PreparedConnection, 0)
	for _, id := range sortedMesoLinkIDs(target.Links) {
		link := target.Links[id]
		if link.Movement().IsZero() {
			continue
		}
		record, _ := source.Movement(link.Movement())
		inbound := layout[record.IncomeLink()]
		outbound := layout[record.OutcomeLink()]
		firstIn := record.StartIncomeLaneSeqID()
		firstOut := record.StartOutcomeLaneSeqID()
		count := record.LanesNum()
		connection := PreparedConnection{linkID: id, movementID: record.ID}
		connection.inbound = append([]int(nil), inbound[len(inbound)-1].numbers[firstIn:firstIn+count]...)
		connection.outbound = append([]int(nil), outbound[0].numbers[firstOut:firstOut+count]...)
		connections = append(connections, connection)
	}
	return connections
}
