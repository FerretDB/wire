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
	"slices"
	"time"

	oldbson "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

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
