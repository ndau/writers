package testwriter_test

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"os"
	"strings"
	"testing"

	"github.com/oneiro-ndev/writers/pkg/testwriter"
	"github.com/stretchr/testify/require"
)

// testing test writers is hard because we can't just mock out a testing.T instance
// therefore, we depend on manual inspection of the output
func TestTestWriter(t *testing.T) {
	verbose := false
	for _, arg := range os.Args {
		if strings.Contains(arg, "test.v") {
			verbose = true
			break
		}
	}
	if !verbose {
		t.Skip("-v flag required to run this test")
	}
	twriter := testwriter.New(t)

	write := func(s string) {
		_, err := twriter.WriteString(s)
		require.NoError(t, err)
	}

	write("this should appear in a log line")
	write("this should appear on one log line\nand this on another")
	write("another log line")
	write("five lines without spacing:\n1\n  2\n3  \n  4  \n\t5\t")
}
