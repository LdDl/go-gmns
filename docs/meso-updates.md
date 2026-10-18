# Local meso updates

`generators.UpdateMesoscopic` applies an explicit macro edit to a prepared meso network. It returns independent macro and meso records, updated lineage and a new immutable `UpdateState()`. Inputs remain unchanged on success, cancellation and failure.

```go
edit := generators.NewMesoEdit()
edit.ChangeRightmostLane(gmns.LinkIDFromInt64(13), -1)

result, err := generators.UpdateMesoscopic(ctx,
    prepared.Macro(), prepared.Meso(), prepared.UpdateState(), edit,
)
if err != nil {
    return err
}
prepared = result.Preparation()
```

Prepare the initial network with `PrepareMesoscopic`, restore saved state with `RestoreMesoUpdateState`, or explicitly initialize supported historical records with `InitializeMesoUpdateState`. Retain its state for previews; adopt the returned preparation only when the caller accepts the complete transaction. The library does not manage UI history or commit caller-owned state.

## Supported edits

| Operation | Behavior |
|---|---|
| `RenameRoad` | Updates road sections; connector names remain independent |
| `SetRoadFreeSpeed`, `ClearRoadFreeSpeed` | Updates sections and upstream connectors; segment overrides, including zero, take precedence |
| `ChangeRightmostLane` | Adds or removes exactly one rightmost lane over a complete directed road, retaining 1..99 lanes |
| `SetMovement` | Explicitly adds a movement or replaces an existing movement's lane ranges; other existing properties must remain unchanged |
| `RemoveMovement` | Removes the supplied movement and its connector; absence is not a legal prohibition |

Multiple roads and movement resolutions form one transaction. Lane-count edits require a road without supplied lane or segment records. Detailed roads support name/speed edits and movement range changes, but local tapers and detailed lane topology edits are rejected. Movement ranges must match active endpoint lanes and the accepted preparation policy.

## Dependencies and identities

Name/speed edits perform no geometry calculations. Lane-count edits recalculate only the selected roads and their incident explicit connectors. Under the current algorithm, reverse pairing depends on unchanged source geometry, and each road's offset and cuts depend on its own lane layout. The opposite road and neighboring sections therefore keep their geometry unless explicitly edited.

All surviving section, node and connector IDs are retained. Explicit new movements allocate connector IDs in typed movement-ID order, reserving all earlier link IDs, including deleted ones. String spelling and arbitrary-size integers are preserved. This profile creates no nodes or shared geometry records. Inputs, catalogs and detailed records are copied and validated globally; offsets and cuts are calculated locally.

`AffectedRoads()` and `AffectedMovements()` identify the source scope. `RecomputedRoads()` and `RecomputedConnections()` list actual geometry work. `CreatedLinks()`, `UpdatedLinks()`, `DeletedLinks()` and `UpdatedNodes()` describe target changes. Returned slices are copies.

## Conflicts and limits

Use `errors.As` with `*generators.MesoUpdateError`, then inspect `Conflicts()`. Unresolved lane scopes report movement IDs, source roads and inbound/outbound sides. Invalid requests, stale sources, incompatible lineage and manual meso changes are rejected without a result. No replacement lanes or turns are inferred.

Local updates require a binding created by preparation, restoration, initialization or a prior accepted update. External manual meso changes, including additional rows, are rejected. Keep records and their state together when branching previews or implementing Undo. Callers must not mutate inputs concurrently.

The full preparation algorithm remains `standard-macro-sections-3`; local updates use `local-macro-roads-1`. Incremental connector IDs may differ from a fresh full build. Compare roads by source link and section index, nodes by road boundary, and connectors by movement ID. Numerical differences must be reported without an implicit tolerance or quantization.
