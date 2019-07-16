package filter

import (
	"io"

	"github.com/oneiro-ndev/writers/pkg/bufio"
	"github.com/oneiro-ndev/writers/pkg/ringbuffer"
)

// Filter implements io.Writer so that it can be passed to a process in place of os.Stdout
// or os.Stderr.
// It assumes that its input is a stream of JSON objects. At initialization, it accepts a number
// of Interpreters. On each call to Write(), it filters the input data through each Interpreter
// in order, and then writes the result (a map of k/v pairs) to its output function.
// Because we can't guarantee that calls to Write map neatly to JSON objects, we use a
// RingBuffer to allow a scanner to retrieve JSON objects independent of the way
// the Write calls work.
type Filter struct {
	Interpreters []Interpreter
	cbuf         *ringbuffer.RingBuffer
}

// static assert that Filter implements Writer
var _ io.Writer = (*Filter)(nil)

// NewFilter accepts a SplitFunc, an output function, and some interpreters and constructs a Filter.
// It spawns a goroutine that uses the splitter to read tokens from the ring buffer,
// and then calls interpreters on the token.
// It accepts a done channel (which may be nil), which will shut down its goroutine when closed.
func NewFilter(splitter bufio.SplitFunc, output func(map[string]interface{}), done chan struct{}, terps ...Interpreter) *Filter {
	fp := &Filter{
		Interpreters: terps,
		cbuf:         ringbuffer.New(4096),
	}

	go func() {
		scanner := bufio.NewScanner(fp.cbuf)
		scanner.Split(splitter)

		for {
			select {
			case <-done:
				// just shut down
				return
			case <-fp.cbuf.C:
				for scanner.Scan() {
					data := scanner.Bytes()
					fields := map[string]interface{}{}
					for _, i := range fp.Interpreters {
						data, fields = i.Interpret(data, fields)
					}
					output(fields)
				}
				// if the scanner fails, emit a standard message to the output
				if err := scanner.Err(); err != nil {
					output(map[string]interface{}{"module": "filter", "level": "error", "error": err.Error()})
				}
			}
		}
	}()

	return fp
}

// Write implements io.Writer on the Filter. It just forwards the writes
// to its ring buffer.
func (f *Filter) Write(b []byte) (int, error) {
	return f.cbuf.Write(b)
}

// NewJSONFilter is a convenience function to construct a Filter that uses a JSON splitter,
// for processes that are known to emit a stream of JSON objects.
// It accepts a done channel (which may be nil), which will shut down its goroutine when closed.
func NewJSONFilter(output func(map[string]interface{}), done chan struct{}, terps ...Interpreter) *Filter {
	return NewFilter(JSONSplit, output, done, terps...)
}

// NewLineFilter is a convenience function to construct a Filter that uses a line splitter,
// for processes that are known to emit lines of text.
// It accepts a done channel (which may be nil), which will shut down its goroutine when closed.
func NewLineFilter(output func(map[string]interface{}), done chan struct{}, terps ...Interpreter) *Filter {
	return NewFilter(bufio.ScanLines, output, done, terps...)
}
