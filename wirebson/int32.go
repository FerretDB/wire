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

// SizeInt32 is a size of the encoding of int32 in bytes.
const SizeInt32 = 4

// EncodeInt32 encodes int32 value v into b.
//
// b must be at least 4 ([SizeInt32]) bytes long; otherwise, EncodeInt32 will panic.
// Only b[0:4] bytes are modified.
func EncodeInt32(b []byte, v int32) {
	binary.LittleEndian.PutUint32(b, uint32(v))
}

// DecodeInt32 decodes int32 value from b.
//
// If there is not enough bytes, DecodeInt32 will return a wrapped [ErrDecodeShortInput].
func DecodeInt32(b []byte) (int32, error) {
	if len(b) < SizeInt32 {
		return 0, fmt.Errorf("DecodeInt32: expected at least %d bytes, got %d: %w", SizeInt32, len(b), ErrDecodeShortInput)
	}

	return int32(binary.LittleEndian.Uint32(b)), nil
}
