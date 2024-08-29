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

// Package wirebson implements encoding and decoding of BSON as defined by https://bsonspec.org/spec.html.
//
// # Types
//
// The following BSON types are supported:
//
//	BSON                Go
//
//	Document/Object     *Document or RawDocument
//	Array               *Array    or RawArray
//
//	Double              float64
//	String              string
//	Binary data         Binary
//	ObjectId            ObjectID
//	Boolean             bool
//	Date                time.Time
//	Null                NullType
//	Regular Expression  Regex
//	32-bit integer      int32
//	Timestamp           Timestamp
//	64-bit integer      int64
//	Decimal128          Decimal128
//
// Composite types (Document and Array) are passed by pointers.
// Raw composite type and scalars are passed by values.
package wirebson

import (
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

//go:generate ../bin/stringer -linecomment -output stringers.go -type decodeMode,tag,BinarySubtype

// ScalarType represents a BSON scalar type.
//
// CString is not included as it is not a real BSON type.
type ScalarType interface {
	float64 | string | Binary | ObjectID | bool | time.Time | NullType | Regex | int32 | Timestamp | int64 | Decimal128
}

// Size returns a size of the encoding of value v in bytes.
func Size[T ScalarType](v T) int {
	return SizeAny(v)
}

// SizeAny returns a size of the encoding of value v in bytes.
//
// It panics if v is not a [ScalarType] (including CString).
// FIXME
func SizeAny(v any) int {
	switch v := v.(type) {
	case float64:
		return SizeFloat64
	case string:
		return SizeString(v)
	case Binary:
		return SizeBinary(v)
	case ObjectID:
		return SizeObjectID
	case bool:
		return SizeBool
	case time.Time:
		return SizeTime
	case NullType:
		return 0
	case Regex:
		return SizeRegex(v)
	case int32:
		return SizeInt32
	case Timestamp:
		return SizeTimestamp
	case int64:
		return SizeInt64
	case Decimal128:
		return SizeDecimal128
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

// Encode encodes value v into b.
//
// b must be at least Size(v) bytes long; otherwise, Encode will panic.
// Only b[0:Size(v)] bytes are modified.
func Encode[T ScalarType](b []byte, v T) {
	EncodeAny(b, v)
}

// EncodeAny encodes value v into b.
//
// b must be at least Size(v) bytes long; otherwise, EncodeAny will panic.
// Only b[0:Size(v)] bytes are modified.
//
// It panics if v is not a [ScalarType] (including CString).
func EncodeAny(b []byte, v any) {
	switch v := v.(type) {
	case float64:
		EncodeFloat64(b, v)
	case string:
		EncodeString(b, v)
	case Binary:
		EncodeBinary(b, v)
	case ObjectID:
		EncodeObjectID(b, v)
	case bool:
		EncodeBool(b, v)
	case time.Time:
		EncodeTime(b, v)
	case NullType:
		// nothing
	case Regex:
		EncodeRegex(b, v)
	case int32:
		EncodeInt32(b, v)
	case Timestamp:
		EncodeTimestamp(b, v)
	case int64:
		EncodeInt64(b, v)
	case Decimal128:
		EncodeDecimal128(b, v)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

// Decode decodes value from b into v.
//
// If there is not enough bytes, Decode will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func Decode[T ScalarType](b []byte, v *T) error {
	return DecodeAny(b, v)
}

// DecodeAny decodes value from b into v.
//
// If there is not enough bytes, DecodeAny will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
//
// It panics if v is not a pointer to [ScalarType] (including CString).
func DecodeAny(b []byte, v any) error {
	var err error
	switch v := v.(type) {
	case *float64:
		*v, err = DecodeFloat64(b)
	case *string:
		*v, err = DecodeString(b)
	case *Binary:
		*v, err = DecodeBinary(b)
	case *ObjectID:
		*v, err = DecodeObjectID(b)
	case *bool:
		*v, err = DecodeBool(b)
	case *time.Time:
		*v, err = DecodeTime(b)
	case *NullType:
		// nothing
	case *Regex:
		*v, err = DecodeRegex(b)
	case *int32:
		*v, err = DecodeInt32(b)
	case *Timestamp:
		*v, err = DecodeTimestamp(b)
	case *int64:
		*v, err = DecodeInt64(b)
	case *Decimal128:
		*v, err = DecodeDecimal128(b)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}

	return err
}

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

// Type represents a BSON type.
type Type interface {
	ScalarType | CompositeType
}

// CompositeType represents a BSON composite type (including raw types).
type CompositeType interface {
	*Document | *Array | RawDocument | RawArray
}

// AnyDocument represents a BSON document type (both [*Document] and [RawDocument]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyDocument interface {
	Encode() (RawDocument, error)
	Decode() (*Document, error)
}

// AnyArray represents a BSON array type (both [*Array] and [RawArray]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyArray interface {
	Encode() (RawArray, error)
	Decode() (*Array, error)
}

// validBSONType checks if v is a valid BSON type (including raw types).
func validBSONType(v any) error {
	switch v := v.(type) {
	case *Document:
	case RawDocument:
	case *Array:
	case RawArray:
	case float64:
	case string:
	case Binary:
	case ObjectID:
	case bool:
	case time.Time:
	case NullType:
	case Regex:
	case int32:
	case Timestamp:
	case int64:
	case Decimal128:

	default:
		return lazyerrors.Errorf("invalid BSON type %T", v)
	}

	return nil
}
