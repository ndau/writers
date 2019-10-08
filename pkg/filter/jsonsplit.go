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
	"bytes"
	"regexp"
	"strconv"
)

// MaxObjectLength is the length after which we just stop looking to close the start of
// a JSON object that started but didn't finish.
const MaxObjectLength = 3000

func wrapMsg(b []byte) []byte {
	return []byte(`{"_msg": ` + strconv.Quote(string(b)) + "}")
}

var startpat = regexp.MustCompile(`{[[:space:]]*"`)

// JSONSplit is compatible with bufio.SplitFunc; it reads a single JSON object
// from the input stream, where "JSON object" is defined as a block of text
// that starts with '{' and is followed by optional whitespace and a '"',
// and ends with a matching '}' (ignoring the contents of quoted strings between).
// The terminating '}' must occur within 3000 characters of the start.
//
// Any non-whitespace content between objects meeting the above definition has
// quotes escaped and then is wrapped in a JSON object containing only
// `{"_msg": "<content>" }`. This will allow it to be post-processed by the
// TendermintInterpreter if desired.
//
// This function is defined to return an error to comply with the SplitFunc signature,
// but in reality it never does -- it simply returns bad results wrapped in JSON.
func JSONSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// there's nothing else to parse, just return
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// pull off everything up to the start pattern; if there's any non-whitespace, return it
	starts := startpat.FindIndex(data)
	if starts == nil {
		return 0, nil, nil
	}
	start := starts[0]
	if start != 0 {
		prefix := bytes.TrimSpace(data[:start])
		if len(prefix) > 0 {
			return start, wrapMsg(prefix), nil
		}
		// it was all whitespace, so we can just continue
	}
	// find the matching brace
	end, ok := matchBrace(data, start+1)
	if !ok {
		// didn't find it, are we at EOF?
		if atEOF {
			// consume the rest of the data
			return len(data), wrapMsg(data), nil
		}
		// we're not at EOF, let's look to see if there's another start somewhere in this buffer
		starts2 := startpat.FindIndex(data[starts[1]:])
		if starts2 != nil {
			// found one, so reject everything up to that point
			return starts2[0], wrapMsg(data[:starts2[0]]), nil
		}
		// we didn't find anything so check the length
		if len(data) > MaxObjectLength {
			// carve off the MaxObjectLength and return it in one big blob
			return MaxObjectLength, wrapMsg(data[:MaxObjectLength]), nil
		}
		// still too short, just go back and look harder
		return 0, nil, nil
	}
	end++
	return end, data[start:end], nil
}

func matchBrace(data []byte, start int) (int, bool) {
	for i := start; i < len(data); i++ {
		switch data[i] {
		case '}':
			return i, true
		case '"':
			newi, ok := matchQuote(data, i+1)
			if !ok {
				return -1, false
			}
			i = newi
		case '{':
			newi, ok := matchBrace(data, i+1)
			if !ok {
				return -1, false
			}
			i = newi
		}
	}
	return -1, false
}

func matchQuote(data []byte, start int) (int, bool) {
	for i := start; i < len(data); i++ {
		switch data[i] {
		case '\\':
			i++ // skip an extra char
		case '"':
			return i, true
		}
	}
	return -1, false
}
