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

// OpReply represent the deprecated OP_REPLY wire protocol message type.
// It stores BSON documents in the raw form.
// Only up to one returned document is supported.
//
// Message is checked during construction by [NewOpReply], [MustOpReply], or [OpReply.UnmarshalBinaryNocopy]
// without decoding BSON documents inside.
type OpReply struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, cursor ID, starting from, documents.

	document     wirebson.RawDocument
	CursorID     int64
	Flags        OpReplyFlags
	StartingFrom int32
}

// NewOpReply creates a new OpReply message.
func NewOpReply(doc wirebson.AnyDocument) (*OpReply, error) {
	raw, err := doc.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	reply := &OpReply{
		document: raw,
	}

	if Debug {
		if err = reply.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return reply, nil
}

// MustOpReply creates a new OpReply message constructed from the given pairs of field names and values.
// It panics on error.
func MustOpReply(pairs ...any) *OpReply {
	reply, err := NewOpReply(wirebson.MustDocument(pairs...))
	if err != nil {
		panic(err)
	}

	return reply
}

func (reply *OpReply) msgbody() {}

// check implements [MsgBody].
func (reply *OpReply) check() error {
	if d := reply.document; d != nil {
		if _, err := d.DecodeDeep(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// UnmarshalBinaryNocopy implements [MsgBody].
func (reply *OpReply) UnmarshalBinaryNocopy(b []byte) error {
	if len(b) < 20 {
		return lazyerrors.Errorf("len=%d", len(b))
	}

	reply.Flags = OpReplyFlags(binary.LittleEndian.Uint32(b[0:4]))
	reply.CursorID = int64(binary.LittleEndian.Uint64(b[4:12]))
	reply.StartingFrom = int32(binary.LittleEndian.Uint32(b[12:16]))
	numberReturned := int32(binary.LittleEndian.Uint32(b[16:20]))
	reply.document = b[20:]

	if numberReturned < 0 || numberReturned > 1 {
		return lazyerrors.Errorf("numberReturned=%d", numberReturned)
	}

	if len(reply.document) == 0 {
		reply.document = nil
	}

	if (numberReturned == 0) != (reply.document == nil) {
		return lazyerrors.Errorf("numberReturned=%d, document=%v", numberReturned, reply.document)
	}

	if Debug {
		if err := reply.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// Size implements [MsgBody].
func (reply *OpReply) Size() int {
	return 20 + len(reply.document)
}

// MarshalBinary implements [MsgBody].
func (reply *OpReply) MarshalBinary() ([]byte, error) {
	b := make([]byte, 20+len(reply.document))

	binary.LittleEndian.PutUint32(b[0:4], uint32(reply.Flags))
	binary.LittleEndian.PutUint64(b[4:12], uint64(reply.CursorID))
	binary.LittleEndian.PutUint32(b[12:16], uint32(reply.StartingFrom))

	if reply.document == nil {
		binary.LittleEndian.PutUint32(b[16:20], uint32(0))
	} else {
		binary.LittleEndian.PutUint32(b[16:20], uint32(1))
		copy(b[20:], reply.document)
	}

	return b, nil
}

// Document returns decoded document, or nil.
// It may be shallowly or deeply decoded.
func (reply *OpReply) Document() (*wirebson.Document, error) {
	if reply.document == nil {
		return nil, nil
	}

	doc, err := reply.document.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// DocumentDeep returns deeply decoded document, or nil.
func (reply *OpReply) DocumentDeep() (*wirebson.Document, error) {
	if reply.document == nil {
		return nil, nil
	}

	doc, err := reply.document.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// DocumentRaw returns raw document (that might be nil).
func (reply *OpReply) DocumentRaw() wirebson.RawDocument {
	return reply.document
}

// Deprecated: use DocumentRaw instead.
func (reply *OpReply) RawDocument() wirebson.RawDocument {
	return reply.DocumentRaw()
}

// logMessage returns a string representation for logging.
func (reply *OpReply) logMessage(logFunc func(v any) string) string {
	if reply == nil {
		return "<nil>"
	}

	m := wirebson.MustDocument(
		"ResponseFlags", reply.Flags.String(),
		"CursorID", reply.CursorID,
		"StartingFrom", reply.StartingFrom,
	)

	if reply.document == nil {
		must.NoError(m.Add("NumberReturned", int32(0)))
	} else {
		must.NoError(m.Add("NumberReturned", int32(1)))

		doc, err := reply.DocumentDeep()
		if err == nil {
			must.NoError(m.Add("Document", doc))
		} else {
			must.NoError(m.Add("DocumentError", err.Error()))
		}
	}

	return logFunc(m)
}

// String returns an string representation for logging.
func (reply *OpReply) String() string {
	return reply.logMessage(wirebson.LogMessage)
}

// StringIndent returns an indented string representation for logging.
func (reply *OpReply) StringIndent() string {
	return reply.logMessage(wirebson.LogMessageIndent)
}

// check interfaces
var (
	_ MsgBody = (*OpReply)(nil)
)
