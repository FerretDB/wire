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
	"encoding/binary"
	"encoding/json"
	"iter"
	"log/slog"
	"slices"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/internal/util/must"
)

// field represents a single Document field in the (partially) decoded form.
type field struct {
	value any
	name  string
}

// Document represents a BSON document a.k.a object in the (partially) decoded form.
//
// It may contain duplicate field names.
type Document struct {
	fields []field
	frozen bool
}

// NewDocument creates a new Document from the given pairs of field names and values.
func NewDocument(pairs ...any) (*Document, error) {
	l := len(pairs)
	if l%2 != 0 {
		return nil, lazyerrors.Errorf("invalid number of arguments: %d", l)
	}

	res := MakeDocument(l / 2)

	for i := 0; i < l; i += 2 {
		name, ok := pairs[i].(string)
		if !ok {
			return nil, lazyerrors.Errorf("invalid field name type: %T", pairs[i])
		}

		value := pairs[i+1]

		if err := res.Add(name, value); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return res, nil
}

// MustDocument is a variant of [NewDocument] that panics on error.
func MustDocument(pairs ...any) *Document {
	res, err := NewDocument(pairs...)
	if err != nil {
		panic(err)
	}

	return res
}

// MakeDocument creates a new empty Document with the given capacity.
func MakeDocument(cap int) *Document {
	return &Document{
		fields: make([]field, 0, cap),
	}
}

// Freeze prevents Document from further field modifications.
// Any methods that would modify Document fields will panic.
//
// It is safe to call Freeze multiple times.
func (doc *Document) Freeze() {
	doc.frozen = true
}

// checkFrozen panics if Document is frozen.
func (doc *Document) checkFrozen() {
	if doc.frozen {
		panic("document is frozen and can't be modified")
	}
}

// Len returns the number of fields in the Document.
func (doc *Document) Len() int {
	return len(doc.fields)
}

// FieldNames returns a slice of field names in the Document in the original order.
//
// If Document contains duplicate field names, the same name will appear multiple times.
func (doc *Document) FieldNames() []string {
	fields := make([]string, len(doc.fields))
	for i, f := range doc.fields {
		fields[i] = f.name
	}

	return fields
}

// Get returns a value of the field with the given name.
//
// It returns nil if the field is not found.
//
// If Document contains duplicate field names, it returns the first one.
// To get other fields, a for/range loop can be used with [Document.All],
// or [Document.Len] with [Document.GetByIndex].
func (doc *Document) Get(name string) any {
	for _, f := range doc.fields {
		if f.name == name {
			return f.value
		}
	}

	return nil
}

// GetByIndex returns the name and the value of the field at the given index (between 0 and [Document.Len]-1).
// It panics if index is out of bounds.
func (doc *Document) GetByIndex(i int) (string, any) {
	f := doc.fields[i]
	return f.name, f.value
}

// All returns an iterator over all field name/value pairs of the Document.
func (doc *Document) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for _, f := range doc.fields {
			if !yield(f.name, f.value) {
				return
			}
		}
	}
}

// Fields returns an iterator over all field names of the Document.
func (doc *Document) Fields() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, f := range doc.fields {
			if !yield(f.name) {
				return
			}
		}
	}
}

// Add adds a new field to the end of the Document.
func (doc *Document) Add(name string, value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Errorf("%q: %w", name, err)
	}

	doc.checkFrozen()

	doc.fields = append(doc.fields, field{
		name:  name,
		value: value,
	})

	return nil
}

// Remove removes the first existing field with the given name.
// It does nothing if the field with that name does not exist.
func (doc *Document) Remove(name string) {
	doc.checkFrozen()

	var found bool
	doc.fields = slices.DeleteFunc(doc.fields, func(f field) bool {
		if f.name != name {
			return false
		}

		if found {
			return false
		}

		found = true

		return true
	})
}

// Replace sets the value for the first existing field with the given name.
// It does nothing if the field with that name does not exist.
func (doc *Document) Replace(name string, value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Errorf("%q: %w", name, err)
	}

	doc.checkFrozen()

	for i, f := range doc.fields {
		if f.name == name {
			doc.fields[i].value = value
			return nil
		}
	}

	return nil
}

// Command returns the first field name. This is often used as a command name.
// It returns an empty string if Document is nil or empty.
func (doc *Document) Command() string {
	if doc == nil || len(doc.fields) == 0 {
		return ""
	}

	return doc.fields[0].name
}

// Encode encodes non-nil Document.
//
// TODO https://github.com/FerretDB/wire/issues/21
// This method should accept a slice of bytes, not return it.
// That would allow to avoid unnecessary allocations.
func (doc *Document) Encode() (RawDocument, error) {
	must.NotBeZero(doc)

	size := sizeDocument(doc)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, f := range doc.fields {
		if err := encodeField(buf, f.name, f.value); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

// MarshalJSON implements the json.Marshaler interface for Document.
// It converts the Document into a JSON object representation while preserving the order of fields.
func (doc *Document) MarshalJSON() ([]byte, error) {
	if doc == nil {
		return nil, lazyerrors.Errorf("nil Document")
	}

	jsonObject := make([]byte, 0)
	jsonObject = append(jsonObject, '{')

	for i, field := range doc.fields {
		key, err := json.Marshal(field.name)
		if err != nil {
			return nil, lazyerrors.Errorf("failed to marshal key: %w", err)
		}

		value, err := json.Marshal(field.value)
		if err != nil {
			return nil, lazyerrors.Errorf("failed to marshal value: %w", err)
		}

		jsonObject = append(jsonObject, key...)
		jsonObject = append(jsonObject, ':')
		jsonObject = append(jsonObject, value...)

		if i < len(doc.fields)-1 {
			jsonObject = append(jsonObject, ',')
		}
	}

	jsonObject = append(jsonObject, '}')

	return jsonObject, nil
}

// Decode returns itself to implement [AnyDocument].
//
// Receiver must not be nil.
func (doc *Document) Decode() (*Document, error) {
	must.NotBeZero(doc)
	return doc, nil
}

// LogValue implements [slog.LogValuer].
func (doc *Document) LogValue() slog.Value {
	return slogValue(doc, 1)
}

// LogMessage implements [AnyDocument].
func (doc *Document) LogMessage() string {
	return LogMessage(doc)
}

// LogMessageIndent implements [AnyDocument].
func (doc *Document) LogMessageIndent() string {
	return LogMessageIndent(doc)
}

// check interfaces
var (
	_ AnyDocument    = (*Document)(nil)
	_ slog.LogValuer = (*Document)(nil)
	_ json.Marshaler = (*Document)(nil)
)
