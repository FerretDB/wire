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
	"bytes"
	"fmt"
	"time"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

// encodeField encodes document/array field.
//
// It panics if v is not a valid type.
func encodeField(buf *bytes.Buffer, name string, v any) error {
	switch v := v.(type) {
	case *Document:
		if err := buf.WriteByte(byte(tagDocument)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		b, err := v.Encode()
		if err != nil {
			return lazyerrors.Error(err)
		}

		if _, err = buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

	case RawDocument:
		if err := buf.WriteByte(byte(tagDocument)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		if _, err := buf.Write(v); err != nil {
			return lazyerrors.Error(err)
		}

	case *Array:
		if err := buf.WriteByte(byte(tagArray)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		b, err := v.Encode()
		if err != nil {
			return lazyerrors.Error(err)
		}

		if _, err = buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

	case RawArray:
		if err := buf.WriteByte(byte(tagArray)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, SizeCString(name))
		EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		if _, err := buf.Write(v); err != nil {
			return lazyerrors.Error(err)
		}

	default:
		return encodeScalarField(buf, name, v)
	}

	return nil
}

// encodeScalarField encodes scalar document field.
//
// It panics if v is not a scalar value.
func encodeScalarField(buf *bytes.Buffer, name string, v any) error {
	switch v := v.(type) {
	case float64:
		buf.WriteByte(byte(tagFloat64))
	case string:
		buf.WriteByte(byte(tagString))
	case Binary:
		buf.WriteByte(byte(tagBinary))
	case ObjectID:
		buf.WriteByte(byte(tagObjectID))
	case bool:
		buf.WriteByte(byte(tagBool))
	case time.Time:
		buf.WriteByte(byte(tagTime))
	case NullType:
		buf.WriteByte(byte(tagNull))
	case Regex:
		buf.WriteByte(byte(tagRegex))
	case int32:
		buf.WriteByte(byte(tagInt32))
	case Timestamp:
		buf.WriteByte(byte(tagTimestamp))
	case int64:
		buf.WriteByte(byte(tagInt64))
	case Decimal128:
		buf.WriteByte(byte(tagDecimal128))
	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}

	b := make([]byte, SizeCString(name))
	EncodeCString(b, name)

	if _, err := buf.Write(b); err != nil {
		return lazyerrors.Error(err)
	}

	b = make([]byte, sizeScalar(v))
	encodeScalarValue(b, v)

	if _, err := buf.Write(b); err != nil {
		return lazyerrors.Error(err)
	}

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
