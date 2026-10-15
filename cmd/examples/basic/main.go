package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/macro"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	network := macro.NewNet()
	config := gmns.NewConfig(gmns.IDTypeString, gmns.WithDatasetName("Basic road"))
	err := network.SetConfig(config)
	if err != nil {
		return fmt.Errorf("set config: %w", err)
	}

	sourceID := gmns.NodeIDFromString("001")
	targetID := gmns.NodeIDFromString("002")
	linkID := gmns.LinkIDFromString("road-1")
	sourcePoint := orb.Point{13.4050, 52.5200}
	targetPoint := orb.Point{13.4060, 52.5200}
	shape := orb.LineString{sourcePoint, targetPoint}
	agents := []types.AgentType{types.AGENT_AUTO}

	source := macro.NewNodeFrom(
		sourceID,
		macro.WithNodeName("Start"),
		macro.WithPointGeom(sourcePoint),
		macro.WithOutcomingLinks(linkID),
	)
	target := macro.NewNodeFrom(
		targetID,
		macro.WithNodeName("End"),
		macro.WithPointGeom(targetPoint),
		macro.WithIncomingLinks(linkID),
	)
	link := macro.NewLinkFrom(
		linkID,
		sourceID,
		targetID,
		macro.WithLinkName("Example road"),
		macro.WithLineGeom(shape),
		macro.WithLengthMeters(geo.Length(shape)),
		macro.WithLanesNum(2),
		macro.WithFreeSpeed(30),
		macro.WithAllowedAgentTypes(agents),
	)
	network.Nodes[source.ID] = source
	network.Nodes[target.ID] = target
	network.Links[link.ID] = link

	err = network.ValidateDataset()
	if err != nil {
		return fmt.Errorf("validate dataset: %w", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(network.GeoFeatureCollection())
	if err != nil {
		return fmt.Errorf("encode GeoJSON: %w", err)
	}
	return nil
}
