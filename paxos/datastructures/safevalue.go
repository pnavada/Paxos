package datastructures

import (
	"sync"
)

type SafeValue[T any] struct {
	value T
	lock  sync.Mutex
}

func (sv *SafeValue[T]) Get() T {
	sv.lock.Lock()
	defer sv.lock.Unlock()
	return sv.value
}

func (sv *SafeValue[T]) Set(value T) {
	sv.lock.Lock()
	defer sv.lock.Unlock()
	sv.value = value
}
func NewSafeValue[T any](initialValue T) *SafeValue[T] {
	return &SafeValue[T]{value: initialValue}
}
