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

package wireclient

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logWriter provides [io.Writer] for [testing.TB].
type logWriter struct {
	tb testing.TB
}

// Write implements [io.Writer].
func (lw *logWriter) Write(p []byte) (int, error) {
	// "logging.go:xx" is added by testing.TB.Log itself; there is nothing we can do about it.
	// lw.tb.Helper() does not help. See:
	// https://github.com/golang/go/issues/59928
	// https://github.com/neilotoole/slogt/tree/v1.1.0?tab=readme-ov-file#deficiency

	// handle the most common escape sequences for request/response bodies
	s := strings.TrimSpace(string(p))
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\"`, `"`)

	lw.tb.Log(s)
	return len(p), nil
}

// logger returns slog test logger.
func logger(tb testing.TB) *slog.Logger {
	h := slog.NewTextHandler(&logWriter{tb: tb}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return slog.New(h)
}

func TestConn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests for -short")
	}

	uri := os.Getenv("MONGODB_URI")
	require.NotEmpty(t, uri, "MONGODB_URI environment variable must be set; set it or run tests with `go test -short`")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	t.Run("Login", func(t *testing.T) {
		t.Run("InvalidUsername", func(t *testing.T) {
			conn := ConnectPing(ctx, uri, logger(t))
			require.NotNil(t, conn)

			t.Cleanup(func() {
				require.NoError(t, conn.Close())
			})

			assert.Error(t, conn.Login(ctx, "invalid", "invalid", "admin"))
		})

		t.Run("Valid", func(t *testing.T) {
			conn := ConnectPing(ctx, uri, logger(t))
			require.NotNil(t, conn)

			t.Cleanup(func() {
				require.NoError(t, conn.Close())
			})

			assert.NoError(t, conn.Login(ctx, "username", "password", "admin"))
		})
	})
}
