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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentials(t *testing.T) {
	t.Parallel()

	cleanURI, userinfo, authSource, authMechanism, err := Credentials(
		"mongodb://username:password@localhost:27017/test?authMechanism=PLAIN&authSource=$external",
	)
	require.NoError(t, err)
	assert.Equal(t, "mongodb://localhost:27017/", cleanURI)
	assert.Equal(t, "username:password", userinfo.String())
	assert.Equal(t, "$external", authSource)
	assert.Equal(t, "PLAIN", authMechanism)

	cleanURI, userinfo, authSource, authMechanism, err = Credentials("mongodb://localhost:27017/test")
	require.NoError(t, err)
	assert.Equal(t, "mongodb://localhost:27017/", cleanURI)
	assert.Equal(t, "", userinfo.String())
	assert.Equal(t, "test", authSource)
	assert.Equal(t, "", authMechanism)
}

func TestLookupSrvURI(t *testing.T) {
	t.Parallel()

	u, err := url.Parse("mongodb+srv://username:password@cts-vcore.mongocluster.cosmos.azure.com/database")
	require.NoError(t, err)

	err = lookupSrvURI(t.Context(), u)
	require.NoError(t, err)
	assert.Equal(t, "mongodb://username:password@fc-f6de9018d614-000.mongocluster.cosmos.azure.com:10260/database", u.String())
}
