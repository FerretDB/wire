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

// Package wiretest provides testing helpers.
package wiretest

import (
	"fmt"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/wire/wirebson"
)

// diff returns a readable form of given values and the difference between them.
func diff(tb testing.TB, expected, actual any) (expectedS, actualS, diff string) {
	tb.Helper()

	expectedS = wirebson.LogMessageIndent(expected)
	actualS = wirebson.LogMessageIndent(actual)

	var err error
	diff, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expectedS),
		FromFile: "expected",
		B:        difflib.SplitLines(actualS),
		ToFile:   "actual",
		Context:  1,
	})
	require.NoError(tb, err)

	return
}

// diffSlices returns a readable form of given slices and the difference between them.
func diffSlices[T any](tb testing.TB, expected, actual []T) (expectedS, actualS, diff string) {
	tb.Helper()

	expectedA := wirebson.MakeArray(len(expected))
	for _, v := range expected {
		require.NoError(tb, expectedA.Add(v))
	}

	actualA := wirebson.MakeArray(len(actual))
	for _, v := range actual {
		require.NoError(tb, actualA.Add(v))
	}

	expectedS = wirebson.LogMessageIndent(expectedA)
	actualS = wirebson.LogMessageIndent(actualA)

	var err error
	diff, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expectedS),
		FromFile: "expected",
		B:        difflib.SplitLines(actualS),
		ToFile:   "actual",
		Context:  1,
	})
	require.NoError(tb, err)

	return
}

// AssertEqual asserts that two BSON values are equal.
func AssertEqual(tb testing.TB, expected, actual any) bool {
	tb.Helper()

	if wirebson.Equal(expected, actual) {
		return true
	}

	expectedS, actualS, diff := diff(tb, expected, actual)

	msg := fmt.Sprintf("Not equal:\n\nexpected:\n%s\n\nactual:\n%s\n\ndiff:\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}

// AssertEqualSlices asserts that two BSON slices are equal.
func AssertEqualSlices[T any](tb testing.TB, expected, actual []T) bool {
	tb.Helper()

	allEqual := len(expected) == len(actual)
	if allEqual {
		for i, e := range expected {
			a := actual[i]
			if !wirebson.Equal(e, a) {
				allEqual = false
				break
			}
		}
	}

	if allEqual {
		return true
	}

	expectedS, actualS, diff := diffSlices(tb, expected, actual)

	msg := fmt.Sprintf("Not equal:\n\nexpected:\n%s\n\nactual:\n%s\n\ndiff:\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}

// FromDriver converts MongoDB driver v2 (and, temporary, v1) value ([bson.D], [bson.A], etc) to wirebson value.
func FromDriver(tb testing.TB, v any) any {
	tb.Helper()

	res, err := wirebson.FromDriver(v)
	require.NoError(tb, err)
	return res
}

// ToDriver converts wirebson value to MongoDB driver v2 value (bson.D, bson.A, etc).
func ToDriver(tb testing.TB, v any) any {
	tb.Helper()

	res, err := wirebson.ToDriver(v)
	require.NoError(tb, err)
	return res
}
