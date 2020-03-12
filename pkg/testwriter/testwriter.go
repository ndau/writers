package testwriter

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"bytes"
	"io"
	"testing"
	"unicode/utf8"

	"github.com/ndau/writers/pkg/linewriter"
)

// TestWriter wraps a testing.T in a linewriter, calling
// t.Log on every newline. It implements io.Writer.
//
// The intent is that within a test suite, you can redirect
// all logging calls to the test log, producing one test-log
// line per log line
type TestWriter struct {
	lw *linewriter.LineWriter
}

// static assert that TestWriter is an io.Writer
var _ io.Writer = (*TestWriter)(nil)

// Write writes some bytes to the test log.
//
// It is expected to produce a new test-log line for every call.
// Therefore, if the log line doesn't end with a newline, one is appended.
func (t *TestWriter) Write(p []byte) (int, error) {
	if p[len(p)-1] != 0x0a { // 0x0a == newline
		p = append(p, 0x0a)
	}
	return t.lw.Write(p)
}

// WriteByte writes a single byte
func (t *TestWriter) WriteByte(c byte) error {
	_, err := t.Write([]byte{c})
	return err
}

// WriteRune writes a single Unicode code point.
//
// It returns the number of bytes written and any error.
func (t *TestWriter) WriteRune(r rune) (size int, err error) {
	buf := make([]byte, utf8.UTFMax)
	nbytes := utf8.EncodeRune(buf, r)
	return t.Write(buf[:nbytes])
}

// WriteString writes a string.
//
// It returns the number of bytes written. If the count is
// less than len(s), it also returns an error explaining
// why the write is short.
func (t *TestWriter) WriteString(s string) (int, error) {
	return t.Write([]byte(s))
}

// New creates a new TestWriter
func New(t *testing.T) *TestWriter {
	return &TestWriter{
		linewriter.New(&testWriterInner{
			t:   t,
			buf: new(bytes.Buffer),
		}),
	}
}

// static assert that TestWriter implements io.Writer
var _ io.Writer = (*TestWriter)(nil)
