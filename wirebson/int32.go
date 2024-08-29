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

// sizeInt32 is a size of the encoding of int32 in bytes.
const sizeInt32 = 4

// encodeInt32 encodes int32 value v into b.
//
// b must be at least 4 ([sizeInt32]) bytes long; otherwise, encodeInt32 will panic.
// Only b[0:4] bytes are modified.
func encodeInt32(b []byte, v int32) {
	binary.LittleEndian.PutUint32(b, uint32(v))
}

// decodeInt32 decodes int32 value from b.
//
// If there is not enough bytes, decodeInt32 will return a wrapped [ErrDecodeShortInput].
func decodeInt32(b []byte) (int32, error) {
	if len(b) < sizeInt32 {
		return 0, fmt.Errorf("DecodeInt32: expected at least %d bytes, got %d: %w", sizeInt32, len(b), ErrDecodeShortInput)
	}

	return int32(binary.LittleEndian.Uint32(b)), nil
}
