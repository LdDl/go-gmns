package macro

import (
	"sort"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/utils"
	"github.com/LdDl/go-gmns/utils/geomath"
)

const (
	RIGHT_MOST_LANES_DEFAULT = 1
	LEFT_MOST_LANES_DEFAULT  = 1
)

type ConnectionPair struct {
	First  int
	Second int
}

// GenerateSpansConnections generates connections between links for the span between roads
func GenerateSpansConnections(outcomingLink *Link, incomingLinksList []*Link) [][]ConnectionPair {
	// Sort outcoming links by angle in descending order (left to right)
	angles := make([]float64, len(incomingLinksList))
	for i, inLink := range incomingLinksList {
		inGeomEuclidean := inLink.GeomEuclidean()
		outGeomEuclidean := outcomingLink.GeomEuclidean()
		if len(inGeomEuclidean) == 0 {
			// If no euclidean geom has been provided, try to get one
			inGeomEuclidean = geomath.LineToEuclidean(inLink.Geom())
		}
		if len(outGeomEuclidean) == 0 {
			// If no euclidean geom has been provided, try to get one
			outGeomEuclidean = geomath.LineToEuclidean(outcomingLink.Geom())
		}
		angles[i] = geomath.AngleBetweenLines(inGeomEuclidean, outGeomEuclidean)
	}
	indicesMap := make(map[gmns.LinkID]int, len(incomingLinksList))
	for index := range incomingLinksList {
		link := incomingLinksList[index]
		indicesMap[link.ID] = index
	}
	indices := make([]int, len(incomingLinksList))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return angles[indices[i]] > angles[indices[j]]
	})
	incomingLinksSorted := make([]*Link, len(incomingLinksList))
	for i := range incomingLinksSorted {
		incomingLinksSorted[i] = incomingLinksList[indices[i]]
	}
	// Evaluate lanes connections
	connections := make([][]ConnectionPair, len(incomingLinksSorted))
	outcomingLanes := outcomingLink.GetIncomingLanes()

	leftLink := incomingLinksSorted[0]
	leftLinkOutcomingLanes := leftLink.GetOutcomingLanes()

	minConnections := min(outcomingLanes, leftLinkOutcomingLanes)
	// In <-> Out
	connections[indicesMap[leftLink.ID]] = []ConnectionPair{
		{leftLinkOutcomingLanes - minConnections, leftLinkOutcomingLanes - 1},
		{0, minConnections - 1},
	}
	for i := range incomingLinksSorted[1:] {
		inLink := incomingLinksSorted[1:][i]
		lanesInfo := inLink.LanesInfo()
		if len(lanesInfo.LanesList) == 0 {
			continue
		}
		inLinkOutcomingLanes := lanesInfo.LanesList[len(lanesInfo.LanesList)-1]
		minConnections := min(outcomingLanes, inLinkOutcomingLanes)
		// In <-> Out
		connections[indicesMap[inLink.ID]] = []ConnectionPair{{0, minConnections - 1}, {outcomingLanes - minConnections, outcomingLanes - 1}}
	}
	return connections
}

// GenerateIntersectionsConnections generates connections between links for the junctions of the roads
func GenerateIntersectionsConnections(incomingLink *Link, outcomingLinks []*Link) [][]ConnectionPair {
	// Sort outcoming links by angle in descending order (left to right)
	angles := make([]float64, len(outcomingLinks))
	for i, outLink := range outcomingLinks {
		inGeomEuclidean := incomingLink.GeomEuclidean()
		outGeomEuclidean := outLink.GeomEuclidean()
		if len(inGeomEuclidean) == 0 {
			// If no euclidean geom has been provided, try to get one
			inGeomEuclidean = geomath.LineToEuclidean(incomingLink.Geom())
		}
		if len(outGeomEuclidean) == 0 {
			// If no euclidean geom has been provided, try to get one
			outGeomEuclidean = geomath.LineToEuclidean(outLink.Geom())
		}
		angles[i] = geomath.AngleBetweenLines(inGeomEuclidean, outGeomEuclidean)
	}
	indicesMap := make(map[gmns.LinkID]int, len(outcomingLinks))
	for index := range outcomingLinks {
		link := outcomingLinks[index]
		indicesMap[link.ID] = index
	}
	indices := make([]int, len(outcomingLinks))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return angles[indices[i]] > angles[indices[j]]
	})
	outcomingLinksSorted := make([]*Link, len(outcomingLinks))
	for i := range outcomingLinksSorted {
		outcomingLinksSorted[i] = outcomingLinks[indices[i]]
	}

	// Evaluate lanes connections
	connections := make([][]ConnectionPair, len(outcomingLinksSorted))
	outcomingLanes := incomingLink.GetOutcomingLanes()
	if outcomingLanes == 1 {
		leftLink := outcomingLinksSorted[0]
		connections[indicesMap[leftLink.ID]] = []ConnectionPair{
			{0, 0},
			{0, 0},
		}
		for i := range outcomingLinksSorted[1:] {
			link := outcomingLinksSorted[1:][i]
			connections[indicesMap[link.ID]] = []ConnectionPair{
				{0, 0},
				{link.GetIncomingLanes() - 1, link.GetIncomingLanes() - 1},
			}
		}
		return connections
	}
	if len(outcomingLinksSorted) == 1 { // Full connection
		link := outcomingLinksSorted[0]
		minConnections := min(outcomingLanes, link.GetIncomingLanes())
		connections[indicesMap[link.ID]] = []ConnectionPair{
			{0, minConnections - 1},
			{0, minConnections - 1},
		}
	} else if len(outcomingLinksSorted) == 2 { // Default right, remaining left
		leftLink := outcomingLinksSorted[0]
		minConnections := min(outcomingLanes-LEFT_MOST_LANES_DEFAULT, leftLink.GetIncomingLanes()) // If link has incoming lanes
		connections[indicesMap[leftLink.ID]] = []ConnectionPair{
			{0, minConnections - 1},
			{0, minConnections - 1},
		}
		rightLink := outcomingLinksSorted[len(outcomingLinksSorted)-1]
		connections[indicesMap[rightLink.ID]] = []ConnectionPair{
			{outcomingLanes - RIGHT_MOST_LANES_DEFAULT, outcomingLanes - 1},
			{rightLink.GetIncomingLanes() - RIGHT_MOST_LANES_DEFAULT, rightLink.GetIncomingLanes() - 1},
		}
	} else { // >= 3, default left, default right, remaining middle
		leftLink := outcomingLinksSorted[0]
		connections[indicesMap[leftLink.ID]] = []ConnectionPair{
			{0, LEFT_MOST_LANES_DEFAULT - 1},
			{0, LEFT_MOST_LANES_DEFAULT - 1},
		}

		middleLinks := outcomingLinksSorted[1 : len(outcomingLinksSorted)-1]
		assignedToMiddle := make([]int, len(middleLinks))
		middleLinksLanes := make([]int, len(middleLinks))
		for i, midLink := range middleLinks {
			middleLinksLanes[i] = midLink.GetIncomingLanes()
		}
		leftLanesNum := outcomingLanes - LEFT_MOST_LANES_DEFAULT - RIGHT_MOST_LANES_DEFAULT
		if leftLanesNum >= len(middleLinks) {
			startLaneNumber := LEFT_MOST_LANES_DEFAULT
			for leftLanesNum > 0 && utils.TotalInts(middleLinksLanes) > 0 {
				for idx := range middleLinks {
					if middleLinksLanes[idx] == 0 {
						continue
					}
					if leftLanesNum == 0 {
						continue
					}
					middleLinksLanes[idx]--
					assignedToMiddle[idx]++
					leftLanesNum--
				}
			}
			for idx, middleLink := range middleLinks {
				connections[indicesMap[middleLink.ID]] = []ConnectionPair{
					{startLaneNumber, startLaneNumber + assignedToMiddle[idx] - 1},
					{middleLink.GetIncomingLanes() - assignedToMiddle[idx], middleLink.GetIncomingLanes() - 1},
				}
				startLaneNumber += assignedToMiddle[idx]
			}
		} else if outcomingLanes < len(middleLinks) {
			laneNumber := -1
			linkIndex := -1
			for laneNumberIdx := 0; laneNumberIdx < outcomingLanes-1; laneNumberIdx++ {
				laneNumber++
				linkIndex = laneNumber
				middleLink := middleLinks[linkIndex]
				connections[indicesMap[middleLink.ID]] = []ConnectionPair{
					{laneNumber, laneNumber},
					{middleLink.GetIncomingLanes() - 1, middleLink.GetIncomingLanes() - 1},
				}
			}
			laneNumber++
			startLinkIndex := linkIndex + 1
			for linkIndexIdx := startLinkIndex; linkIndexIdx < len(middleLinks); linkIndexIdx++ {
				linkIndex++
				middleLink := middleLinks[linkIndex]
				connections[indicesMap[middleLink.ID]] = []ConnectionPair{
					{laneNumber, laneNumber},
					{middleLink.GetIncomingLanes() - 1, middleLink.GetIncomingLanes() - 1},
				}
			}
		} else {
			startLaneNumber := 0
			if outcomingLanes-LEFT_MOST_LANES_DEFAULT == len(middleLinks) {
				startLaneNumber = LEFT_MOST_LANES_DEFAULT
			} else {
				startLaneNumber = 0
			}
			for _, midLink := range middleLinks {
				connections[indicesMap[midLink.ID]] = []ConnectionPair{
					{startLaneNumber, startLaneNumber},
					{midLink.GetIncomingLanes() - 1, midLink.GetIncomingLanes() - 1},
				}
				startLaneNumber++
			}
		}
		rightLink := outcomingLinksSorted[len(outcomingLinksSorted)-1]
		connections[indicesMap[rightLink.ID]] = []ConnectionPair{
			{outcomingLanes - RIGHT_MOST_LANES_DEFAULT, outcomingLanes - 1},
			{rightLink.GetIncomingLanes() - RIGHT_MOST_LANES_DEFAULT, rightLink.GetIncomingLanes() - 1},
		}
	}

	return connections
}
