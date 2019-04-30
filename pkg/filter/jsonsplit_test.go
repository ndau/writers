package filter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

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
