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
	"time"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

// encodeField encodes document/array field.
//
// It returns the number of bytes written.
// It panics if v is not a valid type.
func encodeField(dst []byte, name string, v any) (int, error) {
	var i int
	switch v := v.(type) {
	case *Document:
		dst[i] = byte(tagDocument)
		i++

		EncodeCString(dst[i:], name)
		i += SizeCString(name)

		err := v.EncodeTo(dst[i:])
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		i += sizeDocument(v)

	case RawDocument:
		dst[i] = byte(tagDocument)
		i++

		EncodeCString(dst[i:], name)
		i += SizeCString(name)

		if len(v) > len(dst[i:]) {
			panic(fmt.Sprintf("length of dst should be at least %d bytes, got %d", len(v), len(dst[i:])))
		}

		i += copy(dst[i:], v)

	case *Array:
		dst[i] = byte(tagArray)
		i++

		EncodeCString(dst[i:], name)
		i += SizeCString(name)

		err := v.EncodeTo(dst[i:])
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		i += sizeArray(v)

	case RawArray:
		dst[i] = byte(tagArray)
		i++

		EncodeCString(dst[i:], name)
		i += SizeCString(name)

		if len(v) > len(dst[i:]) {
			panic(fmt.Sprintf("length of dst should be at least %d bytes, got %d", len(v), len(dst[i:])))
		}

		i += copy(dst[i:], v)

	default:
		return i + encodeScalarField(dst[i:], name, v), nil
	}

	return i, nil
}

// encodeScalarField encodes scalar document field.
//
// It returns the number of bytes written.
// It panics if v is not a scalar value.
func encodeScalarField(dst []byte, name string, v any) int {
	var i int
	switch v := v.(type) {
	case float64:
		dst[i] = byte(tagFloat64)
	case string:
		dst[i] = byte(tagString)
	case Binary:
		dst[i] = byte(tagBinary)
	case ObjectID:
		dst[i] = byte(tagObjectID)
	case bool:
		dst[i] = byte(tagBool)
	case time.Time:
		dst[i] = byte(tagTime)
	case NullType:
		dst[i] = byte(tagNull)
	case Regex:
		dst[i] = byte(tagRegex)
	case int32:
		dst[i] = byte(tagInt32)
	case Timestamp:
		dst[i] = byte(tagTimestamp)
	case int64:
		dst[i] = byte(tagInt64)
	case Decimal128:
		dst[i] = byte(tagDecimal128)
	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
	i++

	EncodeCString(dst[i:], name)
	i += SizeCString(name)

	encodeScalarValue(dst[i:], v)
	i += sizeScalar(v)

	return i
}

// encodeScalarValue encodes value v into b.
//
// b must be at least Size(v) bytes long; otherwise, encodeScalarValue will panic.
// Only b[0:Size(v)] bytes are modified.
//
// It panics if v is not a [ScalarType] (including CString).
func encodeScalarValue(dst []byte, v any) {
	switch v := v.(type) {
	case float64:
		encodeFloat64(dst, v)
	case string:
		encodeString(dst, v)
	case Binary:
		encodeBinary(dst, v)
	case ObjectID:
		encodeObjectID(dst, v)
	case bool:
		encodeBool(dst, v)
	case time.Time:
		encodeTime(dst, v)
	case NullType:
		// nothing
	case Regex:
		encodeRegex(dst, v)
	case int32:
		encodeInt32(dst, v)
	case Timestamp:
		encodeTimestamp(dst, v)
	case int64:
		encodeInt64(dst, v)
	case Decimal128:
		encodeDecimal128(dst, v)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}
