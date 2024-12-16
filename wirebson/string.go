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

// sizeString returns the size of the encoding of v string in bytes.
func sizeString(v string) int {
	return len(v) + 5
}

// encodeString encodes string value v into b.
//
// b must be at least len(v)+5 ([sizeString]) bytes long; otherwise, encodeString will panic.
// Only b[0:len(v)+5] bytes are modified.
func encodeString(b []byte, v string) {
	i := len(v) + 1

	// ensure b length early
	b[4+i-1] = 0

	binary.LittleEndian.PutUint32(b, uint32(i))
	copy(b[4:4+i-1], v)
}

// decodeString decodes string value from b.
//
// If there is not enough bytes, decodeString will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func decodeString(b []byte) (string, error) {
	if len(b) < 5 {
		return "", fmt.Errorf("DecodeString: expected at least 5 bytes, got %d: %w", len(b), ErrDecodeShortInput)
	}

	i := int(binary.LittleEndian.Uint32(b))
	if i < 1 {
		return "", fmt.Errorf("DecodeString: expected the prefix to be at least 1, got %d: %w", i, ErrDecodeInvalidInput)
	}
	if e := 4 + i; len(b) < e {
		return "", fmt.Errorf("DecodeString: expected at least %d bytes, got %d: %w", e, len(b), ErrDecodeShortInput)
	}
	if b[4+i-1] != 0 {
		return "", fmt.Errorf("DecodeString: expected the last byte to be 0: %w", ErrDecodeInvalidInput)
	}

	return string(b[4 : 4+i-1]), nil
}
