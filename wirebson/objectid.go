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

// ObjectID represents BSON scalar type ObjectID.
type ObjectID [12]byte

// sizeObjectID is a size of the encoding of [ObjectID] in bytes.
const sizeObjectID = 12

// encodeObjectID encodes [ObjectID] value v into b.
//
// b must be at least 12 ([sizeObjectID]) bytes long; otherwise, encodeObjectID will panic.
// Only b[0:12] bytes are modified.
func encodeObjectID(b []byte, v ObjectID) {
	_ = b[11]
	copy(b, v[:])
}

// decodeObjectID decodes [ObjectID] value from b.
//
// If there is not enough bytes, decodeObjectID will return a wrapped [ErrDecodeShortInput].
func decodeObjectID(b []byte) (ObjectID, error) {
	var res ObjectID

	if len(b) < sizeObjectID {
		return res, fmt.Errorf("DecodeObjectID: expected at least %d bytes, got %d: %w", sizeObjectID, len(b), ErrDecodeShortInput)
	}

	copy(res[:], b)

	return res, nil
}
