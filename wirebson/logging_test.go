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
	"context"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/wire/internal/util/testutil"
)

func TestLoggingNil(t *testing.T) {
	var doc *Document
	assert.Equal(t, doc.LogValue().String(), "Document<nil>")
	assert.Equal(t, LogMessageBlock(doc), "{<nil>}")
	assert.Equal(t, LogMessageFlow(doc), "{<nil>}")

	var arr *Array
	assert.Equal(t, arr.LogValue().String(), "Array<nil>")
	assert.Equal(t, LogMessageBlock(arr), "[<nil>]")
	assert.Equal(t, LogMessageFlow(arr), "[<nil>]")
}

func TestLogging(t *testing.T) {
	ctx := context.Background()

	opts := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if groups != nil {
				return a
			}

			switch a.Key {
			case slog.TimeKey:
				return slog.Attr{}

			case slog.LevelKey:
				return slog.Attr{}

			case slog.MessageKey:
				return slog.Attr{}
			}

			return a
		},
	}
	var tbuf, jbuf bytes.Buffer
	tlog := slog.New(slog.NewTextHandler(&tbuf, opts))
	jlog := slog.New(slog.NewJSONHandler(&jbuf, opts))

	for _, tc := range []struct {
		name string
		doc  any
		t    string
		j    string
		b    string
		f    string
	}{
		{
			name: "Numbers",
			doc: MustDocument(
				"f64", 42.0,
				"inf", float64(math.Inf(1)),
				"neg_inf", float64(math.Inf(-1)),
				"zero", math.Copysign(0, 1),
				"neg_zero", math.Copysign(0, -1),
				"nan", float64(math.NaN()),
				"nan_weird", float64(math.Float64frombits(0x7FF8000F000F0001)),
				"i32", int32(42),
				"i64", int64(42),
			),
			t: `v.f64=42 v.inf=+Inf v.neg_inf=-Inf v.zero=0 v.neg_zero=-0 v.nan=NaN v.nan_weird=NaN v.i32=42 v.i64=42`,
			j: `{"v":{"f64":42,"inf":"+Inf","neg_inf":"-Inf","zero":0,"neg_zero":-0,"nan":"NaN","nan_weird":"NaN","i32":42,"i64":42}}`,
			b: `
			{
			  "f64": 42.0,
			  "inf": +Inf,
			  "neg_inf": -Inf,
			  "zero": 0.0,
			  "neg_zero": -0.0,
			  "nan": NaN,
			  "nan_weird": NaN(111111111111000000000000000111100000000000011110000000000000001),
			  "i32": 42,
			  "i64": int64(42),
			}`,
			f: `{"f64": 42.0, "inf": +Inf, "neg_inf": -Inf, "zero": 0.0, "neg_zero": -0.0, "nan": NaN, ` +
				`"nan_weird": NaN(111111111111000000000000000111100000000000011110000000000000001), ` +
				`"i32": 42, "i64": int64(42)}`,
		},
		{
			name: "Scalars",
			doc: MustDocument(
				"null", Null,
				"id", ObjectID{0x42},
				"bool", true,
				"time", time.Date(2023, 3, 6, 13, 14, 42, 123456789, time.FixedZone("", int(4*time.Hour.Seconds()))),
			),
			t: `v.null=<nil> v.id=ObjectID(420000000000000000000000) v.bool=true v.time=2023-03-06T09:14:42.123Z`,
			j: `{"v":{"null":null,"id":"ObjectID(420000000000000000000000)","bool":true,"time":"2023-03-06T09:14:42.123Z"}}`,
			b: `
			{
			  "null": null,
			  "id": ObjectID(420000000000000000000000),
			  "bool": true,
			  "time": 2023-03-06T09:14:42.123Z,
			}`,
			f: `{"null": null, "id": ObjectID(420000000000000000000000), "bool": true, "time": 2023-03-06T09:14:42.123Z}`,
		},
		{
			name: "Composites",
			doc: MustDocument(
				"doc", MustDocument(
					"foo", "bar",
					"baz", MustDocument("qux", "quux"),
				),
				"doc_raw", RawDocument{0x42},
				"doc_empty", MustDocument(),
				"array", MustArray(
					"foo",
					"bar",
					MustArray("baz", "qux"),
				),
			),
			t: `v.doc.foo=bar v.doc.baz.qux=quux v.doc_raw=RawDocument<1> ` +
				`v.array.0=foo v.array.1=bar v.array.2.0=baz v.array.2.1=qux`,
			j: `{"v":{"doc":{"foo":"bar","baz":{"qux":"quux"}},"doc_raw":"RawDocument<1>",` +
				`"array":{"0":"foo","1":"bar","2":{"0":"baz","1":"qux"}}}}`,
			b: `
				{
				  "doc": {
				    "foo": "bar",
				    "baz": {
				      "qux": "quux",
				    },
				  },
				  "doc_raw": RawDocument<1>,
				  "doc_empty": {},
				  "array": [
				    "foo",
				    "bar",
				    [
				      "baz",
				      "qux",
				    ],
				  ],
				}`,
			f: `{"doc": {"foo": "bar", "baz": {"qux": "quux"}}, "doc_raw": RawDocument<1>, "doc_empty": {}, "array": ["foo", "bar", ["baz", "qux"]]}`,
		},
		{
			name: "Nested",
			doc:  makeNested(false, 20).(*Document),
			t:    `v.f.0.f.0.f.0.f.0.f.0.f.0.f.0.f.0.f.0.f.0=<nil>`,
			j: `{"v":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":` +
				`{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":null}}}}}}}}}}}}}}}}}}}}}`,
			b: `
				{
				  "f": [
				    {
				      "f": [
				        {
				          "f": [
				            {
				              "f": [
				                {
				                  "f": [
				                    {
				                      "f": [
				                        {
				                          "f": [
				                            {
				                              "f": [
				                                {
				                                  "f": [
				                                    {
				                                      "f": [
				                                        null,
				                                      ],
				                                    },
				                                  ],
				                                },
				                              ],
				                            },
				                          ],
				                        },
				                      ],
				                    },
				                  ],
				                },
				              ],
				            },
				          ],
				        },
				      ],
				    },
				  ],
				}`,
			f: `{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [null]}]}]}]}]}]}]}]}]}]}`,
		},
		{
			name: "Raw",
			doc:  RawDocument{42, 7},
			t:    `v=RawDocument<2>`,
			j:    `{"v":"RawDocument<2>"}`,
			b:    `RawDocument<2>`,
			f:    `RawDocument<2>`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tlog.InfoContext(ctx, "", slog.Any("v", tc.doc))
			assert.Equal(t, tc.t+"\n", tbuf.String(), "text output mismatch")
			tbuf.Reset()

			jlog.InfoContext(ctx, "", slog.Any("v", tc.doc))
			assert.Equal(t, tc.j+"\n", jbuf.String(), "json output mismatch")
			jbuf.Reset()

			b := LogMessageBlock(tc.doc)
			assert.Equal(t, testutil.Unindent(tc.b), b, "actual LogMessageBlock result:\n%s", b)

			f := LogMessageFlow(tc.doc)
			assert.Equal(t, testutil.Unindent(tc.f), f, "actual LogMessageFlow result:\n%s", f)
		})
	}
}

// makeNested creates a nested document or array with the given depth.
func makeNested(array bool, depth int) any {
	if depth < 1 {
		panic("depth must be at least 1")
	}

	var child any = Null

	if depth > 1 {
		child = makeNested(!array, depth-1)
	}

	if array {
		return MustArray(child)
	}

	return MustDocument("f", child)
}
