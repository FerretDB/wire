// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package benchpusher

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBenchmarkOutput(t *testing.T) {
	// Sample benchmark output that matches the format from wire benchmarks
	sampleOutput := `goos: linux
goarch: amd64
pkg: github.com/FerretDB/wire/wirebson
cpu: AMD EPYC 7763 64-Core Processor                
BenchmarkDocumentDecode/handshake1-4         	 3148209	       381.6 ns/op	     352 B/op	      10 allocs/op
BenchmarkDocumentDecode/handshake2-4         	 3153933	       381.2 ns/op	     352 B/op	      10 allocs/op
BenchmarkDocumentDecode/nested-4             	10174455	       116.6 ns/op	      88 B/op	       3 allocs/op
PASS
ok  	github.com/FerretDB/wire/wirebson	65.233s`

	logger := slog.Default()
	client := &Client{l: logger}

	results, err := client.ParseBenchmarkOutput(sampleOutput)
	require.NoError(t, err)

	// Debug: print what we got
	t.Logf("Parsed %d results", len(results))
	for i, result := range results {
		t.Logf("Result %d: Name=%s, Iterations=%d, NsPerOp=%f, Metrics=%v", 
			i, result.Name, result.Iterations, result.NsPerOp, result.Metrics)
	}

	// Should parse the benchmark lines
	assert.Greater(t, len(results), 0, "Should parse at least one benchmark result")

	// Check that we can find specific benchmarks
	var handshake1Found, nestedFound bool
	for _, result := range results {
		if strings.Contains(result.Name, "handshake1") {
			handshake1Found = true
			assert.Greater(t, result.NsPerOp, 0.0, "Should have positive ns/op")
			assert.Greater(t, result.Iterations, 0, "Should have positive iterations")
			assert.Contains(t, result.Metrics, "B/op", "Should have B/op metric")
			assert.Contains(t, result.Metrics, "allocs/op", "Should have allocs/op metric")
		}
		if strings.Contains(result.Name, "nested") {
			nestedFound = true
		}
	}

	assert.True(t, handshake1Found, "Should find handshake1 benchmark")
	assert.True(t, nestedFound, "Should find nested benchmark")
}

func TestParseBenchmarkOutput_Empty(t *testing.T) {
	logger := slog.Default()
	client := &Client{l: logger}

	results, err := client.ParseBenchmarkOutput("")
	require.NoError(t, err)
	assert.Empty(t, results, "Empty input should return empty results")
}

func TestParseBenchmarkOutput_InvalidFormat(t *testing.T) {
	logger := slog.Default()
	client := &Client{l: logger}

	// Test with non-benchmark output
	invalidOutput := `This is not benchmark output
Just some random text
No benchmarks here`

	results, err := client.ParseBenchmarkOutput(invalidOutput)
	require.NoError(t, err)
	assert.Empty(t, results, "Invalid format should return empty results")
}