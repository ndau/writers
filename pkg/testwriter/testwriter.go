package testwriter

import (
	"bytes"
	"io"
	"testing"

	"github.com/oneiro-ndev/writers/pkg/linewriter"
)

// TestWriter wraps a testing.T in a linewriter, calling
// t.Log on every newline. It implements io.Writer.
//
// The intent is that within a test suite, you can redirect
// all logging calls to the test log,
type TestWriter struct {
	lw *linewriter.LineWriter
}

// static assert that TestWriter is an io.Writer
var _ io.Writer = (*TestWriter)(nil)

// New creates a new TestWriter
func New(t *testing.T) *TestWriter {
	return &TestWriter{
		lw: linewriter.New(&testWriterInner{
			t:   t,
			buf: new(bytes.Buffer),
		}),
	}
}

// Write writes the contents of p.
//
// It returns the number of bytes written.
// If n < len(p), it also returns an error explaining
// why the write is short.
func (l *TestWriter) Write(p []byte) (n int, err error) {
	return
}
