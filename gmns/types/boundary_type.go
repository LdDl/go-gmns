package types

// BoundaryType is just type alias for the boundary type
type BoundaryType uint16

const (
	BOUNDARY_NONE = BoundaryType(iota)
	BOUNDARY_INCOME_ONLY
	BOUNDARY_OUTCOME_ONLY
	BOUNDARY_INCOME_OUTCOME
)

var boundaryTypeStr = []string{"none", "income_only", "outcome_only", "income_outcome"}

func (iotaIdx BoundaryType) String() string {
	return boundaryTypeStr[iotaIdx]
}
