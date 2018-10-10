package testwriter

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
