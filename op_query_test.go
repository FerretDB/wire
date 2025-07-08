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

package wire

import (
	"testing"

	"github.com/FerretDB/wire/internal/util/testutil"
	"github.com/FerretDB/wire/wirebson"
)

var queryTestCases = []testCase{
	{
		name:    "handshake1",
		headerB: testutil.MustParseDumpFile("testdata", "handshake1_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake1_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 372,
			RequestID:     1,
			ResponseTo:    0,
			OpCode:        OpCodeQuery,
		},
		msgBody: &OpQuery{
			Flags:              0,
			FullCollectionName: "admin.$cmd",
			NumberToSkip:       0,
			NumberToReturn:     -1,
			query: makeRawDocument(
				"ismaster", true,
				"client", wirebson.MustDocument(
					"driver", wirebson.MustDocument(
						"name", "nodejs",
						"version", "4.0.0-beta.6",
					),
					"os", wirebson.MustDocument(
						"type", "Darwin",
						"name", "darwin",
						"architecture", "x64",
						"version", "20.6.0",
					),
					"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
					"application", wirebson.MustDocument(
						"name", "mongosh 1.0.1",
					),
				),
				"compression", wirebson.MustArray("none"),
				"loadBalanced", false,
			),
			returnFieldsSelector: nil,
		},
		si: `
		{
		  "Flags": "[]",
		  "FullCollectionName": "admin.$cmd",
		  "NumberToSkip": 0,
		  "NumberToReturn": -1,
		  "Query": {
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
		  },
		}`,
	},
	{
		name:    "handshake3",
		headerB: testutil.MustParseDumpFile("testdata", "handshake3_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake3_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 372,
			RequestID:     2,
			ResponseTo:    0,
			OpCode:        OpCodeQuery,
		},
		msgBody: &OpQuery{
			Flags:              0,
			FullCollectionName: "admin.$cmd",
			NumberToSkip:       0,
			NumberToReturn:     -1,
			query: makeRawDocument(
				"ismaster", true,
				"client", wirebson.MustDocument(
					"driver", wirebson.MustDocument(
						"name", "nodejs",
						"version", "4.0.0-beta.6",
					),
					"os", wirebson.MustDocument(
						"type", "Darwin",
						"name", "darwin",
						"architecture", "x64",
						"version", "20.6.0",
					),
					"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
					"application", wirebson.MustDocument(
						"name", "mongosh 1.0.1",
					),
				),
				"compression", wirebson.MustArray("none"),
				"loadBalanced", false,
			),
			returnFieldsSelector: nil,
		},
		si: `
		{
		  "Flags": "[]",
		  "FullCollectionName": "admin.$cmd",
		  "NumberToSkip": 0,
		  "NumberToReturn": -1,
		  "Query": {
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
		  },
		}`,
	},
}

func TestQuery(t *testing.T) {
	t.Parallel()
	testMessages(t, queryTestCases)
}

func FuzzQuery(f *testing.F) {
	fuzzMessages(f, queryTestCases)
}

func TestOpQuerySize(t *testing.T) {
	t.Parallel()

	// Test that Size() returns the same value as len(MarshalBinary())
	for _, tc := range queryTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Skip test cases that have nil msgBody (error cases)
			if tc.msgBody == nil {
				t.Skip("Skipping test case with nil msgBody")
			}

			query := tc.msgBody.(*OpQuery)
			
			// Get size from current Size() method
			size := query.Size()
			
			// Get size from MarshalBinary()
			data, err := query.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary failed: %v", err)
			}
			expectedSize := len(data)
			
			if size != expectedSize {
				t.Errorf("Size() = %d, len(MarshalBinary()) = %d", size, expectedSize)
			}
		})
	}
}
