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
	"encoding/binary"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/internal/util/must"
	"github.com/FerretDB/wire/wirebson"
)

// OpQuery represents the deprecated OP_QUERY wire protocol message type.
// It stores BSON documents in the raw form.
//
// Message is checked during construction by [NewOpQuery], [MustOpQuery], or [OpQuery.UnmarshalBinaryNocopy]
// without decoding BSON documents inside.
type OpQuery struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, collection name, number to skip, number to return, query, fields selector.

	FullCollectionName   string
	query                wirebson.RawDocument
	returnFieldsSelector wirebson.RawDocument
	Flags                OpQueryFlags
	NumberToSkip         int32
	NumberToReturn       int32
}

// NewOpQuery creates a new OpQuery message.
func NewOpQuery(doc wirebson.AnyDocument) (*OpQuery, error) {
	raw, err := doc.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	query := &OpQuery{
		query: raw,
	}

	if Debug {
		if err = query.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return query, nil
}

// MustOpQuery creates a new OpQuery message constructed from the given pairs of field names and values.
// It panics on error.
func MustOpQuery(pairs ...any) *OpQuery {
	query, err := NewOpQuery(wirebson.MustDocument(pairs...))
	if err != nil {
		panic(err)
	}

	return query
}

func (query *OpQuery) msgbody() {}

// check implements [MsgBody].
func (query *OpQuery) check() error {
	if d := query.query; d != nil {
		if _, err := d.DecodeDeep(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	if s := query.returnFieldsSelector; s != nil {
		if _, err := s.DecodeDeep(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// UnmarshalBinaryNocopy implements [MsgBody].
func (query *OpQuery) UnmarshalBinaryNocopy(b []byte) error {
	if len(b) < 4 {
		return lazyerrors.Errorf("len=%d", len(b))
	}

	query.Flags = OpQueryFlags(binary.LittleEndian.Uint32(b[0:4]))

	var err error

	query.FullCollectionName, err = wirebson.DecodeCString(b[4:])
	if err != nil {
		return lazyerrors.Error(err)
	}

	numberLow := 4 + wirebson.SizeCString(query.FullCollectionName)
	if len(b) < numberLow+8 {
		return lazyerrors.Errorf("len=%d, can't unmarshal numbers", len(b))
	}

	query.NumberToSkip = int32(binary.LittleEndian.Uint32(b[numberLow : numberLow+4]))
	query.NumberToReturn = int32(binary.LittleEndian.Uint32(b[numberLow+4 : numberLow+8]))

	l, err := wirebson.FindRaw(b[numberLow+8:])
	if err != nil {
		return lazyerrors.Error(err)
	}
	query.query = b[numberLow+8 : numberLow+8+l]

	selectorLow := numberLow + 8 + l
	if len(b) != selectorLow {
		l, err = wirebson.FindRaw(b[selectorLow:])
		if err != nil {
			return lazyerrors.Error(err)
		}

		if len(b) != selectorLow+l {
			return lazyerrors.Errorf("len=%d, expected=%d", len(b), selectorLow+l)
		}
		query.returnFieldsSelector = b[selectorLow:]
	}

	if Debug {
		if err = query.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// Size implements [MsgBody].
func (query *OpQuery) Size() int {
	nameSize := wirebson.SizeCString(query.FullCollectionName)
	return 12 + nameSize + len(query.query) + len(query.returnFieldsSelector)
}

// MarshalBinary implements [MsgBody].
func (query *OpQuery) MarshalBinary() ([]byte, error) {
	nameSize := wirebson.SizeCString(query.FullCollectionName)
	b := make([]byte, 12+nameSize+len(query.query)+len(query.returnFieldsSelector))

	binary.LittleEndian.PutUint32(b[0:4], uint32(query.Flags))

	nameHigh := 4 + nameSize
	wirebson.EncodeCString(b[4:nameHigh], query.FullCollectionName)

	binary.LittleEndian.PutUint32(b[nameHigh:nameHigh+4], uint32(query.NumberToSkip))
	binary.LittleEndian.PutUint32(b[nameHigh+4:nameHigh+8], uint32(query.NumberToReturn))

	queryHigh := nameHigh + 8 + len(query.query)
	copy(b[nameHigh+8:queryHigh], query.query)
	copy(b[queryHigh:], query.returnFieldsSelector)

	return b, nil
}

// Query returns decoded query document.
// It may be shallowly or deeply decoded.
func (query *OpQuery) Query() (*wirebson.Document, error) {
	doc, err := query.query.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// QueryDeep returns deeply decoded query document.
func (query *OpQuery) QueryDeep() (*wirebson.Document, error) {
	doc, err := query.query.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// QueryRaw returns raw query (that might be nil).
func (query *OpQuery) QueryRaw() wirebson.RawDocument {
	return query.query
}

// ReturnFieldsSelector returns decoded returnFieldsSelector document, or nil.
// It may be shallowly or deeply decoded.
func (query *OpQuery) ReturnFieldsSelector() (*wirebson.Document, error) {
	if query.returnFieldsSelector == nil {
		return nil, nil
	}

	doc, err := query.returnFieldsSelector.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// ReturnFieldsSelectorDeep returns decoded returnFieldsSelector document, or nil.
func (query *OpQuery) ReturnFieldsSelectorDeep() (*wirebson.Document, error) {
	if query.returnFieldsSelector == nil {
		return nil, nil
	}

	doc, err := query.returnFieldsSelector.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// ReturnFieldsSelectorRaw returns raw returnFieldsSelector (that might be nil).
func (query *OpQuery) ReturnFieldsSelectorRaw() wirebson.RawDocument {
	return query.returnFieldsSelector
}

// logMessage returns a string representation for logging.
func (query *OpQuery) logMessage(logFunc func(v any) string) string {
	if query == nil {
		return "<nil>"
	}

	m := wirebson.MustDocument(
		"Flags", query.Flags.String(),
		"FullCollectionName", query.FullCollectionName,
		"NumberToSkip", query.NumberToSkip,
		"NumberToReturn", query.NumberToReturn,
	)

	doc, err := query.QueryDeep()
	if err == nil {
		must.NoError(m.Add("Query", doc))
	} else {
		must.NoError(m.Add("QueryError", err.Error()))
	}

	doc, err = query.ReturnFieldsSelectorDeep()
	if err == nil {
		if doc != nil {
			must.NoError(m.Add("ReturnFieldsSelector", doc))
		}
	} else {
		must.NoError(m.Add("ReturnFieldsSelectorError", err.Error()))
	}

	return logFunc(m)
}

// String returns an string representation for logging.
func (query *OpQuery) String() string {
	return query.logMessage(wirebson.LogMessage)
}

// StringIndent returns an indented string representation for logging.
func (query *OpQuery) StringIndent() string {
	return query.logMessage(wirebson.LogMessageIndent)
}

// check interfaces
var (
	_ MsgBody = (*OpQuery)(nil)
)
