package ringbuffer

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRingBufferNoWrap(t *testing.T) {
	// create one and check its initial state
	c := New(100)
	assert.Equal(t, 100, c.Capacity())
	assert.Zero(t, c.Len())
	// now write 5 bytes to it
	n, err := c.Write([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 100, c.Capacity())
	assert.Equal(t, 5, c.Len())
	// and read it back
	buf := make([]byte, c.Len())
	n, err = c.Peek(buf)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("hello"), buf)
	// consume them
	n = c.Consume(5)
	assert.Equal(t, 100, c.Capacity())
	assert.Zero(t, c.Len())

	// now write more bytes to it
	n, err = c.Write([]byte("wassup?"))
	assert.Nil(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, 100, c.Capacity())
	assert.Equal(t, 7, c.Len())
	// and read it back but don't consume it
	buf = make([]byte, c.Len())
	n, err = c.Peek(buf)
	assert.Nil(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, []byte("wassup?"), buf)

	// we should be able to read it again (partially)
	buf = make([]byte, 3)
	n, err = c.Peek(buf)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("was"), buf)
	c.Consume(3)

	// now try the rest
	buf = make([]byte, c.Len())
	n, err = c.Peek(buf)
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("sup?"), buf)
}

func TestRingBufferWrap(t *testing.T) {
	// create one and check its initial state
	c := New(10)
	// now write 5 bytes to it
	n, err := c.Write([]byte("aaaaa"))
	assert.Nil(t, err)
	// throw them away
	c.Consume(5)
	// now write 3 bytes to it
	n, err = c.Write([]byte("bbb"))
	assert.Nil(t, err)
	// and 4 more
	n, err = c.Write([]byte("cccc"))
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	b := make([]byte, 7)
	n, err = c.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, "bbbcccc", string(b))
}

func TestRingBufferRepeat(t *testing.T) {
	c := New(100)
	// create data that's relatively prime to the buffer length (it's 17 bytes)
	data := []byte("<ring buffer>")
	for i := 0; i < 100; i++ {
		n, err := c.Write(data)
		assert.Nil(t, err)
		assert.Equal(t, len(data), n)
		got := make([]byte, len(data))
		n, err = c.Peek(got)
		assert.Nil(t, err)
		assert.Equal(t, len(data), n)
		c.Consume(len(data))
	}
}

func TestRingBuffer(t *testing.T) {
	poem := `
		On A Circle

		I'm up and down, and round about,
		Yet all the world can't find me out;
		Though hundreds have employ'd their leisure,
		They never yet could find my measure.
		I'm found almost in every garden,
		Nay, in the compass of a farthing.
		There's neither chariot, coach, nor mill,
		Can move an inch except I will.
		-- Jonathan Swift	`

	lines := []byte(poem)

	c := New(67)
	output := make([]byte, 0)
	for len(lines) > 0 {
		n := rand.Intn(100)
		if n > len(lines) {
			n = len(lines)
		}
		l := lines[:n]
		_, err := c.Write(l)
		if err == nil {
			// only do this if it worked
			lines = lines[n:]
			continue
		}
		b := make([]byte, c.Len())
		_, err = c.Read(b)
		assert.Nil(t, err)
		output = append(output, b...)
	}
	b := make([]byte, c.Len())
	_, err := c.Read(b)
	assert.Nil(t, err)
	output = append(output, b...)
	assert.Equal(t, poem, string(output))
}

func TestRingBufferWriteBoundary(t *testing.T) {
	c := New(10)
	n, err := c.Write([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, c.Len())
	n, err = c.Write([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 10, c.Len())
	b := make([]byte, 10)
	n, err = c.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, 10, n)

	n, err = c.Write([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, c.Len())
	n, err = c.Write([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 10, c.Len())
	n, err = c.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, 10, n)
}

func TestRingBufferParallel(t *testing.T) {
	c := New(997)
	const n = 100000
	sent := 0
	done := make(chan struct{})
	go func() {
		for sent < n {
			buf := make([]byte, 50+rand.Intn(100))
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			for len(buf) > c.Capacity()-c.Len() {
				time.Sleep(10 * time.Millisecond)
			}
			c.Write(buf)
			sent += len(buf)
		}
		close(done)
	}()

	received := 0
outer:
	for {
		select {
		case <-time.After(1 * time.Millisecond):
			if c.Len() == 0 {
				continue
			}
			buf := make([]byte, c.Len())
			n, err := c.Read(buf)
			assert.Nil(t, err)
			received += n
		case <-done:
			buf := make([]byte, c.Len())
			n, err := c.Read(buf)
			assert.Nil(t, err)
			received += n
			break outer
		}
	}
	if sent != received {
		t.Errorf("sent %d not equal to received %d\n", sent, received)
	}
}

func TestRingBufferClose(t *testing.T) {
	c := New(997)
	const n = 100000
	sent := 0
	go func() {
		for sent < n {
			buf := make([]byte, 50+rand.Intn(100))
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			for len(buf) > c.Capacity()-c.Len() {
				time.Sleep(10 * time.Millisecond)
			}
			c.Write(buf)
			sent += len(buf)
		}
		c.Close()
	}()

	received := 0
	buf := make([]byte, 100)
	for {
		n, err := c.Read(buf)
		if err == io.EOF {
			break
		}
		assert.Nil(t, err)
		received += n
	}

	if sent != received {
		t.Errorf("sent %d not equal to received %d\n", sent, received)
	}
}

func TestRingBufferSelect(t *testing.T) {
	c := New(997)
	const n = 100000
	sent := 0
	go func() {
		for sent < n {
			buf := make([]byte, 50+rand.Intn(100))
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			for len(buf) > c.Capacity()-c.Len() {
				time.Sleep(10 * time.Millisecond)
			}
			c.Write(buf)
			sent += len(buf)
		}
		c.Close()
	}()

	received := 0
	buf := make([]byte, 100)
outer:
	for {
		select {
		case <-c.C:
			n, err := c.Read(buf)
			if err == io.EOF {
				break outer
			}
			assert.Nil(t, err)
			received += n
		}
	}

	if sent != received {
		t.Errorf("sent %d not equal to received %d\n", sent, received)
	}
}

func TestRingBufferReadAll(t *testing.T) {
	c := New(997)
	const n = 100000
	sent := 0
	go func() {
		for sent < n {
			buf := make([]byte, 50+rand.Intn(100))
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			for len(buf) > c.Capacity()-c.Len() {
				time.Sleep(10 * time.Millisecond)
			}
			c.Write(buf)
			sent += len(buf)
		}
		c.Close()
	}()

	buf, err := ioutil.ReadAll(c)
	assert.Nil(t, err)
	received := len(buf)
	if sent != received {
		t.Errorf("sent %d not equal to received %d\n", sent, received)
	}
}

func BenchmarkRingBuffer(b *testing.B) {
	c := New(997)
	go func() {
		for {
			buf := make([]byte, 100)
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			for len(buf) > c.Capacity()-c.Len() {
				time.Sleep(1 * time.Millisecond)
			}
			c.Write(buf)
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total1 := 10000
		total2 := 0
		for i := 0; i < 100; i++ {
			for c.Len() == 0 {
				time.Sleep(1 * time.Millisecond)
			}
			buf := make([]byte, c.Len())
			for _, by := range buf {
				total2 += int(by)
			}
		}
		if total1 == total2 {
			fmt.Printf("Totals were equal: %d, %d\n", total1, total2)
		}
	}
}

func TestRingBufferResizeExplicit(t *testing.T) {
	data := []byte("testing")
	c := New(10)
	c.Write(data)
	assert.Equal(t, 10, c.Capacity())
	assert.Equal(t, 7, c.Len())
	c.resize(20)
	readdata := make([]byte, c.Len())
	n, err := c.Read(readdata)
	assert.Nil(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, data, readdata)
}

func TestRingBufferResizeAuto(t *testing.T) {
	data := []byte("ABC")
	c := New(8)
	// this is going to write 300 bytes total
	for i := 0; i < 100; i++ {
		n, err := c.Write(data)
		assert.Equal(t, 3, n)
		assert.Nil(t, err)
	}
	assert.Equal(t, 512, c.Capacity())
	assert.Equal(t, 300, c.Len())
	readdata := make([]byte, c.Len())
	n, err := c.Read(readdata)
	assert.Nil(t, err)
	assert.Equal(t, 300, n)
}
