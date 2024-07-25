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

package testutil

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
)

// parseDump decodes from hex dump to the byte array.
func parseDump(s string) ([]byte, error) {
	var res []byte

	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(s)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line[len(line)-1] == '|' {
			// go dump
			line = strings.TrimSpace(line[8:60])
			line = strings.Join(strings.Split(line, " "), "")
		} else {
			// wireshark dump
			line = strings.TrimSpace(line[7:54])
			line = strings.Join(strings.Split(line, " "), "")
		}

		b, err := hex.DecodeString(line)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		res = append(res, b...)
	}

	if err := scanner.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// MustParseDumpFile panics if fails to parse file input to byte array.
func MustParseDumpFile(path ...string) []byte {
	b, err := os.ReadFile(filepath.Join(path...))
	if err != nil {
		panic(err)
	}

	b, err = parseDump(string(b))
	if err != nil {
		panic(err)
	}

	return b
}

// Unindent removes the common number of leading tabs from all lines in s.
func Unindent(s string) string {
	if s == "" {
		panic("input must not be empty")
	}

	parts := strings.Split(s, "\n")
	if len(parts) == 0 {
		panic("zero parts")
	}

	if parts[0] == "" {
		parts = parts[1:]
	}

	indent := len(parts[0]) - len(strings.TrimLeft(parts[0], "\t"))
	if indent < 0 {
		panic("invalid indent")
	}

	for i, l := range parts {
		if len(l) <= indent {
			panic(fmt.Sprintf("invalid indent on line %q", l))
		}

		parts[i] = l[indent:]
	}

	return strings.Join(parts, "\n")
}
