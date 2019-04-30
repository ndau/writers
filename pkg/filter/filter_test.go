package filter

import (
	"bufio"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/oneiro-ndev/writers/pkg/ringbuffer"
	"github.com/stretchr/testify/assert"
)

// This demonstrates how to use a RingBuffer with a JSONSplit object and a scanner
// to parse out JSON objects from a stream in real time.
func TestJSONSplitWithBuffer(t *testing.T) {
	c := ringbuffer.New(100)
	// create an object that emits chunks of json
	go func() {
		for i := 0; i < 10; i++ {
			j := buildJSON(5)
			c.Write(j)
			time.Sleep(50 * time.Millisecond)
		}
		c.Close()
	}()

	scanner := bufio.NewScanner(c)
	scanner.Split(JSONSplit)
outer:
	for {
		select {
		case <-c.C:
			if !scanner.Scan() {
				break outer
			}
			// fmt.Printf("GOT %s\n", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Invalid input: %s", err)
	}
}

func TestJSONFilter_Basic(t *testing.T) {
	// create a function to consume our output
	mutex := sync.Mutex{}
	ma := make([]map[string]interface{}, 0)
	outputter := func(m map[string]interface{}) {
		mutex.Lock()
		defer mutex.Unlock()
		ma = append(ma, m)
	}

	// create an object that emits chunks of json
	done := make(chan struct{})
	var w io.Writer = NewJSONFilter(outputter, done, JSONInterpreter{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 10; i++ {
			j := buildJSON(5)
			n, err := w.Write(j)
			assert.Equal(t, len(j), n)
			assert.Nil(t, err)
			time.Sleep(5 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()
	close(done)
	mutex.Lock()
	defer mutex.Unlock()
	assert.Equal(t, 10, len(ma))
	for n := 0; n < 10; n++ {
		assert.Equal(t, 5, len(ma[n]))
	}

}
