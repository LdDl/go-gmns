package micro

import "fmt"

var (
	ErrLinkNotFound = fmt.Errorf("link not found")
	ErrNodeNotFound = fmt.Errorf("node not found")
)
