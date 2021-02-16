package main

import (
	"errors"
	"sync"
)

// Slice type that can be safely shared between goroutines
type ConcurrentSlice struct {
	sync.RWMutex
	items []interface{}
}

// Concurrent slice item
type ConcurrentSliceItem struct {
	Index int
	Value interface{}
}

// Append an item to the concurrent slice
func (cs *ConcurrentSlice) append(item interface{}) {
	cs.Lock()
	defer cs.Unlock()

	cs.items = append(cs.items, item)
}

// Remove the last item from the current slice
func (cs *ConcurrentSlice) remove() (interface{}, error) {
	cs.Lock()
	defer cs.Unlock()
	if len(cs.items) > 0 {
		tr := cs.items[len(cs.items)-1]
		cs.items = cs.items[:len(cs.items)-1]
		return tr, nil
	}
	return nil, errors.New("Empty slice")
}

// Iterate over the items in the concurrent slice
func (cs *ConcurrentSlice) iter() <-chan ConcurrentSliceItem {
	c := make(chan ConcurrentSliceItem)

	f := func() {
		cs.Lock()
		defer cs.Unlock()
		for index, value := range cs.items {
			c <- ConcurrentSliceItem{index, value}
		}
		close(c)
	}
	go f()

	return c
}
