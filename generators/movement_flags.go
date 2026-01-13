package generators

import (
	"math"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/movement"
	"github.com/LdDl/go-gmns/utils/geomath"
)

// MovementFlags contains precomputed flags for macro nodes and links
// used during meso/micro network generation
type MovementFlags struct {
	// Indicates which macro nodes need movement links (i.e. are intersections)
	// True means the node is an intersection, false means it's a pass-through node
	NodesNeedMovement map[gmns.NodeID]bool
	// Indicates which macro links have their downstream end as the target
	// for movement shortcuts (when movementIsNeeded is false)
	DownstreamIsTarget map[gmns.LinkID]bool
	// Indicates which macro links have their upstream end as the target
	// for movement shortcuts (when movementIsNeeded is false)
	UpstreamIsTarget map[gmns.LinkID]bool
}

// ComputeMovementFlags calculates movement flags for macro network
// based on topology, angles, and movement patterns
func ComputeMovementFlags(macroNet *macro.Net, movements movement.MovementsStorage) *MovementFlags {
	flags := &MovementFlags{
		NodesNeedMovement:  make(map[gmns.NodeID]bool),
		DownstreamIsTarget: make(map[gmns.LinkID]bool),
		UpstreamIsTarget:   make(map[gmns.LinkID]bool),
	}

	// Group movements by macro node
	macroNodesMovements := make(map[gmns.NodeID][]*movement.Movement)
	for _, mvmt := range movements {
		macroNodesMovements[mvmt.MacroNode()] = append(macroNodesMovements[mvmt.MacroNode()], mvmt)
	}

	// Assume all nodes need movements by default (i.e. all are intersections)
	for _, node := range macroNet.Nodes {
		flags.NodesNeedMovement[node.ID] = true
	}

	// Initialize all links as not being targets
	for _, link := range macroNet.Links {
		flags.DownstreamIsTarget[link.ID] = false
		flags.UpstreamIsTarget[link.ID] = false
	}

	// Process each macro node
	for _, macroNode := range macroNet.Nodes {
		// Skip signal-controlled nodes
		if macroNode.ControlType() == types.CONTROL_TYPE_IS_SIGNAL {
			continue
		}

		incomingMacroLinks := macroNode.IncomingLinks()
		outcomingMacroLinks := macroNode.OutcomingLinks()

		// Case 1: Only one incoming link
		if len(incomingMacroLinks) == 1 && len(outcomingMacroLinks) >= 1 {
			incomingMacroLinkID := incomingMacroLinks[0]
			incomingMacroLink, ok := macroNet.Links[incomingMacroLinkID]
			if !ok {
				continue
			}

			// Check angles
			badAngle := true
			for _, outcomingMacroLinkID := range outcomingMacroLinks {
				outcomingMacroLink, ok := macroNet.Links[outcomingMacroLinkID]
				if !ok {
					continue
				}
				angle := geomath.AngleBetweenLines(incomingMacroLink.GeomEuclidean(), outcomingMacroLink.GeomEuclidean())
				if angle > 0.75*math.Pi || angle < -0.75*math.Pi {
					badAngle = false
					break
				}
			}
			if !badAngle {
				continue
			}

			// Check for multiple connections to same outcoming link
			hasMultipleConnections := false
			outcomingMacroLinksObserved := make(map[gmns.LinkID]struct{})
			if macroNodeMvmts, ok := macroNodesMovements[macroNode.ID]; ok {
				for _, mvmt := range macroNodeMvmts {
					outcomeMacroLinkID := mvmt.OutcomeMacroLink()
					if _, ok := outcomingMacroLinksObserved[outcomeMacroLinkID]; ok {
						hasMultipleConnections = true
						break
					}
					outcomingMacroLinksObserved[outcomeMacroLinkID] = struct{}{}
				}
			}
			if hasMultipleConnections {
				continue
			}

			// Mark node as not needing movement and set target flag
			flags.NodesNeedMovement[macroNode.ID] = false
			flags.DownstreamIsTarget[incomingMacroLinkID] = true

		} else if len(incomingMacroLinks) >= 1 && len(outcomingMacroLinks) == 1 {
			// Case 2: Only one outcoming link
			outcomingMacroLinkID := outcomingMacroLinks[0]
			outcomingMacroLink, ok := macroNet.Links[outcomingMacroLinkID]
			if !ok {
				continue
			}

			// Check angles
			badAngle := true
			for _, incomingMacroLinkID := range incomingMacroLinks {
				incomingMacroLink, ok := macroNet.Links[incomingMacroLinkID]
				if !ok {
					continue
				}
				angle := geomath.AngleBetweenLines(incomingMacroLink.GeomEuclidean(), outcomingMacroLink.GeomEuclidean())
				if angle > 0.75*math.Pi || angle < -0.75*math.Pi {
					badAngle = false
					break
				}
			}
			if !badAngle {
				continue
			}

			// Check for multiple connections from same incoming link
			hasMultipleConnections := false
			incomingMacroLinksObserved := make(map[gmns.LinkID]struct{})
			if macroNodeMvmts, ok := macroNodesMovements[macroNode.ID]; ok {
				for _, mvmt := range macroNodeMvmts {
					incomeMacroLinkID := mvmt.IncomeMacroLink()
					if _, ok := incomingMacroLinksObserved[incomeMacroLinkID]; ok {
						hasMultipleConnections = true
						break
					}
					incomingMacroLinksObserved[incomeMacroLinkID] = struct{}{}
				}
			}
			if hasMultipleConnections {
				continue
			}

			// Mark node as not needing movement and set target flag
			flags.NodesNeedMovement[macroNode.ID] = false
			flags.UpstreamIsTarget[outcomingMacroLinkID] = true
		}
	}

	return flags
}
