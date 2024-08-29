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

// Timestamp represents BSON scalar type timestamp.
type Timestamp uint64

// sizeTimestamp is a size of the encoding of [Timestamp] in bytes.
const sizeTimestamp = 8

// encodeTimestamp encodes [Timestamp] value v into b.
//
// b must be at least 8 ([sizeTimestamp]) bytes long; otherwise, encodeTimestamp will panic.
// Only b[0:8] bytes are modified.
func encodeTimestamp(b []byte, v Timestamp) {
	binary.LittleEndian.PutUint64(b, uint64(v))
}

// decodeTimestamp decodes [Timestamp] value from b.
//
// If there is not enough bytes, decodeTimestamp will return a wrapped [ErrDecodeShortInput].
func decodeTimestamp(b []byte) (Timestamp, error) {
	if len(b) < sizeTimestamp {
		return 0, fmt.Errorf("DecodeTimestamp: expected at least %d bytes, got %d: %w", sizeTimestamp, len(b), ErrDecodeShortInput)
	}

	return Timestamp(binary.LittleEndian.Uint64(b)), nil
}
