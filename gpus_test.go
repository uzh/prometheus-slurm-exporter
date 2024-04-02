/* Copyright 2017 Victor Penso, Matteo Dessalvi

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
	"io"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseGPUsMetrics(t *testing.T) {
	// Read the input data from a file
	path := "test_data/sinfo_gpus.txt"
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open test file %s: %v", path, err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read the test file %s: %v", path, err)
	}
	metrics := ParseGPUsMetrics(string(data))
	t.Logf("%+v", metrics)
	assert.Equal(t, metrics.alloc, uint64(52))
	assert.Equal(t, metrics.idle, uint64(61))
	assert.Equal(t, metrics.total, uint64(113))
	// Not sure what's the right way to check a float
	assert.Greater(t, metrics.utilization, 51.0 / 113.0)
	assert.Less(t, metrics.utilization, 53.0 / 113.0)
}

func TestGPUsGetMetrics(t *testing.T) {
	utility := "sinfo"
	if ! UtilityAvailable(utility) {
		t.Logf("%s is not available. Skipping...", utility)
	} else {
		t.Logf("%+v", GPUsGetMetrics())
	}
}
