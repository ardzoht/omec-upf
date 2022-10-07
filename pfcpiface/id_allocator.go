// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Intel Corporation
package pfcpiface

import (
	"errors"
	"sync"
)

type IDAllocator struct {
	lock       sync.Mutex
	minValue   uint32
	maxValue   uint32
	valueRange uint32
	id         uint32
	usedMap    map[uint32]bool
}

// NewIDAllocator with minValue and maxValue.
func NewIDAllocator(minValue, maxValue uint32) *IDAllocator {
	idAllocator := &IDAllocator{}
	idAllocator.init(minValue, maxValue)
	return idAllocator
}

func (idAllocator *IDAllocator) init(minValue, maxValue uint32) {
	idAllocator.id = minValue
	idAllocator.minValue = minValue
	idAllocator.maxValue = maxValue
	idAllocator.valueRange = maxValue - minValue + 1
	idAllocator.usedMap = make(map[uint32]bool)
}

// Allocate and return an id in range [minValue, maxValue]
func (idAllocator *IDAllocator) Allocate() (uint32, error) {
	idAllocator.lock.Lock()
	defer idAllocator.lock.Unlock()

	id := idAllocator.id
	if idAllocator.id == idAllocator.maxValue+1 {
		return 0, errors.New("no available value range to allocate id")
	}
	idAllocator.usedMap[idAllocator.id] = true
	idAllocator.id++
	return id, nil
}

// Free releases an already allocated ID
func (idAllocator *IDAllocator) Free(id uint32) {
	if id < idAllocator.minValue || id > idAllocator.maxValue {
		return
	}
	idAllocator.lock.Lock()
	defer idAllocator.lock.Unlock()
	delete(idAllocator.usedMap, id-idAllocator.minValue)
}
