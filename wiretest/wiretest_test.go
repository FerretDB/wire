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

package wiretest

import (
	"math"
	"testing"

	_ "github.com/FerretDB/wire/internal/util/testutil"
)

func TestEqual(t *testing.T) {
	t.Parallel()

	AssertEqualSlices(t,
		[]any{0.0, math.Copysign(0, +1), math.NaN()},
		[]any{0.0, math.Copysign(0, +1), math.Float64frombits(0x7ff8000f000f0001)},
	)
}
