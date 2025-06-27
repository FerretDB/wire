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
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wiretest"
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

// setup waits for FerretDB or MongoDB to start and returns the URI.
func setup(tb testing.TB) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping integration tests for -short")
	}

	uri := os.Getenv("MONGODB_URI")
	require.NotEmpty(tb, uri, "MONGODB_URI environment variable must be set; set it or run tests with `go test -short`")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn := ConnectPing(ctx, uri, logger(tb))
	require.NotNil(tb, conn)

	err := conn.Close()
	require.NoError(tb, err)

	return uri
}

// databaseName returns the database name for the test.
func databaseName(tb testing.TB) string {
	tb.Helper()

	return strings.ReplaceAll(tb.Name(), "/", "_")
}

func TestLookupSrvURI(t *testing.T) {
	t.Parallel()

	u, err := url.Parse("mongodb+srv://username:password@cts-vcore.mongocluster.cosmos.azure.com/database")
	require.NoError(t, err)

	err = lookupSrvURI(t.Context(), u)
	require.NoError(t, err)
	assert.Equal(t, "mongodb://username:password@fc-f6de9018d614-000.mongocluster.cosmos.azure.com:10260/database", u.String())
}

func TestConn(t *testing.T) {
	t.Parallel()

	uri := setup(t)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
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

			u, err := url.Parse(uri)
			require.NoError(t, err)

			username, password := "username", "password"

			if u.User.Username() != "" {
				username = u.User.Username()
			}

			if pass, ok := u.User.Password(); ok {
				password = pass
			}

			assert.NoError(t, conn.Login(ctx, username, password, ""))
		})
	})
}

func TestTypes(t *testing.T) {
	t.Parallel()

	uri := setup(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	var conn *Conn
	var mConn *mongo.Client

	// avoid shadowing err in subtests
	{
		var err error

		conn = ConnectPing(ctx, uri, logger(t))
		require.NotNil(t, conn)

		err = conn.Login(ctx, "username", "password", "admin")
		require.NoError(t, err)

		opts := options.Client().ApplyURI(uri).SetAuth(options.Credential{Username: "username", Password: "password"})
		mConn, err = mongo.Connect(opts)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, conn.Close())
			require.NoError(t, mConn.Disconnect(ctx))
		})
	}

	t.Run("Decimal128", func(t *testing.T) {
		d := wirebson.Decimal128{H: 13, L: 42}
		md := bson.NewDecimal128(13, 42)

		db := mConn.Database(databaseName(t))
		require.NoError(t, db.Drop(ctx))

		_, body, err := conn.Request(ctx, wire.MustOpMsg(
			"insert", "test",
			"documents", wirebson.MustArray(wirebson.MustDocument("_id", "d", "v", d)),
			"$db", db.Name(),
		))
		require.NoError(t, err)

		doc, err := body.(*wire.OpMsg).DocumentDeep()
		require.NoError(t, err)
		require.Equal(t, 1.0, doc.Get("ok"))

		_, err = db.Collection("test").InsertOne(ctx, bson.D{{"_id", "md"}, {"v", md}})
		require.NoError(t, err)

		_, body, err = conn.Request(ctx, wire.MustOpMsg(
			"find", "test",
			"sort", wirebson.MustDocument("_id", int32(1)),
			"$db", db.Name(),
		))
		require.NoError(t, err)

		doc, err = body.(*wire.OpMsg).DocumentDeep()
		require.NoError(t, err)
		require.Equal(t, 1.0, doc.Get("ok"))

		expected := wirebson.MustArray(
			wirebson.MustDocument("_id", "d", "v", d),
			wirebson.MustDocument("_id", "md", "v", d),
		)
		wiretest.AssertEqual(t, expected, doc.Get("cursor").(*wirebson.Document).Get("firstBatch"))

		c, err := db.Collection("test").Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
		require.NoError(t, err)

		var res bson.A
		err = c.All(ctx, &res)
		require.NoError(t, err)

		mExpected := bson.A{
			bson.D{{"_id", "d"}, {"v", md}},
			bson.D{{"_id", "md"}, {"v", md}},
		}
		assert.Equal(t, mExpected, res)
	})

	t.Run("Timestamp", func(t *testing.T) {
		ts := wirebson.NewTimestamp(100, 500)
		mts := bson.Timestamp{T: 100, I: 500}

		require.EqualValues(t, 100, ts.T())
		require.EqualValues(t, 500, ts.I())

		db := mConn.Database(databaseName(t))
		require.NoError(t, db.Drop(ctx))

		_, body, err := conn.Request(ctx, wire.MustOpMsg(
			"insert", "test",
			"documents", wirebson.MustArray(wirebson.MustDocument("_id", "ts", "v", ts)),
			"$db", db.Name(),
		))
		require.NoError(t, err)

		doc, err := body.(*wire.OpMsg).DocumentDeep()
		require.NoError(t, err)
		require.Equal(t, 1.0, doc.Get("ok"))

		_, err = db.Collection("test").InsertOne(ctx, bson.D{{"_id", "mts"}, {"v", mts}})
		require.NoError(t, err)

		_, body, err = conn.Request(ctx, wire.MustOpMsg(
			"find", "test",
			"sort", wirebson.MustDocument("_id", int32(-1)),
			"$db", db.Name(),
		))
		require.NoError(t, err)

		doc, err = body.(*wire.OpMsg).DocumentDeep()
		require.NoError(t, err)
		require.Equal(t, 1.0, doc.Get("ok"))

		expected := wirebson.MustArray(
			wirebson.MustDocument("_id", "ts", "v", ts),
			wirebson.MustDocument("_id", "mts", "v", ts),
		)
		wiretest.AssertEqual(t, expected, doc.Get("cursor").(*wirebson.Document).Get("firstBatch"))

		c, err := db.Collection("test").Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", -1}}))
		require.NoError(t, err)

		var res bson.A
		err = c.All(ctx, &res)
		require.NoError(t, err)

		mExpected := bson.A{
			bson.D{{"_id", "ts"}, {"v", mts}},
			bson.D{{"_id", "mts"}, {"v", mts}},
		}
		assert.Equal(t, mExpected, res)
	})
}
