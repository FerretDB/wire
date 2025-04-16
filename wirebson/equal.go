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
	"iter"
	"math"
	"time"
)

// Equal returns true if two BSON values are equal.
//
// It panics if invalid BSON type or decoding error is encountered.
// For that reason, it should not be used with user input.
//
// See also the wiretest package for test helpers.
//
//   - Documents and arrays are compared by their length and then by content.
//     They are decoded as needed; raw values are compared directly without decoding.
//   - float64 values are compared by their bits.
//     That handles all values, including various NaNs, infinities, negative zeros, etc.
//   - time.Time values are compared using the Equal method.
func Equal(v1, v2 any) bool {
	if err := validBSONType(v1); err != nil {
		panic(err)
	}

	if err := validBSONType(v2); err != nil {
		panic(err)
	}

	switch v1 := v1.(type) {
	case AnyDocument:
		d, ok := v2.(AnyDocument)
		if !ok {
			return false
		}

		return equalDocuments(v1, d)

	case AnyArray:
		a, ok := v2.(AnyArray)
		if !ok {
			return false
		}

		return equalArrays(v1, a)

	default:
		return equalScalars(v1, v2)
	}
}

// equalDocuments returns true if documents are equal.
func equalDocuments(d1, d2 AnyDocument) bool {
	raw1, _ := d1.(RawDocument)
	raw2, _ := d2.(RawDocument)

	if raw1 != nil && raw2 != nil {
		return bytes.Equal(raw1, raw2)
	}

	doc1, err := d1.Decode()
	if err != nil {
		panic(err)
	}

	doc2, err := d2.Decode()
	if err != nil {
		panic(err)
	}

	if doc1.Len() != doc2.Len() {
		return false
	}

	next, stop := iter.Pull2(doc2.All())
	defer stop()

	for k1, v1 := range doc1.All() {
		k2, v2, ok := next()
		if !ok {
			panic("broken iterator")
		}
		if k1 != k2 {
			return false
		}
		if !Equal(v1, v2) {
			return false
		}
	}

	_, _, ok := next()
	if ok {
		panic("broken iterator")
	}

	return true
}

// equalArrays returns true if arrays are equal.
func equalArrays(a1, a2 AnyArray) bool {
	raw1, _ := a1.(RawArray)
	raw2, _ := a2.(RawArray)

	if raw1 != nil && raw2 != nil {
		return bytes.Equal(raw1, raw2)
	}

	arr1, err := a1.Decode()
	if err != nil {
		panic(err)
	}

	arr2, err := a2.Decode()
	if err != nil {
		panic(err)
	}

	if arr1.Len() != arr2.Len() {
		return false
	}

	next, stop := iter.Pull(arr2.Values())
	defer stop()

	for v1 := range arr1.Values() {
		v2, ok := next()
		if !ok {
			panic("broken iterator")
		}
		if !Equal(v1, v2) {
			return false
		}
	}

	_, ok := next()
	if ok {
		panic("broken iterator")
	}

	return true
}

// equalScalars returns true if scalar values are equal.
func equalScalars(v1, v2 any) bool {
	switch s1 := v1.(type) {
	case float64:
		s2, ok := v2.(float64)
		if !ok {
			return false
		}

		// handles all values, including various NaNs
		return math.Float64bits(s1) == math.Float64bits(s2)

	case string:
		s2, ok := v2.(string)
		if !ok {
			return false
		}

		return s1 == s2

	case Binary:
		s2, ok := v2.(Binary)
		if !ok {
			return false
		}

		return s1.Subtype == s2.Subtype && bytes.Equal(s1.B, s2.B)

	// FIXME
	// case UndefinedType:
	// 	_, ok := v2.(UndefinedType)
	// 	return ok

	case ObjectID:
		s2, ok := v2.(ObjectID)
		if !ok {
			return false
		}

		return s1 == s2

	case bool:
		s2, ok := v2.(bool)
		if !ok {
			return false
		}

		return s1 == s2

	case time.Time:
		s2, ok := v2.(time.Time)
		if !ok {
			return false
		}

		return s1.Equal(s2)

	case NullType:
		_, ok := v2.(NullType)
		return ok

	case Regex:
		s2, ok := v2.(Regex)
		if !ok {
			return false
		}

		return s1.Pattern == s2.Pattern && s1.Options == s2.Options

	case int32:
		s2, ok := v2.(int32)
		if !ok {
			return false
		}

		return s1 == s2

	case Timestamp:
		s2, ok := v2.(Timestamp)
		if !ok {
			return false
		}

		return s1 == s2

	case int64:
		s2, ok := v2.(int64)
		if !ok {
			return false
		}

		return s1 == s2

	case Decimal128:
		s2, ok := v2.(Decimal128)
		if !ok {
			return false
		}

		return s1 == s2

	default:
		panic("not reached")
	}
}
