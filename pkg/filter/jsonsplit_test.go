package filter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

var sampleTmJson = `{"_msg":"Block{\n  Header{\n    Version:        {10 0}\n    ChainID:        localnet\n    Height:         2\n    Time:           2019-04-27 01:13:43.232704 +0000 UTC\n    NumTxs:         0\n    TotalTxs:       1\n    LastBlockID:    528F0CCA2BC8CE9FDAD1394BDCBCF544B69961845DF80847B8DFED5E3EA3C59A:1:3BD8D1307A95\n    LastCommit:     4765C8140D5F6D1E463DD3185CA3C468E7D1B7CCC41C37DAE6B60669AE856D0C\n    Data:           \n    Validators:     D736B1878F42508E2535F245CE040E793368FFA8331D9685C44A28B13C831C18\n    NextValidators: D736B1878F42508E2535F245CE040E793368FFA8331D9685C44A28B13C831C18\n    App:            457BAB38A80A871BCF08AB0154232F7B2021AB58\n    Consensus:       048091BC7DDC283F77BFBF91D73C44DA58C3DF8A9CBC867405D8B7F3DAADA22F\n    Results:        6E340B9CFFB37A989CA544E6BB780A2C78901D3FB33738768511A30617AFA01D\n    Evidence:       \n    Proposer:       497B1D7E8CD2C6D43C9326145E6C3819179EFE9E\n  }#F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD\n  Data{\n    \n  }#\n  EvidenceData{\n    \n  }#\n  Commit{\n    BlockID:    528F0CCA2BC8CE9FDAD1394BDCBCF544B69961845DF80847B8DFED5E3EA3C59A:1:3BD8D1307A95\n    Precommits:\n      Vote{0:2D0AA78150B6 1/00/2(Precommit) 528F0CCA2BC8 9B76B58D8E6E @ 2019-04-27T01:13:43.336014Z}\n      Vote{1:497B1D7E8CD2 1/00/2(Precommit) 528F0CCA2BC8 D932148F1631 @ 2019-04-27T01:13:43.232704Z}\n  }#4765C8140D5F6D1E463DD3185CA3C468E7D1B7CCC41C37DAE6B60669AE856D0C\n}#F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD","level":"info","module":"consensus"}`

func TestJSONSplit(t *testing.T) {
	type args struct {
		data  string
		atEOF bool
	}
	tests := []struct {
		name        string
		args        args
		wantAdvance int
		wantToken   string
		wantErr     bool
	}{
		{"empty EOF", args{"", true}, 0, "", false},
		{"simple EOF", args{`{"a":1}`, true}, 7, `{"a":1}`, false},
		{"empty", args{"", false}, 0, "", false},
		{"simple", args{`{"a":1}`, false}, 7, `{"a":1}`, false},
		{"nested", args{`{"a":{"b":17}}`, false}, 14, `{"a":{"b":17}}`, false},
		{"indented", args{`  {"a":{"b":17}}`, false}, 16, `{"a":{"b":17}}`, false},
		{"embedded quote", args{`{"a":"\"I am\", I said"}`, false}, 24, `{"a":"\"I am\", I said"}`, false},
		{"unmatched nesting", args{`{"a":{"b":17}`, true}, 13, `{"_msg": "{\"a\":{\"b\":17}"}`, false},
		{"unmatched quote", args{`{"a":"}`, true}, 7, `{"_msg": "{\"a\":\"}"}`, false},
		{"tendermint", args{sampleTmJson, false}, len(sampleTmJson), sampleTmJson, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdvance, gotToken, err := JSONSplit([]byte(tt.args.data), tt.args.atEOF)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONSplit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("JSONSplit() gotAdvance = %v, want %v", gotAdvance, tt.wantAdvance)
			}
			if !reflect.DeepEqual(string(gotToken), tt.wantToken) {
				t.Errorf("JSONSplit() gotToken = %v, want %v", string(gotToken), tt.wantToken)
			}
		})
	}
}

func buildJSON(n int) []byte {
	r := make(map[string]interface{})
	for f := 0; f < n; f++ {
		k := fmt.Sprintf("%x", rand.Intn(65536))
		v := fmt.Sprintf("%x", rand.Intn(65536))
		n, err := strconv.Atoi(v)
		if err != nil {
			r[k] = v
		} else {
			r[k] = n
		}
	}
	j, _ := json.Marshal(r)
	return j
}

// The set of benchmarks below are tests of whether it's viable to send output from the writer to a channel
// and then read the channel in a separate goroutine.
// We compared writing one byte a time to the channel vs writing a slice of 100 bytes to a channel.
// The slice is about 15-20x faster, and the buffered slice is a couple of times faster still, so
// we're going to work that way.

// The results from one run looked like this:
// BenchmarkByteAtATime-8              	     100	  12796310 ns/op	       1 B/op	       0 allocs/op
// BenchmarkBunchAtATime-8             	    2000	    646602 ns/op	   11200 B/op	     100 allocs/op
// BenchmarkBunchAtATimeBuffered1-8    	    3000	    559832 ns/op	   11200 B/op	     100 allocs/op
// BenchmarkBunchAtATimeBuffered10-8   	   10000	    219604 ns/op	   11200 B/op	     100 allocs/op

func BenchmarkByteAtATime(b *testing.B) {
	ch := make(chan byte)
	go func() {
		ix := 0
		for {
			ch <- byte(ix & 0xFF)
			ix++
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total1 := 12345
		total2 := 0
		for i := 0; i < 10000; i++ {
			by := <-ch
			total2 += int(by)
		}
		if total1 == total2 {
			fmt.Printf("Totals were equal: %d, %d\n", total1, total2)
		}
	}
}

func BenchmarkBunchAtATime(b *testing.B) {
	ch := make(chan []byte)
	go func() {
		for {
			buf := make([]byte, 100)
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			ch <- buf
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total1 := 10000
		total2 := 0
		for i := 0; i < 100; i++ {
			ba := <-ch
			for _, by := range ba {
				total2 += int(by)
			}
		}
		if total1 == total2 {
			fmt.Printf("Totals were equal: %d, %d\n", total1, total2)
		}
	}
}

func BenchmarkBunchAtATimeBuffered1(b *testing.B) {
	ch := make(chan []byte, 1)
	go func() {
		for {
			buf := make([]byte, 100)
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			ch <- buf
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total1 := 10000
		total2 := 0
		for i := 0; i < 100; i++ {
			ba := <-ch
			for _, by := range ba {
				total2 += int(by)
			}
		}
		if total1 == total2 {
			fmt.Printf("Totals were equal: %d, %d\n", total1, total2)
		}
	}
}

func BenchmarkBunchAtATimeBuffered10(b *testing.B) {
	ch := make(chan []byte, 10)
	go func() {
		for {
			buf := make([]byte, 100)
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i)
			}
			ch <- buf
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		total1 := 10000
		total2 := 0
		for i := 0; i < 100; i++ {
			ba := <-ch
			for _, by := range ba {
				total2 += int(by)
			}
		}
		if total1 == total2 {
			fmt.Printf("Totals were equal: %d, %d\n", total1, total2)
		}
	}
}
