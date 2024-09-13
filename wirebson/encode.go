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
func encodeField(i int, buf []byte, name string, v any) (int, error) {
	switch v := v.(type) {
	case *Document:
		writeByte(buf, byte(tagDocument), i)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(buf, b, i)
		i += len(b)

		size := sizeDocument(v)
		b = make([]byte, size)

		err := v.Encode(b)
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		write(buf, b, i)
		i += len(b)

	case RawDocument:
		writeByte(buf, byte(tagDocument), i)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(buf, b, i)
		i += len(b)

		write(buf, v, i)
		i += len(b)

	case *Array:
		writeByte(buf, byte(tagArray), i)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(buf, b, i)
		i += len(b)

		b, err := v.Encode()
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		write(buf, b, i)
		i += len(b)

	case RawArray:
		writeByte(buf, byte(tagArray), i)
		i++

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(buf, b, i)
		i += len(b)

		write(buf, v, i)
		i += len(b)

	default:
		written, err := encodeScalarField(buf[i:], name, v)
		return i + written, err
	}

	return i, nil
}

func writeByte(b []byte, v byte, offset int) {
	b[offset] = v
}

// returns number of bytes written
func write(b []byte, v []byte, offset int) int {
	copy(b[offset:], v)
	return len(v)
}

// encodeScalarField encodes scalar document field.
//
// It panics if v is not a scalar value.
func encodeScalarField(b []byte, name string, v any) (int, error) {
	var i int
	switch v := v.(type) {
	case float64:
		writeByte(b, byte(tagFloat64), i)
	case string:
		writeByte(b, byte(tagString), i)
	case Binary:
		writeByte(b, byte(tagBinary), i)
	case ObjectID:
		writeByte(b, byte(tagObjectID), i)
	case bool:
		writeByte(b, byte(tagBool), i)
	case time.Time:
		writeByte(b, byte(tagTime), i)
	case NullType:
		writeByte(b, byte(tagNull), i)
	case Regex:
		writeByte(b, byte(tagRegex), i)
	case int32:
		writeByte(b, byte(tagInt32), i)
	case Timestamp:
		writeByte(b, byte(tagTimestamp), i)
	case int64:
		writeByte(b, byte(tagInt64), i)
	case Decimal128:
		writeByte(b, byte(tagDecimal128), i)
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
