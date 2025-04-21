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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/wire/internal/util/testutil"
)

// normalTestCase represents a single test case for successful decoding/encoding.
//
//nolint:govet // for readability
type normalTestCase struct {
	name string
	raw  RawDocument
	doc  *Document
	mi   string
	j    string
}

// decodeTestCase represents a single test case for unsuccessful decoding.
//
//nolint:govet // for readability
type decodeTestCase struct {
	name string
	raw  RawDocument

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
		mi: `
		{
		  "ismaster": true,
		  "client": {
		    "driver": {
		      "name": "nodejs",
		      "version": "4.0.0-beta.6",
		    },
		    "os": {
		      "type": "Darwin",
		      "name": "darwin",
		      "architecture": "x64",
		      "version": "20.6.0",
		    },
		    "platform": "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
		    "application": {
		      "name": "mongosh 1.0.1",
		    },
		  },
		  "compression": [
		    "none",
		  ],
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
		mi: `
		{
		  "ismaster": true,
		  "client": {
		    "driver": {
		      "name": "nodejs",
		      "version": "4.0.0-beta.6",
		    },
		    "os": {
		      "type": "Darwin",
		      "name": "darwin",
		      "architecture": "x64",
		      "version": "20.6.0",
		    },
		    "platform": "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
		    "application": {
		      "name": "mongosh 1.0.1",
		    },
		  },
		  "compression": [
		    "none",
		  ],
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
		mi: `
		{
		  "buildInfo": 1,
		  "lsid": {
		    "id": Binary(uuid:oxnytKF1QMe456OjLsJWvg==),
		  },
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
		mi: `
		{
		  "version": "5.0.0",
		  "gitVersion": "1184f004a99660de6f5e745573419bda8a28c0e9",
		  "modules": [],
		  "allocator": "tcmalloc",
		  "javascriptEngine": "mozjs",
		  "sysInfo": "deprecated",
		  "versionArray": [
		    5,
		    0,
		    0,
		    0,
		  ],
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
		  "storageEngines": [
		    "devnull",
		    "ephemeralForTest",
		    "wiredTiger",
		  ],
		  "ok": 1.0,
		}`,
	},
	{
		name: "all",
		raw:  testutil.MustParseDumpFile("testdata", "all.hex"),
		doc: MustDocument(
			"document", MustArray(
				MustDocument("", "foo", "bar", "baz", "", "qux"),
				MustDocument(),
			),
			"array", MustArray(
				MustArray("foo"),
				MustArray(),
			),
			"float64", MustArray(
				42.13,
				math.Copysign(0, +1), math.Copysign(0, -1),
				math.Inf(+1), math.Inf(-1),
				math.NaN(), math.Float64frombits(0x7ff8000f000f0001),
			),
			"string", MustArray("foo", ""),
			"binary", MustArray(
				Binary{Subtype: BinaryUser, B: []byte{0x42}},
				Binary{Subtype: BinaryGeneric, B: []byte{}},
			),
			"undefined", MustArray(Undefined),
			"objectID", MustArray(ObjectID{0x42}, ObjectID{}),
			"bool", MustArray(true, false),
			"datetime", MustArray(
				time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC),
				time.Time{},
			),
			"null", MustArray(Null),
			"regex", MustArray(Regex{Pattern: "p", Options: "o"}, Regex{}),
			"int32", MustArray(int32(42), int32(0)),
			"timestamp", MustArray(Timestamp(42), Timestamp(0)),
			"int64", MustArray(int64(42), int64(0)),
			"decimal128", MustArray(Decimal128{H: 13, L: 42}, Decimal128{}),
		),
		mi: `
		{
		  "document": [
		    {
		      "": "foo",
		      "bar": "baz",
		      "": "qux",
		    },
		    {},
		  ],
		  "array": [
		    [
		      "foo",
		    ],
		    [],
		  ],
		  "float64": [
		    42.13,
		    0.0,
		    -0.0,
		    +Inf,
		    -Inf,
		    NaN,
		    NaN,
		  ],
		  "string": [
		    "foo",
		    "",
		  ],
		  "binary": [
		    Binary(user:Qg==),
		    Binary(generic:),
		  ],
		  "undefined": [
		    undefined,
		  ],
		  "objectID": [
		    ObjectID(420000000000000000000000),
		    ObjectID(000000000000000000000000),
		  ],
		  "bool": [
		    true,
		    false,
		  ],
		  "datetime": [
		    2021-07-27T09:35:42.123Z,
		    0001-01-01T00:00:00Z,
		  ],
		  "null": [
		    null,
		  ],
		  "regex": [
		    /p/o,
		    //,
		  ],
		  "int32": [
		    42,
		    0,
		  ],
		  "timestamp": [
		    Timestamp(42),
		    Timestamp(0),
		  ],
		  "int64": [
		    int64(42),
		    int64(0),
		  ],
		  "decimal128": [
		    Decimal128(H:13,L:42),
		    Decimal128(H:0,L:0),
		  ],
		}`,
		j: `
		{
		  "document": [
		    {
		      "": "foo",
		      "bar": "baz",
		      "": "qux"
		    },
		    {}
		  ],
		  "array": [
		    [
		      "foo"
		    ],
		    []
		  ],
		  "float64": [
		    {
		      "$numberDouble": "42.13"
		    },
		    {
		      "$numberDouble": "0.0"
		    },
		    {
		      "$numberDouble": "-0.0"
		    },
		    {
		      "$numberDouble": "Infinity"
		    },
		    {
		      "$numberDouble": "-Infinity"
		    },
		    {
		      "$numberDouble": "NaN"
		    },
		    {
		      "$numberDouble": "NaN"
		    }
		  ],
		  "string": [
		    "foo",
		    ""
		  ],
		  "binary": [
		    {
		      "$binary": {
		        "base64": "Qg==",
		        "subType": "80"
		      }
		    },
		    {
		      "$binary": {
		        "base64": "",
		        "subType": "00"
		      }
		    }
		  ],
		  "undefined": [
		    {
		      "$undefined": true
		    }
		  ],
		  "objectID": [
		    {
		      "$oid": "420000000000000000000000"
		    },
		    {
		      "$oid": "000000000000000000000000"
		    }
		  ],
		  "bool": [
		    true,
		    false
		  ],
		  "datetime": [
		    {
		      "$date": {
		        "$numberLong": "1627378542123"
		      }
		    },
		    {
		      "$date": {
		        "$numberLong": "-62135596800000"
		      }
		    }
		  ],
		  "null": [
		    null
		  ],
		  "regex": [
		    {
		      "$regularExpression": {
		        "pattern": "p",
		        "options": "o"
		      }
		    },
		    {
		      "$regularExpression": {
		        "pattern": "",
		        "options": ""
		      }
		    }
		  ],
		  "int32": [
		    {
		      "$numberInt": "42"
		    },
		    {
		      "$numberInt": "0"
		    }
		  ],
		  "timestamp": [
		    {
		      "$timestamp": {
		        "t": 0,
		        "i": 42
		      }
		    },
		    {
		      "$timestamp": {
		        "t": 0,
		        "i": 0
		      }
		    }
		  ],
		  "int64": [
		    {
		      "$numberLong": "42"
		    },
		    {
		      "$numberLong": "0"
		    }
		  ],
		  "decimal128": [
		    {
		      "$numberDecimal": "2.39807672958224171050E-6156"
		    },
		    {
		      "$numberDecimal": "0E-6176"
		    }
		  ]
		}`,
	},
	{
		name: "nested",
		raw:  testutil.MustParseDumpFile("testdata", "nested.hex"),
		doc:  makeNested(false, 150).(*Document),
		mi: `
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
		                                        {...},
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
		mi: `
		{
		  "f": 3.141592653589793,
		}`,
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
		mi: `
		{
		  "f": "v",
		}`,
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
		mi: `
		{
		  "f": Binary(user:dg==),
		}`,
	},
	{
		name: "undefinedDoc",
		raw: RawDocument{
			0x08, 0x00, 0x00, 0x00,
			0x06, 0x66, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", Undefined,
		),
		mi: `
		{
		  "f": undefined,
		}`,
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
		mi: `
		{
		  "f": ObjectID(6256c5ba182d4454fb210940),
		}`,
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
		mi: `
		{
		  "f": true,
		}`,
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
		mi: `
		{
		  "f": 2024-01-17T17:40:42.123Z,
		}`,
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
		mi: `
		{
		  "f": null,
		}`,
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
		mi: `
		{
		  "f": /p/o,
		}`,
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
		mi: `
		{
		  "f": 314159265,
		}`,
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
		mi: `
		{
		  "f": Timestamp(42),
		}`,
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
		mi: `
		{
		  "f": int64(3141592653589793),
		}`,
	},
	{
		name: "decimal128Doc",
		raw: RawDocument{
			0x18, 0x00, 0x00, 0x00,
			0x13, 0x66, 0x00,
			0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00,
		},
		doc: MustDocument(
			"f", Decimal128{H: 13, L: 42},
		),
		mi: `
		{
		  "f": Decimal128(H:13,L:42),
		}`,
		j: `
		{
		  "f": {
		    "$numberDecimal": "2.39807672958224171050E-6156"
		  }
		}`,
	},
	{
		name: "decimal128DocPrec",
		raw: RawDocument{
			0x18, 0x00, 0x00, 0x00,
			0x13, 0x66, 0x00,
			0x30, 0x30, 0x31, 0x30, 0x30, 0x31, 0x30, 0x30,
			0x31, 0x31, 0x30, 0x31, 0x30, 0xff, 0x31, 0x30,
			0x00,
		},
		doc: MustDocument(
			"f", Decimal128{H: 3472837370128118065, L: 3472329395739373616},
		),
		mi: `
		{
		  "f": Decimal128(H:3472837370128118065,L:3472329395739373616),
		}`,
		j: `
		{
		  "f": {
		    "$numberDecimal": "103681294822929121827017235.39812400"
		  }
		}`,
	},
	{
		name: "emptyDoc",
		raw: RawDocument{
			0x05, 0x00, 0x00, 0x00, // document length
			0x00, // end of document
		},
		doc: MustDocument(),
		mi:  `{}`,
		j:   `{}`,
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
		mi: `
		{
		  "foo": {},
		}`,
		j: `
		{
		  "foo": {}
		}`,
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
		mi: `
		{
		  "foo": [],
		}`,
		j: `
		{
		  "foo": []
		}`,
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
		mi: `
		{
		  "": false,
		  "": true,
		}`,
		j: `
		{
		  "": false,
		  "": true
		}`,
	},
	{
		name: "RegexEscape", // https://jira.mongodb.org/browse/GODRIVER-3476
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x0b, 0x22, 0x60, 0x00, 0x22, 0x60, 0x00, 0x22, 0x60, 0x00, // weird regex
			0x00, // end of document
		},
		doc: MustDocument(
			`"`+"`", Regex{Pattern: `"` + "`", Options: `"` + "`"},
		),
		mi: "{\n  \"\\\"`\": /\"`/\"`,\n}",
		j:  "{\n  \"\\\"`\": {\n    \"$regularExpression\": {\n      \"pattern\": \"\\\"`\",\n      \"options\": \"\\\"`\"\n    }\n  }\n}",
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

				assert.NotEmpty(t, tc.raw.LogMessage())
				assert.NotEmpty(t, tc.raw.LogMessageIndent())

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

				assert.NotEmpty(t, doc.LogMessage())
				assert.NotEmpty(t, doc.LogMessageIndent())

				raw, err := doc.Encode()
				require.NoError(t, err)
				assert.Equal(t, tc.raw, raw)
			})

			t.Run("DecodeDeepEncode", func(t *testing.T) {
				doc, err := tc.raw.DecodeDeep()
				require.NoError(t, err)
				assertEqual(t, tc.doc, doc)

				ls := doc.LogValue().Resolve().String()
				assert.NotContains(t, ls, "panicked")
				assert.NotContains(t, ls, "called too many times")

				assert.NotEmpty(t, doc.LogMessage())
				require.NotEmpty(t, tc.mi)
				mi := testutil.Unindent(tc.mi)
				if !strings.Contains(tc.mi, "`") {
					mi = strings.ReplaceAll(mi, `"`, "`")
				}
				assert.Equal(t, mi, doc.LogMessageIndent())

				raw, err := tc.doc.Encode()
				require.NoError(t, err)
				assert.Equal(t, tc.raw, raw, "actual:\n"+hex.Dump(raw))
			})

			t.Run("MarshalUnmarshal", func(t *testing.T) {
				// We should set all tc.j and remove this Skip.
				// TODO https://github.com/FerretDB/wire/issues/49
				if tc.j == "" {
					t.Skip("https://github.com/FerretDB/wire/issues/49")
				}

				b, err := json.MarshalIndent(tc.doc, "", "  ")
				require.NoError(t, err)
				assert.Equal(t, testutil.Unindent(tc.j), string(b))

				var doc *Document
				err = json.Unmarshal([]byte(tc.j), &doc)
				require.NoError(t, err)

				// TODO https://github.com/FerretDB/wire/issues/49
				// https://jira.mongodb.org/browse/GODRIVER-3531
				if strings.Contains(tc.j, `$numberDecimal`) {
					t.Skip("https://github.com/FerretDB/wire/issues/49")
				}

				assertEqual(t, tc.doc, doc)
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

				assert.NotEmpty(t, tc.raw.LogMessage())
				assert.NotEmpty(t, tc.raw.LogMessageIndent())

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

func TestJSONNull(t *testing.T) {
	var doc *Document
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	require.Equal(t, "null", string(b))

	err = json.Unmarshal(b, &doc)
	require.NoError(t, err)
	assert.Nil(t, doc)
}

var drain any

func BenchmarkDocumentDecode(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()

				var err error
				for range b.N {
					drain, err = tc.raw.Decode()
				}

				b.StopTimer()

				require.NoError(b, err)
				require.NotNil(b, drain)
			})
		}
	}
}

func BenchmarkDocumentDecodeDeep(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()

				var err error
				for range b.N {
					drain, err = tc.raw.DecodeDeep()
				}

				b.StopTimer()

				require.NoError(b, err)
				require.NotNil(b, drain)
			})
		}
	}
}

func BenchmarkDocumentEncode(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain, err = doc.Encode()
				}

				b.StopTimer()

				require.NoError(b, err)
				assert.NotNil(b, drain)
			})
		}
	}
}

func BenchmarkDocumentEncodeDeep(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain, err = doc.Encode()
				}

				b.StopTimer()

				require.NoError(b, err)
				assert.NotNil(b, drain)
			})
		}
	}
}

func BenchmarkDocumentLogValue(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain = doc.LogValue().Resolve().String()
				}

				b.StopTimer()

				assert.NotEmpty(b, drain)
				assert.NotContains(b, drain, "panicked")
				assert.NotContains(b, drain, "called too many times")
			})
		}
	}
}

func BenchmarkDocumentLogValueDeep(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain = doc.LogValue().Resolve().String()
				}

				b.StopTimer()

				assert.NotEmpty(b, drain)
				assert.NotContains(b, drain, "panicked")
				assert.NotContains(b, drain, "called too many times")
			})
		}
	}
}

func BenchmarkDocumentLogMessage(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain = doc.LogMessage()
				}

				b.StopTimer()

				assert.NotEmpty(b, drain)
			})
		}
	}
}

func BenchmarkDocumentLogMessageDeep(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain = doc.LogMessage()
				}

				b.StopTimer()

				assert.NotEmpty(b, drain)
			})
		}
	}
}

func BenchmarkDocumentLogMessageIndent(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.Decode()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain = doc.LogMessageIndent()
				}

				b.StopTimer()

				assert.NotEmpty(b, drain)
			})
		}
	}
}

func BenchmarkDocumentLogMessageIndentDeep(b *testing.B) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			b.Run(tc.name, func(b *testing.B) {
				doc, err := tc.raw.DecodeDeep()
				require.NoError(b, err)

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					drain = doc.LogMessageIndent()
				}

				b.StopTimer()

				assert.NotEmpty(b, drain)
			})
		}
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

		assert.NotEmpty(t, rawDoc.LogMessage())
		assert.NotEmpty(t, rawDoc.LogMessageIndent())

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

		assert.NotEmpty(t, doc.LogMessage())
		assert.NotEmpty(t, doc.LogMessageIndent())

		raw, err := doc.Encode()
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

		assert.NotEmpty(t, doc.LogMessage())
		assert.NotEmpty(t, doc.LogMessageIndent())

		raw, err := doc.Encode()
		require.NoError(t, err)
		assert.Equal(t, rawDoc, raw)
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		doc, err := rawDoc.DecodeDeep()
		if err != nil {
			return
		}

		b, err := json.Marshal(doc)
		j := string(b)
		d, _ := ToDriver(doc)
		require.NoError(t, err, "%s\n%#v", doc.LogMessage(), d)

		var doc2 *Document
		err = json.Unmarshal(b, &doc2)
		if err != nil {
			if strings.Contains(err.Error(), "$invalid $numberDecimal string") {
				// TODO https://github.com/FerretDB/wire/issues/49
				// See https://jira.mongodb.org/browse/GODRIVER-3531
				t.Skip()
			}
		}

		require.NoError(t, err, "%s\n%s", doc.LogMessage(), b)

		// invalid UTF-8 bytes can't survive marshaling/unmarshaling
		if strings.Contains(j, `\ufffd`) { // Unicode replacement rune
			t.Skip()
		}

		// TODO https://github.com/FerretDB/wire/issues/49
		// https://jira.mongodb.org/browse/GODRIVER-3531
		if strings.Contains(j, `$numberDecimal`) {
			t.Skip()
		}

		assertEqual(t, doc, doc2)
	})
}

func FuzzDocument(f *testing.F) {
	for _, tc := range normalTestCases {
		if tc.raw != nil {
			f.Add([]byte(tc.raw))
		}
	}

	for _, tc := range decodeTestCases {
		f.Add([]byte(tc.raw))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		rawDoc := RawDocument(b)

		t.Run("TestRawDocument", func(t *testing.T) {
			testRawDocument(t, rawDoc)
		})
	})
}

// assertEqual asserts that two BSON values are equal.
// It is copied from the wiretest package to avoid a circular dependency.
func assertEqual(tb testing.TB, expected, actual any) bool {
	tb.Helper()

	if assert.True(tb, Equal(expected, actual)) {
		return true
	}

	expectedS := LogMessageIndent(expected)
	actualS := LogMessageIndent(actual)

	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expectedS),
		FromFile: "expected",
		B:        difflib.SplitLines(actualS),
		ToFile:   "actual",
		Context:  1,
	})
	require.NoError(tb, err)

	msg := fmt.Sprintf("Not equal:\n\nexpected:\n%s\n\nactual:\n%s\n\ndiff:\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}
