package gmns

// GeometryID identifies a record in the optional GMNS geometry table.
type GeometryID int

// DirectionFlag describes shape orientation relative to a link's endpoints.
// It is independent of macro.DirectionType used when constructing OSM links.
type DirectionFlag int8

const (
	DIR_FLAG_REVERSE DirectionFlag = -1
	DIR_FLAG_NONE    DirectionFlag = 0
	DIR_FLAG_FORWARD DirectionFlag = 1
)
