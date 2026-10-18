# Saving local meso state

Local editing state is application metadata, separate from GMNS tables. Store it alongside the exact macro and meso records. Preparation and accepted updates create the snapshot; saving only serializes it.

```go
data, err := generators.SaveMesoUpdateState(ctx, prepared.UpdateState())
if err != nil {
    return err
}

restored, err := generators.RestoreMesoUpdateState(ctx, source, target, data)
if err != nil {
    return err
}
```

Restoration validates the records, preparation policies and lineage, then reconstructs disposable indexes. It does not generate road geometry, cut sections or rebuild connectors. Inputs remain unchanged. Save records and state together; keep the same pair through previews and Undo/Redo.

| Operation | History |
|---|---|
| `RestoreMesoUpdateState` | Continues a saved history, including all reserved deleted link IDs |
| `InitializeMesoUpdateState` | Explicitly starts a new history from current records and trusted `MesoArchiveMetadata` |

Initialization supports the `standard-macro-sections-3` profile with complete section and explicit connector associations. It validates topology, lane scopes and inherited attributes. Existing geometry becomes the accepted baseline: old provenance cannot prove that geometry was never changed. Only surviving IDs can be reserved; unknown deleted IDs from the previous history may be reused. Do not initialize if external consumers still depend on that old allocation history. Ordinary OSM generator ancestry alone is insufficient because its shortcut topology uses a different profile.

`MesoUpdateError.Conflicts()` reports unsupported profiles, missing or contradictory lineage, stale records and malformed state. Manual refinement merging and detailed lane topology edits remain outside the local update scope.

## Portable representation

`MesoArchiveSnapshot` uses the application format `meso-update-state-1`. IDs are exact strings interpreted by `IDType`. It stores policies, lineage, fixed offset decisions, reservations and per-record SHA-256 fingerprints instead of duplicate network geometry. Section interval numbers use 16 lowercase hexadecimal binary64 digits. Fingerprints distinguish absence, supplied zero and negative zero.

Record serialization must also preserve exact numbers and representation fields. Ordinary JavaScript `JSON.stringify` loses negative zero; use a serializer that emits the raw JSON token `-0`. Do not round geometry, convert IDs through floating point, choose a different inline/shared representation or fill optional values between saving and restoring. A mismatch is rejected without repair. Checksums detect inconsistencies; they are not signatures or evidence of historical authenticity.

The versioned field order and token encoding are defined in `generators/meso_archive_records.go` and the archive structs. Canonical records exclude disposable projected coordinates, adjacency and generator caches. Those can be reconstructed from current records and verified lineage. Validation and copying visit the whole dataset; geometry generation remains absent during restoration.
