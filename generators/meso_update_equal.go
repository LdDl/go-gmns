package generators

import (
	"math"
	"reflect"
)

type mesoComparedPointers struct {
	first  uintptr
	second uintptr
	kind   reflect.Type
}

// equalMesoRecords also distinguishes signed zero when binding source and target records.
func equalMesoRecords(first, second any) bool {
	if !reflect.DeepEqual(first, second) {
		return false
	}
	visited := make(map[mesoComparedPointers]bool)
	return equalMesoFloatBits(reflect.ValueOf(first), reflect.ValueOf(second), visited)
}

func equalMesoFloatBits(first, second reflect.Value, visited map[mesoComparedPointers]bool) bool {
	if !first.IsValid() {
		return true
	}
	switch first.Kind() {
	case reflect.Float64:
		return math.Float64bits(first.Float()) == math.Float64bits(second.Float())
	case reflect.Float32:
		return math.Float32bits(float32(first.Float())) == math.Float32bits(float32(second.Float()))
	case reflect.Pointer:
		if first.IsNil() {
			return true
		}
		pointers := mesoComparedPointers{first: first.Pointer(), second: second.Pointer(), kind: first.Type()}
		if visited[pointers] {
			return true
		}
		visited[pointers] = true
		return equalMesoFloatBits(first.Elem(), second.Elem(), visited)
	case reflect.Interface:
		if first.IsNil() {
			return true
		}
		return equalMesoFloatBits(first.Elem(), second.Elem(), visited)
	case reflect.Struct:
		for index := 0; index < first.NumField(); index++ {
			if !equalMesoFloatBits(first.Field(index), second.Field(index), visited) {
				return false
			}
		}
	case reflect.Array, reflect.Slice:
		for index := 0; index < first.Len(); index++ {
			if !equalMesoFloatBits(first.Index(index), second.Index(index), visited) {
				return false
			}
		}
	case reflect.Map:
		iterator := first.MapRange()
		for iterator.Next() {
			if !equalMesoFloatBits(iterator.Value(), second.MapIndex(iterator.Key()), visited) {
				return false
			}
		}
	}
	return true
}
