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
	"errors"
	"math"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
)

// checkNaN returns error if float64 NaN is encountered.
func checkNaN(v any) error {
	switch v := v.(type) {
	case *wirebson.Document:
		for _, f := range v.FieldNames() {
			if err := checkNaN(v.Get(f)); err != nil {
				return err
			}
		}

	case wirebson.RawDocument:
		doc, err := v.Decode()
		if err != nil {
			return lazyerrors.Error(err)
		}

		return checkNaN(doc)

	case *wirebson.Array:
		for i := range v.Len() {
			if err := checkNaN(v.Get(i)); err != nil {
				return err
			}
		}

	case wirebson.RawArray:
		arr, err := v.Decode()
		if err != nil {
			return lazyerrors.Error(err)
		}

		return checkNaN(arr)

	case float64:
		if math.IsNaN(v) {
			return errors.New("NaN is not supported")
		}
	}

	return nil
}
