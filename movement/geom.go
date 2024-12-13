package movement

import (
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

const (
	indentationThreshold = 8.0
)

// FindMovementType extracts movement description. Possible descriptions are northbound left (NBL), northbound through (NBT), westbound left (WBL), westbound through (WBT), southbound left (SBL), southbound through (SBT), eastbound left (EBL) and eastbound through (EBT) movements;
// ibLine - The line with coordinates in EPSG:3857. Line represents source of movement;
// obLine - The line with coordinates in EPSG:3857. Line represents target of movement;
// Returns one of corresponding values: NBL, NBT, NBR, NBU, SBL, SBT, SBR, SBU, EBL, EBT, EBR, EBU, WBL, WBT, WBR, WBU along with corresponding movement with possible values: thru, right, left, uturn;
// Warning: use it for Euclidean space only (or EPSG:3857).
func FindMovementType(ibLine orb.LineString, obLine orb.LineString) (MovementCompositeType, MovementType) {
	startIB, endIB := ibLine[0], ibLine[len(ibLine)-1]
	endOB := obLine[len(obLine)-1]

	var direction DirectionType

	angleIB := math.Atan2(endIB.Y()-startIB.Y(), endIB.X()-startIB.X())

	if -0.75*math.Pi <= angleIB && angleIB < -0.25*math.Pi {
		direction = DIRECTION_TYPE_SB
	} else if -0.25*math.Pi <= angleIB && angleIB < 0.25*math.Pi {
		direction = DIRECTION_TYPE_EB
	} else if 0.25*math.Pi <= angleIB && angleIB < 0.75*math.Pi {
		direction = DIRECTION_TYPE_NB
	} else {
		direction = DIRECTION_TYPE_WB
	}

	angleOB := math.Atan2(endOB.Y()-endIB.Y(), endOB.X()-endIB.X())

	angleDiff := angleOB - angleIB

	if angleDiff <= -1*math.Pi { // '<=' instead of '<' because of floating point number precision
		angleDiff += 2 * math.Pi
	}
	if angleDiff > math.Pi {
		angleDiff -= 2 * math.Pi
	}

	var movementShortType MovementShortType
	var movementType MovementType
	if -0.25*math.Pi <= angleDiff && angleDiff <= 0.25*math.Pi {
		movementShortType = MOVEMENT_SHORT_TYPE_THRU
		movementType = MOVEMENT_TYPE_THRU
	} else if angleDiff < -0.25*math.Pi {
		movementShortType = MOVEMENT_SHORT_TYPE_RIGHT
		movementType = MOVEMENT_TYPE_RIGHT
	} else if angleDiff <= 0.75*math.Pi {
		movementShortType = MOVEMENT_SHORT_TYPE_LEFT
		movementType = MOVEMENT_TYPE_LEFT
	} else {
		movementShortType = MOVEMENT_SHORT_TYPE_U_TURN
		movementType = MOVEMENT_TYPE_U_TURN
	}

	return movementTextIDsMatch[direction.String()+movementShortType.String()], movementType
}

// FindMovementGeom returns movement geometry for given lines pair;
// ibLine - The line represents source of movement;
// obLine - The line represents target of movement;
// Notice: panics if number of points in any line is less than 2.
func FindMovementGeom(ibLine orb.LineString, obLine orb.LineString) orb.LineString {
	indentIB := indentationThreshold
	lengthIB := geo.Length(ibLine)
	if lengthIB <= indentIB {
		indentIB = lengthIB / 2.0
	}
	pointIB, _ := geo.PointAtDistanceAlongLine(ibLine, lengthIB-indentIB) // Ident from link end

	indentOB := indentationThreshold
	lengthOB := geo.Length(obLine)
	if lengthOB <= indentOB {
		indentOB = lengthOB / 2.0
	}

	pointOB, _ := geo.PointAtDistanceAlongLine(obLine, indentOB)
	return orb.LineString{pointIB, pointOB}
}
