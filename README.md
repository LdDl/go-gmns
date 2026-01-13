# GMNS - General Modeling Network Specification

## About
Go implementation of basic data in GMNS (General Modeling Network Specification). Contains data structures and generators utils for multi-resolution transportation networks.

This package provides:
- Data structures for macroscopic, mesoscopic, and microscopic networks
- Generators for meso and micro networks from macro networks
- Movement generation at intersections/junctions
- GeoJSON export for  debugging and visualization

This package has been created for [osm2gmns](https://github.com/LdDl/osm2gmns) mostly, but can be used independently.

## Installation

```bash
go get github.com/LdDl/go-gmns
```

## Current state

### Network levels

- [x] **Macroscopic network** (`macro/`)
    - [x] Links with lane information
    - [x] Nodes with control/boundary types
    - [x] Network container
    - [x] GeoJSON export

- [x] **Movements** a.k.a. allowed maneuvers at junctions (`movement/`)
    - [x] Movement types (THRU, LEFT, RIGHT, UTURN)
    - [x] Composite movement classification
    - [x] Geometry utilities
    - [x] GeoJSON export

- [x] **Mesoscopic network** (`meso/`)
    - [x] Lane-level links
    - [x] Lane-level nodes
    - [x] Network container
    - [x] GeoJSON export

- [x] **Microscopic network** (`micro/`)
    - [x] Cell-based links (forward, lane-change)
    - [x] Cell vertex nodes
    - [x] Network container
    - [x] GeoJSON export

### Generators (`generators/`)

- [x] **Movements** - turn movements at intersections
- [x] **Mesoscopic data** - expands macro network to lane-level
- [x] **Microscopic data** - cell-based decomposition of meso network

### Basic stuff

- [x] **Types** (`gmns/types/`)
    - [x] LinkType, LinkClass, LinkConnectionType
    - [x] ControlType, BoundaryType
    - [x] AgentType (auto, bike, walk, etc.)
    - [x] CellType (forward, lane_change)
    - [x] ActivityType, AccessType
    - [x] NetworkType, HighwayType

- [x] **Utils** (`utils/geomath/`)
    - [x] Coordinate transformations (WGS84/Euclidean)
    - [x] Line offset calculations
    - [x] Distance calculations

### further work:

- [ ] Traffic signal timing
- [ ] Extended test coverage
- [ ] Performance benchmarks

## Usage Example

The best thing to get idea is to explore this tool: https://github.com/LdDl/osm2gmns.

## References
Lu, J., & Zhou, X.S. (2023). Virtual track networks: A hierarchical modeling framework and open-source tools for simplified and efficient connected and automated mobility (CAM) system design based on general modeling network specification (GMNS). Transportation Research Part C: Emerging Technologies, 153, 104223. [paper link](https://linkinghub.elsevier.com/retrieve/pii/S0968090X23002127)
