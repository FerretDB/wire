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
	"time"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

//go:generate ../bin/stringer -linecomment -output stringers.go -type decodeMode,tag,BinarySubtype

// Type represents a BSON type.
type Type interface {
	CompositeType | ScalarType
}

// CompositeType represents a BSON composite type (including raw types).
type CompositeType interface {
	*Document | *Array | RawDocument | RawArray
}

// ScalarType represents a BSON scalar type.
//
// CString is not included as it is not a real BSON type.
type ScalarType interface {
	float64 | string | Binary | ObjectID | bool | time.Time | NullType | Regex | int32 | Timestamp | int64 | Decimal128
}

// AnyDocument represents a BSON document type (both [*Document] and [RawDocument]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyDocument interface {
	Encode(RawDocument) error
	Decode() (*Document, error)
}

// AnyArray represents a BSON array type (both [*Array] and [RawArray]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyArray interface {
	Encode(RawArray) error
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
