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
	"strconv"
	"time"
)

// Size returns a Size of the encoding of value v in bytes.
//
// It panics for invalid types.
func Size(v any) int {
	switch v := v.(type) {
	case *Document:
		return sizeDocument(v)
	case RawDocument:
		return len(v)
	case *Array:
		return sizeArray(v)
	case RawArray:
		return len(v)
	default:
		return sizeScalar(v)
	}
}

// sizeDocument returns a size of the encoding of Document doc in bytes.
func sizeDocument(doc *Document) int {
	res := 5

	for _, f := range doc.fields {
		res += 1 + SizeCString(f.name) + Size(f.value)
	}

	return res
}

// sizeArray returns a size of the encoding of Array arr in bytes.
func sizeArray(arr *Array) int {
	res := 5

	for i, v := range arr.values {
		res += 1 + SizeCString(strconv.Itoa(i)) + Size(v)
	}

	return res
}

// sizeScalar returns a size of the encoding of scalar value v in bytes.
//
// It panics if v is not a [ScalarType] (including CString).
func sizeScalar(v any) int {
	switch v := v.(type) {
	case float64:
		return sizeFloat64
	case string:
		return sizeString(v)
	case Binary:
		return sizeBinary(v)
	case ObjectID:
		return sizeObjectID
	case bool:
		return sizeBool
	case time.Time:
		return sizeTime
	case NullType:
		return 0
	case Regex:
		return sizeRegex(v)
	case int32:
		return sizeInt32
	case Timestamp:
		return sizeTimestamp
	case int64:
		return sizeInt64
	case Decimal128:
		return sizeDecimal128
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}
