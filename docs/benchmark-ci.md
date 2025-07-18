# Benchmark CI

This document describes the benchmark CI setup for the FerretDB Wire repository.

## Overview

The repository includes automated benchmark running that:
1. Runs benchmarks using Go's built-in benchmark framework
2. Parses benchmark results using `golang.org/x/perf/benchfmt`
3. Converts results to BSON format
4. Pushes results to a MongoDB-compatible database for visualization in Grafana

## Components

### Internal Package: `internal/benchpusher`

This package provides:
- Benchmark result parsing using the modern `golang.org/x/perf/benchfmt` library
- MongoDB client for pushing structured benchmark data
- BSON document creation compatible with FerretDB

### CLI Tool: `cmd/benchpush`

Command-line tool that:
- Runs `go test -bench` with configurable parameters
- Parses the output and extracts benchmark metrics
- Optionally pushes results to MongoDB if URI is provided

Usage:
```bash
# Run benchmarks without pushing to database
go run ./cmd/benchpush -bench=BenchmarkDocumentDecode -count=1 -benchtime=100ms

# Run benchmarks and push to MongoDB
go run ./cmd/benchpush -bench=BenchmarkDocumentDecode -count=5 -benchtime=1s -uri="mongodb://..."
```

### Task Integration

Added `bench-ci` task to `Taskfile.yml`:
```yaml
bench-ci:
  desc: "Run benchmarks and push results to MongoDB (for CI)"
  cmds:
    - go run ./cmd/benchpush -bench='{{.BENCH}}' -count={{.BENCH_COUNT}} -benchtime={{.BENCH_TIME}} -uri='{{.BENCHMARK_MONGODB_URI}}'
```

### GitHub Actions Workflow

Added `benchmark` job to `.github/workflows/go.yml` that:
- Runs on main branch pushes, scheduled runs, and manual workflow dispatch
- Uses the `bench-ci` task to run benchmarks and push results
- Requires `BENCHMARK_MONGODB_URI` secret to be configured

## Configuration

### Environment Variables

- `BENCHMARK_MONGODB_URI`: MongoDB URI for pushing benchmark results
- `RUNNER_NAME`: GitHub Actions runner name (automatically set)
- `GITHUB_REPOSITORY`: Repository name (automatically set)

### Benchmark Parameters

Configurable via Taskfile.yml variables:
- `BENCH`: Benchmark regex pattern (default: `Benchmark.*`)
- `BENCH_TIME`: Benchmark duration (default: `1s`)
- `BENCH_COUNT`: Number of benchmark runs (default: `5`)

## Data Format

Benchmark results are stored in MongoDB with this structure:

```json
{
  "time": "2025-07-18T14:30:12.077Z",
  "env": {
    "runner": "runner-name",
    "hostname": "hostname",
    "repository": "FerretDB/wire"
  },
  "benchmarks": {
    "DocumentDecode_handshake1-4": {
      "iterations": 3148209,
      "ns_per_op": 381.6,
      "metrics": {
        "B/op": "352.00",
        "allocs/op": "10.00"
      }
    }
  }
}
```

## Manual Execution

To manually trigger benchmark runs:

1. Go to the repository's Actions tab
2. Select "Go" workflow
3. Click "Run workflow"
4. Check "Run benchmarks" option
5. Click "Run workflow"

## Local Testing

To test locally without MongoDB:
```bash
# Install task
go generate -x tools/tools.go

# Run short benchmarks
BENCHMARK_MONGODB_URI="" bin/task bench-ci
```

## Integration with Grafana

The benchmark results pushed to MongoDB can be visualized in Grafana by:
1. Configuring MongoDB as a data source
2. Creating dashboards that query the `benchmarks` collection
3. Displaying metrics like ns/op, memory allocations, and throughput over time