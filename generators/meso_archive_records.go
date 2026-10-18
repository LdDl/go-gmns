package generators

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math"
	"reflect"
	"sort"
	"strconv"

	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/macro"
	"github.com/LdDl/go-gmns/meso"
)

// ArchiveRecordHash identifies one exact canonical record without copying its data.
type ArchiveRecordHash struct {
	Table string
	ID    string
	Hash  string
}

type archiveIdentifier interface {
	String() string
	IsZero() bool
}

// The ordered getter lists are the versioned portable fingerprint schema.
var archiveFields = map[string][]string{
	"config":         {"DatasetName", "IDType", "Currency", "VersionNumber", "ShortLength", "LongLength", "Speed", "CRS"},
	"node":           {"Name", "NodeType", "ControlType", "Geom", "AltitudeMeters", "ParentNode"},
	"link":           {"Name", "SourceNode", "TargetNode", "Directed", "LanesNum", "LengthMeters", "FreeSpeed", "Capacity", "GeometryID", "DirFlag", "Geom", "ParentLink", "FacilityType", "BikeFacility", "PedFacility", "Parking", "Jurisdiction", "AllowedUses", "Grade", "RightOfWayWidthMeters", "Toll"},
	"movement":       {"Name", "Node", "IncomeLink", "OutcomeLink", "IncomeLaneStart", "IncomeLaneEnd", "OutcomeLaneStart", "OutcomeLaneEnd", "Type", "MvmtTextID", "ControlType", "Penalty", "Capacity", "AllowedUses", "Geom"},
	"geometry":       {"Geom"},
	"zone":           {"Name", "Boundary", "SuperZone"},
	"use_definition": {"Use", "Description", "PersonsPerVehicle", "PCE", "SpecialConditions"},
	"use_group":      {"Name", "Uses", "Description"},
	"lane":           {"Link", "LaneNum", "AllowedUses", "LeftBarrier", "RightBarrier", "WidthMeters"},
	"segment":        {"Link", "RefNode", "StartMeters", "EndMeters", "Grade", "Capacity", "FreeSpeed", "LanesNum", "LeftLanesAdded", "RightLanesAdded", "BikeFacility", "PedFacility", "Parking", "AllowedUses", "Toll", "Jurisdiction", "RightOfWayWidthMeters"},
	"segment_lane":   {"Segment", "LaneNum", "ParentLane", "AllowedUses", "LeftBarrier", "RightBarrier", "WidthMeters"},
}

func archiveToken(output hash.Hash, text string) {
	fmt.Fprintf(output, "%d:", len([]byte(text)))
	output.Write([]byte(text))
}

func archiveValue(output hash.Hash, value reflect.Value) {
	if !value.IsValid() {
		archiveToken(output, "null")
		return
	}
	if value.CanInterface() {
		id, ok := value.Interface().(archiveIdentifier)
		if ok {
			if id.IsZero() {
				archiveToken(output, "null")
			} else {
				archiveToken(output, "id:"+id.String())
			}
			return
		}
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			archiveToken(output, "null")
		} else {
			archiveValue(output, value.Elem())
		}
	case reflect.String:
		archiveToken(output, "s:"+value.String())
	case reflect.Bool:
		archiveToken(output, "b:"+strconv.FormatBool(value.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		archiveToken(output, "i:"+strconv.FormatInt(value.Int(), 10))
	case reflect.Float64:
		archiveToken(output, fmt.Sprintf("f:%016x", math.Float64bits(value.Float())))
	case reflect.Slice, reflect.Array:
		archiveToken(output, "a:"+strconv.Itoa(value.Len()))
		for index := 0; index < value.Len(); index++ {
			archiveValue(output, value.Index(index))
		}
	case reflect.Struct:
		archiveToken(output, "o:"+strconv.Itoa(value.NumField()))
		for index := 0; index < value.NumField(); index++ {
			archiveToken(output, value.Type().Field(index).Name)
			archiveValue(output, value.Field(index))
		}
	default:
		panic("unsupported archive fingerprint value: " + value.Type().String())
	}
}

func archiveDigest(value interface{}) string {
	output := sha256.New()
	archiveValue(output, reflect.ValueOf(value))
	return hex.EncodeToString(output.Sum(nil))
}

func archiveRecord(table string, record interface{}, level string) ArchiveRecordHash {
	output := sha256.New()
	value := reflect.ValueOf(record)
	key := ""
	field := value.Elem().FieldByName("ID")
	if field.IsValid() {
		key = fmt.Sprint(field.Interface())
	}
	if table == "use_definition" {
		key = value.MethodByName("Use").Call(nil)[0].String()
	}
	if table == "use_group" {
		key = value.MethodByName("Name").Call(nil)[0].String()
	}
	archiveToken(output, table)
	archiveToken(output, key)
	fields := append([]string(nil), archiveFields[table]...)
	if table == "node" {
		if level == "macro" {
			fields = append(fields, "Zone")
		} else {
			fields = append(fields, "MacroZone")
		}
	}
	for _, name := range fields {
		archiveToken(output, name)
		values := value.MethodByName(name).Call(nil)
		if len(values) == 2 && !values[1].Bool() {
			archiveToken(output, "null")
		} else {
			archiveValue(output, values[0])
		}
	}
	return ArchiveRecordHash{Table: table, ID: key, Hash: hex.EncodeToString(output.Sum(nil))}
}

func archiveRecords(source *macro.Net, target *meso.Net) []ArchiveRecordHash {
	level := "macro"
	var net interface{} = source
	if target != nil {
		level = "meso"
		net = target
	}
	value := reflect.ValueOf(net)
	records := []ArchiveRecordHash{archiveRecord("config", value.MethodByName("Config").Call(nil)[0].Interface(), level)}
	for _, table := range []string{"node", "link"} {
		name := "Nodes"
		if table == "link" {
			name = "Links"
		}
		collection := value.Elem().FieldByName(name)
		for _, key := range collection.MapKeys() {
			records = append(records, archiveRecord(table, collection.MapIndex(key).Interface(), level))
		}
	}
	methods := map[string]string{"geometry": "GeometryRecords", "zone": "ZoneRecords", "use_definition": "UseDefinitionRecords", "use_group": "UseGroupRecords", "movement": "MovementRecords", "lane": "LaneRecords", "segment": "SegmentRecords", "segment_lane": "SegmentLaneRecords"}
	for table, method := range methods {
		collection := value.MethodByName(method).Call(nil)[0]
		for index := 0; index < collection.Len(); index++ {
			records = append(records, archiveRecord(table, collection.Index(index).Interface(), level))
		}
	}
	sort.Slice(records, func(first, second int) bool {
		if records[first].Table != records[second].Table {
			return records[first].Table < records[second].Table
		}
		return records[first].ID < records[second].ID
	})
	return records
}

func archiveFloat(value float64) string {
	return fmt.Sprintf("%016x", math.Float64bits(value))
}

func archiveParseFloat(value string) (float64, error) {
	bits, err := strconv.ParseUint(value, 16, 64)
	if err != nil || len(value) != 16 {
		return 0, fmt.Errorf("expected 16 hexadecimal binary64 digits")
	}
	number := math.Float64frombits(bits)
	if !gmns.Finite(number) || archiveFloat(number) != value {
		return 0, fmt.Errorf("noncanonical or nonfinite binary64 value")
	}
	return number, nil
}

func archiveRecordConflicts(code string, expected, actual []ArchiveRecordHash, kind gmns.IDType, level string) error {
	records := make(map[string]ArchiveRecordHash)
	for _, record := range expected {
		records[record.Table+"\x00"+record.ID] = record
	}
	changed := make(map[string]ArchiveRecordHash)
	for _, record := range actual {
		key := record.Table + "\x00" + record.ID
		old, exists := records[key]
		if !exists || old.Hash != record.Hash {
			changed[key] = record
		}
		delete(records, key)
	}
	for key, record := range records {
		changed[key] = record
	}
	keys := make([]string, 0, len(changed))
	for key := range changed {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	conflicts := make([]MesoUpdateConflict, 0, len(keys))
	for _, key := range keys {
		record := changed[key]
		conflict := MesoUpdateConflict{code: code, message: fmt.Sprintf("%s %s %q differs from the saved state", level, record.Table, record.ID)}
		if record.Table == "link" {
			id, _ := gmns.ParseLinkID(record.ID, kind)
			if level == "macro" {
				conflict.source = id
			} else {
				conflict.target = id
			}
		}
		if record.Table == "node" && level == "meso" {
			conflict.targetNode, _ = gmns.ParseNodeID(record.ID, kind)
		}
		if record.Table == "movement" {
			conflict.movement, _ = gmns.ParseMovementID(record.ID, kind)
		}
		conflicts = append(conflicts, conflict)
	}
	if len(conflicts) == 0 {
		return mesoUpdateFailure("invalid_state", "record fingerprint order or multiplicity differs")
	}
	return &MesoUpdateError{conflicts: conflicts}
}
