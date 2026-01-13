package types

// CellType represents the type of microscopic link (cell)
type CellType uint8

const (
	// CELL_UNDEFINED is undefined cell type
	CELL_UNDEFINED CellType = iota
	// CELL_FORWARD is a forward traveling cell (normal movement within lane)
	CELL_FORWARD
	// CELL_LANE_CHANGE is a lane changing cell (movement between lanes)
	CELL_LANE_CHANGE
)

var cellTypeStr = [...]string{"undefined", "forward", "lane_change"}

// String returns the string representation of CellType
func (iotaIdx CellType) String() string {
	return cellTypeStr[iotaIdx]
}
