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
	"errors"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

var (
	// ErrDecodeShortInput is returned wrapped by Decode functions if the input bytes slice is too short.
	ErrDecodeShortInput = errors.New("wirebson: short input")

	// ErrDecodeInvalidInput is returned wrapped by Decode functions if the input bytes slice is invalid.
	ErrDecodeInvalidInput = errors.New("wirebson: invalid input")
)

// decodeMode represents a mode for decoding BSON.
type decodeMode int

const (
	_ decodeMode = iota

	// DecodeShallow represents a mode in which only top-level fields/elements are decoded;
	// nested documents and arrays are converted to RawDocument and RawArray respectively,
	// using raw's subslices without copying.
	decodeShallow

	// DecodeDeep represents a mode in which nested documents and arrays are decoded recursively;
	// RawDocuments and RawArrays are never returned.
	decodeDeep
)

// FindRaw finds the first raw BSON document or array in b and returns its length l.
// It should start from the first byte of b.
// RawDocument(b[:l]) / RawArray(b[:l]) might not be valid. It is the caller's responsibility to check it.
//
// Use RawDocument(b) / RawArray(b) conversion instead if b contains exactly one document/array and no extra bytes.
func FindRaw(b []byte) (int, error) {
	bl := len(b)
	if bl < 5 {
		return 0, lazyerrors.Errorf("len(b) = %d: %w", bl, ErrDecodeShortInput)
	}

	dl := int(binary.LittleEndian.Uint32(b))
	if dl < 5 {
		return 0, lazyerrors.Errorf("dl = %d: %w", dl, ErrDecodeInvalidInput)
	}

	if bl < dl {
		return 0, lazyerrors.Errorf("len(b) = %d, dl = %d: %w", bl, dl, ErrDecodeShortInput)
	}

	if b[dl-1] != 0 {
		return 0, lazyerrors.Errorf("invalid last byte: %w", ErrDecodeInvalidInput)
	}

	return dl, nil
}

// decodeCheckOffset checks that b has enough bytes to decode size bytes starting from offset.
func decodeCheckOffset(b []byte, offset, size int) error {
	if l := len(b); l < offset+size {
		return lazyerrors.Errorf("len(b) = %d, offset = %d, size = %d: %w", l, offset, size, ErrDecodeShortInput)
	}

	return nil
}

func decodeScalarField(b []byte, t tag) (v any, size int, err error) {
	switch t {
	case tagDocument, tagArray:
		err = lazyerrors.Errorf("non-scalar tag: %s", t)

	case tagFloat64:
		var f float64
		f, err = decodeFloat64(b)
		v = f
		size = sizeFloat64

	case tagString:
		var s string
		s, err = decodeString(b)
		v = s
		size = sizeString(s)

	case tagBinary:
		var bin Binary
		bin, err = decodeBinary(b)
		v = bin
		size = sizeBinary(bin)

	case tagUndefined:
		v = Undefined

	case tagObjectID:
		v, err = decodeObjectID(b)
		size = sizeObjectID

	case tagBool:
		v, err = decodeBool(b)
		size = sizeBool

	case tagTime:
		v, err = decodeTime(b)
		size = sizeTime

	case tagNull:
		v = Null

	case tagRegex:
		var re Regex
		re, err = decodeRegex(b)
		v = re
		size = sizeRegex(re)

	case tagDBPointer, tagJavaScript, tagSymbol, tagJavaScriptScope:
		err = lazyerrors.Errorf("unsupported tag %s: %w", t, ErrDecodeInvalidInput)

	case tagInt32:
		v, err = decodeInt32(b)
		size = sizeInt32

	case tagTimestamp:
		v, err = decodeTimestamp(b)
		size = sizeTimestamp

	case tagInt64:
		v, err = decodeInt64(b)
		size = sizeInt64

	case tagDecimal128:
		v, err = decodeDecimal128(b)
		size = sizeDecimal128

	case tagMinKey, tagMaxKey:
		err = lazyerrors.Errorf("unsupported tag %s: %w", t, ErrDecodeInvalidInput)

	default:
		err = lazyerrors.Errorf("unexpected tag %s: %w", t, ErrDecodeInvalidInput)
	}

	return
}
