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

// Package wireclient provides low-level wire protocol client.
package wireclient

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Credentials extracts user credentials, authSource, and authMechanism suitable for [Conn.Login]
// from the given MongoDB URI.
// It also returns a clean URI suitable for [Connect].
//
// If both authSource query parameter and URI path are present, the query parameter takes precedence.
// If both are empty, it does not defaults to "admin".
// The caller should handle this case if needed.
func Credentials(uri string) (cleanURI string, credentials *url.Userinfo, authSource, authMechanism string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}

	credentials = u.User
	u.User = nil

	q := u.Query()

	if q.Has("authMechanism") {
		v := q["authMechanism"]
		if l := len(v); l != 1 {
			err = fmt.Errorf("%q: expected 1 authMechanism, got %d", uri, l)
			return
		}

		authMechanism = v[0]
		q.Del("authMechanism")
	}

	if q.Has("authSource") {
		v := q["authSource"]
		if l := len(v); l != 1 {
			err = fmt.Errorf("%q: expected 1 authSource, got %d", uri, l)
			return
		}

		authSource = v[0]
		q.Del("authSource")
	}

	if authSource == "" {
		authSource = strings.TrimPrefix(u.Path, "/")
	}
	u.Path = "/"

	u.RawQuery = q.Encode()
	cleanURI = u.String()
	return
}

// lookupSrvURI converts mongodb+srv:// URI to mongodb:// URI, performing the simplest SRV lookup.
func lookupSrvURI(ctx context.Context, u *url.URL) error {
	_, srvs, err := net.DefaultResolver.LookupSRV(ctx, "mongodb", "tcp", u.Hostname())
	if err != nil {
		return fmt.Errorf("lookupSrvURI: SRV lookup failed: %w", err)
	}

	if len(srvs) != 1 {
		return fmt.Errorf("lookupSrvURI: expected exactly one SRV record, got %d", len(srvs))
	}

	srv := srvs[0]
	u.Host = net.JoinHostPort(strings.TrimSuffix(srv.Target, "."), strconv.Itoa(int(srv.Port)))
	u.Scheme = "mongodb"

	return nil
}

// sleep waits until the given duration is over or the context is canceled.
func sleep(ctx context.Context, d time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	<-ctx.Done()
}
