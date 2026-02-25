package scheduler_func

import (
	"sort"
	"sync"
	"time"

	"github.com/gaohao-creator/turbopool/errors"
)

type WorkersStackWithFunc struct {
	data []WorkerWithFunc
	size int
	lock *sync.Mutex
}

// Get stack length.
func (s *WorkersStackWithFunc) Len() int {
	return len(s.data)
}

// Get stack is empty.
func (s *WorkersStackWithFunc) IsEmpty() bool {
	return s.Len() == 0
}

// Push a worker to stack.
func (s *WorkersStackWithFunc) Push(w WorkerWithFunc) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.Len() < s.size {
		s.data = append(s.data, w)
		return nil
	}
	return errors.ErrorsWorkerStackFull
}

// Get a worker and remove it.
func (s *WorkersStackWithFunc) Pop() (WorkerWithFunc, error) {
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
func (s *WorkersStackWithFunc) Clear() error {
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
func (s *WorkersStackWithFunc) ClearExpired(t time.Time) (int, error) {
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
func (s *WorkersStackWithFunc) Scale(cap int32) error {
	return nil
}

// Create a workers, serving as a worker container
func NewWorkersStackWithFunc(size int) (WorkersWithFunc, error) {
	s := WorkersStackWithFunc{
		data: make([]WorkerWithFunc, 0, size),
		size: size,
		lock: &sync.Mutex{},
	}
	return &s, nil
}
