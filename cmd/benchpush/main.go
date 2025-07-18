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

// Package main provides a command-line tool for running benchmarks and pushing results to MongoDB.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/FerretDB/wire/internal/benchpusher"
)

func main() {
	var (
		mongoURI   = flag.String("uri", "", "MongoDB URI for pushing results (if empty, only parse and print)")
		benchRegex = flag.String("bench", "Benchmark.*", "Benchmark regex pattern")
		benchTime  = flag.String("benchtime", "1s", "Benchmark time")
		benchCount = flag.String("count", "5", "Benchmark count")
		pkg        = flag.String("pkg", "./wirebson", "Package to benchmark")
		timeout    = flag.Duration("timeout", 10*time.Minute, "Benchmark timeout")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Run the benchmarks
	logger.Info("Running benchmarks...", 
		slog.String("package", *pkg), 
		slog.String("pattern", *benchRegex),
		slog.String("benchtime", *benchTime),
		slog.String("count", *benchCount))

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", 
		"-bench="+*benchRegex, 
		"-count="+*benchCount, 
		"-benchtime="+*benchTime, 
		"-timeout=60m", 
		*pkg)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.Error("Benchmark command failed", 
				slog.String("error", err.Error()),
				slog.String("stderr", string(exitErr.Stderr)))
		} else {
			logger.Error("Failed to run benchmark command", slog.String("error", err.Error()))
		}
		os.Exit(1)
	}

	outputStr := string(output)
	logger.Info("Benchmark completed", slog.Int("output_length", len(outputStr)))

	// Parse the benchmark output
	var client *benchpusher.Client
	if *mongoURI != "" {
		var err error
		client, err = benchpusher.New(*mongoURI, logger)
		if err != nil {
			logger.Error("Failed to create MongoDB client", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer client.Close()
	}

	results, err := benchpusher.ParseBenchmarkOutput(outputStr)
	if err != nil {
		logger.Error("Failed to parse benchmark output", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Parsed benchmark results", slog.Int("count", len(results)))

	// Print results summary
	for _, result := range results {
		logger.Info("Benchmark result",
			slog.String("name", result.Name),
			slog.Int("iterations", result.Iterations),
			slog.Float64("ns_per_op", result.NsPerOp),
			slog.Any("metrics", result.Metrics))
	}

	// Push to MongoDB if URI is provided
	if *mongoURI != "" && len(results) > 0 {
		logger.Info("Pushing results to MongoDB...")
		if err := client.Push(context.Background(), results); err != nil {
			logger.Error("Failed to push results to MongoDB", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("Successfully pushed results to MongoDB")
	} else if *mongoURI == "" {
		logger.Info("No MongoDB URI provided, skipping push to database")
	} else {
		logger.Info("No benchmark results to push")
	}
}