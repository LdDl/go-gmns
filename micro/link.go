package micro

import (
	"github.com/LdDl/go-gmns/gmns"
	"github.com/LdDl/go-gmns/gmns/types"
	"github.com/LdDl/go-gmns/movement"
	"github.com/paulmach/orb"
)

// Link represents a microscopic network link (cell)
type Link struct {
	lanesNum *int

	toll         *float64
	facilityType string
	allowedUses  string

	rightOfWayWidthMeters *float64

	bikeFacility string
	pedFacility  string
	parking      string

	gradePercent *float64
	jurisdiction string

	directed bool

	parentLinkID *gmns.LinkID

	name string

	ID gmns.LinkID

	geom          orb.LineString
	geomEuclidean orb.LineString
	geometryID    *gmns.GeometryID
	dirFlag       *gmns.DirectionFlag

	lengthMeters *float64

	sourceNodeID gmns.NodeID
	targetNodeID gmns.NodeID

	// Parent references
	mesoLinkID  gmns.LinkID
	macroLinkID gmns.LinkID
	macroNodeID gmns.NodeID

	// Cell properties
	cellType types.CellType
	laneID   int // Lane number this cell belongs to

	// Movement properties
	isFirstMovementCell   bool
	movementCompositeType movement.MovementCompositeType
	additionalTravelCost  float64

	// Inherited from parent data
	mesoLinkType      types.LinkType
	controlType       types.ControlType
	freeSpeed         *float64
	capacity          *float64
	allowedAgentTypes []types.AgentType
}

// NewLinkFrom creates pointer to the new microscopic Link
func NewLinkFrom(id gmns.LinkID, sourceNodeID, targetNodeID gmns.NodeID, options ...func(*Link)) *Link {
	newLink := &Link{
		directed:             true,
		ID:                   id,
		geom:                 orb.LineString{},
		geomEuclidean:        orb.LineString{},
		sourceNodeID:         sourceNodeID,
		targetNodeID:         targetNodeID,
		mesoLinkID:           gmns.LinkID{},
		macroLinkID:          gmns.LinkID{},
		macroNodeID:          gmns.NodeID{},
		cellType:             types.CELL_FORWARD,
		laneID:               0,
		isFirstMovementCell:  false,
		additionalTravelCost: 0.0,
		mesoLinkType:         types.LINK_UNDEFINED,
		allowedAgentTypes:    []types.AgentType{},
	}
	for _, option := range options {
		option(newLink)
	}
	return newLink
}

// Geom returns corresponding geometry [WGS84]
func (link *Link) Geom() orb.LineString {
	return link.geom
}

// GeomEuclidean returns corresponding geometry [Euclidean]
func (link *Link) GeomEuclidean() orb.LineString {
	return link.geomEuclidean
}

// GeometryID returns the optional reference to the GMNS geometry table.
// The boolean distinguishes an absent reference from geometry ID zero.
func (link *Link) GeometryID() (gmns.GeometryID, bool) {
	if link.geometryID == nil {
		return gmns.GeometryID{}, false
	}
	return *link.geometryID, true
}

// DirFlag returns the optional GMNS shape orientation: 1 follows the link's
// endpoints, -1 opposes them, and 0 means undirected or no geometry information.
// An absent flag is distinct from an explicitly supplied zero.
func (link *Link) DirFlag() (gmns.DirectionFlag, bool) {
	if link.dirFlag == nil {
		return gmns.DIR_FLAG_NONE, false
	}
	return *link.dirFlag, true
}

// LengthMeters returns the supplied length in meters and whether it is set.
func (link *Link) LengthMeters() (float64, bool) {
	if link.lengthMeters == nil {
		return 0, false
	}
	return *link.lengthMeters, true
}

// SourceNode returns identifier of source node
func (link *Link) SourceNode() gmns.NodeID {
	return link.sourceNodeID
}

// TargetNode returns identifier of target node
func (link *Link) TargetNode() gmns.NodeID {
	return link.targetNodeID
}

// MesoLink returns identifier of parent mesoscopic link. Returns a zero ID when absent.
func (link *Link) MesoLink() gmns.LinkID {
	return link.mesoLinkID
}

// MacroLink returns identifier of parent macroscopic link. Returns a zero ID when absent.
func (link *Link) MacroLink() gmns.LinkID {
	return link.macroLinkID
}

// MacroNode returns identifier of parent macroscopic node. Returns a zero ID when absent.
func (link *Link) MacroNode() gmns.NodeID {
	return link.macroNodeID
}

// CellType returns the cell type (forward or lane_change)
func (link *Link) CellType() types.CellType {
	return link.cellType
}

// LaneID returns the lane number this cell belongs to
func (link *Link) LaneID() int {
	return link.laneID
}

// IsFirstMovementCell returns true if this is the first cell in a movement
func (link *Link) IsFirstMovementCell() bool {
	return link.isFirstMovementCell
}

// MovementCompositeType returns the movement composite type
func (link *Link) MovementCompositeType() movement.MovementCompositeType {
	return link.movementCompositeType
}

// AdditionalTravelCost returns additional travel cost for this cell
func (link *Link) AdditionalTravelCost() float64 {
	return link.additionalTravelCost
}

// MesoLinkType returns the parent mesoscopic link type
func (link *Link) MesoLinkType() types.LinkType {
	return link.mesoLinkType
}

// ControlType returns control type
func (link *Link) ControlType() types.ControlType {
	return link.controlType
}

// FreeSpeed returns the supplied free-flow speed in km/h and whether it is set.
func (link *Link) FreeSpeed() (float64, bool) {
	if link.freeSpeed == nil {
		return 0, false
	}
	return *link.freeSpeed, true
}

// Capacity returns the supplied saturation capacity in PCE/hour/lane and whether it is set.
func (link *Link) Capacity() (float64, bool) {
	if link.capacity == nil {
		return 0, false
	}
	return *link.capacity, true
}

// AllowedAgentTypes returns set of allowed agent types
func (link *Link) AllowedAgentTypes() []types.AgentType {
	return link.allowedAgentTypes
}

// SetSourceNode updates the source node ID
func (link *Link) SetSourceNode(nodeID gmns.NodeID) {
	link.sourceNodeID = nodeID
}

// SetTargetNode updates the target node ID
func (link *Link) SetTargetNode(nodeID gmns.NodeID) {
	link.targetNodeID = nodeID
}

// WithLineGeom sets geometry [WGS84] for the link
func WithLineGeom(geom orb.LineString) func(*Link) {
	return func(link *Link) {
		link.geom = geom
	}
}

// WithLineGeomEuclidean sets geometry [Euclidean] for the link
func WithLineGeomEuclidean(geomEuclidean orb.LineString) func(*Link) {
	return func(link *Link) {
		link.geomEuclidean = geomEuclidean
	}
}

// WithLengthMeters sets length in meters
func WithLengthMeters(lengthMeters float64) func(*Link) {
	return func(link *Link) {
		value := lengthMeters
		link.lengthMeters = &value
	}
}

// WithMesoLinkID sets parent mesoscopic link identifier
func WithMesoLinkID(mesoLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.mesoLinkID = mesoLinkID
	}
}

// WithMacroLinkID sets parent macroscopic link identifier
func WithMacroLinkID(macroLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		link.macroLinkID = macroLinkID
	}
}

// WithMacroNodeID sets parent macroscopic node identifier
func WithMacroNodeID(macroNodeID gmns.NodeID) func(*Link) {
	return func(link *Link) {
		link.macroNodeID = macroNodeID
	}
}

// WithCellType sets the cell type
func WithCellType(cellType types.CellType) func(*Link) {
	return func(link *Link) {
		link.cellType = cellType
	}
}

// WithLaneID sets the lane number
func WithLaneID(laneID int) func(*Link) {
	return func(link *Link) {
		link.laneID = laneID
	}
}

// WithIsFirstMovementCell marks the link as first movement cell
func WithIsFirstMovementCell(isFirst bool) func(*Link) {
	return func(link *Link) {
		link.isFirstMovementCell = isFirst
	}
}

// WithMovementCompositeType sets the movement composite type
func WithMovementCompositeType(mct movement.MovementCompositeType) func(*Link) {
	return func(link *Link) {
		link.movementCompositeType = mct
	}
}

// WithAdditionalTravelCost sets additional travel cost
func WithAdditionalTravelCost(cost float64) func(*Link) {
	return func(link *Link) {
		link.additionalTravelCost = cost
	}
}

// WithMesoLinkType sets the parent mesoscopic link type
func WithMesoLinkType(linkType types.LinkType) func(*Link) {
	return func(link *Link) {
		link.mesoLinkType = linkType
	}
}

// WithControlType sets the control type
func WithControlType(controlType types.ControlType) func(*Link) {
	return func(link *Link) {
		link.controlType = controlType
	}
}

// WithFreeSpeed stores the supplied free-flow speed in km/h.
func WithFreeSpeed(freeSpeed float64) func(*Link) {
	return func(link *Link) {
		value := freeSpeed
		link.freeSpeed = &value
	}
}

// WithCapacity stores the supplied saturation capacity in PCE/hour/lane.
func WithCapacity(capacity float64) func(*Link) {
	return func(link *Link) {
		value := capacity
		link.capacity = &value
	}
}

// WithAllowedAgentTypes sets allowed agent types (copies the slice)
func WithAllowedAgentTypes(allowedAgentTypes []types.AgentType) func(*Link) {
	return func(link *Link) {
		link.allowedAgentTypes = make([]types.AgentType, len(allowedAgentTypes))
		copy(link.allowedAgentTypes, allowedAgentTypes)
	}
}

// Name returns the supplied link name, or an empty string when not supplied.
func (link *Link) Name() string {
	return link.name
}

// WithLinkName sets the link name; an empty string clears the value.
func WithLinkName(name string) func(*Link) {
	return func(link *Link) {
		link.name = name
	}
}

// ParentLink returns the optional associated link ID within the same network.
// This reference is independent of source provenance across network levels.
func (link *Link) ParentLink() (gmns.LinkID, bool) {
	if link.parentLinkID == nil {
		return gmns.LinkID{}, false
	}
	return *link.parentLinkID, true
}

// WithParentLinkID sets an associated link reference within the same network.
// Zero and negative IDs are stored as supplied values, not as absence.
func WithParentLinkID(parentLinkID gmns.LinkID) func(*Link) {
	return func(link *Link) {
		value := parentLinkID
		link.parentLinkID = &value
	}
}

// WithGeometryID sets a reference to a shared shape without changing Geom().
func WithGeometryID(geometryID gmns.GeometryID) func(*Link) {
	return func(link *Link) {
		value := geometryID
		link.geometryID = &value
	}
}

// WithDirFlag sets GMNS shape orientation without reversing Geom(). For a
// shared canonical shape opposite directed links normally use 1 and -1.
// Existing per-link geometries already follow travel order and use 1.
func WithDirFlag(dirFlag gmns.DirectionFlag) func(*Link) {
	return func(link *Link) {
		value := dirFlag
		link.dirFlag = &value
	}
}

// Directed reports whether travel is restricted to source-to-target direction.
func (link *Link) Directed() bool {
	return link.directed
}

// WithDirected sets whether the link is directed. NewLinkFrom defaults to true.
// It does not change geometry, endpoints or source-road bidirectionality.
func WithDirected(directed bool) func(*Link) {
	return func(link *Link) {
		link.directed = directed
	}
}

// RightOfWayWidthMeters returns the supplied width of the entire right-of-way in meters.
// The width covers both directions and is independent of individual lane widths.
func (link *Link) RightOfWayWidthMeters() (float64, bool) {
	if link.rightOfWayWidthMeters == nil {
		return 0, false
	}
	return *link.rightOfWayWidthMeters, true
}

// WithRightOfWayWidthMeters stores the entire right-of-way width without deriving it from lanes.
func WithRightOfWayWidthMeters(rightOfWayWidthMeters float64) func(*Link) {
	return func(link *Link) {
		value := rightOfWayWidthMeters
		link.rightOfWayWidthMeters = &value
	}
}

// Grade returns the supplied grade in percent and whether it was supplied.
// Negative values describe downhill travel.
func (link *Link) Grade() (float64, bool) {
	if link.gradePercent == nil {
		return 0, false
	}
	return *link.gradePercent, true
}

// WithGrade stores a grade in percent without calculating it from geometry.
func WithGrade(gradePercent float64) func(*Link) {
	return func(link *Link) {
		value := gradePercent
		link.gradePercent = &value
	}
}

// Jurisdiction returns the supplied owner/operator, or an empty string when absent.
func (link *Link) Jurisdiction() string {
	return link.jurisdiction
}

// WithJurisdiction sets the owner/operator; an empty string clears the value.
func WithJurisdiction(jurisdiction string) func(*Link) {
	return func(link *Link) {
		link.jurisdiction = jurisdiction
	}
}

// BikeFacility returns the supplied bicycle infrastructure, or an empty string when absent.
func (link *Link) BikeFacility() string {
	return link.bikeFacility
}

// WithBikeFacility stores the bicycle infrastructure; an empty string clears the value.
// Suggested GMNS values are available in gmns/types; custom strings are supported.
func WithBikeFacility(bikeFacility string) func(*Link) {
	return func(link *Link) {
		link.bikeFacility = bikeFacility
	}
}

// PedFacility returns the supplied pedestrian infrastructure, or an empty string when absent.
func (link *Link) PedFacility() string {
	return link.pedFacility
}

// WithPedFacility stores the pedestrian infrastructure; an empty string clears the value.
// Suggested GMNS values are available in gmns/types; custom strings are supported.
func WithPedFacility(pedFacility string) func(*Link) {
	return func(link *Link) {
		link.pedFacility = pedFacility
	}
}

// Parking returns the supplied curbside parking type, or an empty string when absent.
func (link *Link) Parking() string {
	return link.parking
}

// WithParking stores the curbside parking type; an empty string clears the value.
// Suggested GMNS values are available in gmns/types; custom strings are supported.
func WithParking(parking string) func(*Link) {
	return func(link *Link) {
		link.parking = parking
	}
}

// Toll returns the supplied amount in the dataset's currency units and its presence.
func (link *Link) Toll() (float64, bool) {
	if link.toll == nil {
		return 0, false
	}
	return *link.toll, true
}

// WithToll stores an amount without currency conversion or automatic calculation.
func WithToll(toll float64) func(*Link) {
	return func(link *Link) {
		value := toll
		link.toll = &value
	}
}

// FacilityType returns the supplied facility classification, or empty if absent.
func (link *Link) FacilityType() string {
	return link.facilityType
}

// WithFacilityType stores or clears a classification independently of OSM link types.
func WithFacilityType(facilityType string) func(*Link) {
	return func(link *Link) {
		link.facilityType = facilityType
	}
}

// AllowedUses returns the supplied comma-separated use references, or empty if absent.
func (link *Link) AllowedUses() string {
	return link.allowedUses
}

// WithAllowedUses stores or clears use references without deriving agent permissions.
// Definition lookup and syntax validation belong to dataset validation.
func WithAllowedUses(allowedUses string) func(*Link) {
	return func(link *Link) {
		link.allowedUses = allowedUses
	}
}

// LanesNum returns the supplied motor-vehicle lane count and its presence.
// The count is independent of the generated cell's LaneID.
func (link *Link) LanesNum() (int, bool) {
	if link.lanesNum == nil {
		return 0, false
	}
	return *link.lanesNum, true
}

// WithLanesNum stores a supplied count without inferring it from a cell's lane index.
func WithLanesNum(lanesNum int) func(*Link) {
	return func(link *Link) {
		value := lanesNum
		link.lanesNum = &value
	}
}

// WithoutToll clears the supplied value, restoring absence.
func WithoutToll() func(*Link) {
	return func(link *Link) {
		link.toll = nil
	}
}

// WithoutLengthMeters clears the supplied value, restoring absence.
func WithoutLengthMeters() func(*Link) {
	return func(link *Link) {
		link.lengthMeters = nil
	}
}

// WithoutFreeSpeed clears the supplied value, restoring absence.
func WithoutFreeSpeed() func(*Link) {
	return func(link *Link) {
		link.freeSpeed = nil
	}
}

// WithoutCapacity clears the supplied value, restoring absence.
func WithoutCapacity() func(*Link) {
	return func(link *Link) {
		link.capacity = nil
	}
}

// WithoutLanesNum clears the supplied value, restoring absence.
func WithoutLanesNum() func(*Link) {
	return func(link *Link) {
		link.lanesNum = nil
	}
}

// WithoutGrade clears the supplied value, restoring absence.
func WithoutGrade() func(*Link) {
	return func(link *Link) {
		link.gradePercent = nil
	}
}

// WithoutRightOfWayWidthMeters clears the supplied value, restoring absence.
func WithoutRightOfWayWidthMeters() func(*Link) {
	return func(link *Link) {
		link.rightOfWayWidthMeters = nil
	}
}

// WithoutGeometryID clears the supplied value, restoring absence.
func WithoutGeometryID() func(*Link) {
	return func(link *Link) {
		link.geometryID = nil
	}
}

// WithoutDirFlag clears the supplied value, restoring absence.
func WithoutDirFlag() func(*Link) {
	return func(link *Link) {
		link.dirFlag = nil
	}
}

// WithoutParentLinkID clears the supplied value, restoring absence.
func WithoutParentLinkID() func(*Link) {
	return func(link *Link) {
		link.parentLinkID = nil
	}
}
