package filter

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Interpreter is an interface that accepts an []byte and a map of fields,
// extracts as much data into the set of fields, and returns the remainder
// of the data and the updated map. Interpreters can never error; they must
// simply pass on any uninterpretable data to their output.
type Interpreter interface {
	Interpret(data []byte,
		fields map[string]interface{}) ([]byte, map[string]interface{})
}

// JSONInterpreter attempts to parse the data as JSON and then extracts
// the fields it finds. If any error occurs, it simply passes the entire
// collection to the output unchanged.
// The assumption is that data is a single json object; use JSONSplit and a
// scanner to read the appropriate data from a Reader.
type JSONInterpreter struct{}

var _ Interpreter = JSONInterpreter{}

// Interpret implements Interpreter for JSONInterpreter
func (JSONInterpreter) Interpret(data []byte,
	fields map[string]interface{}) ([]byte, map[string]interface{}) {
	var parsed map[string]interface{}
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		// if it wasn't json, just do nothing
		return data, fields
	}
	for k, v := range parsed {
		fields[k] = v
	}
	return nil, fields
}

// LastChanceInterpreter should be the last Interpreter in the chain. It
// takes any remaining data bytes and escapes them into a string, and sets
// the field _other to the result. We're assuming since this was intended
// to be a log entry that the data is mostly string-like, but there is
// the option to run an Escaper over it so that we don't try to print gibberish.
type LastChanceInterpreter struct {
	Escaper func([]byte) string
}

var _ Interpreter = LastChanceInterpreter{}

// Interpret implements Interpreter for LastChanceInterpreter
func (i LastChanceInterpreter) Interpret(data []byte,
	fields map[string]interface{}) ([]byte, map[string]interface{}) {
	if len(data) != 0 {
		if i.Escaper != nil {
			fields["_other"] = i.Escaper(data)
		} else {
			fields["_other"] = string(data)
		}
	}
	return nil, fields
}

// RequiredFieldsInterpreter simply copies its default fields into the
// destination and then passes on its input data unexamined.
// You can control whether these fields override existing fields
// or are overridden by where this sits in the stack of interpreters.
type RequiredFieldsInterpreter struct {
	Defaults map[string]interface{}
}

var _ Interpreter = RequiredFieldsInterpreter{}

// Interpret implements Interpreter for RequiredFieldsInterpreter
func (i RequiredFieldsInterpreter) Interpret(data []byte,
	fields map[string]interface{}) ([]byte, map[string]interface{}) {
	for k, v := range i.Defaults {
		fields[k] = v
	}
	return data, fields
}

// TendermintInterpreter looks at the specific keys specified
// and attempts to interpret them further by parsing them for
// things that look like "name: value". You'd generally want to
// put this in the list after a JSONInterpreter has split the
// file up.
type TendermintInterpreter struct {
	Keys []string
}

// NewTendermintInterpreter constructs a TendermintInterpreter with the
// one field that is currently worth searching: _msg
func NewTendermintInterpreter() Interpreter {
	return TendermintInterpreter{Keys: []string{"_msg"}}
}

var _ Interpreter = TendermintInterpreter{}

func findFields(v string, fields map[string]interface{}) map[string]interface{} {
	// pattern for matching lines that have Key: value as long as that line doesn't end in curly brace.
	// This pattern is specific to some odd data that Tendermint shoves into a single log message
	// without using the JSON logging. It's not intended to be a general-purpose key/value matcher.
	lpat := regexp.MustCompile(`^([A-Z][A-Za-z0-9]+):[ \t]*(.*[^{])$`)
	// pattern for splitting up lines including trailing and leading whitespace
	spat := regexp.MustCompile(`[ \t]*\n[ \t]*`)
	ss := spat.Split(v, -1)
	for _, s := range ss {
		r := lpat.FindStringSubmatch(s)
		if r != nil {
			n, err := strconv.Atoi(r[2])
			if err != nil {
				fields[r[1]] = r[2]
			} else {
				fields[r[1]] = n
			}
		}
	}
	return fields
}

// Interpret implements Interpreter for TendermintInterpreter
func (i TendermintInterpreter) Interpret(data []byte,
	fields map[string]interface{}) ([]byte, map[string]interface{}) {
	for _, k := range i.Keys {
		v, ok := fields[k]
		if !ok {
			continue
		}
		switch s := v.(type) {
		case string:
			fields = findFields(s, fields)
		case []byte:
			fields = findFields(string(s), fields)
		}
	}
	return data, fields
}

// RedisInterpreter parses the redis logs into useful fields
// Format is documented here: http://build47.com/redis-log-format-levels/
// Example:
// 66940:C 18 Apr 2019 15:18:28.565 # Configuration loaded
// pid:role timestamp loglevel message
type RedisInterpreter struct{}

var _ Interpreter = RedisInterpreter{}

// Interpret implements Interpreter for RedisInterpreter
func (i RedisInterpreter) Interpret(data []byte,
	fields map[string]interface{}) ([]byte, map[string]interface{}) {
	pat := regexp.MustCompile("^([0-9]+):([XCSM]) " +
		"([0-9]+ [A-Za-z]+ [0-9]+ [0-9:.]+) ([.*#-]) (.*)$")
	s := strings.TrimSpace(string(data))
	// don't do anything for empty strings
	if s == "" {
		return nil, fields
	}
	matches := pat.FindStringSubmatch(s)
	if matches == nil {
		// if the match failed, just save the raw message
		// but we still say we processed all the data
		fields["_txt"] = s
		return nil, fields
	}

	// ok, we did get a match, now divide it up
	fields["pid"] = matches[1]
	switch matches[2] {
	case "X":
		fields["role"] = "sentinel"
	case "C":
		fields["role"] = "child"
	case "S":
		fields["role"] = "slave"
	case "M":
		fields["role"] = "master"
	}

	ts := matches[3]
	t, err := time.Parse("02 Jan 2006 15:04:05.000", ts)
	if err != nil {
		fields["timestamp"] = ts
	} else {
		fields["timestamp"] = t.Format(time.RFC3339Nano)
	}

	// redis log levels are as follows
	// . debug
	// - verbose
	// * notice
	// # warning
	// there is no "error" level, so we map "verbose" to "debug"
	switch matches[4] {
	case ".":
		fields["level"] = "debug"
	case "-":
		fields["level"] = "debug"
	case "*":
		fields["level"] = "info"
	case "#":
		fields["level"] = "warn"
	}
	fields["msg"] = matches[5]
	return nil, fields
}
