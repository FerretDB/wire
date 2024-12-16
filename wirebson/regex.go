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

package wirebson

import (
	"fmt"
)

// Regex represents BSON scalar type regular expression.
type Regex struct {
	Pattern string
	Options string
}

// sizeRegex returns the size of the encoding of v [Regex] in bytes.
func sizeRegex(v Regex) int {
	return len(v.Pattern) + len(v.Options) + 2
}

// encodeRegex encodes [Regex] value v into b.
//
// b must be at least len(v.Pattern)+len(v.Options)+2 ([sizeRegex]) bytes long; otherwise, encodeRegex will panic.
// Only b[0:len(v.Pattern)+len(v.Options)+2] bytes are modified.
func encodeRegex(b []byte, v Regex) {
	// ensure b length early
	b[len(v.Pattern)+len(v.Options)+1] = 0

	copy(b, v.Pattern)
	b[len(v.Pattern)] = 0
	copy(b[len(v.Pattern)+1:], v.Options)
}

// decodeRegex decodes [Regex] value from b.
//
// If there is not enough bytes, decodeRegex will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func decodeRegex(b []byte) (Regex, error) {
	var res Regex

	if len(b) < 2 {
		return res, fmt.Errorf("DecodeRegex: expected at least 2 bytes, got %d: %w", len(b), ErrDecodeShortInput)
	}

	p, o := -1, -1
	for i, b := range b {
		if b == 0 {
			if p == -1 {
				p = i
			} else {
				o = i
				break
			}
		}
	}

	if o == -1 {
		return res, fmt.Errorf("DecodeRegex: expected two 0 bytes: %w", ErrDecodeShortInput)
	}

	res.Pattern = string(b[:p])
	res.Options = string(b[p+1 : o])

	return res, nil
}
