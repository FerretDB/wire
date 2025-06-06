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
//	BSON                Go                        $type alias
//
//	Document/Object     *Document or RawDocument  object
//	Array               *Array or RawArray        array
//
//	Double              float64                   double
//	String              string                    string
//	Binary data         Binary                    binData
//	Undefined           UndefinedType             undefined
//	ObjectId            ObjectID                  objectId
//	Boolean             bool                      bool
//	Date                time.Time                 date
//	Null                NullType                  null
//	Regular Expression  Regex                     regex
//	32-bit integer      int32                     int
//	Timestamp           Timestamp                 timestamp
//	64-bit integer      int64                     long
//	Decimal128          Decimal128                decimal
//
// Composite types (Document and Array) are passed by pointers.
// Raw composite type and scalars are passed by values.
package wirebson

import (
	"slices"
	"time"

	oldbson "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

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
	float64 | string | Binary | UndefinedType | ObjectID | bool | time.Time | NullType | Regex | int32 | Timestamp | int64 | Decimal128
}

// AnyDocument represents a BSON document type (both [*Document] and [RawDocument]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyDocument interface {
	Encode() (RawDocument, error)
	Decode() (*Document, error)
	LogMessage() string
	LogMessageIndent() string
}

// AnyArray represents a BSON array type (both [*Array] and [RawArray]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyArray interface {
	Encode() (RawArray, error)
	Decode() (*Array, error)
	LogMessage() string
	LogMessageIndent() string
}

// validBSONType checks if v is a valid BSON type (including raw types).
func validBSONType(v any) error {
	switch v := v.(type) {
	case AnyDocument:
	case AnyArray:
	case float64:
	case string:
	case Binary:
	case UndefinedType:
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

// FromDriver converts MongoDB driver v2 (and, temporary, v1) value ([bson.D], [bson.A], etc) to wirebson value.
func FromDriver(v any) (any, error) {
	switch v := v.(type) {
	case bson.D:
		doc := MakeDocument(len(v))
		for _, e := range v {
			val, err := FromDriver(e.Value)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if err = doc.Add(e.Key, val); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		return doc, nil

	case oldbson.D:
		d := make(bson.D, len(v))
		for i, e := range v {
			d[i] = bson.E{Key: e.Key, Value: e.Value}
		}

		return FromDriver(d)

	case bson.A:
		arr := MakeArray(len(v))
		for _, e := range v {
			val, err := FromDriver(e)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if err = arr.Add(val); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		return arr, nil

	case oldbson.A:
		return FromDriver(bson.A(v))

	case float64:
		return v, nil
	case string:
		return v, nil
	case bson.Binary:
		return Binary{B: slices.Clip(slices.Clone(v.Data)), Subtype: BinarySubtype(v.Subtype)}, nil
	case bson.Undefined:
		return Undefined, nil
	case bson.ObjectID:
		return ObjectID(v), nil
	case bool:
		return v, nil
	case bson.DateTime:
		return v.Time().UTC(), nil
	case bson.Null, nil:
		return Null, nil
	case bson.Regex:
		return Regex{Pattern: v.Pattern, Options: v.Options}, nil
	case int32:
		return v, nil
	case bson.Timestamp:
		return NewTimestamp(v.T, v.I), nil
	case int64:
		return v, nil
	case bson.Decimal128:
		h, l := v.GetBytes()
		return Decimal128{H: h, L: l}, nil

	case oldbson.Binary:
		return Binary{B: slices.Clip(slices.Clone(v.Data)), Subtype: BinarySubtype(v.Subtype)}, nil
	case oldbson.Undefined:
		return Undefined, nil
	case oldbson.ObjectID:
		return ObjectID(v), nil
	case oldbson.DateTime:
		return v.Time().UTC(), nil
	case oldbson.Null:
		return Null, nil
	case oldbson.Regex:
		return Regex{Pattern: v.Pattern, Options: v.Options}, nil
	case oldbson.Timestamp:
		return NewTimestamp(v.T, v.I), nil
	case oldbson.Decimal128:
		h, l := v.GetBytes()
		return Decimal128{H: h, L: l}, nil

	default:
		return nil, lazyerrors.Errorf("invalid BSON type %T", v)
	}
}

// ToDriver converts wirebson value to MongoDB driver v2 value (bson.D, bson.A, etc).
func ToDriver(v any) (any, error) {
	switch v := v.(type) {
	case *Document:
		doc := make(bson.D, 0, v.Len())
		for k, v := range v.All() {
			val, err := ToDriver(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			doc = append(doc, bson.E{Key: k, Value: val})
		}

		return doc, nil

	case *Array:
		arr := make(bson.A, v.Len())
		for i, v := range v.All() {
			val, err := ToDriver(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			arr[i] = val
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case Binary:
		return bson.Binary{
			Subtype: byte(v.Subtype),
			Data:    slices.Clip(slices.Clone(v.B)),
		}, nil
	case UndefinedType:
		return bson.Undefined{}, nil
	case ObjectID:
		return bson.ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return bson.NewDateTimeFromTime(v), nil
	case NullType:
		return bson.Null{}, nil
	case Regex:
		return bson.Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case Timestamp:
		return bson.Timestamp{T: v.T(), I: v.I()}, nil
	case int64:
		return v, nil
	case Decimal128:
		return bson.NewDecimal128(v.H, v.L), nil

	default:
		return nil, lazyerrors.Errorf("invalid BSON type %T", v)
	}
}
