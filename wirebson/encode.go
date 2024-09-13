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
// It panics if v is not a valid type.
func encodeField(buf []byte, name string, v any) (int, error) {
	var i int
	switch v := v.(type) {
	case *Document:
		buf[i] = byte(tagDocument)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		i += write(buf[i:], b)

		size := sizeDocument(v)
		b = make([]byte, size)

		err := v.Encode(b)
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		i += write(buf[i:], b)

	case RawDocument:

		buf[i] = byte(tagDocument)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		i += write(buf[i:], b)

		i += write(buf[i:], v)

	case *Array:
		buf[i] = byte(tagArray)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		i += write(buf[i:], b)

		b, err := v.Encode()
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		i += write(buf[i:], b)

	case RawArray:
		buf[i] = byte(tagArray)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		i += write(buf[i:], b)

		i += write(buf[i:], v)

	default:
		written, err := encodeScalarField(buf[i:], name, v)
		return i + written, err
	}

	return i, nil
}

// returns number of bytes written
func write(b []byte, v []byte) int {
	if len(v) > len(b) {
		panic("write impossible")
	}

	copy(b, v)
	return len(v)
}

// encodeScalarField encodes scalar document field.
//
// It panics if v is not a scalar value.
func encodeScalarField(b []byte, name string, v any) (int, error) {
	var i int
	switch v := v.(type) {
	case float64:
		b[i] = byte(tagFloat64)
	case string:
		b[i] = byte(tagString)
	case Binary:
		b[i] = byte(tagBinary)
	case ObjectID:
		b[i] = byte(tagObjectID)
	case bool:
		b[i] = byte(tagBool)
	case time.Time:
		b[i] = byte(tagTime)
	case NullType:
		b[i] = byte(tagNull)
	case Regex:
		b[i] = byte(tagRegex)
	case int32:
		b[i] = byte(tagInt32)
	case Timestamp:
		b[i] = byte(tagTimestamp)
	case int64:
		b[i] = byte(tagInt64)
	case Decimal128:
		b[i] = byte(tagDecimal128)
	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
	i++

	EncodeCString(b[i:], name)
	i += SizeCString(name)

	encodeScalarValue(b[i:], v)
	i += sizeScalar(v)

	return i, nil
}

// encodeScalarValue encodes value v into b.
//
// b must be at least Size(v) bytes long; otherwise, encodeScalarValue will panic.
// Only b[0:Size(v)] bytes are modified.
//
// It panics if v is not a [ScalarType] (including CString).
func encodeScalarValue(b []byte, v any) {
	switch v := v.(type) {
	case float64:
		encodeFloat64(b, v)
	case string:
		encodeString(b, v)
	case Binary:
		encodeBinary(b, v)
	case ObjectID:
		encodeObjectID(b, v)
	case bool:
		encodeBool(b, v)
	case time.Time:
		encodeTime(b, v)
	case NullType:
		// nothing
	case Regex:
		encodeRegex(b, v)
	case int32:
		encodeInt32(b, v)
	case Timestamp:
		encodeTimestamp(b, v)
	case int64:
		encodeInt64(b, v)
	case Decimal128:
		encodeDecimal128(b, v)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}
