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

// Package wireclient provides low-level wire protocol client.
package wireclient

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/xdg-go/scram"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
)

// nextRequestID stores the last generated request ID.
var nextRequestID atomic.Int32

// Conn represents a single client connection.
//
// It is not safe for concurrent use.
type Conn struct {
	c net.Conn
	r *bufio.Reader
	w *bufio.Writer
	l *slog.Logger // debug-level only
}

// New wraps the given connection.
//
// The passed logger will be used only for debug-level messages.
func New(c net.Conn, l *slog.Logger) *Conn {
	return &Conn{
		c: c,
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
		l: l,
	}
}

// Connect creates a new connection for the given MongoDB URI.
//
// Context can be used to cancel the connection attempt.
// Canceling the context after the connection is established has no effect.
//
// The passed logger will be used only for debug-level messages.
func Connect(ctx context.Context, uri string, l *slog.Logger) (*Conn, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("wireclient.Connect: %w", err)
	}

	if u.Scheme != "mongodb" {
		return nil, fmt.Errorf("wireclient.Connect: invalid scheme %q", u.Scheme)
	}

	if u.Opaque != "" {
		return nil, fmt.Errorf("wireclient.Connect: invalid URI %q", uri)
	}

	if u.Path != "/" {
		return nil, fmt.Errorf("wireclient.Connect: unsupported path %q", u.Path)
	}

	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		return nil, fmt.Errorf("wireclient.Connect: %w", err)
	}

	for k := range u.Query() {
		switch k {
		case "replicaSet":
			// safe to ignore

		default:
			return nil, fmt.Errorf("wireclient.Connect: query parameter %q is not supported", k)
		}
	}

	l.DebugContext(ctx, "Connecting", slog.String("uri", uri))

	d := net.Dialer{}

	c, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return nil, fmt.Errorf("wireclient.Connect: %w", err)
	}

	return New(c, l), nil
}

// ConnectPing uses a combination of [Connect] and [Conn.Ping] to establish a working connection.
//
// nil is returned on context expiration.
func ConnectPing(ctx context.Context, uri string, l *slog.Logger) *Conn {
	for ctx.Err() == nil {
		conn, err := Connect(ctx, uri, l)
		if err != nil {
			sleep(ctx, time.Second)
			continue
		}

		if err = conn.Ping(ctx); err != nil {
			_ = conn.Close()
			sleep(ctx, time.Second)
			continue
		}

		return conn
	}

	return nil
}

// Close closes the connection.
func (c *Conn) Close() error {
	c.l.Debug("Closing")

	if err := c.c.Close(); err != nil {
		return fmt.Errorf("wireclient.Conn.Close: %w", err)
	}

	return nil
}

// Read reads the next message from the connection.
//
// Passed context's deadline is honored if set.
func (c *Conn) Read(ctx context.Context) (*wire.MsgHeader, wire.MsgBody, error) {
	d, _ := ctx.Deadline()
	c.c.SetReadDeadline(d)

	header, body, err := wire.ReadMessage(c.r)
	if err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Read: %w", err)
	}

	c.l.DebugContext(
		ctx,
		fmt.Sprintf("<<<\n%s\n", body.StringBlock()),
		slog.Int("length", int(header.MessageLength)),
		slog.Int("id", int(header.RequestID)),
		slog.Int("response_to", int(header.ResponseTo)),
		slog.String("opcode", header.OpCode.String()),
	)

	return header, body, nil
}

// Write writes the given message to the connection.
//
// Passed context's deadline is honored if set.
func (c *Conn) Write(ctx context.Context, header *wire.MsgHeader, body wire.MsgBody) error {
	c.l.DebugContext(
		ctx,
		fmt.Sprintf(">>>\n%s\n", body.StringBlock()),
		slog.Int("length", int(header.MessageLength)),
		slog.Int("id", int(header.RequestID)),
		slog.Int("response_to", int(header.ResponseTo)),
		slog.String("opcode", header.OpCode.String()),
	)

	if d, ok := ctx.Deadline(); ok {
		c.c.SetWriteDeadline(d)
	}

	if err := wire.WriteMessage(c.w, header, body); err != nil {
		return fmt.Errorf("wireclient.Conn.Write: %w", err)
	}

	if err := c.w.Flush(); err != nil {
		return fmt.Errorf("wireclient.Conn.Write: %w", err)
	}

	return nil
}

// WriteRaw writes the given raw bytes to the connection.
//
// Passed context's deadline is honored if set.
func (c *Conn) WriteRaw(ctx context.Context, b []byte) error {
	c.l.DebugContext(ctx, ">>> raw bytes", slog.Int("length", len(b)))

	d, _ := ctx.Deadline()
	c.c.SetWriteDeadline(d)

	if _, err := c.w.Write(b); err != nil {
		return fmt.Errorf("wireclient.Conn.WriteRaw: %w", err)
	}

	if err := c.w.Flush(); err != nil {
		return fmt.Errorf("wireclient.Conn.WriteRaw: %w", err)
	}

	return nil
}

// Request sends the given request to the connection and returns the response.
// The header is generated automatically.
//
// Passed context's deadline is honored if set.
//
// It returns errors only for request/response parsing or connection issues.
// All protocol-level errors are stored inside response.
func (c *Conn) Request(ctx context.Context, body wire.MsgBody) (*wire.MsgHeader, wire.MsgBody, error) {
	b, err := body.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Request: %w", err)
	}

	header := &wire.MsgHeader{
		MessageLength: int32(len(b) + wire.MsgHeaderLen),
		RequestID:     nextRequestID.Add(1),
	}

	switch body.(type) {
	case *wire.OpMsg:
		header.OpCode = wire.OpCodeMsg
	case *wire.OpQuery:
		header.OpCode = wire.OpCodeQuery
	default:
		return nil, nil, fmt.Errorf("wireclient.Conn.Request:unsupported body type %T", body)
	}

	if err = c.Write(ctx, header, body); err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Request: %w", err)
	}

	resHeader, resBody, err := c.Read(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("wireclient.Conn.Request: %w", err)
	}

	if resHeader.ResponseTo != header.RequestID {
		err = fmt.Errorf(
			"wireclient.Conn.Request: response's response_to=%d is not equal to request's request_id=%d",
			resHeader.ResponseTo,
			header.RequestID,
		)
	}

	return resHeader, resBody, err
}

// Ping sends a ping command.
// It returns error for unexpected server response or any other error.
func (c *Conn) Ping(ctx context.Context) error {
	cmd := wire.MustOpMsg("ping", int32(1), "$db", "test")

	_, resBody, err := c.Request(ctx, cmd)
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Ping: %w", err)
	}

	resRaw, err := resBody.(*wire.OpMsg).RawDocument()
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Ping: %w", err)
	}

	res, err := resRaw.Decode()
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Ping: %w", err)
	}

	if ok := res.Get("ok"); ok != 1.0 {
		return fmt.Errorf("wireclient.Conn.Ping: failed (ok was %v)", ok)
	}

	return nil
}

// Login authenticates the connection with the given credentials.
//
// It should not be used to test various authentication scenarios.
func (c *Conn) Login(ctx context.Context, username, password, authDB string) error {
	s, err := scram.SHA256.NewClient(username, password, "")
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Login: %w", err)
	}

	conv := s.NewConversation()

	payload, err := conv.Step("")
	if err != nil {
		return fmt.Errorf("wireclient.Conn.Login: %w", err)
	}

	cmd := wirebson.MustDocument(
		"saslStart", int32(1),
		"mechanism", "SCRAM-SHA-256",
		"payload", wirebson.Binary{B: []byte(payload)},
		"$db", authDB,
	)

	for step := range 3 {
		c.l.DebugContext(
			ctx, "Login",
			slog.Int("step", step), slog.Bool("done", conv.Done()), slog.Bool("valid", conv.Valid()),
		)

		var body *wire.OpMsg
		if body, err = wire.NewOpMsg(cmd); err != nil {
			return fmt.Errorf("wireclient.Conn.Login: %w", err)
		}

		_, resBody, err := c.Request(ctx, body)
		if err != nil {
			return fmt.Errorf("wireclient.Conn.Login: %w", err)
		}

		resRaw, err := resBody.(*wire.OpMsg).RawDocument()
		if err != nil {
			return fmt.Errorf("wireclient.Conn.Login: %w", err)
		}

		res, err := resRaw.Decode()
		if err != nil {
			return fmt.Errorf("wireclient.Conn.Login: %w", err)
		}

		if ok := res.Get("ok"); ok != 1.0 {
			return fmt.Errorf("wireclient.Conn.Login: %s failed (ok was %v)", cmd.Command(), ok)
		}

		if res.Get("done").(bool) {
			if !conv.Valid() {
				return fmt.Errorf("wireclient.Conn.Login: conversation is not valid")
			}

			return nil
		}

		payload, err = conv.Step(string(res.Get("payload").(wirebson.Binary).B))
		if err != nil {
			return fmt.Errorf("wireclient.Conn.Login: %w", err)
		}

		cmd = wirebson.MustDocument(
			"saslContinue", int32(1),
			"conversationId", int32(1),
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", authDB,
		)
	}

	return fmt.Errorf("wireclient.Conn.Login: too many steps")
}

// sleep waits until the given duration is over or the context is canceled.
func sleep(ctx context.Context, d time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	<-ctx.Done()
}
