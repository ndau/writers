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
	"strings"
	"testing"
)

type testWriterInner struct {
	t   *testing.T
	buf *bytes.Buffer
}

// static assert that testWriterInner is an io.Writer
var _ io.Writer = (*testWriterInner)(nil)

// Write writes the contents of p.
//
// It returns the number of bytes written.
// If n < len(p), it also returns an error explaining
// why the write is short.
func (l *testWriterInner) Write(p []byte) (n int, err error) {
	n, err = l.buf.Write(p)
	if err != nil {
		return
	}
	l.t.Log(strings.TrimSpace(l.buf.String()))
	l.buf.Reset()
	return
}
