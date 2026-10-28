# Network generation

[Back to README](../README.md)

## Contents

- [Input requirements](#input-requirements)
- [Preparing a macro network](#preparing-a-macro-network)
- [Calling the generators](#calling-the-generators)
- [Movement generation](#movement-generation)
- [Meso generation](#meso-generation)
- [Micro generation](#micro-generation)
- [Micro options](#micro-options)

The library transforms prepared macro networks into movements, meso road segments and micro lane cells. [osm2gmns](https://github.com/LdDl/osm2gmns) handles OSM parsing and macro construction. The [basic example](../cmd/examples/basic/README.md) demonstrates direct macro construction and GeoJSON output.

## Input requirements

| Input | Requirement |
|---|---|
| IDs and references | One ID kind per dataset; valid endpoints, local references and source provenance |
| Direction | Directed links; use two opposite links for a two-way road |
| Adjacency | Macro node incoming/outgoing lists must match the links |
| Geometry | WGS84 line geometry and corresponding prepared projected geometry |
| Length | Positive finite supplied road lengths for segmentation and cell generation |
| Lanes | Positive macro counts and prepared `LanesInfo` for movement/meso generation |
| Meso input to micro | Prepared meso topology and source references, normally produced by `GenerateMesoscopic` |
| Configuration | When present, it and its owned dataset records must pass validation |

The generators reject inputs they cannot transform faithfully: supplied lane/segment/segment-lane records, link `allowed_uses` mappings and nonzero tolls. Meso generation also rejects movement `allowed_uses`; micro generation rejects meso-owned movement records. These inputs can still be stored in the data model. Unsupported transformations return `ErrUnsupportedGenerationInput`.

Micro generation can consume prepared zero-motor-lane bike/walk roads with `SeparateBikeWalk`; the movement/meso generators still require positive macro lane counts. Supplying `geometry_id` alone does not prepare the inline geometry needed for generation.

## Preparing a macro network

1. Create nodes and directed links, using a consistent ID kind and endpoint references.
2. Set WGS84 geometry, lengths in meters, lane counts and allowed built-in agents.
3. Populate node adjacency with `macro.WithIncomingLinks` and `macro.WithOutcomingLinks`.
4. Prepare projected geometry using `geomath.PointToEuclidean` and `geomath.LineToEuclidean` from `utils/geomath`.
5. For a uniform road, call `macro.NewLanesInfo(link)` and apply the result with `macro.WithLanesInfo`. For changing lane counts, supply the lane layout and distance breakpoints explicitly.
6. Set the dataset configuration and call `ValidateDataset()` to check supplied values and references. Generator-specific preparation requirements are checked separately.

For example, after setting a link's WGS84 geometry, length and lane count, prepare its calculation data inside an error-returning function:

```go
projected := geomath.LineToEuclidean(link.Geom())
macro.WithLineGeomEuclidean(projected)(link)

lanesInfo, err := macro.NewLanesInfo(link)
if err != nil {
    return err
}
macro.WithLanesInfo(lanesInfo)(link)
```

Projected geometry is an internal calculation representation; exported coordinates remain WGS84. Supplied `gmns.Lane` and `gmns.Segment` records do not replace `LanesInfo`.

## Calling the generators

The following function uses the `generators`, `macro` and `micro` packages from this module. Its argument must meet the preparation requirements above.

```go
func buildMicro(preparedMacro *macro.Net) (*micro.Net, error) {
    movements, err := generators.GenerateMovements(preparedMacro)
    if err != nil {
        return nil, err
    }

    mesoNet, err := generators.GenerateMesoscopic(preparedMacro, movements)
    if err != nil {
        return nil, err
    }

    options := generators.DefaultMicroGenOptions()
    options.CellLength = 5
    microNet, err := generators.GenerateMicroscopic(preparedMacro, mesoNet, movements, options)
    if err != nil {
        return nil, err
    }
    return microNet, nil
}
```

`GenerateMovements` returns a separate `MovementsStorage`. To retain those records in the macro dataset, add them with `macroNet.AddMovement`. The same macro movement storage is passed to meso and micro generation; the generators do not create meso-owned or micro-owned movement records.

Generated IDs are allocated within each result using the input ID kind. Use explicit source references to associate generated objects with their inputs rather than deriving that relationship from ID values.

## Movement generation

At nodes with incoming and outgoing links, the generator selects compatible approaches/departures, maps lane connections and classifies turns from geometry. It uses prepared topology and built-in agent permissions; it does not read OSM restriction relations or calculate turn penalties.

If the macro network already owns movement records, `GenerateMovements` validates and returns copies of those records instead of generating additional maneuvers. Supplied categories and penalties are retained.

## Meso generation

Each macro road link is split into segments at lane-count changes. The generator builds segment endpoints, offsets road geometry where needed and connects segments through macro movements. Each meso road link can carry multiple lanes.

At eligible uncontrolled junctions, the generator joins endpoints directly. Controlled junctions retain explicit movement connectors; custom control categories preserve the junction without interpreting traffic rules. A macro node can therefore correspond to multiple meso nodes.

Road attributes and endpoint metadata are carried into generated records where applicable. Configuration, zone and use catalogs are copied to the result. Shared geometry records are not copied as references to newly offset shapes.

## Micro generation

Each meso road segment becomes a row of cells for each lane. Forward links connect consecutive cells in the same lane; lane-change links connect to the next cell in an adjacent lane. Junction connectors produce movement cell chains with source movement codes.

Cell length is a target: the number of cells is rounded from road length, with at least one cell. Consequently individual cells need not have exactly the requested length. Graph size depends on road length, lane counts and options; there is no fixed macro-to-micro size ratio.

When enabled, separate bicycle and pedestrian paths follow the allowed agents. Supplied altitude is retained at corresponding road endpoints without interpolating it onto intermediate cells. The result owns its configuration and catalogs, and keeps source network references separately from local IDs.

## Micro options

Start with `generators.DefaultMicroGenOptions()` and change the fields you need. Passing a `MicroGenOptions` value replaces the defaults as a whole; omitted struct fields are not filled automatically. Supply positive finite lengths and widths.

| Option | Default | Meaning |
|---|---|---|
| `CellLength` | `4.5` m | Target cell length |
| `LaneWidth` | `3.5` m | Motor-vehicle lane width used for offsets |
| `BikeLaneWidth` | `0.5` m | Bicycle path spacing used for offsets |
| `WalkLaneWidth` | `0.5` m | Pedestrian path spacing used for offsets |
| `SeparateBikeWalk` | `false` | Build separate bike/walk paths when those agents are allowed |
| `Verbose` | `false` | Print micro generation progress |

All physical dimensions use the metric system. These settings control graph geometry and resolution; they do not configure a vehicle simulator.

## Altitude

Node altitude remains independent of two-dimensional geometry. Generators copy endpoint altitude only when the generated point has exactly the source coordinates; cut, offset and intermediate points have no inferred height. Source records remain unchanged. No vertical interpolation or datum conversion is performed.

Micro generation rejects supplied meso altitude that is not the same value at its exact macro source point. Local meso lane edits reject moving a node with supplied altitude. Existing saved meso records retain their heights on restoration, including values inherited by earlier generators. Attribute-only updates remain available. A numerical vertical profile is required before these unsupported transformations can be defined.
