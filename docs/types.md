# Types and constants

[Back to README](../README.md)

## Contents

- [Control and node types](#control-and-node-types)
- [Movement types and codes](#movement-types-and-codes)
- [Agents and cells](#agents-and-cells)
- [Road categories](#road-categories)
- [Boundary and activity types](#boundary-and-activity-types)
- [Infrastructure](#infrastructure)
- [Geometry direction](#geometry-direction)

These are the library's built-in values, including GMNS categories and application-specific classifications. They are not all mandatory GMNS enumerations. Tables show string values or the output of `String()`; numeric enum ordinals are not interchange codes.

## Control and node types

Package: `github.com/LdDl/go-gmns/gmns/types`.

`ControlType` accepts custom strings. Empty means absent; `no_control` explicitly describes an uncontrolled node or link.

| Constant | Value | Meaning |
|---|---|---|
| `CONTROL_TYPE_NOT_SIGNAL` | `common` | Legacy non-signal classification |
| `CONTROL_TYPE_IS_SIGNAL` | `signal` | Traffic signal |
| `CONTROL_TYPE_NO_CONTROL` | `no_control` | Explicitly uncontrolled |
| `CONTROL_TYPE_YIELD` | `yield` | Yield control |
| `CONTROL_TYPE_STOP` | `stop` | Stop control |
| `CONTROL_TYPE_STOP_2_WAY` | `stop_2_way` | Two-way stop control |
| `CONTROL_TYPE_STOP_4_WAY` | `stop_4_way` | Four-way stop control |
| `CONTROL_TYPE_SIGNAL_WITH_RTOR` | `signal_with_RTOR` | Traffic signal with right turn on red |

For example, use `macro.WithNodeControlType(types.CONTROL_TYPE_STOP)`. These categories describe control, not signal timing or executable traffic rules. Supplied control categories other than empty, `common` and `no_control` preserve explicit junctions during meso generation.

Node types are descriptive strings. The following constants come from GMNS examples; custom values are accepted and empty means absent.

| Constant | Value | Meaning |
|---|---|---|
| `NODE_TYPE_INTERSECTION` | `intersection` | Road intersection |
| `NODE_TYPE_TRANSIT_STATION` | `transit station` | Transit station |
| `NODE_TYPE_PARK_AND_RIDE` | `park & ride` | Park-and-ride facility |

## Movement types and codes

Package: `github.com/LdDl/go-gmns/movement`.

`MovementType` accepts custom strings. Empty means absent and is distinct from explicit `undefined`.

| Constant | Value | Meaning |
|---|---|---|
| `MOVEMENT_TYPE_UNDEFINED` | `undefined` | Explicitly unclassified |
| `MOVEMENT_TYPE_THRU` | `thru` | Through movement |
| `MOVEMENT_TYPE_RIGHT` | `right` | Right turn |
| `MOVEMENT_TYPE_LEFT` | `left` | Left turn |
| `MOVEMENT_TYPE_U_TURN` | `uturn` | U-turn |
| `MOVEMENT_TYPE_MERGE` | `merge` | Merge |
| `MOVEMENT_TYPE_DIVERGE` | `diverge` | Diverge |

`MovementCompositeType` also accepts custom strings. The built-in codes combine the incoming travel direction with the maneuver: eastbound means traveling toward the east, not arriving from the east.

| Constant | Value | Direction | Maneuver |
|---|---|---|---|
| `MOVEMENT_UNDEFINED` | `undefined` | Unclassified | Unclassified |
| `MOVEMENT_EBT` | `EBT` | Eastbound | Through |
| `MOVEMENT_EBL` | `EBL` | Eastbound | Left |
| `MOVEMENT_EBR` | `EBR` | Eastbound | Right |
| `MOVEMENT_EBU` | `EBU` | Eastbound | U-turn |
| `MOVEMENT_WBT` | `WBT` | Westbound | Through |
| `MOVEMENT_WBL` | `WBL` | Westbound | Left |
| `MOVEMENT_WBR` | `WBR` | Westbound | Right |
| `MOVEMENT_WBU` | `WBU` | Westbound | U-turn |
| `MOVEMENT_NBT` | `NBT` | Northbound | Through |
| `MOVEMENT_NBL` | `NBL` | Northbound | Left |
| `MOVEMENT_NBR` | `NBR` | Northbound | Right |
| `MOVEMENT_NBU` | `NBU` | Northbound | U-turn |
| `MOVEMENT_SBT` | `SBT` | Southbound | Through |
| `MOVEMENT_SBL` | `SBL` | Southbound | Left |
| `MOVEMENT_SBR` | `SBR` | Southbound | Right |
| `MOVEMENT_SBU` | `SBU` | Southbound | U-turn |

## Agents and cells

Package: `github.com/LdDl/go-gmns/gmns/types`. These are numeric enums with the listed built-in values; they do not accept custom strings.

| Constant | String value | Meaning |
|---|---|---|
| `AGENT_UNDEFINED` | `undefined` | Unclassified agent |
| `AGENT_AUTO` | `auto` | Motor vehicle |
| `AGENT_BIKE` | `bike` | Bicycle |
| `AGENT_WALK` | `walk` | Pedestrian |
| `CELL_UNDEFINED` | `undefined` | Unclassified micro link |
| `CELL_FORWARD` | `forward` | Travel along a lane |
| `CELL_LANE_CHANGE` | `lane_change` | Travel to an adjacent lane |

There are no separate left/right `CellType` constants. The connected nodes determine which lane the link reaches. `AllowedUses()` refers to dataset use definitions/groups and is independent of the built-in `AgentType` enum.

## Road categories

Package: `github.com/LdDl/go-gmns/gmns/types`. `LinkType`, `LinkClass` and `LinkConnectionType` are numeric enums; use the defined constants.

| LinkType constant | String value | Meaning |
|---|---|---|
| `LINK_UNDEFINED` | `undefined` | Unclassified link |
| `LINK_MOTORWAY` | `motorway` | Motorway |
| `LINK_TRUNK` | `trunk` | Trunk road |
| `LINK_PRIMARY` | `primary` | Primary road |
| `LINK_SECONDARY` | `secondary` | Secondary road |
| `LINK_TERTIARY` | `tertiary` | Tertiary road |
| `LINK_RESIDENTIAL` | `residential` | Residential road |
| `LINK_LIVING_STREET` | `living_street` | Living street |
| `LINK_SERVICE` | `service` | Service road |
| `LINK_CYCLEWAY` | `cycleway` | Cycleway |
| `LINK_FOOTWAY` | `footway` | Footway |
| `LINK_TRACK` | `track` | Track |
| `LINK_UNCLASSIFIED` | `unclassified` | OSM unclassified road category |
| `LINK_CONNECTOR` | `connector` | Network connector |
| `LINK_RAILWAY` | `railway` | Railway |
| `LINK_AEROWAY` | `aeroway` | Aeroway |

| Constant | String value | Meaning |
|---|---|---|
| `LINK_CLASS_UNDEFINED` | `undefined` | Unclassified transport class |
| `LINK_CLASS_HIGHWAY` | `highway` | Road network |
| `LINK_CLASS_RAILWAY` | `railway` | Rail network |
| `LINK_CLASS_AEROWAY` | `aeroway` | Airfield network |
| `NOT_A_LINK` | `no` | Ordinary way |
| `IS_LINK` | `yes` | Connecting way, such as an OSM ramp |

`facility_type` is a separate descriptive string and has no fixed list of constants.

## Boundary and activity types

Package: `github.com/LdDl/go-gmns/gmns/types`. These are numeric enums used for network classification.

| Constant | String value | Meaning |
|---|---|---|
| `BOUNDARY_NONE` | `none` | Not classified as a boundary |
| `BOUNDARY_INCOME_ONLY` | `income_only` | Incoming-only boundary classification |
| `BOUNDARY_OUTCOME_ONLY` | `outcome_only` | Outgoing-only boundary classification |
| `BOUNDARY_INCOME_OUTCOME` | `income_outcome` | Both boundary directions |
| `ACTIVITY_NONE` | `none` | No activity classification |
| `ACTIVITY_POI` | `poi` | Point-of-interest activity |
| `ACTIVITY_LINK` | `link` | Activity classified from connected links |

## Infrastructure

Package: `github.com/LdDl/go-gmns/gmns/types`. Bicycle, pedestrian and parking categories accept custom strings. Empty means absent; explicit `none` and `unknown` retain their stated meaning.

| Bicycle constant | Value | Meaning |
|---|---|---|
| `BIKE_FACILITY_UNSEPARATED_LANE` | `unseparated bike lane` | Unseparated bicycle lane |
| `BIKE_FACILITY_BUFFERED_LANE` | `buffered bike lane` | Buffered bicycle lane |
| `BIKE_FACILITY_SEPARATED_LANE` | `separated bike lane` | Separated bicycle lane |
| `BIKE_FACILITY_COUNTER_FLOW_LANE` | `counter-flow bike lane` | Bicycle lane against the main flow |
| `BIKE_FACILITY_PAVED_SHOULDER` | `paved shoulder` | Paved shoulder |
| `BIKE_FACILITY_SHARED_LANE` | `shared lane` | Lane shared with other traffic |
| `BIKE_FACILITY_SHARED_USE_PATH` | `shared use path` | Shared-use path |
| `BIKE_FACILITY_OFF_ROAD_UNPAVED_TRAIL` | `off-road unpaved trail` | Unpaved trail away from the road |
| `BIKE_FACILITY_OTHER` | `other` | Other bicycle facility |
| `BIKE_FACILITY_NONE` | `none` | No bicycle facility |

| Pedestrian constant | Value | Meaning |
|---|---|---|
| `PED_FACILITY_UNKNOWN` | `unknown` | Unknown facility |
| `PED_FACILITY_NONE` | `none` | No pedestrian facility |
| `PED_FACILITY_SHOULDER` | `shoulder` | Shoulder |
| `PED_FACILITY_SIDEWALK` | `sidewalk` | Sidewalk |
| `PED_FACILITY_OFFSTREET_PATH` | `offstreet_path` | Off-street path |
| `PED_FACILITY_CROSSWALK` | `crosswalk` | Crosswalk |

| Parking constant | Value | Meaning |
|---|---|---|
| `PARKING_UNKNOWN` | `unknown` | Unknown parking arrangement |
| `PARKING_NONE` | `none` | No curbside parking |
| `PARKING_PARALLEL` | `parallel` | Parallel parking |
| `PARKING_ANGLE` | `angle` | Angled parking |
| `PARKING_OTHER` | `other` | Other parking arrangement |

## Geometry direction

Package: `github.com/LdDl/go-gmns/gmns`.

| Constant | Value | Meaning |
|---|---|---|
| `DIR_FLAG_FORWARD` | `1` | Referenced shape follows source-to-target order |
| `DIR_FLAG_REVERSE` | `-1` | Referenced shape runs opposite to source-to-target order |
| `DIR_FLAG_NONE` | `0` | Unspecified sentinel; not valid as a supplied dataset `dir_flag` |

Omit the direction option for absence. `dir_flag` describes geometry orientation; `Directed()` describes whether link travel is directed. See [geometry references](data-model.md#geometry-references).
