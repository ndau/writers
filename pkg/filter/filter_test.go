package filter

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/oneiro-ndev/writers/pkg/bufio"
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

	scanner := bufio.NewScanner(c, JSONSplit)
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

	// Give the scanner some time to process the json.
	time.Sleep(500 * time.Millisecond)

	close(done)
	mutex.Lock()
	defer mutex.Unlock()
	assert.Equal(t, 10, len(ma))
	for n := 0; n < len(ma); n++ {
		assert.Equal(t, 5, len(ma[n]))
	}
}

func TestErrNoProgress(t *testing.T) {
	outputter := func(m map[string]interface{}) {
		// The "error" used to contain io.ErrNoProgress.  Here we are intolerant of any error.
		if e, ok := m["error"]; ok {
			assert.FailNow(t, "scanner error", e)
		}
	}
	filter := NewFilter(JSONSplit, outputter, nil, JSONInterpreter{})

	// Do multiple zero-byte Write()s to make the Scan() in NewFilter() output an error.
	for i := 0; i < 10; i++ {
		// Put something in the buffer.  addLen() will send it through the "C" channel.
		n1, err1 := filter.Write([]byte("x"))

		// Call Write() again to send more through the channel, but w/o offering more data.
		n2, err2 := filter.Write([]byte(""))

		// Make the above happen as quickly as possible to stress the scanner; assert after.
		assert.Nil(t, err1)
		assert.Equal(t, 1, n1)
		assert.Nil(t, err2)
		assert.Equal(t, 0, n2)

		// Give some cycles to the scanner go routine so it can process the above writes.
		time.Sleep(5 * time.Millisecond)
	}
}

func TestEmptyData(t *testing.T) {
	outputter := func(m map[string]interface{}) {
		if len(m) == 0 {
			assert.FailNow(t, "empty map found", m)
		}
	}
	filter := NewFilter(JSONSplit, outputter, nil, JSONInterpreter{})

	for i := 0; i < 10; i++ {
		// Put something in the buffer.  addLen() will send it through the "C" channel.
		n1, err1 := filter.Write([]byte("x"))

		// Call Write() again to send more through the channel, but w/o offering more data.
		n2, err2 := filter.Write([]byte(""))

		// Make the above happen as quickly as possible to stress the scanner; assert after.
		assert.Nil(t, err1)
		assert.Equal(t, 1, n1)
		assert.Nil(t, err2)
		assert.Equal(t, 0, n2)

		// Give some cycles to the scanner go routine so it can process the above writes.
		time.Sleep(5 * time.Millisecond)
	}
}

func TestSingleJSON(t *testing.T) {
	mut := sync.Mutex{}
	j := buildJSON(5)
	count := 0

	outputter := func(m map[string]interface{}) {
		mut.Lock()
		defer mut.Unlock()
		p, err := json.Marshal(m)
		assert.Nil(t, err)
		assert.Equal(t, j, p)
		count++
	}
	filter := NewFilter(JSONSplit, outputter, nil, JSONInterpreter{})
	filter.Write(j)

	// Give the scanner some time to process the json.
	time.Sleep(500 * time.Millisecond)

	// Make sure the outputter was called.
	mut.Lock()
	assert.Equal(t, 1, count)
	mut.Unlock()
}

func TestDoubleJSON(t *testing.T) {
	mut := sync.Mutex{}
	j1 := buildJSON(5)
	j2 := buildJSON(5)
	count := 0

	outputter := func(m map[string]interface{}) {
		mut.Lock()
		defer mut.Unlock()
		p, err := json.Marshal(m)
		assert.Nil(t, err)
		if count == 0 {
			assert.Equal(t, j1, p)
		} else {
			assert.Equal(t, j2, p)
		}
		count++
	}
	filter := NewFilter(JSONSplit, outputter, nil, JSONInterpreter{})
	filter.Write(j1)
	filter.Write(j2)

	// Give the scanner some time to process the json.
	time.Sleep(500 * time.Millisecond)

	// Make sure the outputter was called as many times as we expected.
	mut.Lock()
	assert.Equal(t, 2, count)
	mut.Unlock()
}

func TestTendermintJSON(t *testing.T) {
	mut := sync.Mutex{}
	count := 0

	interpreters := []Interpreter{
		RequiredFieldsInterpreter{
			Defaults: map[string]interface{}{
				"bin":     "tendermint",
				"node_id": "mainnet-0",
			},
		},
		JSONInterpreter{},
		NewTendermintInterpreter(),
		LastChanceInterpreter{},
	}

	outputter := func(m map[string]interface{}) {
		mut.Lock()
		defer mut.Unlock()
		expected := []string{
			// Standard tendermint fields.
			"_msg", "level", "module",
			// Required fields.
			"bin", "node_id",
			// Embedded fields.
			"App", "BlockID", "ChainID", "Consensus", "Height", "LastBlockID",
			"LastCommit", "NextValidators", "NumTxs", "Proposer", "Results",
			"Time", "TotalTxs", "Validators", "Version",
		}
		for _, e := range expected {
			if _, ok := m[e]; !ok {
				t.Errorf("got = %#v, expected it to have %s\n(parsing %s)", m, e, sampleTmJson)
			}
		}
		count++
	}

	filter := NewFilter(JSONSplit, outputter, nil, interpreters...)
	filter.Write([]byte(sampleTmJson))

	// Give the scanner some time to process the json.
	time.Sleep(500 * time.Millisecond)

	// Make sure the outputter was called as many times as we expected.
	mut.Lock()
	assert.Equal(t, 1, count)
	mut.Unlock()
}
