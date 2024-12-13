package movement

var (
	movementsTypes       = []string{"undefined", "thru", "right", "left", "uturn"}
	movementsShortTypes  = []string{"undefined", "T", "R", "L", "U"}
	directionTypes       = []string{"undefined", "SB", "EB", "NB", "WB"}
	movementsTextIDs     = []string{"undefined", "SBT", "SBR", "SBL", "SBU", "EBT", "EBR", "EBL", "EBU", "NBT", "NBR", "NBL", "NBU", "WBT", "WBR", "WBL", "WBU"}
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

type MovementType uint16

const (
	MOVEMENT_TYPE_UNDEFINED = MovementType(iota)
	MOVEMENT_TYPE_THRU
	MOVEMENT_TYPE_RIGHT
	MOVEMENT_TYPE_LEFT
	MOVEMENT_TYPE_U_TURN
)

func (iotaIdx MovementType) String() string {
	return movementsTypes[iotaIdx]
}

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

type MovementCompositeType uint16

const (
	MOVEMENT_UNDEFINED = MovementCompositeType(iota)
	MOVEMENT_SBT
	MOVEMENT_SBR
	MOVEMENT_SBL
	MOVEMENT_SBU
	MOVEMENT_EBT
	MOVEMENT_EBR
	MOVEMENT_EBL
	MOVEMENT_EBU
	MOVEMENT_NBT
	MOVEMENT_NBR
	MOVEMENT_NBL
	MOVEMENT_NBU
	MOVEMENT_WBT
	MOVEMENT_WBR
	MOVEMENT_WBL
	MOVEMENT_WBU
)

func (iotaIdx MovementCompositeType) String() string {
	return movementsTextIDs[iotaIdx]
}
