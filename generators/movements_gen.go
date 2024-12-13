package generators

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/pkg/errors"
)

// GenerateMovements generates movements for the given macroscopic network
func GenerateMovements(macroNet macro.Net) (movement.MovementsStorage, error) {
	ans := movement.NewMovementsStorage()
	for i := range macroNet.Nodes {
		node := macroNet.Nodes[i]
		movements, err := findMovements(node, macroNet.Links)
		if err != nil {
			return nil, errors.Wrapf(err, "Can't find movements for macro node with ID: '%d' (OSM ID: '%d')", node.ID, node.OSMNode())
		}
		for j := range movements {
			mvmt := movements[j]
			ans[mvmt.ID] = mvmt
		}
	}
	return ans, nil
}

// findMovements generates array of movements for the given macroscopic node [this function is not exported yet]
func findMovements(macroNode *macro.Node, links map[gmns.LinkID]*macro.Link) ([]*movement.Movement, error) {
	movements := []*movement.Movement{}

	macroIncomingLinks := macroNode.IncomingLinks()
	macroOutcomingLinks := macroNode.OutcomingLinks()

	income := len(macroIncomingLinks)
	outcome := len(macroOutcomingLinks)
	if income == 0 || outcome == 0 {
		return movements, nil
	}

	if outcome == 1 {
		// Merge
		incomingLinksList := []*macro.Link{}
		outcomingLinkID := macroOutcomingLinks[0]
		outcomingLink, ok := links[outcomingLinkID]
		if !ok {
			return nil, errors.Wrapf(macro.ErrLinkNotFound, "Span outcoming Link ID: %d. Node: %d", outcomingLinkID, macroNode.ID)
		}
		for i := range macroIncomingLinks {
			incomingLinkID := macroIncomingLinks[i]
			incomingLink, ok := links[incomingLinkID]
			if !ok {
				return nil, errors.Wrapf(macro.ErrLinkNotFound, "Span incoming Link ID: %d. Node: %d", incomingLinkID, macroNode.ID)
			}
			if incomingLink.SourceNode() != outcomingLink.TargetNode() { // Ignore all reverse directions
				incomingLinksList = append(incomingLinksList, incomingLink)
			}
		}
		if len(incomingLinksList) == 0 {
			return movements, nil
		}

		connections := macro.GenerateSpansConnections(outcomingLink, incomingLinksList)
		incomingLaneIndices := outcomingLink.GetOutcomingLaneIndices()
		for i := range incomingLinksList {
			incomingLink := incomingLinksList[i]
			incomeLaneIndexStart := connections[i][0].First
			incomeLaneIndexEnd := connections[i][0].Second
			outcomeLaneIndexStart := connections[i][1].First
			outcomeLaneIndexEnd := connections[i][1].Second
			lanesNum := incomeLaneIndexEnd - incomeLaneIndexStart + 1

			outcomingLaneIndices := incomingLink.GetOutcomingLaneIndices()
			mvmtTextID, mvmtType := movement.FindMovementType(incomingLink.GeomEuclidean(), outcomingLink.GeomEuclidean())
			mvmtGeom := movement.FindMovementGeom(incomingLink.Geom(), outcomingLink.Geom())
			mvmt := movement.NewMovement(
				movement.GenMovementID(),
				macroNode.ID, incomingLink.ID, outcomingLinkID, mvmtTextID, mvmtType,
				movement.WithOSMNodeID(macroNode.OSMNode()),
				movement.WithSourceOSMNodeID(incomingLink.SourceOSMNode()),
				movement.WithTargetOSMNodeID(outcomingLink.TargetOSMNode()),
				movement.WithControlType(macroNode.ControlType()),
				movement.WithAllowedAgentTypes(incomingLink.AllowedAgentTypes()),
				movement.WithLanesNum(lanesNum),
				movement.WithIncomeLane(outcomingLaneIndices[incomeLaneIndexStart], outcomingLaneIndices[incomeLaneIndexEnd]),
				movement.WithIncomeLaneSequence(incomeLaneIndexStart, incomeLaneIndexEnd),
				movement.WithOutcomeLane(incomingLaneIndices[outcomeLaneIndexStart], incomingLaneIndices[outcomeLaneIndexEnd]),
				movement.WithOutcomeLaneSequence(outcomeLaneIndexStart, outcomeLaneIndexEnd),
				movement.WithGeom(mvmtGeom),
			)
			movements = append(movements, mvmt)
		}
	} else {
		// Diverge
		// Intersections
		for i := range macroIncomingLinks {
			incomingLinkID := macroIncomingLinks[i]
			incomingLink, ok := links[incomingLinkID]
			if !ok {
				return nil, errors.Wrapf(macro.ErrLinkNotFound, "Intersection incoming Link ID: %d. Node: %d", incomingLinkID, macroNode.ID)
			}
			outcomingLinksList := []*macro.Link{}
			for j := range macroOutcomingLinks {
				outcomingLinkID := macroOutcomingLinks[j]
				outcomingLink, ok := links[outcomingLinkID]
				if !ok {
					return nil, errors.Wrapf(macro.ErrLinkNotFound, "Intersection outcoming Link ID: %d. Node: %d", outcomingLinkID, macroNode.ID)
				}
				if incomingLink.SourceNode() != outcomingLink.TargetNode() { // Ignore all reverse directions
					outcomingLinksList = append(outcomingLinksList, outcomingLink)
				}
			}
			if len(outcomingLinksList) == 0 {
				// @todo: can just return?
				return movements, nil
			}
			connections := macro.GenerateIntersectionsConnections(incomingLink, outcomingLinksList)
			outcomingLaneIndices := incomingLink.GetOutcomingLaneIndices()

			for i := range outcomingLinksList {
				outcomingLink := outcomingLinksList[i]
				incomeLaneIndexStart := connections[i][0].First
				incomeLaneIndexEnd := connections[i][0].Second
				outcomeLaneIndexStart := connections[i][1].First
				outcomeLaneIndexEnd := connections[i][1].Second
				lanesNum := incomeLaneIndexEnd - incomeLaneIndexStart + 1

				incomingLaneIndices := outcomingLink.GetOutcomingLaneIndices()
				mvmtTextID, mvmtType := movement.FindMovementType(incomingLink.GeomEuclidean(), outcomingLink.GeomEuclidean())
				mvmtGeom := movement.FindMovementGeom(incomingLink.Geom(), outcomingLink.Geom())
				mvmt := movement.NewMovement(
					movement.GenMovementID(),
					macroNode.ID, incomingLinkID, outcomingLink.ID, mvmtTextID, mvmtType,
					movement.WithOSMNodeID(macroNode.OSMNode()),
					movement.WithSourceOSMNodeID(incomingLink.SourceOSMNode()),
					movement.WithTargetOSMNodeID(outcomingLink.TargetOSMNode()),
					movement.WithControlType(macroNode.ControlType()),
					movement.WithAllowedAgentTypes(incomingLink.AllowedAgentTypes()),
					movement.WithLanesNum(lanesNum),
					movement.WithIncomeLane(outcomingLaneIndices[incomeLaneIndexStart], outcomingLaneIndices[incomeLaneIndexEnd]),
					movement.WithIncomeLaneSequence(incomeLaneIndexStart, incomeLaneIndexEnd),
					movement.WithOutcomeLane(incomingLaneIndices[outcomeLaneIndexStart], incomingLaneIndices[outcomeLaneIndexEnd]),
					movement.WithOutcomeLaneSequence(outcomeLaneIndexStart, outcomeLaneIndexEnd),
					movement.WithGeom(mvmtGeom),
				)
				movements = append(movements, mvmt)
			}
		}
	}

	return movements, nil
}
