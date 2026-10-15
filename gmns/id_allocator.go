package gmns

import (
	"fmt"
	"math/big"
)

// ErrDuplicateID reports an identifier already reserved in the same namespace.
var ErrDuplicateID = fmt.Errorf("duplicate ID")

// IDAllocator allocates decimal identifiers in one resource namespace.
// Use separate instances for nodes, links and movements. The zero value allocates integers.
// It is not safe for concurrent use. Reserve existing IDs before allocating into a populated resource.
type IDAllocator struct {
	kind IDType
	next big.Int
	used map[string]struct{}
}

func NewIDAllocator(kind IDType) (*IDAllocator, error) {
	err := kind.Validate()
	if err != nil {
		return nil, err
	}
	return &IDAllocator{kind: kind}, nil
}

func (allocator *IDAllocator) idType() IDType {
	if allocator.kind == "" {
		return IDTypeInteger
	}
	return allocator.kind
}

func (allocator *IDAllocator) reserve(value string, kind IDType) error {
	err := validateID(value, kind)
	if err != nil {
		return err
	}
	if allocator.used == nil {
		allocator.used = make(map[string]struct{})
	}
	if _, exists := allocator.used[value]; exists {
		return fmt.Errorf("ID %q: %w", value, ErrDuplicateID)
	}
	allocator.used[value] = struct{}{}
	return nil
}

func (allocator *IDAllocator) peek() string {
	for {
		candidate := allocator.next.String()
		if _, exists := allocator.used[candidate]; !exists {
			return candidate
		}
		allocator.next.Add(&allocator.next, big.NewInt(1))
	}
}

func (allocator *IDAllocator) take() string {
	candidate := allocator.peek()
	if allocator.used == nil {
		allocator.used = make(map[string]struct{})
	}
	allocator.used[candidate] = struct{}{}
	allocator.next.Add(&allocator.next, big.NewInt(1))
	return candidate
}

// SetNextInteger sets the starting counter, without releasing reserved identifiers.
func (allocator *IDAllocator) SetNextInteger(value string) error {
	canonical, err := parseID(value, IDTypeInteger)
	if err != nil {
		return err
	}
	integer, _ := new(big.Int).SetString(canonical, 10)
	if integer.Sign() < 0 {
		return fmt.Errorf("allocation counter must be nonnegative")
	}
	allocator.next.Set(integer)
	return nil
}

func (allocator *IDAllocator) ReserveNodeID(identifier NodeID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextNodeID() NodeID {
	return NodeID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekNodeID() NodeID {
	return NodeID{value: allocator.peek(), kind: allocator.idType()}
}

func (allocator *IDAllocator) ReserveLinkID(identifier LinkID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextLinkID() LinkID {
	return LinkID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekLinkID() LinkID {
	return LinkID{value: allocator.peek(), kind: allocator.idType()}
}

func (allocator *IDAllocator) ReserveMovementID(identifier MovementID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextMovementID() MovementID {
	return MovementID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekMovementID() MovementID {
	return MovementID{value: allocator.peek(), kind: allocator.idType()}
}

func (allocator *IDAllocator) ReserveGeometryID(identifier GeometryID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextGeometryID() GeometryID {
	return GeometryID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekGeometryID() GeometryID {
	return GeometryID{value: allocator.peek(), kind: allocator.idType()}
}

func (allocator *IDAllocator) ReserveLaneID(identifier LaneID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextLaneID() LaneID {
	return LaneID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekLaneID() LaneID {
	return LaneID{value: allocator.peek(), kind: allocator.idType()}
}

func (allocator *IDAllocator) ReserveSegmentID(identifier SegmentID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextSegmentID() SegmentID {
	return SegmentID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekSegmentID() SegmentID {
	return SegmentID{value: allocator.peek(), kind: allocator.idType()}
}

func (allocator *IDAllocator) ReserveSegmentLaneID(identifier SegmentLaneID) error {
	return allocator.reserve(identifier.value, identifier.kind)
}

func (allocator *IDAllocator) NextSegmentLaneID() SegmentLaneID {
	return SegmentLaneID{value: allocator.take(), kind: allocator.idType()}
}

func (allocator *IDAllocator) PeekSegmentLaneID() SegmentLaneID {
	return SegmentLaneID{value: allocator.peek(), kind: allocator.idType()}
}
