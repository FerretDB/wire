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
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/wire/internal/util/testutil"
)

func TestConn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests for -short")
	}

	uri := os.Getenv("MONGODB_URI")
	require.NotEmpty(t, uri, "MONGODB_URI environment variable must be set; set it or run tests with `go test -short`")

	ctx := context.Background()

	conn, err := Connect(ctx, uri, testutil.Logger(t))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close())
	})

	t.Run("SomeTest", func(t *testing.T) {
		// TODO https://github.com/FerretDB/wire/issues/1
	})
}
