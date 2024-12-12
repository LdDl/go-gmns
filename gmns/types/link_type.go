package types

type LinkType uint16

const (
	LINK_UNDEFINED = LinkType(iota)
	LINK_MOTORWAY
	LINK_TRUNK
	LINK_PRIMARY
	LINK_SECONDARY
	LINK_TERTIARY
	LINK_RESIDENTIAL
	LINK_LIVING_STREET
	LINK_SERVICE
	LINK_CYCLEWAY
	LINK_FOOTWAY
	LINK_TRACK
	LINK_UNCLASSIFIED
	LINK_CONNECTOR
	LINK_RAILWAY
	LINK_AEROWAY
)

func (iotaIdx LinkType) String() string {
	return [...]string{"undefined", "motorway", "trunk", "primary", "secondary", "tertiary", "residential", "living_street", "service", "cycleway", "footway", "track", "unclassified", "connector", "railway", "aeroway"}[iotaIdx]
}

type linkComposition struct {
	LinkType           LinkType
	LinkConnectionType LinkConnectionType
}

var (
	onewayDefaultByLink = map[LinkType]bool{
		LINK_MOTORWAY:      false,
		LINK_TRUNK:         false,
		LINK_PRIMARY:       false,
		LINK_SECONDARY:     false,
		LINK_TERTIARY:      false,
		LINK_RESIDENTIAL:   false,
		LINK_LIVING_STREET: false,
		LINK_SERVICE:       false,
		LINK_CYCLEWAY:      true,
		LINK_FOOTWAY:       true,
		LINK_TRACK:         true,
		LINK_UNCLASSIFIED:  false,
		LINK_CONNECTOR:     false,
		LINK_RAILWAY:       true,
		LINK_AEROWAY:       true,
	}
	defaultLanesByLinkType = map[LinkType]int{
		LINK_MOTORWAY:     4,
		LINK_TRUNK:        3,
		LINK_PRIMARY:      3,
		LINK_SECONDARY:    2,
		LINK_TERTIARY:     2,
		LINK_RESIDENTIAL:  1,
		LINK_SERVICE:      1,
		LINK_CYCLEWAY:     1,
		LINK_FOOTWAY:      1,
		LINK_TRACK:        1,
		LINK_UNCLASSIFIED: 1,
		LINK_CONNECTOR:    2,
	}
	defaultSpeedByLinkType = map[LinkType]float64{
		LINK_MOTORWAY:     120,
		LINK_TRUNK:        100,
		LINK_PRIMARY:      80,
		LINK_SECONDARY:    60,
		LINK_TERTIARY:     40,
		LINK_RESIDENTIAL:  30,
		LINK_SERVICE:      30,
		LINK_CYCLEWAY:     5,
		LINK_FOOTWAY:      5,
		LINK_TRACK:        30,
		LINK_UNCLASSIFIED: 30,
		LINK_CONNECTOR:    120,
	}
	defaultCapacityByLinkType = map[LinkType]int{
		LINK_MOTORWAY:     2300,
		LINK_TRUNK:        2200,
		LINK_PRIMARY:      1800,
		LINK_SECONDARY:    1600,
		LINK_TERTIARY:     1200,
		LINK_RESIDENTIAL:  1000,
		LINK_SERVICE:      800,
		LINK_CYCLEWAY:     800,
		LINK_FOOTWAY:      800,
		LINK_TRACK:        800,
		LINK_UNCLASSIFIED: 800,
		LINK_CONNECTOR:    9999,
	}

	// it IS bad ranking currently. Need to research correct ranking
	priorityRank = map[LinkType]int{
		LINK_UNDEFINED:     0,
		LINK_MOTORWAY:      15,
		LINK_TRUNK:         14,
		LINK_PRIMARY:       13,
		LINK_SECONDARY:     12,
		LINK_TERTIARY:      11,
		LINK_RESIDENTIAL:   10,
		LINK_LIVING_STREET: 9,
		LINK_SERVICE:       8,
		LINK_CYCLEWAY:      7,
		LINK_FOOTWAY:       6,
		LINK_TRACK:         5,
		LINK_UNCLASSIFIED:  4,
		LINK_CONNECTOR:     3,
		LINK_RAILWAY:       2,
		LINK_AEROWAY:       1,
	}
)

func NewOnewayDefault(lt LinkType) bool {
	if found, ok := onewayDefaultByLink[lt]; ok {
		return found
	}
	return false
}

func NewCapacityDefault(lt LinkType) int {
	if defaultCap, ok := defaultCapacityByLinkType[lt]; ok {
		return defaultCap
	}
	return -1
}

func NewSpeedDefault(lt LinkType) float64 {
	if defaultSpeed, ok := defaultSpeedByLinkType[lt]; ok {
		return defaultSpeed
	}
	return -1
}

func NewLanesDefault(lt LinkType) int {
	if defaultLanes, ok := defaultLanesByLinkType[lt]; ok {
		return defaultLanes
	}
	return -1
}

func FindPriorLinkType(linkTypes []LinkType) LinkType {
	maxPriority := -1
	var maxPriorityLink LinkType
	for _, linkType := range linkTypes {
		if priority, ok := priorityRank[linkType]; ok && priority > maxPriority {
			maxPriority = priority
			maxPriorityLink = linkType
		}
	}
	return maxPriorityLink
}
