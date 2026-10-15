package movement

var (
	movementsShortTypes  = []string{"undefined", "T", "R", "L", "U"}
	directionTypes       = []string{"undefined", "SB", "EB", "NB", "WB"}
	movementTextIDsMatch = map[string]MovementCompositeType{
		"SBT": MOVEMENT_SBT,
		"SBR": MOVEMENT_SBR,
		"SBL": MOVEMENT_SBL,
		"SBU": MOVEMENT_SBU,
		"EBT": MOVEMENT_EBT,
		"EBR": MOVEMENT_EBR,
		"EBL": MOVEMENT_EBL,
		"EBU": MOVEMENT_EBU,
		"NBT": MOVEMENT_NBT,
		"NBR": MOVEMENT_NBR,
		"NBL": MOVEMENT_NBL,
		"NBU": MOVEMENT_NBU,
		"WBT": MOVEMENT_WBT,
		"WBR": MOVEMENT_WBR,
		"WBL": MOVEMENT_WBL,
		"WBU": MOVEMENT_WBU,
	}
)

// MovementType stores a supplied movement category. An empty value means absent.
type MovementType string

const (
	MOVEMENT_TYPE_UNDEFINED MovementType = "undefined"
	MOVEMENT_TYPE_THRU      MovementType = "thru"
	MOVEMENT_TYPE_RIGHT     MovementType = "right"
	MOVEMENT_TYPE_LEFT      MovementType = "left"
	MOVEMENT_TYPE_U_TURN    MovementType = "uturn"
	MOVEMENT_TYPE_MERGE     MovementType = "merge"
	MOVEMENT_TYPE_DIVERGE   MovementType = "diverge"
)

func (movementType MovementType) String() string {
	return string(movementType)
}

// MovementShortType is just type alias for the movement type in shorten form
type MovementShortType uint16

const (
	MOVEMENT_SHORT_TYPE_UNDEFINED = MovementShortType(iota)
	MOVEMENT_SHORT_TYPE_THRU
	MOVEMENT_SHORT_TYPE_RIGHT
	MOVEMENT_SHORT_TYPE_LEFT
	MOVEMENT_SHORT_TYPE_U_TURN
)

func (iotaIdx MovementShortType) String() string {
	return movementsShortTypes[iotaIdx]
}

// DirectionType is just type alias for the direction of a movement (it does distinct from DirectionType in package "macro")
type DirectionType uint16

const (
	DIRECTION_TYPE_UNDEFINED = DirectionType(iota)
	DIRECTION_TYPE_SB
	DIRECTION_TYPE_EB
	DIRECTION_TYPE_NB
	DIRECTION_TYPE_WB
)

func (iotaIdx DirectionType) String() string {
	return directionTypes[iotaIdx]
}

// MovementCompositeType stores a supplied movement code. An empty value means absent.
// Constants also retain legacy generator codes; storage does not validate code syntax.
type MovementCompositeType string

const (
	MOVEMENT_UNDEFINED MovementCompositeType = "undefined"
	MOVEMENT_SBT       MovementCompositeType = "SBT"
	MOVEMENT_SBR       MovementCompositeType = "SBR"
	MOVEMENT_SBL       MovementCompositeType = "SBL"
	MOVEMENT_SBU       MovementCompositeType = "SBU"
	MOVEMENT_EBT       MovementCompositeType = "EBT"
	MOVEMENT_EBR       MovementCompositeType = "EBR"
	MOVEMENT_EBL       MovementCompositeType = "EBL"
	MOVEMENT_EBU       MovementCompositeType = "EBU"
	MOVEMENT_NBT       MovementCompositeType = "NBT"
	MOVEMENT_NBR       MovementCompositeType = "NBR"
	MOVEMENT_NBL       MovementCompositeType = "NBL"
	MOVEMENT_NBU       MovementCompositeType = "NBU"
	MOVEMENT_WBT       MovementCompositeType = "WBT"
	MOVEMENT_WBR       MovementCompositeType = "WBR"
	MOVEMENT_WBL       MovementCompositeType = "WBL"
	MOVEMENT_WBU       MovementCompositeType = "WBU"
)

func (movementType MovementCompositeType) String() string {
	return string(movementType)
}
