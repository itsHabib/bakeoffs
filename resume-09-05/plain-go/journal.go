package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const definition = "plain-artifact-workflow-v1"

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func encode(v any) []byte {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return b
}
func durable(path string, b []byte) error {
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	c := f.Close()
	if e != nil {
		return e
	}
	return c
}

type record struct {
	Seq                        int
	Definition, Previous, Kind string
	Data                       json.RawMessage
	Hash                       string
}
type journal struct {
	path    string
	Records []record
}

func loadJournal(path string) (*journal, error) {
	j := &journal{path: path}
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return j, nil
	}
	if e != nil {
		return nil, e
	}
	prefix := bytes.LastIndexByte(b, '\n') + 1
	scanner := bufio.NewScanner(bytes.NewReader(b[:prefix]))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	previous := ""
	for scanner.Scan() {
		var r record
		if e := json.Unmarshal(scanner.Bytes(), &r); e != nil {
			return nil, fmt.Errorf("REFUSED_CORRUPT_JOURNAL: %w", e)
		}
		h := r.Hash
		r.Hash = ""
		if r.Definition != definition {
			return nil, fmt.Errorf("REFUSED_DEFINITION: %s", r.Definition)
		}
		if r.Seq != len(j.Records)+1 || r.Previous != previous || digest(encode(r)) != h {
			return nil, fmt.Errorf("REFUSED_CORRUPT_JOURNAL: chain or checksum")
		}
		r.Hash = h
		j.Records = append(j.Records, r)
		previous = h
	}
	if e := scanner.Err(); e != nil {
		return nil, e
	}
	if prefix < len(b) {
		if e := os.Truncate(path, int64(prefix)); e != nil {
			return nil, e
		}
	}
	return j, nil
}
func (j *journal) add(kind string, v any) error {
	r := record{Seq: len(j.Records) + 1, Definition: definition, Kind: kind, Data: encode(v)}
	if len(j.Records) > 0 {
		r.Previous = j.Records[len(j.Records)-1].Hash
	}
	r.Hash = digest(encode(r))
	f, e := os.OpenFile(j.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(append(encode(r), '\n'))
	if e == nil {
		e = f.Sync()
	}
	c := f.Close()
	if e != nil {
		return e
	}
	if c != nil {
		return c
	}
	j.Records = append(j.Records, r)
	return nil
}
func decode[T any](b []byte) T {
	var v T
	if e := json.Unmarshal(b, &v); e != nil {
		panic(e)
	}
	return v
}
