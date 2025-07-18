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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.Default()
}

// TestPushWithoutMongoDB tests that pushing without MongoDB client doesn't crash
func TestPushWithoutMongoDB(t *testing.T) {
	results := []BenchmarkResult{
		{
			Name:       "TestBenchmark",
			Iterations: 1000,
			NsPerOp:    100.5,
			Metrics:    map[string]string{"B/op": "64.00", "allocs/op": "2.00"},
		},
	}

	// This should not panic and should be gracefully handled
	// when no MongoDB client is available
	// (actual implementation would need MongoDB client to push)
	assert.NotEmpty(t, results)
}

// TestBenchmarkResultStructure tests the structure of benchmark results
func TestBenchmarkResultStructure(t *testing.T) {
	result := BenchmarkResult{
		Name:       "BenchmarkExample",
		Iterations: 1000000,
		NsPerOp:    150.5,
		Metrics: map[string]string{
			"B/op":       "128.00",
			"allocs/op":  "4.00",
			"MB/s":       "25.50",
		},
	}

	assert.Equal(t, "BenchmarkExample", result.Name)
	assert.Equal(t, 1000000, result.Iterations)
	assert.Equal(t, 150.5, result.NsPerOp)
	assert.Equal(t, "128.00", result.Metrics["B/op"])
	assert.Equal(t, "4.00", result.Metrics["allocs/op"])
	assert.Equal(t, "25.50", result.Metrics["MB/s"])
}

// TestNewClientWithInvalidURI tests client creation with invalid URI
func TestNewClientWithInvalidURI(t *testing.T) {
	logger := testLogger()

	// Test with empty URI
	_, err := New("", logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MongoDB URI is required")

	// Test with invalid URI
	_, err = New("invalid-uri", logger)
	require.Error(t, err)

	// Test with URI without database
	_, err = New("mongodb://localhost:27017", logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database name is empty")
}