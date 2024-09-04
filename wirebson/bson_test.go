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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/wire/internal/util/testutil"
)

// normalTestCase represents a single test case for successful decoding/encoding.
//
//nolint:vet // for readability
type normalTestCase struct {
	name string
	raw  RawDocument
	doc  *Document
	m    string
}

// decodeTestCase represents a single test case for unsuccessful decoding.
//
//nolint:vet // for readability
type decodeTestCase struct {
	name string
	raw  RawDocument

	oldOk bool

	findRawErr    error
	findRawL      int
	decodeErr     error
	decodeDeepErr error // defaults to decodeErr
}

// normalTestCases represents test cases for successful decoding/encoding.
//
//nolint:lll // for readability
var normalTestCases = []normalTestCase{
	{
		name: "handshake1",
		raw:  testutil.MustParseDumpFile("testdata", "handshake1.hex"),
		doc: MustDocument(
			"ismaster", true,
			"client", MustDocument(
				"driver", MustDocument(
					"name", "nodejs",
					"version", "4.0.0-beta.6",
				),
				"os", MustDocument(
					"type", "Darwin",
					"name", "darwin",
					"architecture", "x64",
					"version", "20.6.0",
				),
				"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
				"application", MustDocument(
					"name", "mongosh 1.0.1",
				),
			),
			"compression", MustArray("none"),
			"loadBalanced", false,
		),
		m: `
		{
		  "ismaster": true,
		  "client": {
		    "driver": {"name": "nodejs", "version": "4.0.0-beta.6"},
		    "os": {
		      "type": "Darwin",
		      "name": "darwin",
		      "architecture": "x64",
		      "version": "20.6.0",
		    },
		    "platform": "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
		    "application": {"name": "mongosh 1.0.1"},
		  },
		  "compression": ["none"],
		  "loadBalanced": false,
		}`,
	},
	{
		name: "handshake2",
		raw:  testutil.MustParseDumpFile("testdata", "handshake2.hex"),
		doc: MustDocument(
			"ismaster", true,
			"client", MustDocument(
				"driver", MustDocument(
					"name", "nodejs",
					"version", "4.0.0-beta.6",
				),
				"os", MustDocument(
					"type", "Darwin",
					"name", "darwin",
					"architecture", "x64",
					"version", "20.6.0",
				),
				"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
				"application", MustDocument(
					"name", "mongosh 1.0.1",
				),
			),
			"compression", MustArray("none"),
			"loadBalanced", false,
		),
		m: `
		{
		  "ismaster": true,
		  "client": {
		    "driver": {"name": "nodejs", "version": "4.0.0-beta.6"},
		    "os": {
		      "type": "Darwin",
		      "name": "darwin",
		      "architecture": "x64",
		      "version": "20.6.0",
		    },
		    "platform": "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
		    "application": {"name": "mongosh 1.0.1"},
		  },
		  "compression": ["none"],
		  "loadBalanced": false,
		}`,
	},
	{
		name: "handshake3",
		raw:  testutil.MustParseDumpFile("testdata", "handshake3.hex"),
		doc: MustDocument(
			"buildInfo", int32(1),
			"lsid", MustDocument(
				"id", Binary{
					Subtype: BinaryUUID,
					B: []byte{
						0xa3, 0x19, 0xf2, 0xb4, 0xa1, 0x75, 0x40, 0xc7,
						0xb8, 0xe7, 0xa3, 0xa3, 0x2e, 0xc2, 0x56, 0xbe,
					},
				},
			),
			"$db", "admin",
		),
		m: `
		{
		  "buildInfo": 1,
		  "lsid": {"id": Binary(uuid:oxnytKF1QMe456OjLsJWvg==)},
		  "$db": "admin",
		}`,
	},
	{
		name: "handshake4",
		raw:  testutil.MustParseDumpFile("testdata", "handshake4.hex"),
		doc: MustDocument(
			"version", "5.0.0",
			"gitVersion", "1184f004a99660de6f5e745573419bda8a28c0e9",
			"modules", MustArray(),
			"allocator", "tcmalloc",
			"javascriptEngine", "mozjs",
			"sysInfo", "deprecated",
			"versionArray", MustArray(int32(5), int32(0), int32(0), int32(0)),
			"openssl", MustDocument(
				"running", "OpenSSL 1.1.1f  31 Mar 2020",
				"compiled", "OpenSSL 1.1.1f  31 Mar 2020",
			),
			"buildEnvironment", MustDocument(
				"distmod", "ubuntu2004",
				"distarch", "x86_64",
				"cc", "/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0",
				"ccflags", "-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables -ggdb "+
					"-Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer "+
					"-fno-strict-aliasing -O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 "+
					"-Wno-unused-local-typedefs -Wno-unused-function -Wno-deprecated-declarations "+
					"-Wno-unused-const-variable -Wno-unused-but-set-variable -Wno-missing-braces "+
					"-fstack-protector-strong -Wa,--nocompress-debug-sections -fno-builtin-memcmp",
				"cxx", "/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0",
				"cxxflags", "-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17",
				"linkflags", "-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong "+
					"-Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack "+
					"-Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags",
				"target_arch", "x86_64",
				"target_os", "linux",
				"cppdefines", "SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 _GNU_SOURCE "+
					"_REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 BOOST_THREAD_USES_DATETIME "+
					"BOOST_SYSTEM_NO_DEPRECATED BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS "+
					"BOOST_ENABLE_ASSERT_DEBUG_HANDLER BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG "+
					"BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS",
			),
			"bits", int32(64),
			"debug", false,
			"maxBsonObjectSize", int32(16777216),
			"storageEngines", MustArray("devnull", "ephemeralForTest", "wiredTiger"),
			"ok", float64(1),
		),
		m: `
		{
		  "version": "5.0.0",
		  "gitVersion": "1184f004a99660de6f5e745573419bda8a28c0e9",
		  "modules": [],
		  "allocator": "tcmalloc",
		  "javascriptEngine": "mozjs",
		  "sysInfo": "deprecated",
		  "versionArray": [5, 0, 0, 0],
		  "openssl": {
		    "running": "OpenSSL 1.1.1f  31 Mar 2020",
		    "compiled": "OpenSSL 1.1.1f  31 Mar 2020",
		  },
		  "buildEnvironment": {
		    "distmod": "ubuntu2004",
		    "distarch": "x86_64",
		    "cc": "/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0",
		    "ccflags": "-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables -ggdb -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer -fno-strict-aliasing -O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 -Wno-unused-local-typedefs -Wno-unused-function -Wno-deprecated-declarations -Wno-unused-const-variable -Wno-unused-but-set-variable -Wno-missing-braces -fstack-protector-strong -Wa,--nocompress-debug-sections -fno-builtin-memcmp",
		    "cxx": "/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0",
		    "cxxflags": "-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17",
		    "linkflags": "-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong -Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack -Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags",
		    "target_arch": "x86_64",
		    "target_os": "linux",
		    "cppdefines": "SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 _GNU_SOURCE _REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 BOOST_THREAD_USES_DATETIME BOOST_SYSTEM_NO_DEPRECATED BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS BOOST_ENABLE_ASSERT_DEBUG_HANDLER BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS",
		  },
		  "bits": 64,
		  "debug": false,
		  "maxBsonObjectSize": 16777216,
		  "storageEngines": ["devnull", "ephemeralForTest", "wiredTiger"],
		  "ok": 1.0,
		}`,
	},
	{
		name: "all",
		raw:  testutil.MustParseDumpFile("testdata", "all.hex"),
		doc: MustDocument(
			"array", MustArray(
				MustArray(""),
				MustArray("foo"),
			),
			"binary", MustArray(
				Binary{Subtype: BinaryUser, B: []byte{0x42}},
				Binary{Subtype: BinaryGeneric, B: []byte{}},
			),
			"bool", MustArray(true, false),
			"datetime", MustArray(
				time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
				time.Time{}.Local(),
			),
			"document", MustArray(
				MustDocument("foo", ""),
				MustDocument("", "foo"),
			),
			"double", MustArray(42.13, 0.0),
			"int32", MustArray(int32(42), int32(0)),
			"int64", MustArray(int64(42), int64(0)),
			"objectID", MustArray(ObjectID{0x42}, ObjectID{}),
			"string", MustArray("foo", ""),
			"timestamp", MustArray(Timestamp(42), Timestamp(0)),
			"decimal128", MustArray(Decimal128{L: 42, H: 13}),
		),
		m: `
		{
		  "array": [[""], ["foo"]],
		  "binary": [Binary(user:Qg==), Binary(generic:)],
		  "bool": [true, false],
		  "datetime": [2021-07-27T09:35:42.123Z, 0001-01-01T00:00:00Z],
		  "document": [{"foo": ""}, {"": "foo"}],
		  "double": [42.13, 0.0],
		  "int32": [42, 0],
		  "int64": [int64(42), int64(0)],
		  "objectID": [ObjectID(420000000000000000000000), ObjectID(000000000000000000000000)],
		  "string": ["foo", ""],
		  "timestamp": [Timestamp(42), Timestamp(0)],
		  "decimal128": [Decimal128(42,13)],
		}`,
	},
	{
		name: "nested",
		raw:  testutil.MustParseDumpFile("testdata", "nested.hex"),
		doc:  makeNested(false, 150).(*Document),
		m: `
		{
		  "f": [
		    {
		      "f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{"f": [{...}]}]}]}]}]}]}]}]}],
		    },
		  ],
		}`,
	},
	{
		name: "float64Doc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x01, 0x66, 0x00,
			0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40,
			0x00,
		},
		doc: MustDocument(
			"f", float64(3.141592653589793),
		),
		m: `{"f": 3.141592653589793}`,
	},
	{
		name: "stringDoc",
		raw: RawDocument{
			0x0e, 0x00, 0x00, 0x00,
			0x02, 0x66, 0x00,
			0x02, 0x00, 0x00, 0x00,
			0x76, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", "v",
		),
		m: `{"f": "v"}`,
	},
	{
		name: "binaryDoc",
		raw: RawDocument{
			0x0e, 0x00, 0x00, 0x00,
			0x05, 0x66, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x80,
			0x76,
			0x00,
		},
		doc: MustDocument(
			"f", Binary{B: []byte("v"), Subtype: BinaryUser},
		),
		m: `{"f": Binary(user:dg==)}`,
	},
	{
		name: "objectIDDoc",
		raw: RawDocument{
			0x14, 0x00, 0x00, 0x00,
			0x07, 0x66, 0x00,
			0x62, 0x56, 0xc5, 0xba, 0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40,
			0x00,
		},
		doc: MustDocument(
			"f", ObjectID{0x62, 0x56, 0xc5, 0xba, 0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40},
		),
		m: `{"f": ObjectID(6256c5ba182d4454fb210940)}`,
	},
	{
		name: "boolDoc",
		raw: RawDocument{
			0x09, 0x00, 0x00, 0x00,
			0x08, 0x66, 0x00,
			0x01,
			0x00,
		},
		doc: MustDocument(
			"f", true,
		),
		m: `{"f": true}`,
	},
	{
		name: "timeDoc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x09, 0x66, 0x00,
			0x0b, 0xce, 0x82, 0x18, 0x8d, 0x01, 0x00, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", time.Date(2024, 1, 17, 17, 40, 42, 123000000, time.UTC),
		),
		m: `{"f": 2024-01-17T17:40:42.123Z}`,
	},
	{
		name: "nullDoc",
		raw: RawDocument{
			0x08, 0x00, 0x00, 0x00,
			0x0a, 0x66, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", Null,
		),
		m: `{"f": null}`,
	},
	{
		name: "regexDoc",
		raw: RawDocument{
			0x0c, 0x00, 0x00, 0x00,
			0x0b, 0x66, 0x00,
			0x70, 0x00,
			0x6f, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", Regex{Pattern: "p", Options: "o"},
		),
		m: `{"f": /p/o}`,
	},
	{
		name: "int32Doc",
		raw: RawDocument{
			0x0c, 0x00, 0x00, 0x00,
			0x10, 0x66, 0x00,
			0xa1, 0xb0, 0xb9, 0x12,
			0x00,
		},
		doc: MustDocument(
			"f", int32(314159265),
		),
		m: `{"f": 314159265}`,
	},
	{
		name: "timestampDoc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x11, 0x66, 0x00,
			0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", Timestamp(42),
		),
		m: `{"f": Timestamp(42)}`,
	},
	{
		name: "int64Doc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x12, 0x66, 0x00,
			0x21, 0x6d, 0x25, 0x0a, 0x43, 0x29, 0x0b, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", int64(3141592653589793),
		),
		m: `{"f": int64(3141592653589793)}`,
	},
	{
		name: "decimal128Doc",
		raw: RawDocument{
			0x18, 0x00, 0x00, 0x00,
			0x13, 0x66, 0x00,
			0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", Decimal128{L: 42, H: 13},
		),
		m: `{"f": Decimal128(42,13)}`,
	},
	{
		name: "smallDoc",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x05, 0x00, 0x00, 0x00, 0x00, // subdocument length and end of subdocument
			0x00, // end of document
		},
		doc: MustDocument(
			"foo", MustDocument(),
		),
		m: `{"foo": {}}`,
	},
	{
		name: "smallArray",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x04, 0x66, 0x6f, 0x6f, 0x00, // subarray "foo"
			0x05, 0x00, 0x00, 0x00, 0x00, // subarray length and end of subarray
			0x00, // end of document
		},
		doc: MustDocument(
			"foo", MustArray(),
		),
		m: `{"foo": []}`,
	},
	{
		name: "duplicateKeys",
		raw: RawDocument{
			0x0b, 0x00, 0x00, 0x00, // document length
			0x08, 0x00, 0x00, // "": false
			0x08, 0x00, 0x01, // "": true
			0x00, // end of document
		},
		doc: MustDocument(
			"", false,
			"", true,
		),
		m: `{"": false, "": true}`,
	},
}

// decodeTestCases represents test cases for unsuccessful decoding.
var decodeTestCases = []decodeTestCase{
	{
		name:       "EOF",
		raw:        RawDocument{0x00},
		findRawErr: ErrDecodeShortInput,
		decodeErr:  ErrDecodeShortInput,
	},
	{
		name: "invalidLength",
		raw: RawDocument{
			0x00, 0x00, 0x00, 0x00, // invalid document length
			0x00, // end of document
		},
		findRawErr: ErrDecodeInvalidInput,
		decodeErr:  ErrDecodeInvalidInput,
	},
	{
		name: "missingByte",
		raw: RawDocument{
			0x06, 0x00, 0x00, 0x00, // document length
			0x00, // end of document
		},
		findRawErr: ErrDecodeShortInput,
		decodeErr:  ErrDecodeShortInput,
	},
	{
		name: "extraByte",
		raw: RawDocument{
			0x05, 0x00, 0x00, 0x00, // document length
			0x00, // end of document
			0x00, // extra byte
		},
		oldOk:     true,
		findRawL:  5,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "unexpectedTag",
		raw: RawDocument{
			0x06, 0x00, 0x00, 0x00, // document length
			0xdd, // unexpected tag
			0x00, // end of document
		},
		findRawL:  6,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "invalidTag",
		raw: RawDocument{
			0x06, 0x00, 0x00, 0x00, // document length
			0x00, // invalid tag
			0x00, // end of document
		},
		findRawL:  6,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "shortDoc",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x06, 0x00, 0x00, 0x00, // invalid subdocument length
			0x00, // end of subdocument
			0x00, // end of document
		},
		findRawL:      15,
		decodeErr:     ErrDecodeShortInput,
		decodeDeepErr: ErrDecodeInvalidInput,
	},
	{
		name: "invalidDoc",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x05, 0x00, 0x00, 0x00, // subdocument length
			0x30, // invalid end of subdocument
			0x00, // end of document
		},
		findRawL:  15,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "invalidDocTag",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x06, 0x00, 0x00, 0x00, // subdocument length
			0x00, // invalid tag
			0x00, // end of subdocument
			0x00, // end of document
		},
		findRawL:      16,
		decodeDeepErr: ErrDecodeInvalidInput,
	},
}

func TestNormal(t *testing.T) {
	for _, tc := range normalTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("FindRaw", func(t *testing.T) {
				ls := tc.raw.LogValue().Resolve().String()
				assert.NotContains(t, ls, "panicked")
				assert.NotContains(t, ls, "called too many times")

				assert.NotEmpty(t, LogMessage(tc.raw))
				assert.NotEmpty(t, LogMessageBlock(tc.raw))
				assert.NotEmpty(t, LogMessageFlow(tc.raw))

				l, err := FindRaw(tc.raw)
				require.NoError(t, err)
				require.Len(t, tc.raw, l)
			})

			t.Run("DecodeEncode", func(t *testing.T) {
				doc, err := tc.raw.Decode()
				require.NoError(t, err)

				ls := doc.LogValue().Resolve().String()
				assert.NotContains(t, ls, "panicked")
				assert.NotContains(t, ls, "called too many times")

				assert.NotEmpty(t, LogMessage(doc))
				assert.NotEmpty(t, LogMessageBlock(doc))
				assert.NotEmpty(t, LogMessageFlow(doc))

				raw := make([]byte, Size(doc))
				err = doc.Encode(raw)
				require.NoError(t, err)

				assert.Equal(t, tc.raw, raw)
			})

			t.Run("DecodeDeepEncode", func(t *testing.T) {
				doc, err := tc.raw.DecodeDeep()
				require.NoError(t, err)

				ls := doc.LogValue().Resolve().String()
				assert.NotContains(t, ls, "panicked")
				assert.NotContains(t, ls, "called too many times")

				assert.Equal(t, testutil.Unindent(tc.m), LogMessage(doc))
				assert.NotEmpty(t, LogMessageBlock(doc))
				assert.NotEmpty(t, LogMessageFlow(doc))

				raw := make([]byte, Size(doc))
				err = doc.Encode(raw)
				require.NoError(t, err)
				assert.Equal(t, tc.raw, raw)
			})
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tc := range decodeTestCases {
		if tc.decodeDeepErr == nil {
			tc.decodeDeepErr = tc.decodeErr
		}

		require.NotNil(t, tc.decodeDeepErr, "invalid test case %q", tc.name)

		t.Run(tc.name, func(t *testing.T) {
			t.Run("FindRaw", func(t *testing.T) {
				ls := tc.raw.LogValue().Resolve().String()
				assert.NotContains(t, ls, "panicked")
				assert.NotContains(t, ls, "called too many times")

				assert.NotEmpty(t, LogMessage(tc.raw))
				assert.NotEmpty(t, LogMessageBlock(tc.raw))
				assert.NotEmpty(t, LogMessageFlow(tc.raw))

				l, err := FindRaw(tc.raw)

				if tc.findRawErr != nil {
					require.ErrorIs(t, err, tc.findRawErr)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.findRawL, l)
			})

			t.Run("Decode", func(t *testing.T) {
				_, err := tc.raw.Decode()

				if tc.decodeErr != nil {
					require.ErrorIs(t, err, tc.decodeErr)
					return
				}

				require.NoError(t, err)
			})

			t.Run("DecodeDeep", func(t *testing.T) {
				_, err := tc.raw.DecodeDeep()
				require.ErrorIs(t, err, tc.decodeDeepErr)
			})
		})
	}
}

func BenchmarkDocument(b *testing.B) {
	for _, tc := range normalTestCases {
		b.Run(tc.name, func(b *testing.B) {
			var doc *Document
			var raw []byte
			var m string
			var err error

			b.Run("Decode", func(b *testing.B) {
				b.ReportAllocs()

				for range b.N {
					doc, err = tc.raw.Decode()
				}

				b.StopTimer()

				require.NoError(b, err)
				require.NotNil(b, doc)
			})

			b.Run("Encode", func(b *testing.B) {
				doc, err = tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					raw = make([]byte, Size(doc))
					err = doc.Encode(raw)
				}

				b.StopTimer()

				require.NoError(b, err)
				assert.NotNil(b, raw)
			})

			b.Run("LogValue", func(b *testing.B) {
				doc, err = tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					m = doc.LogValue().Resolve().String()
				}

				b.StopTimer()

				assert.NotEmpty(b, m)
				assert.NotContains(b, m, "panicked")
				assert.NotContains(b, m, "called too many times")
			})

			b.Run("LogMessage", func(b *testing.B) {
				doc, err = tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					m = LogMessage(doc)
				}

				b.StopTimer()

				assert.NotEmpty(b, m)
			})

			b.Run("DecodeDeep", func(b *testing.B) {
				b.ReportAllocs()

				for range b.N {
					doc, err = tc.raw.DecodeDeep()
				}

				b.StopTimer()

				require.NoError(b, err)
				require.NotNil(b, doc)
			})

			b.Run("EncodeDeep", func(b *testing.B) {
				doc, err = tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					raw := make([]byte, Size(doc))
					err = doc.Encode(raw)
				}

				b.StopTimer()

				require.NoError(b, err)
				assert.NotNil(b, raw)
			})

			b.Run("LogValueDeep", func(b *testing.B) {
				doc, err = tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					m = doc.LogValue().Resolve().String()
				}

				b.StopTimer()

				assert.NotEmpty(b, m)
				assert.NotContains(b, m, "panicked")
				assert.NotContains(b, m, "called too many times")
			})

			b.Run("LogMessageDeep", func(b *testing.B) {
				doc, err = tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					m = LogMessage(doc)
				}

				b.StopTimer()

				assert.NotEmpty(b, m)
			})
		})
	}
}

// testRawDocument tests a single RawDocument (that might or might not be valid).
// It is adapted from tests above.
func testRawDocument(t *testing.T, rawDoc RawDocument) {
	t.Helper()

	t.Run("FindRaw", func(t *testing.T) {
		ls := rawDoc.LogValue().Resolve().String()
		assert.NotContains(t, ls, "panicked")
		assert.NotContains(t, ls, "called too many times")

		assert.NotEmpty(t, LogMessage(rawDoc))
		assert.NotEmpty(t, LogMessageBlock(rawDoc))
		assert.NotEmpty(t, LogMessageFlow(rawDoc))

		_, _ = FindRaw(rawDoc)
	})

	t.Run("DecodeEncode", func(t *testing.T) {
		doc, err := rawDoc.Decode()
		if err != nil {
			_, err = rawDoc.DecodeDeep()
			assert.Error(t, err) // it might be different

			return
		}

		ls := doc.LogValue().Resolve().String()
		assert.NotContains(t, ls, "panicked")
		assert.NotContains(t, ls, "called too many times")

		assert.NotEmpty(t, LogMessage(doc))
		assert.NotEmpty(t, LogMessageBlock(doc))
		assert.NotEmpty(t, LogMessageFlow(doc))

		raw := make([]byte, Size(doc))
		err = doc.Encode(raw)

		if err == nil {
			assert.Equal(t, rawDoc, raw)
		}
	})

	t.Run("DecodeDeepEncode", func(t *testing.T) {
		doc, err := rawDoc.DecodeDeep()
		if err != nil {
			return
		}

		ls := doc.LogValue().Resolve().String()
		assert.NotContains(t, ls, "panicked")
		assert.NotContains(t, ls, "called too many times")

		assert.NotEmpty(t, LogMessage(doc))
		assert.NotEmpty(t, LogMessageBlock(doc))
		assert.NotEmpty(t, LogMessageFlow(doc))

		raw := make([]byte, Size(doc))
		err = doc.Encode(raw)
		require.NoError(t, err)
		assert.Equal(t, rawDoc, raw)
	})
}

func FuzzDocument(f *testing.F) {
	for _, tc := range normalTestCases {
		f.Add([]byte(tc.raw))
	}

	for _, tc := range decodeTestCases {
		f.Add([]byte(tc.raw))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		testRawDocument(t, RawDocument(b))

		l, err := FindRaw(b)
		if err == nil {
			testRawDocument(t, RawDocument(b[:l]))
		}
	})
}
