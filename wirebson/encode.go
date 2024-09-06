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

func writeByte(d []byte, b byte) {
	// TODO handle overflow
	d[getIndex(d)] = b
}

func write(d []byte, b []byte) {
	// TODO handle overflow
	i := getIndex(d)
	copy(d[i:], b)
}

func getIndex(d []byte) int {
	// TODO handle overflow
	return cap(d) - len(d)
}

// encodeField encodes document/array field.
//
// It panics if v is not a valid type.
func encodeField(d []byte, name string, v any) error {
	switch v := v.(type) {
	case *Document:
		writeByte(d, byte(tagDocument))
		//if err := buf.WriteByte(byte(tagDocument)); err != nil {
		//	return lazyerrors.Error(err)
		//}

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(d, b)

		v.Encode(b)

		b = make([]byte, 0, Size(v))
		write(d, b)

		//if _, err := buf.Write(b); err != nil {
		//	return lazyerrors.Error(err)
		//}

		//b = make([]byte, Size(v))

		//if err := v.Encode(b); err != nil {
		//	return lazyerrors.Error(err)
		//}

		//if _, err := buf.Write(b); err != nil {
		//	return lazyerrors.Error(err)
		//}

	case RawDocument:
		writeByte(d, byte(tagDocument))

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(d, b)

		write(d, v)

	case *Array:
		writeByte(d, byte(tagArray))

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(d, b)

		b = make([]byte, 0, Size(v))

		err := v.Encode(b)
		if err != nil {
			return lazyerrors.Error(err)
		}

		write(d, b)

	case RawArray:
		writeByte(d, byte(tagArray))

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		write(d, b)

		write(d, v)

	default:
		return encodeScalarField(d, name, v)
	}

	return nil
}

// encodeScalarField encodes scalar document field.
//
// It panics if v is not a scalar value.
func encodeScalarField(d []byte, name string, v any) error {
	switch v := v.(type) {
	case float64:
		writeByte(d, byte(tagFloat64))
	case string:
		writeByte(d, byte(tagString))
	case Binary:
		writeByte(d, byte(tagBinary))
	case ObjectID:
		writeByte(d, byte(tagObjectID))
	case bool:
		writeByte(d, byte(tagBool))
	case time.Time:
		writeByte(d, byte(tagTime))
	case NullType:
		writeByte(d, byte(tagNull))
	case Regex:
		writeByte(d, byte(tagRegex))
	case int32:
		writeByte(d, byte(tagInt32))
	case Timestamp:
		writeByte(d, byte(tagTimestamp))
	case int64:
		writeByte(d, byte(tagInt64))
	case Decimal128:
		writeByte(d, byte(tagDecimal128))
	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}

	b := make([]byte, SizeCString(name))
	EncodeCString(b, name)

	write(d, b)

	b = make([]byte, sizeScalar(v))
	encodeScalarValue(b, v)

	write(d, b)

	return nil
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
