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

// ObjectID represents BSON scalar type ObjectID.
type ObjectID [12]byte

// SizeObjectID is a size of the encoding of [ObjectID] in bytes.
const SizeObjectID = 12

// EncodeObjectID encodes [ObjectID] value v into b.
//
// "b" must be at least 12 ([SizeObjectID]) bytes long; otherwise, EncodeObjectID will panic.
// Only b[0:12] bytes are modified.
func EncodeObjectID(b []byte, v ObjectID) {
	_ = b[11]
	copy(b, v[:])
}

// DecodeObjectID decodes [ObjectID] value from b.
//
// If there is not enough bytes, DecodeObjectID will return a wrapped [ErrDecodeShortInput].
func DecodeObjectID(b []byte) (ObjectID, error) {
	var res ObjectID

	if len(b) < SizeObjectID {
		return res, fmt.Errorf(
			"DecodeObjectID: expected at least %d bytes, got %d: %w",
			SizeObjectID, len(b), ErrDecodeShortInput,
		)
	}

	copy(res[:], b)

	return res, nil
}
