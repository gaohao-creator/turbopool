package scheduler_generic

import (
	"sort"
	"sync"
	"time"

	"github.com/gaohao-creator/turbopool/errors"
)

type WorkersStack[T any] struct {
	data []Worker[T]
	size int
	lock *sync.Mutex
}

// Get stack length.
func (s *WorkersStack[T]) Len() int {
	return len(s.data)
}

// Get stack is empty.
func (s *WorkersStack[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Push a worker to stack.
func (s *WorkersStack[T]) Push(w Worker[T]) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.Len() < s.size {
		s.data = append(s.data, w)
		return nil
	}
	return errors.ErrorsWorkerStackFull
}

// Get a worker and remove it.
func (s *WorkersStack[T]) Pop() (Worker[T], error) {
	s.lock.Lock()
	if s.IsEmpty() {
		s.lock.Unlock()
		return nil, errors.ErrorWorkersIsEmpty
	}
	i := len(s.data) - 1
	item := s.data[i]
	s.data[i] = nil
	s.data = s.data[:i]
	s.lock.Unlock()
	return item, nil
}

// Clear all worker.
func (s *WorkersStack[T]) Clear() error {
	s.lock.Lock()
	if s.IsEmpty() {
		s.lock.Unlock()
		return nil
	}
	for i := 0; i < s.Len(); i++ {
		w := s.data[i]
		s.data[i] = nil
		w.Finish()
	}
	s.data = nil // free old array, and no need make new slice.
	s.lock.Unlock()
	return nil
}

// Clear expired worker.
func (s *WorkersStack[T]) ClearExpired(t time.Time) (int, error) {
	s.lock.Lock()
	if s.IsEmpty() {
		s.lock.Unlock()
		return 0, nil
	}

	// Binary search first not expired index.
	index := sort.Search(len(s.data), func(i int) bool {
		return t.Before(s.data[i].GetUsedTime())
	})
	if index < 0 {
		s.lock.Unlock()
		return 0, nil
	}

	// Finish worker and free memory.
	for i := 0; i < index; i++ {
		w := s.data[i]
		s.data[i] = nil
		w.Finish()
	}

	// Move not expired worker to index 0
	j := copy(s.data, s.data[index:])
	s.data = s.data[:j] // slice data reset len.
	s.lock.Unlock()
	return j, nil
}

// Scale capacity.
func (s *WorkersStack[T]) Scale(cap int32) error {
	return nil
}

// Create a workers, serving as a worker container
func NewWorkersStack[T any](size int) (Workers[T], error) {
	s := WorkersStack[T]{
		data: make([]Worker[T], 0, size),
		size: size,
		lock: &sync.Mutex{},
	}
	return &s, nil
}
