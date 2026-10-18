package generators

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// MesoArchiveVersion identifies application metadata, not a GMNS table or schema version.
const MesoArchiveVersion = "meso-update-state-1"

// MesoArchiveOptions records explicit preparation policies.
type MesoArchiveOptions struct {
	LaneLayout     string
	MovementPolicy string
	DefaultLanes   *int
}

// MesoArchiveSection uses exact ID text and binary64 hexadecimal interval endpoints.
type MesoArchiveSection struct {
	Link           string
	Source         string
	Segment        string
	Start          string
	End            string
	Lanes          []int
	LaneIDs        []string
	SegmentLaneIDs []string
}

// MesoArchiveConnection associates an explicit source movement and ordered lane scopes.
type MesoArchiveConnection struct {
	Link     string
	Movement string
	Inbound  []int
	Outbound []int
}

// MesoArchiveAssumption retains an explicitly accepted missing lane count.
type MesoArchiveAssumption struct {
	Link  string
	Lanes int
}

// MesoArchiveOffset retains the fixed reverse-pair decision for a source road.
type MesoArchiveOffset struct {
	Link   string
	Offset bool
}

// MesoArchiveMetadata is supplied for explicit initialization of a new history.
// It does not prove that historical geometry was unmodified after preparation.
type MesoArchiveMetadata struct {
	Algorithm   string
	Options     MesoArchiveOptions
	Sections    []MesoArchiveSection
	Connections []MesoArchiveConnection
	Assumptions []MesoArchiveAssumption
}

// MesoArchiveSnapshot is a portable history continuation envelope.
// Numeric IDs are decimal strings interpreted using IDType. Hashes retain signed zero.
type MesoArchiveSnapshot struct {
	Version       string
	Algorithm     string
	IDType        gmns.IDType
	History       string
	Metadata      MesoArchiveMetadata
	Offsets       []MesoArchiveOffset
	ReservedLinks []string
	Source        []ArchiveRecordHash
	Target        []ArchiveRecordHash
	Integrity     string
}

// SaveMesoUpdateState serializes the snapshot prepared by the accepted transaction.
// It performs no record traversal, geometry calculation or state mutation.
func SaveMesoUpdateState(ctx context.Context, state *MesoUpdateState) ([]byte, error) {
	if ctx == nil || state == nil || state.snapshot == nil {
		return nil, mesoUpdateFailure("invalid_request", "saving requires a context and completed update state")
	}
	err := ctx.Err()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state.snapshot)
}

// RestoreMesoUpdateState continues a saved history against unchanged canonical records.
// It never generates geometry or changes the supplied datasets.
func RestoreMesoUpdateState(ctx context.Context, source *macro.Net, target *meso.Net, data []byte) (*MesoPreparationResult, error) {
	if ctx == nil {
		return nil, mesoUpdateFailure("invalid_request", "restoration requires a context")
	}
	err := ctx.Err()
	if err != nil {
		return nil, err
	}
	err = archiveUniqueJSON(data)
	if err != nil {
		return nil, mesoUpdateFailure("invalid_state", err.Error())
	}
	snapshot := MesoArchiveSnapshot{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&snapshot)
	if err != nil {
		return nil, mesoUpdateFailure("invalid_state", err.Error())
	}
	integrity := snapshot.Integrity
	snapshot.Integrity = ""
	if integrity == "" || archiveDigest(snapshot) != integrity {
		return nil, mesoUpdateFailure("invalid_state", "snapshot integrity does not match")
	}
	snapshot.Integrity = integrity
	if snapshot.Version != MesoArchiveVersion || snapshot.Algorithm != MesoUpdateAlgorithm {
		return nil, mesoUpdateFailure("unsupported_profile", "unsupported archive or local update algorithm")
	}
	if snapshot.History != "complete" && snapshot.History != "initialized-current-records" {
		return nil, mesoUpdateFailure("invalid_state", "unknown history origin")
	}
	err = validateArchiveDatasets(source, target)
	if err != nil {
		return nil, err
	}
	if source.Config().IDType() != snapshot.IDType {
		return nil, mesoUpdateFailure("invalid_state", "snapshot identifier kind differs from records")
	}
	if !reflect.DeepEqual(snapshot.Source, archiveRecords(source, nil)) {
		return nil, archiveRecordConflicts("stale_source", snapshot.Source, archiveRecords(source, nil), snapshot.IDType, "macro")
	}
	if !reflect.DeepEqual(snapshot.Target, archiveRecords(nil, target)) {
		return nil, archiveRecordConflicts("manual_meso_conflict", snapshot.Target, archiveRecords(nil, target), snapshot.IDType, "meso")
	}
	prepared, layout, err := restoreArchiveRecords(ctx, source, target, &snapshot.Metadata)
	if err != nil {
		return nil, err
	}
	offsets := make(map[gmns.LinkID]bool)
	for _, entry := range snapshot.Offsets {
		id, parseErr := archiveLinkID(entry.Link, snapshot.IDType)
		if parseErr != nil || source.Links[id] == nil {
			return nil, mesoUpdateFailure("invalid_state", "offset references an invalid source link")
		}
		_, duplicate := offsets[id]
		if duplicate {
			return nil, mesoUpdateFailure("invalid_state", "duplicate offset decision")
		}
		offsets[id] = entry.Offset
	}
	if len(offsets) != len(source.Links) {
		return nil, mesoUpdateFailure("invalid_state", "missing fixed offset decisions")
	}
	for id, processing := range mesoRoadProcessing(prepared.macroNet) {
		if offsets[id] != processing.needsOffset {
			return nil, mesoUpdateFailure("invalid_state", fmt.Sprintf("source link %v: fixed offset decision contradicts source pairing", id))
		}
	}
	reserved := make(map[gmns.LinkID]bool)
	for _, text := range snapshot.ReservedLinks {
		id, parseErr := archiveLinkID(text, snapshot.IDType)
		if parseErr != nil || reserved[id] {
			return nil, mesoUpdateFailure("invalid_state", "invalid or duplicate reserved link ID")
		}
		reserved[id] = true
	}
	for id := range target.Links {
		if !reserved[id] {
			return nil, mesoUpdateFailure("invalid_state", fmt.Sprintf("missing reservation for meso link %v", id))
		}
	}
	previous := &MesoUpdateState{offsets: offsets, reservedLinks: reserved, snapshot: &snapshot}
	bindMesoPreparation(prepared, layout, previous)
	prepared.updateState.importedSource = source.Clone()
	prepared.updateState.importedTarget = target.Clone()
	err = ctx.Err()
	if err != nil {
		return nil, err
	}
	return prepared, nil
}

// InitializeMesoUpdateState explicitly starts a new history from trusted preparation metadata.
// Surviving IDs are reserved. Deleted historical IDs are unknowable and are not recovered.
// The caller accepts existing geometry as the initial baseline after structural validation.
func InitializeMesoUpdateState(ctx context.Context, source *macro.Net, target *meso.Net, metadata *MesoArchiveMetadata) (*MesoPreparationResult, error) {
	prepared, layout, err := restoreArchiveRecords(ctx, source, target, metadata)
	if err != nil {
		return nil, err
	}
	bindMesoPreparation(prepared, layout, nil)
	prepared.updateState.snapshot.History = "initialized-current-records"
	prepared.updateState.snapshot.Integrity = ""
	prepared.updateState.snapshot.Integrity = archiveDigest(*prepared.updateState.snapshot)
	prepared.updateState.importedSource = source.Clone()
	prepared.updateState.importedTarget = target.Clone()
	err = ctx.Err()
	if err != nil {
		return nil, err
	}
	return prepared, nil
}

func archiveLinkID(text string, kind gmns.IDType) (gmns.LinkID, error) {
	id, err := gmns.ParseLinkID(text, kind)
	if err != nil {
		return id, err
	}
	if id.String() != text {
		return id, fmt.Errorf("noncanonical ID")
	}
	return id, nil
}

func makeMesoArchive(state *MesoUpdateState, previous *MesoUpdateState) *MesoArchiveSnapshot {
	prepared := state.prepared
	snapshot := &MesoArchiveSnapshot{Version: MesoArchiveVersion, Algorithm: MesoUpdateAlgorithm, IDType: prepared.macroNet.Config().IDType(), History: "complete"}
	if previous != nil && previous.snapshot != nil {
		snapshot.History = previous.snapshot.History
	}
	snapshot.Metadata.Algorithm = prepared.algorithm
	snapshot.Metadata.Options.LaneLayout = "uniform"
	if prepared.options.detailedLaneLayout {
		snapshot.Metadata.Options.LaneLayout = "detailed"
	}
	snapshot.Metadata.Options.MovementPolicy = "single"
	if prepared.options.orderedLaneConnections {
		snapshot.Metadata.Options.MovementPolicy = "ordered"
	}
	if prepared.options.defaultLanes != nil {
		count := *prepared.options.defaultLanes
		snapshot.Metadata.Options.DefaultLanes = &count
	}
	for _, section := range prepared.sections {
		entry := MesoArchiveSection{Link: section.linkID.String(), Source: section.sourceLinkID.String(), Segment: section.segmentID.String(), Start: archiveFloat(section.start), End: archiveFloat(section.end), Lanes: section.LaneNumbers()}
		for _, id := range section.laneIDs {
			entry.LaneIDs = append(entry.LaneIDs, id.String())
		}
		for _, id := range section.segmentLaneIDs {
			entry.SegmentLaneIDs = append(entry.SegmentLaneIDs, id.String())
		}
		snapshot.Metadata.Sections = append(snapshot.Metadata.Sections, entry)
	}
	for _, connection := range prepared.connections {
		snapshot.Metadata.Connections = append(snapshot.Metadata.Connections, MesoArchiveConnection{Link: connection.linkID.String(), Movement: connection.movementID.String(), Inbound: connection.InboundLanes(), Outbound: connection.OutboundLanes()})
	}
	for _, assumption := range prepared.laneAssumptions {
		snapshot.Metadata.Assumptions = append(snapshot.Metadata.Assumptions, MesoArchiveAssumption{Link: assumption.linkID.String(), Lanes: assumption.lanes})
	}
	for _, link := range macroLinksToSlice(prepared.macroNet.Links) {
		id := link.ID
		snapshot.Offsets = append(snapshot.Offsets, MesoArchiveOffset{Link: id.String(), Offset: state.offsets[id]})
	}
	ids := make([]gmns.LinkID, 0, len(state.reservedLinks))
	for id := range state.reservedLinks {
		ids = append(ids, id)
	}
	sortMesoUpdateLinkIDs(ids)
	for _, id := range ids {
		snapshot.ReservedLinks = append(snapshot.ReservedLinks, id.String())
	}
	snapshot.Source = archiveRecords(prepared.macroNet, nil)
	snapshot.Target = archiveRecords(nil, prepared.mesoNet)
	snapshot.Integrity = archiveDigest(*snapshot)
	return snapshot
}

// Reject duplicate keys and trailing JSON instead of accepting last-key-wins state.
func archiveUniqueJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	err := archiveJSONValue(decoder, reflect.TypeOf(MesoArchiveSnapshot{}))
	if err != nil {
		return err
	}
	_, err = decoder.Token()
	if err != io.EOF {
		return fmt.Errorf("trailing archive data")
	}
	return nil
}

func archiveJSONValue(decoder *json.Decoder, schema reflect.Type) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	if schema.Kind() == reflect.Pointer {
		schema = schema.Elem()
	}
	if delimiter == '{' && schema.Kind() != reflect.Struct || delimiter == '[' && schema.Kind() != reflect.Slice {
		return fmt.Errorf("unexpected JSON structure")
	}
	seen := make(map[string]bool)
	for decoder.More() {
		var child reflect.Type
		if delimiter == '{' {
			key, keyErr := decoder.Token()
			if keyErr != nil {
				return keyErr
			}
			name, valid := key.(string)
			if !valid || seen[name] {
				return fmt.Errorf("duplicate or invalid JSON key")
			}
			field, exists := schema.FieldByName(name)
			if !exists {
				return fmt.Errorf("unknown archive field %q", name)
			}
			child = field.Type
			seen[name] = true
		} else {
			child = schema.Elem()
		}
		err = archiveJSONValue(decoder, child)
		if err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
