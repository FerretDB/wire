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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

// logMaxDepth is the maximum depth of a recursive representation of a BSON value.
const logMaxDepth = 20

// nanBits is the most common pattern of a NaN float64 value, the same as math.Float64bits(math.NaN()).
const nanBits = 0b111111111111000000000000000000000000000000000000000000000000001

// slogValue returns a compact representation of any BSON value as [slog.Value].
// It may change over time.
//
// The result is optimized for small values such as function parameters.
// Some information is lost;
// for example, both int32 and int64 values are returned with [slog.KindInt64],
// arrays are treated as documents, and empty documents are omitted.
// More information is subsequently lost in handlers output;
// for example, float64(42), int32(42), and int64(42) values would all look the same
// (`f64=42 i32=42 i64=42` or `{"f64":42,"i32":42,"i64":42}`).
func slogValue(v any, depth int) slog.Value {
	switch v := v.(type) {
	case *Document:
		if v == nil {
			return slog.StringValue("Document<nil>")
		}

		if depth > logMaxDepth {
			return slog.StringValue("Document<...>")
		}

		attrs := make([]slog.Attr, len(v.fields))

		for i, f := range v.fields {
			attrs[i] = slog.Attr{Key: f.name, Value: slogValue(f.value, depth+1)}
		}

		return slog.GroupValue(attrs...)

	case RawDocument:
		if v == nil {
			return slog.StringValue("RawDocument<nil>")
		}

		return slog.StringValue("RawDocument<" + strconv.Itoa(len(v)) + ">")

	case *Array:
		if v == nil {
			return slog.StringValue("Array<nil>")
		}

		if depth > logMaxDepth {
			return slog.StringValue("Array<...>")
		}

		attrs := make([]slog.Attr, len(v.values))

		for i, v := range v.values {
			attrs[i] = slog.Attr{Key: strconv.Itoa(i), Value: slogValue(v, depth+1)}
		}

		return slog.GroupValue(attrs...)

	case RawArray:
		if v == nil {
			return slog.StringValue("RawArray<nil>")
		}

		return slog.StringValue("RawArray<" + strconv.Itoa(len(v)) + ">")

	case float64:
		// for JSON handler to work
		switch {
		case math.IsNaN(v):
			return slog.StringValue("NaN")
		case math.IsInf(v, 1):
			return slog.StringValue("+Inf")
		case math.IsInf(v, -1):
			return slog.StringValue("-Inf")
		}

		return slog.Float64Value(v)

	case string:
		return slog.StringValue(v)

	case Binary:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	case ObjectID:
		return slog.StringValue("ObjectID(" + hex.EncodeToString(v[:]) + ")")

	case bool:
		return slog.BoolValue(v)

	case time.Time:
		return slog.TimeValue(v.Truncate(time.Millisecond).UTC())

	case NullType:
		return slog.Value{}

	case Regex:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	case int32:
		return slog.Int64Value(int64(v))

	case Timestamp:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	case int64:
		return slog.Int64Value(v)

	case Decimal128:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

// LogMessage returns a representation as a string.
// It may change over time.
func LogMessage(v any) string {
	var b strings.Builder
	logMessage(v, -1, 1, &b)
	return b.String()
}

// LogMessageIndent returns a representation as an indented string.
// It may change over time.
func LogMessageIndent(v any) string {
	var b strings.Builder
	logMessage(v, 0, 1, &b)
	return b.String()
}

// logMessage returns an indented representation of any BSON value as a string,
// somewhat similar (but not identical) to JSON or Go syntax.
// It may change over time.
//
// The result is optimized for large values such as full request documents.
// All information is preserved.
func logMessage(v any, indent, depth int, b *strings.Builder) {
	switch v := v.(type) {
	case *Document:
		if v == nil {
			b.WriteString("{<nil>}")
			return
		}

		l := len(v.fields)
		if l == 0 {
			b.WriteString("{}")
			return
		}

		if depth > logMaxDepth {
			b.WriteString("{...}")
			return
		}

		if indent < 0 {
			b.WriteByte('{')

			for i, f := range v.fields {
				fmt.Fprintf(b, "%#q: ", f.name)

				logMessage(f.value, -1, depth+1, b)

				if i != l-1 {
					b.WriteString(", ")
				}
			}

			b.WriteByte('}')
			return
		}

		b.WriteString("{\n")

		for _, f := range v.fields {
			b.WriteString(strings.Repeat("  ", indent+1))

			fmt.Fprintf(b, "%#q: ", f.name)

			logMessage(f.value, indent+1, depth+1, b)

			b.WriteString(",\n")
		}

		b.WriteString(strings.Repeat("  ", indent))
		b.WriteByte('}')

	case RawDocument:
		fmt.Fprintf(b, "RawDocument<%d>", len(v))

	case *Array:
		if v == nil {
			b.WriteString("[<nil>]")
			return
		}

		l := len(v.values)
		if l == 0 {
			b.WriteString("[]")
			return
		}

		if depth > logMaxDepth {
			b.WriteString("[...]")
			return
		}

		if indent < 0 {
			b.WriteByte('[')

			for i, e := range v.values {
				logMessage(e, -1, depth+1, b)

				if i != l-1 {
					b.WriteString(", ")
				}
			}

			b.WriteRune(']')
			return
		}

		b.WriteString("[\n")

		for _, e := range v.values {
			b.WriteString(strings.Repeat("  ", indent+1))

			logMessage(e, indent+1, depth+1, b)

			b.WriteString(",\n")
		}

		b.WriteString(strings.Repeat("  ", indent))
		b.WriteByte(']')

	case RawArray:
		fmt.Fprintf(b, "RawArray<%d>", len(v))

	case float64:
		switch {
		case math.IsNaN(v):
			if bits := math.Float64bits(v); bits != nanBits {
				fmt.Fprintf(b, "NaN(%b)", bits)
				return
			}

			b.WriteString("NaN")

		case math.IsInf(v, 1):
			b.WriteString("+Inf")

		case math.IsInf(v, -1):
			b.WriteString("-Inf")

		default:
			res := strconv.FormatFloat(v, 'f', -1, 64)
			if !strings.Contains(res, ".") {
				res += ".0"
			}

			b.WriteString(res)
		}

	case string:
		fmt.Fprintf(b, "%#q", v)

	case Binary:
		b.WriteString("Binary(")
		b.WriteString(v.Subtype.String())
		b.WriteByte(':')
		b.WriteString(base64.StdEncoding.EncodeToString(v.B))
		b.WriteByte(')')

	case ObjectID:
		b.WriteString("ObjectID(")
		b.WriteString(hex.EncodeToString(v[:]))
		b.WriteByte(')')

	case bool:
		b.WriteString(strconv.FormatBool(v))

	case time.Time:
		b.WriteString(v.Truncate(time.Millisecond).UTC().Format(time.RFC3339Nano))

	case NullType:
		b.WriteString("null")

	case Regex:
		b.WriteByte('/')
		b.WriteString(v.Pattern)
		b.WriteByte('/')
		b.WriteString(v.Options)

	case int32:
		b.WriteString(strconv.FormatInt(int64(v), 10))

	case Timestamp:
		b.WriteString("Timestamp(")
		b.WriteString(strconv.FormatUint(uint64(v), 10))
		b.WriteByte(')')

	case int64:
		b.WriteString("int64(")
		b.WriteString(strconv.FormatInt(int64(v), 10))
		b.WriteByte(')')

	case Decimal128:
		b.WriteString("Decimal128(H:")
		b.WriteString(strconv.FormatUint(uint64(v.H), 10))
		b.WriteString(",L:")
		b.WriteString(strconv.FormatUint(uint64(v.L), 10))
		b.WriteByte(')')

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}
