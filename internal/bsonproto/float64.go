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

package bsonproto

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SizeFloat64 is a size of the encoding of float64 in bytes.
const SizeFloat64 = 8

// EncodeFloat64 encodes float64 value v into b.
//
// b must be at least 8 ([SizeFloat64]) bytes long; otherwise, EncodeFloat64 will panic.
// Only b[0:8] bytes are modified.
//
// Infinities, NaNs, negative zeros are preserved.
func EncodeFloat64(b []byte, v float64) {
	binary.LittleEndian.PutUint64(b, math.Float64bits(float64(v)))
}

// DecodeFloat64 decodes float64 value from b.
//
// If there is not enough bytes, DecodeFloat64 will return a wrapped [ErrDecodeShortInput].
//
// Infinities, NaNs, negative zeros are preserved.
func DecodeFloat64(b []byte) (float64, error) {
	if len(b) < SizeFloat64 {
		return 0, fmt.Errorf("DecodeFloat64: expected at least %d bytes, got %d: %w", SizeFloat64, len(b), ErrDecodeShortInput)
	}

	return math.Float64frombits(binary.LittleEndian.Uint64(b)), nil
}
