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
	"fmt"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/internal/util/must"
	"github.com/FerretDB/wire/wirebson"
)

// OpMsg is the main wire protocol message type.
type OpMsg struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, sections, optional checksum.

	sections []opMsgSection
	Flags    OpMsgFlags
	checksum uint32
}

// NewOpMsg creates a message with a single section of kind 0 with a single document.
func NewOpMsg(doc wirebson.AnyDocument) (*OpMsg, error) {
	raw, err := doc.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sections := []opMsgSection{{documents: []wirebson.RawDocument{raw}}}
	if err = checkSections(sections); err != nil {
		return nil, lazyerrors.Error(err)
	}

	msg := OpMsg{
		sections: sections,
	}

	if Debug || CheckNaNs {
		if err = msg.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return &msg, nil
}

// MustOpMsg creates a message with a single section of kind 0 with a single document
// constructed from the given pairs of field names and values.
// It panics on error.
func MustOpMsg(pairs ...any) *OpMsg {
	msg, err := NewOpMsg(wirebson.MustDocument(pairs...))
	if err != nil {
		panic(err)
	}

	return msg
}

func (msg *OpMsg) msgbody() {}

// check implements [MsgBody].
func (msg *OpMsg) check() error {
	if err := checkSections(msg.sections); err != nil {
		return err
	}

	for _, s := range msg.sections {
		for _, d := range s.documents {
			doc, err := d.DecodeDeep()
			if err != nil {
				return lazyerrors.Error(err)
			}

			if !CheckNaNs {
				continue
			}

			if err = checkNaN(doc); err != nil {
				return err
			}
		}
	}

	return nil
}

// UnmarshalBinaryNocopy implements [MsgBody].
func (msg *OpMsg) UnmarshalBinaryNocopy(b []byte) error {
	if len(b) < 6 {
		return lazyerrors.Errorf("len=%d", len(b))
	}

	msg.Flags = OpMsgFlags(binary.LittleEndian.Uint32(b[0:4]))

	offset := 4

	for {
		var section opMsgSection
		section.kind = b[offset]
		offset++

		switch section.kind {
		case 0:
			l, err := wirebson.FindRaw(b[offset:])
			if err != nil {
				return lazyerrors.Error(err)
			}

			section.documents = []wirebson.RawDocument{b[offset : offset+l]}
			offset += l

		case 1:
			if len(b) < offset+4 {
				return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
			}

			secSize := int(binary.LittleEndian.Uint32(b[offset:offset+4])) - 4
			if secSize < 5 {
				return lazyerrors.Errorf("size = %d", secSize)
			}

			offset += 4

			var err error

			if len(b) < offset {
				return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
			}

			section.identifier, err = wirebson.DecodeCString(b[offset:])
			if err != nil {
				return lazyerrors.Error(err)
			}

			offset += wirebson.SizeCString(section.identifier)
			secSize -= wirebson.SizeCString(section.identifier)

			for secSize != 0 {
				if secSize < 0 {
					return lazyerrors.Errorf("size = %d", secSize)
				}

				if len(b) < offset {
					return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
				}

				var l int
				if l, err = wirebson.FindRaw(b[offset:]); err != nil {
					return lazyerrors.Error(err)
				}

				section.documents = append(section.documents, b[offset:offset+l])
				offset += l
				secSize -= l
			}

		default:
			return lazyerrors.Errorf("kind is %d", section.kind)
		}

		msg.sections = append(msg.sections, section)

		if msg.Flags.FlagSet(OpMsgChecksumPresent) {
			if offset == len(b)-4 {
				break
			}
		} else {
			if offset == len(b) {
				break
			}
		}
	}

	if msg.Flags.FlagSet(OpMsgChecksumPresent) {
		// Move checksum validation here. It needs header data to be available.
		// TODO https://github.com/FerretDB/FerretDB/issues/2690
		msg.checksum = binary.LittleEndian.Uint32(b[offset:])
	}

	if err := checkSections(msg.sections); err != nil {
		return lazyerrors.Error(err)
	}

	if Debug || CheckNaNs {
		if err := msg.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// MarshalBinary writes an OpMsg to a byte array.
func (msg *OpMsg) MarshalBinary() ([]byte, error) {
	if Debug {
		if err := msg.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	b := make([]byte, 4, 16)

	binary.LittleEndian.PutUint32(b, uint32(msg.Flags))

	for _, section := range msg.sections {
		b = append(b, section.kind)

		switch section.kind {
		case 0:
			b = append(b, section.documents[0]...)

		case 1:
			sec := make([]byte, wirebson.SizeCString(section.identifier))
			wirebson.EncodeCString(sec, section.identifier)

			for _, doc := range section.documents {
				sec = append(sec, doc...)
			}

			var size [4]byte
			binary.LittleEndian.PutUint32(size[:], uint32(len(sec)+4))
			b = append(b, size[:]...)
			b = append(b, sec...)

		default:
			return nil, lazyerrors.Errorf("kind is %d", section.kind)
		}
	}

	if msg.Flags.FlagSet(OpMsgChecksumPresent) {
		// Calculate checksum before writing it. It needs header data to be ready and available here.
		// TODO https://github.com/FerretDB/FerretDB/issues/2690
		var checksum [4]byte
		binary.LittleEndian.PutUint32(checksum[:], msg.checksum)
		b = append(b, checksum[:]...)
	}

	return b, nil
}

// Document returns the value of msg as decoded frozen [*wirebson.Document].
// Only top-level fields are decoded.
//
// The error is returned if msg contains anything other than a single section of kind 0
// with a single document.
func (msg *OpMsg) Document() (*wirebson.Document, error) {
	raw, err := msg.DocumentRaw()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := raw.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc.Freeze()
	return doc, nil
}

// DocumentDeep returns the value of msg as deeply decoded frozen [*wirebson.Document].
//
// The error is returned if msg contains anything other than a single section of kind 0
// with a single document.
func (msg *OpMsg) DocumentDeep() (*wirebson.Document, error) {
	raw, err := msg.DocumentRaw()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := raw.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc.Freeze()
	return doc, nil
}

// DocumentRaw returns the value of msg as a [wirebson.DocumentRaw].
//
// The error is returned if msg contains anything other than a single section of kind 0
// with a single document.
func (msg *OpMsg) DocumentRaw() (wirebson.RawDocument, error) {
	if l := len(msg.sections); l != 1 {
		return nil, lazyerrors.Errorf("expected 1 section, got %d", l)
	}

	s := msg.sections[0]
	if s.kind != 0 || s.identifier != "" {
		return nil, lazyerrors.Errorf(`expected section 0/"", got %d/%q`, s.kind, s.identifier)
	}

	return s.documents[0], nil
}

// Deprecated: use DocumentRaw instead.
func (msg *OpMsg) RawDocument() (wirebson.RawDocument, error) {
	return msg.DocumentRaw()
}

// RawSection0 returns the value of the section with kind 0.
//
// Most callers should use [OpMsg.DocumentRaw] instead.
func (msg *OpMsg) RawSection0() wirebson.RawDocument {
	for _, s := range msg.sections {
		if s.kind == 0 {
			return s.documents[0]
		}
	}

	return nil
}

// RawSections returns the value of section with kind 0 and the value of all sections with kind 1.
//
// Most callers should use [OpMsg.DocumentRaw] instead.
func (msg *OpMsg) RawSections() (wirebson.RawDocument, []byte) {
	var spec wirebson.RawDocument
	var seq []byte

	for _, s := range msg.sections {
		switch s.kind {
		case 0:
			spec = s.documents[0]

		case 1:
			for _, d := range s.documents {
				seq = append(seq, d...)
			}
		}
	}

	return spec, seq
}

// logMessage returns a string representation for logging.
func (msg *OpMsg) logMessage(logFunc func(v any) string) string {
	if msg == nil {
		return "<nil>"
	}

	m := wirebson.MustDocument(
		"FlagBits", msg.Flags.String(),
		"Checksum", int64(msg.checksum),
	)

	sections := wirebson.MakeArray(len(msg.sections))
	for _, section := range msg.sections {
		s := wirebson.MustDocument(
			"Kind", int32(section.kind),
		)

		switch section.kind {
		case 0:
			doc, err := section.documents[0].DecodeDeep()
			if err == nil {
				must.NoError(s.Add("Document", doc))
			} else {
				must.NoError(s.Add("DocumentError", err.Error()))
			}

		case 1:
			must.NoError(s.Add("Identifier", section.identifier))
			docs := wirebson.MakeArray(len(section.documents))

			for _, d := range section.documents {
				doc, err := d.DecodeDeep()
				if err == nil {
					must.NoError(docs.Add(doc))
				} else {
					must.NoError(docs.Add(wirebson.MustDocument("error", err.Error())))
				}
			}

			must.NoError(s.Add("Documents", docs))

		default:
			panic(fmt.Sprintf("unknown kind %d", section.kind))
		}

		must.NoError(sections.Add(s))
	}

	must.NoError(m.Add("Sections", sections))

	return logFunc(m)
}

// String returns an string representation for logging.
func (msg *OpMsg) String() string {
	return msg.logMessage(wirebson.LogMessage)
}

// StringIndent returns an indented string representation for logging.
func (msg *OpMsg) StringIndent() string {
	return msg.logMessage(wirebson.LogMessageIndent)
}

// check interfaces
var (
	_ MsgBody = (*OpMsg)(nil)
)
