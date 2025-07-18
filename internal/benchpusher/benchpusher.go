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

// Package benchpusher provides functionality to parse benchmark results and push them to MongoDB.
package benchpusher

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/perf/benchfmt"
)

// Client represents a MongoDB client for pushing benchmark results.
type Client struct {
	l            *slog.Logger
	c            *mongo.Client
	pingerCancel context.CancelFunc
	pingerDone   chan struct{}
	database     string
	hostname     string
	runner       string
	repository   string
}

// BenchmarkResult represents a parsed benchmark result.
type BenchmarkResult struct {
	Name       string            `bson:"name"`
	Iterations int               `bson:"iterations"`
	NsPerOp    float64           `bson:"ns_per_op"`
	Metrics    map[string]string `bson:"metrics"`
}

// New creates a new MongoDB client for pushing benchmark results.
func New(uri string, l *slog.Logger) (*Client, error) {
	if uri == "" {
		return nil, fmt.Errorf("MongoDB URI is required")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return nil, fmt.Errorf("database name is empty in the URL")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	opts := options.Client().ApplyURI(uri)
	opts.SetDirect(true)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetHeartbeatInterval(3 * time.Second)
	opts.SetMaxConnIdleTime(0)
	opts.SetMinPoolSize(1)
	opts.SetMaxPoolSize(1)
	opts.SetMaxConnecting(1)

	l.InfoContext(ctx, "Connecting to MongoDB URI to push benchmark results...", slog.String("uri", u.Redacted()))

	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	pingerCtx, pingerCancel := context.WithCancel(ctx)

	res := &Client{
		l:            l,
		c:            c,
		pingerCancel: pingerCancel,
		pingerDone:   make(chan struct{}),
		database:     database,
		hostname:     hostname,
		runner:       os.Getenv("RUNNER_NAME"),
		repository:   os.Getenv("GITHUB_REPOSITORY"),
	}

	go func() {
		res.ping(pingerCtx)
		close(res.pingerDone)
	}()

	return res, nil
}

// ping pings the database until connection is established or ctx is canceled.
func (c *Client) ping(ctx context.Context) {
	for ctx.Err() == nil {
		pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)

		err := c.c.Ping(pingCtx, nil)
		if err == nil {
			c.l.InfoContext(pingCtx, "Ping successful")
			pingCancel()
			return
		}

		c.l.WarnContext(pingCtx, "Ping failed", slog.String("error", err.Error()))

		// always wait, even if ping returns immediately
		<-pingCtx.Done()
		pingCancel()
	}
}

// ParseBenchmarkOutput parses benchmark output and returns structured results.
func ParseBenchmarkOutput(output string) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	reader := benchfmt.NewReader(strings.NewReader(output), "")
	for reader.Scan() {
		record := reader.Result()
		if record == nil {
			continue
		}

		// Only process Result records, not Config records
		result, ok := record.(*benchfmt.Result)
		if !ok {
			continue
		}

		benchResult := BenchmarkResult{
			Name:       string(result.Name),
			Iterations: result.Iters,
			Metrics:    make(map[string]string),
		}

		// Extract standard benchmark metrics
		for _, metric := range result.Values {
			switch metric.Unit {
			case "ns/op", "sec/op":
				if metric.Unit == "sec/op" {
					// Convert seconds to nanoseconds
					benchResult.NsPerOp = metric.Value * 1e9
				} else {
					benchResult.NsPerOp = metric.Value
				}
			case "B/op", "allocs/op", "MB/s":
				benchResult.Metrics[metric.Unit] = fmt.Sprintf("%.2f", metric.Value)
			default:
				benchResult.Metrics[metric.Unit] = fmt.Sprintf("%.2f", metric.Value)
			}
		}

		results = append(results, benchResult)
	}

	if err := reader.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse benchmark output: %w", err)
	}

	return results, nil
}

// ParseBenchmarkOutput parses benchmark output and returns structured results (method version).
func (c *Client) ParseBenchmarkOutput(output string) ([]BenchmarkResult, error) {
	return ParseBenchmarkOutput(output)
}

// Push pushes benchmark results to MongoDB.
func (c *Client) Push(ctx context.Context, results []BenchmarkResult) error {
	if len(results) == 0 {
		c.l.InfoContext(ctx, "No benchmark results to push")
		return nil
	}

	var benchmarks bson.D
	for _, result := range results {
		// Replace dots with underscores to make it compatible with FerretDB v1
		name := strings.ReplaceAll(result.Name, ".", "_")
		benchmarks = append(benchmarks, bson.E{Key: name, Value: bson.D{
			{"iterations", result.Iterations},
			{"ns_per_op", result.NsPerOp},
			{"metrics", result.Metrics},
		}})
	}

	doc := bson.D{
		{"time", time.Now()},
		{"env", bson.D{
			{"runner", c.runner},
			{"hostname", c.hostname},
			{"repository", c.repository},
		}},
		{"benchmarks", benchmarks},
	}

	c.l.InfoContext(ctx, "Pushing benchmark results to MongoDB...", slog.Int("count", len(results)))

	c.ping(ctx)

	_, err := c.c.Database(c.database).Collection("benchmarks").InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert benchmark results: %w", err)
	}

	return nil
}

// Close closes all connections.
func (c *Client) Close() {
	c.pingerCancel()
	<-c.pingerDone

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.c.Disconnect(ctx)
}