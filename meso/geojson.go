package meso

import (
	"github.com/paulmach/orb/geojson"
)

// GeoFeature returns GeoJSON LineString feature for the given link
func (link *Link) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(nil)
	f.ID = link.ID
	panic("@tbd")
	return f
}

// GeoFeature returns GeoJSON Point feature for the given node
func (node *Node) GeoFeature() *geojson.Feature {
	f := geojson.NewFeature(nil)
	f.ID = node.ID
	panic("@tbd")
	return f
}

// GeoFeatureCollection returns GeoJSON FeatureCollection consisting of Linestring and Point features for the given road network
func (net *Net) GeoFeatureCollection() *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()
	for _, node := range net.Nodes {
		fc.Append(node.GeoFeature())
	}
	for _, link := range net.Links {
		fc.Append(link.GeoFeature())
	}
	return fc
}
