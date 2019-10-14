package linewriter_test

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bytes"
	"testing"

	"github.com/oneiro-ndev/writers/pkg/linewriter"
	"github.com/stretchr/testify/require"
)

func TestLinewriterPassthrough(t *testing.T) {
	buffer := new(bytes.Buffer)
	writer := linewriter.New(buffer)

	writer.WriteString("hello ")
	require.Equal(t, "", buffer.String())
	writer.WriteString("world!\n")
	require.Equal(t, "hello world!\n", buffer.String())
}

type writesCounter struct {
	bytes.Buffer
	writes int
}

func (w *writesCounter) Write(p []byte) (int, error) {
	w.writes++
	return w.Buffer.Write(p)
}

func TestLinewriterMultiline(t *testing.T) {
	text := `
I met a traveller from an antique land,
Who said—“Two vast and trunkless legs of stone
Stand in the desert. . . . Near them, on the sand,
Half sunk a shattered visage lies, whose frown,
And wrinkled lip, and sneer of cold command,
Tell that its sculptor well those passions read
Which yet survive, stamped on these lifeless things,
The hand that mocked them, and the heart that fed;
And on the pedestal, these words appear:
My name is Ozymandias, King of Kings;
Look on my Works, ye Mighty, and despair!
Nothing beside remains. Round the decay
Of that colossal Wreck, boundless and bare
The lone and level sands stretch far away.”
	`

	counter := new(writesCounter)
	writer := linewriter.New(counter)
	_, err := writer.WriteString(text)
	require.NoError(t, err)
	require.Equal(t, 15, counter.writes)
}

func TestLinewriterDoesntFlushWithoutNewline(t *testing.T) {
	buffer := new(bytes.Buffer)
	writer := linewriter.New(buffer)

	bytes := make([]byte, 512)
	for idx := range bytes {
		b := idx % 256 // to produce bytes which fit
		if b != 0x0a { // newline
			bytes[idx] = byte(b)
		}
	}

	// still waiting for a newline
	writer.Write(bytes)
	require.Empty(t, buffer.Bytes())

	// given a  newline
	writer.WriteByte(0x0a)
	require.NotEmpty(t, buffer.Bytes())
}
