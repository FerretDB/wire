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

// Package wire provides [MongoDB wire protocol] implementation.
//
// [MongoDB wire protocol]: https://www.mongodb.com/docs/manual/reference/mongodb-wire-protocol/
package wire

//go:generate ./bin/stringer -linecomment -output stringers.go -type OpCode,OpMsgFlagBit,OpQueryFlagBit,OpReplyFlagBit

// Debug set to true performs additional slow checks during encoding/decoding that are not normally required.
// It is exposed mainly to simplify testing.
var Debug bool

// CheckNaNs set to true returns an error if float64 NaN value is present in wire messages.
//
// TODO https://github.com/FerretDB/wire/issues/73
var CheckNaNs bool
