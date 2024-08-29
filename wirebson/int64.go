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
	"encoding/binary"
	"fmt"
)

// sizeInt64 is a size of the encoding of int64 in bytes.
const sizeInt64 = 8

// encodeInt64 encodes int64 value v into b.
//
// b must be at least 8 ([sizeInt64]) bytes long; otherwise, encodeInt64 will panic.
// Only b[0:8] bytes are modified.
func encodeInt64(b []byte, v int64) {
	binary.LittleEndian.PutUint64(b, uint64(v))
}

// decodeInt64 decodes int64 value from b.
//
// If there is not enough bytes, decodeInt64 will return a wrapped [ErrDecodeShortInput].
func decodeInt64(b []byte) (int64, error) {
	if len(b) < sizeInt64 {
		return 0, fmt.Errorf("DecodeInt64: expected at least %d bytes, got %d: %w", sizeInt64, len(b), ErrDecodeShortInput)
	}

	return int64(binary.LittleEndian.Uint64(b)), nil
}
