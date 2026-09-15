package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Request struct {
	Action  string  `json:"action"`
	Mode    string  `json:"mode"`
	Receipt Receipt `json:"receipt"`
	Hold    bool    `json:"hold"`
}
type Response struct {
	Receipt *Receipt `json:"receipt,omitempty"`
	Error   string   `json:"error,omitempty"`
}

func callSink(socket string, req Request) (*Receipt, error) {
	c, e := net.DialTimeout("unix", socket, 3*time.Second)
	if e != nil {
		return nil, e
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(12 * time.Second))
	if e = json.NewEncoder(c).Encode(req); e != nil {
		return nil, e
	}
	var res Response
	if e = json.NewDecoder(c).Decode(&res); e != nil {
		return nil, e
	}
	if res.Error != "" {
		return nil, fmt.Errorf("sink refusal: %s", res.Error)
	}
	return res.Receipt, nil
}
func sinkRecords(dir string) ([]Receipt, error) {
	f, e := os.Open(filepath.Join(dir, "effects.ndjson"))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	defer f.Close()
	var rs []Receipt
	s := bufio.NewScanner(f)
	for s.Scan() {
		var r Receipt
		if e = json.Unmarshal(s.Bytes(), &r); e != nil {
			return nil, e
		}
		rs = append(rs, r)
	}
	return rs, s.Err()
}

// This process is the sole writer. The worker receives only the socket address.
func serveSink(dir, socket, mode string) error {
	if mode != "queryable" && mode != "opaque" {
		return fmt.Errorf("bad sink mode")
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	rs, e := sinkRecords(dir)
	if e != nil {
		return e
	}
	listener, e := net.Listen("unix", socket)
	if e != nil {
		return e
	}
	defer listener.Close()
	fmt.Println("ready")
	for {
		c, e := listener.Accept()
		if e != nil {
			return e
		}
		_ = c.SetDeadline(time.Now().Add(15 * time.Second))
		var req Request
		e = json.NewDecoder(c).Decode(&req)
		res := Response{}
		if e != nil {
			res.Error = e.Error()
		} else {
			res = handleRequest(dir, mode, req, &rs)
		}
		if req.Hold && res.Receipt != nil && res.Error == "" {
			// Notification is on the controller pipe, never the worker connection.
			fmt.Println("reply_withheld " + res.Receipt.Key)
			var one [1]byte
			_, _ = c.Read(one[:]) // EOF when the controller kills its worker.
		} else {
			_ = json.NewEncoder(c).Encode(res)
		}
		c.Close()
	}
}
func handleRequest(dir, mode string, req Request, rs *[]Receipt) Response {
	if req.Mode != mode {
		return Response{Error: "mode mismatch"}
	}
	if req.Action == "query" {
		if mode == "opaque" {
			return Response{Error: "opaque sink offers no query"}
		}
		for _, r := range *rs {
			if r.Key == req.Receipt.Key {
				copy := r
				return Response{Receipt: &copy}
			}
		}
		return Response{}
	}
	if req.Action != "put" {
		return Response{Error: "unknown action"}
	}
	r := req.Receipt
	if r.Key == "" || r.Input != digest(r.Payload) {
		return Response{Error: "payload digest mismatch"}
	}
	if mode == "queryable" {
		for _, old := range *rs {
			if old.Key == r.Key {
				if old.Input != r.Input || old.Run != r.Run || old.Step != r.Step || old.Source != r.Source || old.Subject != r.Subject {
					return Response{Error: "key conflicts with retained effect"}
				}
				copy := old
				return Response{Receipt: &copy}
			}
		}
	}
	r.Number = len(*rs) + 1
	r.ID = ""
	r.ID = digest(encoded(r))
	f, e := os.OpenFile(filepath.Join(dir, "effects.ndjson"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return Response{Error: e.Error()}
	}
	b := append(encoded(r), '\n')
	n, e := f.Write(b)
	if e == nil && n != len(b) {
		e = fmt.Errorf("short effect write")
	}
	if e == nil {
		e = f.Sync()
	}
	f.Close()
	if e != nil {
		return Response{Error: e.Error()}
	}
	*rs = append(*rs, r)
	fmt.Println("committed " + r.Key)
	return Response{Receipt: &r}
}
