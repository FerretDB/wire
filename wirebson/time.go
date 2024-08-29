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
	"time"
)

// sizeTime is a size of the encoding of [time.Time] in bytes.
const sizeTime = 8

// encodeTime encodes [time.Time] value v into b.
//
// b must be at least 8 ([sizeTime]) byte long; otherwise, encodeTime will panic.
// Only b[0:8] bytes are modified.
func encodeTime(b []byte, v time.Time) {
	binary.LittleEndian.PutUint64(b, uint64(v.UnixMilli()))
}

// decodeTime decodes [time.Time] value from b.
//
// If there is not enough bytes, decodeTime will return a wrapped [ErrDecodeShortInput].
func decodeTime(b []byte) (time.Time, error) {
	var res time.Time

	if len(b) < sizeTime {
		return res, fmt.Errorf("DecodeTime: expected at least %d bytes, got %d: %w", sizeTime, len(b), ErrDecodeShortInput)
	}

	res = time.UnixMilli(int64(binary.LittleEndian.Uint64(b))).UTC()

	return res, nil
}
