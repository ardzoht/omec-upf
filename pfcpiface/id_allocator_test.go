// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Intel Corporation
package pfcpiface

import (
	"fmt"
	"sync"
	"testing"
)

func TestAllocate(t *testing.T) {
	testCases := []struct {
		minValue uint32
		maxValue uint32
	}{
		{1, 200},
		{11, 5000},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("minValue: %d, maxValue: %d", testCase.minValue, testCase.maxValue), func(t *testing.T) {
			idGenerator := NewIDAllocator(testCase.minValue, testCase.maxValue)

			for i := testCase.minValue; i <= testCase.maxValue; i++ {
				id, err := idGenerator.Allocate()
				if id != i {
					t.Errorf("expected id: %d, output id: %d", i, id)
					t.FailNow()
				} else if err != nil {
					t.Error(err)
					t.FailNow()
				}
			}

			for i := testCase.minValue; i <= testCase.maxValue; i++ {
				idGenerator.Free(i)
			}
		})
	}
}

func TestAllocatorConcurrency(t *testing.T) {
	var usedMap sync.Map

	idGenerator := NewIDAllocator(1, 50000)

	wg := sync.WaitGroup{}
	for routineID := 1; routineID <= 10; routineID++ {
		wg.Add(1)
		go func(routineID int) {
			for i := 0; i < 1000; i++ {
				id, err := idGenerator.Allocate()
				if err != nil {
					t.Errorf("idGenerator.Allocate fail: %+v", err)
				}
				if value, ok := usedMap.Load(id); ok {
					t.Errorf("ID %d has been allocated at routine[%d], concurrent test failed", id, value)
				} else {
					usedMap.Store(id, routineID)
				}
			}
			usedMap.Range(func(key, value interface{}) bool {
				id := key.(uint32)
				idGenerator.Free(id)
				return true
			})
			wg.Done()
		}(routineID)
	}
	wg.Wait()
}

func TestTriggerNoSpaceToAllocateError(t *testing.T) {
	testCases := []struct {
		minValue uint32
		maxValue uint32
	}{
		{1, 10},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("minValue: %d, maxValue: %d", testCase.minValue, testCase.maxValue), func(t *testing.T) {
			valueRange := int(testCase.maxValue - testCase.minValue + 1)
			idGenerator := NewIDAllocator(testCase.minValue, testCase.maxValue)

			for i := 0; i < valueRange; i++ {
				_, err := idGenerator.Allocate()
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			}

			_, err := idGenerator.Allocate()
			if err == nil {
				t.Error("expect return error, but error is nil")
				t.FailNow()
			}
		})
	}
}
