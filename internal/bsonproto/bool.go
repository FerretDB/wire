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
	"fmt"
)

// SizeBool is a size of the encoding of bool in bytes.
const SizeBool = 1

// EncodeBool encodes bool value v into b.
//
// b must be at least 1 ([SizeBool]) byte long; otherwise, EncodeBool will panic.
// Only b[0] is modified.
func EncodeBool(b []byte, v bool) {
	if v {
		b[0] = 0x01
	} else {
		b[0] = 0x00
	}
}

// DecodeBool decodes bool value from b.
//
// If there is not enough bytes, DecodeBool will return a wrapped [ErrDecodeShortInput].
func DecodeBool(b []byte) (bool, error) {
	if len(b) == 0 {
		return false, fmt.Errorf("DecodeBool: expected at least 1 byte, got 0: %w", ErrDecodeShortInput)
	}

	switch b[0] {
	case 0x00:
		return false, nil
	case 0x01:
		return true, nil
	default:
		return false, fmt.Errorf("DecodeBool: expected 0x00 or 0x01, got 0x%02x: %w", b[0], ErrDecodeInvalidInput)
	}
}
