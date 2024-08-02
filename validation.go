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
	"github.com/FerretDB/wire/wirebson"
	"math"
)

// validateNan returns error if float Nan was encountered.
func validateNan(v any) error {
	switch v := v.(type) {
	case *wirebson.Document:
		for _, f := range v.FieldNames() {
			if err := validateNan(v.Get(f)); err != nil {
				return err
			}
		}

	case *wirebson.Array:
		for i := range v.Len() {
			if err := validateNan(v.Get(i)); err != nil {
				return err
			}
		}

	case float64:
		if math.IsNaN(v) {
			return errors.New("NaN is not supported")
		}
	}

	return nil
}
