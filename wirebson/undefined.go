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

// UndefinedType represents BSON scalar type undefined.
type UndefinedType struct{}

// Undefined represents BSON scalar value undefined.
//
// Its usage is deprecated, but it is still used in a few places.
// See https://github.com/FerretDB/FerretDB/issues/2286 for an example.
var Undefined = UndefinedType{}
