# Meso preparation

`generators.PrepareMesoscopic` builds meso from a `*macro.Net` containing standard records and supplied movements. It rebuilds internal geometry and lane data without OSM attributes or precomputed caches. The input remains unchanged on success and failure.

```go
result, err := generators.PrepareMesoscopic(source,
    generators.WithUniformLaneLayout(),
    generators.WithOrderedLaneConnections(),
)
if err != nil {
    return err
}
mesoNet := result.Meso()
```

## Options

Choose one lane layout. Defaults and ordered connections are optional.

| Option | Behavior |
|---|---|
| `WithUniformLaneLayout()` | Constant lane count along each link, numbered 1..lanes; rejects supplied detailed tables |
| `WithDetailedLaneLayout()` | Applies supported `lane`, `segment` and `segment_lane` records; uncovered intervals retain the base layout |
| `WithDefaultMesoLanes(n)` | Fills only absent link counts with an explicit default in 1..99; never replaces supplied zeroes or negatives |
| `WithOrderedLaneConnections()` | Pairs equally sized movement ranges from left to right; otherwise only single-lane scopes are accepted |

## Inputs and behavior

A configuration with metric units and WGS84 is required. Links must be directed, with positive lengths in meters, 1..99 lanes and valid inline or shared geometry. Shared-only shapes require an explicit direction. Missing shapes, consecutive duplicate points, geometry shorter than 0.01 meter and coordinates at the poles are rejected.

Movements require start lanes; a blank end means one lane and stays blank in the source records. References must match active lanes at the connected road ends. Each movement produces an explicit connector; missing movements and replacement lanes are never inferred. Meso's own movement collection stays empty.

Detailed preparation supports nonoverlapping positive segment intervals, measured from either link endpoint, and consecutive edge lane additions/drops. Base lane numbers are 1..lanes; additions on the left use negative numbers, while zero marks a dropped `segment_lane`. Surviving base numbers remain unchanged. Segment counts must agree with base counts and side changes. When both side changes are absent, they are derived from `segment_lane`; otherwise an absent side means no change.

Supplied segment attributes override link attributes; absent values inherit them. Connector speed and capacity follow the upstream section. Connector permissions intersect the supplied permissions of both endpoint sections and the movement, expanding use groups; an empty intersection is an error. Movement penalty and capacity remain on the source movement.

## Result

`Macro()` returns the completed source dataset, including original detailed records and accepted lane defaults. `Meso()` returns the generated network. `LaneAssumptions()`, `Sections()` and `Connections()` return copies of defaults, source-to-section associations and ordered lane pairs. Section intervals describe the source road before junction cuts. No standard lane IDs are fabricated from counts.

Identical inputs and options produce repeatable results within Go. Generated IDs may change after source edits; cross-language numerical equality is not guaranteed.

## Limits

Preparation rejects overlapping/nested segments, interior lane drops, parent lane renumbering, nonconsecutive additions, unequal movement range sizes, nonzero tolls, supplied lane barriers, lane widths other than 3.5 meters and lane permissions differing from their section. These records can still be stored by the library. Use conditions remain opaque text; signal timing and legal turn restrictions are not evaluated.
