package filter

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestJSONInterpreter_Interpret(t *testing.T) {
	r := JSONInterpreter{}

	tests := []struct {
		name    string
		input   string
		fields  map[string]interface{}
		wantlen int
		wantf   map[string]interface{}
	}{
		{"not json at all", "hi", map[string]interface{}{"a": "hi"}, 2, map[string]interface{}{"a": "hi"}},
		{"not a json object", `"hi"`, map[string]interface{}{"c": "abc"}, 4, map[string]interface{}{"c": "abc"}},
		{"empty json", "{}", map[string]interface{}{"a": "abc"}, 0, map[string]interface{}{"a": "abc"}},
		{"simple", `{"b":"hi"}`, map[string]interface{}{"a": "abc"}, 0, map[string]interface{}{"a": "abc", "b": "hi"}},
		{"several", `{"b":"hi", "msg":"lots of things"}`, map[string]interface{}{"a": "abc"}, 0, map[string]interface{}{
			"a": "abc", "b": "hi", "msg": "lots of things",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotbytes, gotfields := r.Interpret([]byte(tt.input), tt.fields)
			if len(gotbytes) != tt.wantlen {
				t.Errorf("RedisInterpreter.Interpret() return %d bytes, expected %d", len(gotbytes), tt.wantlen)
			}
			if !reflect.DeepEqual(gotfields, tt.wantf) {
				t.Errorf("JSONInterpreter.Interpret() got1 = %v, want %v", gotfields, tt.wantf)
			}
		})
	}
}

func TestLastChanceInterpreter_Interpret(t *testing.T) {
	r := LastChanceInterpreter{
		Escaper: func(data []byte) string { return hex.EncodeToString(data) },
	}

	tests := []struct {
		name    string
		input   string
		fields  map[string]interface{}
		wantlen int
		wantf   map[string]interface{}
	}{
		{"basic", "hi", map[string]interface{}{}, 0, map[string]interface{}{"_other": "6869"}},
		{"additive", "hi", map[string]interface{}{"c": "abc"}, 0, map[string]interface{}{"_other": "6869", "c": "abc"}},
		{"override", "ddd", map[string]interface{}{"a": "abc"}, 0, map[string]interface{}{"a": "abc", "_other": "646464"}},
		{"empty", "", map[string]interface{}{"a": "abc"}, 0, map[string]interface{}{"a": "abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotbytes, gotfields := r.Interpret([]byte(tt.input), tt.fields)
			if len(gotbytes) != tt.wantlen {
				t.Errorf("RedisInterpreter.Interpret() return %d bytes, expected %d", len(gotbytes), tt.wantlen)
			}
			if !reflect.DeepEqual(gotfields, tt.wantf) {
				t.Errorf("JSONInterpreter.Interpret() got1 = %v, want %v", gotfields, tt.wantf)
			}
		})
	}
}

func TestRequiredFieldsInterpreterBasic(t *testing.T) {
	r := RequiredFieldsInterpreter{
		Defaults: map[string]interface{}{
			"a": 1,
			"b": "buzz",
		},
	}

	tests := []struct {
		name    string
		input   string
		fields  map[string]interface{}
		wantlen int
		wantf   map[string]interface{}
	}{
		{"basic", "hi", map[string]interface{}{}, 2, map[string]interface{}{"a": 1, "b": "buzz"}},
		{"additive", "hi", map[string]interface{}{"c": "hello"}, 2, map[string]interface{}{"a": 1, "b": "buzz", "c": "hello"}},
		{"override", "whee", map[string]interface{}{"a": "hello"}, 4, map[string]interface{}{"a": 1, "b": "buzz"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotbytes, gotfields := r.Interpret([]byte(tt.input), tt.fields)
			if len(gotbytes) != tt.wantlen {
				t.Errorf("RedisInterpreter.Interpret() return %d bytes, expected %d", len(gotbytes), tt.wantlen)
			}
			if !reflect.DeepEqual(gotfields, tt.wantf) {
				t.Errorf("JSONInterpreter.Interpret() got1 = %v, want %v", gotfields, tt.wantf)
			}
		})
	}
}

var sampleTmLogs = `
{"_msg":"enterNewRound(2/0): Invalid args. Current step: 2/0/RoundStepPrecommit","height":2,"level":"debug","module":"consensus","round":0}
{"_msg":"Send","channel":32,"conn":"MConn{127.0.0.1:26660}","level":"debug","module":"p2p","msgBytes":"1919B3D5080218022001","peer":"d204a43cbf19cbb16c49748e31b65828be8f2901@127.0.0.1:26660"}
{"_msg":"enterPrecommit(2/0): Invalid args. Current step: 2/0/RoundStepPrecommit","height":2,"level":"debug","module":"consensus","round":0}
{"_msg":"enterCommit(2/0). Current: 2/0/RoundStepPrecommit","commitRound":0,"height":2,"level":"info","module":"consensus"}
{"_msg":"Commit is for locked block. Set ProposalBlock=LockedBlock","blockHash":"F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD","commitRound":0,"height":2,"level":"info","module":"consensus"}
{"_msg":"Broadcast","channel":32,"level":"debug","module":"p2p","msgBytes":"C96A6FA808021808"}
{"_msg":"Send","channel":32,"conn":"MConn{127.0.0.1:26660}","level":"debug","module":"p2p","msgBytes":"C96A6FA808021808","peer":"d204a43cbf19cbb16c49748e31b65828be8f2901@127.0.0.1:26660"}
{"_msg":"Finalizing commit of block with 0 txs","hash":"F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD","height":2,"level":"info","module":"consensus","root":"457BAB38A80A871BCF08AB0154232F7B2021AB58"}
{"_msg":"Block{\n  Header{\n    Version:        {10 0}\n    ChainID:        localnet\n    Height:         2\n    Time:           2019-04-27 01:13:43.232704 +0000 UTC\n    NumTxs:         0\n    TotalTxs:       1\n    LastBlockID:    528F0CCA2BC8CE9FDAD1394BDCBCF544B69961845DF80847B8DFED5E3EA3C59A:1:3BD8D1307A95\n    LastCommit:     4765C8140D5F6D1E463DD3185CA3C468E7D1B7CCC41C37DAE6B60669AE856D0C\n    Data:           \n    Validators:     D736B1878F42508E2535F245CE040E793368FFA8331D9685C44A28B13C831C18\n    NextValidators: D736B1878F42508E2535F245CE040E793368FFA8331D9685C44A28B13C831C18\n    App:            457BAB38A80A871BCF08AB0154232F7B2021AB58\n    Consensus:       048091BC7DDC283F77BFBF91D73C44DA58C3DF8A9CBC867405D8B7F3DAADA22F\n    Results:        6E340B9CFFB37A989CA544E6BB780A2C78901D3FB33738768511A30617AFA01D\n    Evidence:       \n    Proposer:       497B1D7E8CD2C6D43C9326145E6C3819179EFE9E\n  }#F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD\n  Data{\n    \n  }#\n  EvidenceData{\n    \n  }#\n  Commit{\n    BlockID:    528F0CCA2BC8CE9FDAD1394BDCBCF544B69961845DF80847B8DFED5E3EA3C59A:1:3BD8D1307A95\n    Precommits:\n      Vote{0:2D0AA78150B6 1/00/2(Precommit) 528F0CCA2BC8 9B76B58D8E6E @ 2019-04-27T01:13:43.336014Z}\n      Vote{1:497B1D7E8CD2 1/00/2(Precommit) 528F0CCA2BC8 D932148F1631 @ 2019-04-27T01:13:43.232704Z}\n  }#4765C8140D5F6D1E463DD3185CA3C468E7D1B7CCC41C37DAE6B60669AE856D0C\n}#F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD","level":"info","module":"consensus"}
{"_msg":"Executed block","height":2,"invalidTxs":0,"level":"info","module":"state","validTxs":0}
{"_msg":"Committed state","appHash":"457BAB38A80A871BCF08AB0154232F7B2021AB58","height":2,"level":"info","module":"state","txs":0}
{"_msg":"Indexed block","height":2,"level":"info","module":"txindex"}
`

var sampleTmLogsExtraFields = `
{"_msg":"Block{\n  Header{\n    Version:        {10 0}\n    ChainID:        localnet\n    Height:         2\n    Time:           2019-04-27 01:13:43.232704 +0000 UTC\n    NumTxs:         0\n    TotalTxs:       1\n    LastBlockID:    528F0CCA2BC8CE9FDAD1394BDCBCF544B69961845DF80847B8DFED5E3EA3C59A:1:3BD8D1307A95\n    LastCommit:     4765C8140D5F6D1E463DD3185CA3C468E7D1B7CCC41C37DAE6B60669AE856D0C\n    Data:           \n    Validators:     D736B1878F42508E2535F245CE040E793368FFA8331D9685C44A28B13C831C18\n    NextValidators: D736B1878F42508E2535F245CE040E793368FFA8331D9685C44A28B13C831C18\n    App:            457BAB38A80A871BCF08AB0154232F7B2021AB58\n    Consensus:       048091BC7DDC283F77BFBF91D73C44DA58C3DF8A9CBC867405D8B7F3DAADA22F\n    Results:        6E340B9CFFB37A989CA544E6BB780A2C78901D3FB33738768511A30617AFA01D\n    Evidence:       \n    Proposer:       497B1D7E8CD2C6D43C9326145E6C3819179EFE9E\n  }#F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD\n  Data{\n    \n  }#\n  EvidenceData{\n    \n  }#\n  Commit{\n    BlockID:    528F0CCA2BC8CE9FDAD1394BDCBCF544B69961845DF80847B8DFED5E3EA3C59A:1:3BD8D1307A95\n    Precommits:\n      Vote{0:2D0AA78150B6 1/00/2(Precommit) 528F0CCA2BC8 9B76B58D8E6E @ 2019-04-27T01:13:43.336014Z}\n      Vote{1:497B1D7E8CD2 1/00/2(Precommit) 528F0CCA2BC8 D932148F1631 @ 2019-04-27T01:13:43.232704Z}\n  }#4765C8140D5F6D1E463DD3185CA3C468E7D1B7CCC41C37DAE6B60669AE856D0C\n}#F4006F1F2544906BC057B8AEFB1B5305264605F1456D78B5DC48C66D84823BBD","level":"info","module":"consensus"}
`

func TestTendermintInterpreter_Interpret(t *testing.T) {
	expected := []string{
		"_msg", "level", "module",
	}
	testTendermintInterpreter(t, sampleTmLogs, expected)
}

func TestTendermintInterpreter_InterpretExtraFields(t *testing.T) {
	expected := []string{
		// Standard tendermint fields.
		"_msg", "level", "module",
		// Embedded fields.
		"App", "BlockID", "ChainID", "Consensus", "Height", "LastBlockID",
		"LastCommit", "NextValidators", "NumTxs", "Proposer", "Results",
		"Time", "TotalTxs", "Validators", "Version",
	}
	testTendermintInterpreter(t, sampleTmLogsExtraFields, expected)
}

func testTendermintInterpreter(t *testing.T, logs string, expected []string) {
	r := NewTendermintInterpreter()
	j := JSONInterpreter{}

	for _, line := range strings.Split(logs, "\n") {
		if len(line) == 0 {
			continue
		}
		f := make(map[string]interface{})
		gotbytes, gotfields := j.Interpret([]byte(line), f)
		gotbytes, gotfields = r.Interpret(gotbytes, gotfields)
		if len(gotbytes) != 0 {
			t.Errorf("TendermintInterpreter.Interpret() got = %v, expected nothing", gotbytes)
		}
		for _, e := range expected {
			if _, ok := gotfields[e]; !ok {
				t.Errorf("TendermintInterpreter.Interpret() got = %#v, expected it to have %s\n(parsing %s)", gotfields, e, line)
			}
		}
		if strings.Contains(gotfields["_msg"].(string), "{") {
			if len(gotfields) < 5 {
				for k, v := range gotfields {
					switch n := v.(type) {
					case int:
						fmt.Printf("%20s: %d\n", k, n)
					default:
						fmt.Printf("%20s: %s\n", k, v)
					}
				}
				t.Errorf("Expected at least 5 fields, got that ^")
			}
		}
	}
}

func TestRedisInterpreterBasic(t *testing.T) {
	r := RedisInterpreter{}
	emptyf := make(map[string]interface{})
	expected := []string{
		"pid", "role", "timestamp", "level", "msg",
	}

	tests := []struct {
		name    string
		input   string
		fields  map[string]interface{}
		wantlen int
		wantf   []string
	}{
		{"basic", "66940:C 18 Apr 2019 15:18:28.565 # Configuration loaded", emptyf, 0, expected},
		{"bad date", "23434:M 18-Apr-2019 14:12:28.565 . bad date", emptyf, 0, expected},
		{"short pid", "5:C 23 Jul 2020 15:18:28.032 - 342asdfuj2", emptyf, 0, expected},
		{"fail", "this is a failure", emptyf, 0, []string{"_txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotbytes, gotfields := r.Interpret([]byte(tt.input), tt.fields)
			if len(gotbytes) != tt.wantlen {
				t.Errorf("RedisInterpreter.Interpret() return %d bytes, expected %d", len(gotbytes), tt.wantlen)
			}
			for _, e := range tt.wantf {
				if _, ok := gotfields[e]; !ok {
					t.Errorf("RedisInterpreter.Interpret() got = %#v, expected it to have %s", gotfields, e)
				}
			}
		})
	}
}

var sampleRedisLogs = `
66940:C 18 Apr 2019 15:18:28.565 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
66940:C 18 Apr 2019 15:18:28.565 # Redis version=5.0.4, bits=64, commit=00000000, modified=0, pid=66940, just started
66940:C 18 Apr 2019 15:18:28.565 # Configuration loaded
66940:M 18 Apr 2019 15:18:28.566 # You requested maxclients of 10000 requiring at least 10032 max file descriptors.
66940:M 18 Apr 2019 15:18:28.566 # Server can't set maximum open files to 10032 because of OS error: Operation not permitted.
66940:M 18 Apr 2019 15:18:28.566 # Current maximum open files is 1024. maxclients has been reduced to 992 to compensate for low ulimit. If you need higher maxclients increase 'ulimit -n'.
66940:M 18 Apr 2019 15:18:28.567 * Running mode=standalone, port=6380.
66940:M 18 Apr 2019 15:18:28.567 # Server initialized
66940:M 18 Apr 2019 15:18:28.569 * DB loaded from disk: 0.001 seconds
66940:M 18 Apr 2019 15:18:28.569 * Ready to accept connections
66940:M 18 Apr 2019 15:19:29.084 * 1 changes in 60 seconds. Saving...
66940:M 18 Apr 2019 15:19:29.085 * Background saving started by pid 67252
67252:C 18 Apr 2019 15:19:29.087 * DB saved on disk
66940:M 18 Apr 2019 15:19:29.190 * Background saving terminated with success
66940:M 18 Apr 2019 15:20:30.003 * 1 changes in 60 seconds. Saving...
66940:M 18 Apr 2019 15:20:30.011 * Background saving started by pid 67489
67489:C 18 Apr 2019 15:20:30.015 * DB saved on disk
66940:M 18 Apr 2019 15:20:30.111 * Background saving terminated with success
`

func TestRedisInterpreterReal(t *testing.T) {
	r := RedisInterpreter{}
	f := make(map[string]interface{})
	expected := []string{
		"pid", "role", "timestamp", "level", "msg",
	}

	for _, line := range strings.Split(sampleRedisLogs, "\n") {
		gotbytes, gotfields := r.Interpret([]byte(line), f)
		if len(gotbytes) == 0 && len(gotfields) == 0 {
			continue
		}
		if len(gotbytes) != 0 {
			t.Errorf("RedisInterpreter.Interpret() got = %v, expected nothing", gotbytes)
		}
		for _, e := range expected {
			if _, ok := gotfields[e]; !ok {
				t.Errorf("RedisInterpreter.Interpret() got = %#v, expected it to have %s\n(parsing %s)", gotfields, e, line)
			}
		}
	}
}
