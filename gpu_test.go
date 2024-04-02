/*

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

func TestParseGPUMetrics(t *testing.T) {
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
	metrics := ParseGPUMetrics(string(data))
	t.Logf("%+v", metrics)

	assert.Equal(t, len(metrics), 3)
	expectedKeys := map[string]bool{"T4": true, "V100": true, "A100": true}
	for k := range metrics {
		if _, ok := expectedKeys[k]; !ok {
			t.Errorf("Unexpected key found in metrics: %s", k)
		}
		delete(expectedKeys, k)
	}
	if len(expectedKeys) != 0 {
		t.Errorf("Missing keys in map: %v", expectedKeys)
	}

	assert.Equal(t, metrics["T4"].alloc, uint64(17))
	assert.Equal(t, metrics["T4"].idle, uint64(9))
	assert.Equal(t, metrics["T4"].total, uint64(26))

	assert.Equal(t, metrics["V100"].alloc, uint64(21))
	assert.Equal(t, metrics["V100"].idle, uint64(43))
	assert.Equal(t, metrics["V100"].total, uint64(64))

	assert.Equal(t, metrics["A100"].alloc, uint64(14))
	assert.Equal(t, metrics["A100"].idle, uint64(9))
	assert.Equal(t, metrics["A100"].total, uint64(23))
}

func TestGPUGetMetrics(t *testing.T) {
	utility := "sinfo"
	if ! UtilityAvailable(utility) {
		t.Logf("%s is not available. Skipping...", utility)
	} else {
		t.Logf("%+v", GPUGetMetrics())
	}
}

