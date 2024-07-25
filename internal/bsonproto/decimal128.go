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
)

// Decimal128 represents BSON scalar type decimal128.
type Decimal128 struct {
	L uint64
	H uint64
}

// SizeDecimal128 is a size of the encoding of [Decimal128] in bytes.
const SizeDecimal128 = 16

// EncodeDecimal128 encodes [Decimal128] value v into b.
//
// b must be at least 16 ([SizeDecimal128]) bytes long; otherwise, EncodeDecimal128 will panic.
// Only b[0:16] bytes are modified.
func EncodeDecimal128(b []byte, v Decimal128) {
	binary.LittleEndian.PutUint64(b, uint64(v.L))
	binary.LittleEndian.PutUint64(b[8:], uint64(v.H))
}

// DecodeDecimal128 decodes [Decimal128] value from b.
//
// If there is not enough bytes, DecodeDecimal128 will return a wrapped [ErrDecodeShortInput].
func DecodeDecimal128(b []byte) (Decimal128, error) {
	var res Decimal128

	if len(b) < SizeDecimal128 {
		return res, fmt.Errorf("DecodeDecimal128: expected at least %d bytes, got %d: %w", SizeDecimal128, len(b), ErrDecodeShortInput)
	}

	res.L = binary.LittleEndian.Uint64(b[:8])
	res.H = binary.LittleEndian.Uint64(b[8:])

	return res, nil
}
