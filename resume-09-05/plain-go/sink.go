package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Sink private storage is read only by the sink and test controller, never recovery.
func sink(dir, mode, action, key, payload string, hold bool) error {
	if mode != "queryable" && mode != "opaque" {
		return fmt.Errorf("unknown sink mode")
	}
	if action != "apply" && action != "query" {
		return fmt.Errorf("unknown sink action")
	}
	if mode == "opaque" && action == "query" {
		return fmt.Errorf("OPAQUE_QUERY_UNSUPPORTED")
	}
	path := filepath.Join(dir, "sink-private", "effects.jsonl")
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	// Exclusion refuses concurrent calls; this POC's controller serializes requests.
	lock := filepath.Join(dir, "sink-private", "lock")
	if e := os.Mkdir(lock, 0700); e != nil {
		return fmt.Errorf("sink busy: %w", e)
	}
	defer os.Remove(lock)
	var effects []receipt
	f, e := os.Open(path)
	if e == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			var r receipt
			if e = json.Unmarshal(sc.Bytes(), &r); e != nil {
				f.Close()
				return e
			}
			effects = append(effects, r)
		}
		e = sc.Err()
		f.Close()
		if e != nil {
			return e
		}
	} else if !os.IsNotExist(e) {
		return e
	}
	if mode == "queryable" {
		for _, r := range effects {
			if r.Key == key {
				if action == "apply" && r.Payload != payload {
					return fmt.Errorf("REFUSED_KEY_CONFLICT")
				}
				fmt.Println(string(encode(r)))
				return nil
			}
		}
	}
	if action == "query" {
		fmt.Println(string(encode(receipt{})))
		return nil
	}
	r := receipt{Key: key, Payload: payload, ID: fmt.Sprintf("effect-%d", len(effects)+1), Mode: mode}
	f, e = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(append(encode(r), '\n'))
	if e == nil {
		e = f.Sync()
	}
	f.Close()
	if e != nil {
		return e
	}
	if hold {
		fmt.Println("SINK_COMMITTED_REPLY_WITHHELD")
		var one [1]byte
		if _, e = os.Stdin.Read(one[:]); e != nil {
			return nil
		}
	}
	fmt.Println(string(encode(r)))
	return nil
}
func callSink(c config, action string, s subject, hold bool) (receipt, error) {
	exe, e := os.Executable()
	if e != nil {
		return receipt{}, e
	}
	args := []string{"sink", "--dir", c.Dir, "--mode", c.Mode, "--action", action, "--key", s.Key, "--payload", s.Candidate}
	if hold {
		args = append(args, "--hold")
	}
	cmd := exec.Command(exe, args...)
	cmd.Stderr = os.Stderr
	input, e := cmd.StdinPipe()
	if e != nil {
		return receipt{}, e
	}
	defer input.Close()
	output, e := cmd.StdoutPipe()
	if e != nil {
		return receipt{}, e
	}
	if e = cmd.Start(); e != nil {
		return receipt{}, e
	}
	reader := bufio.NewReader(output)
	line, e := reader.ReadString('\n')
	if e != nil {
		cmd.Wait()
		return receipt{}, e
	}
	if hold && line == "SINK_COMMITTED_REPLY_WITHHELD\n" {
		// No receipt has crossed the worker boundary. SIGKILL closes input; the sink exits.
		if e = boundary(c, "sink_committed"); e != nil {
			input.Close()
			cmd.Wait()
			return receipt{}, e
		}
		if _, e = input.Write([]byte("r")); e != nil {
			cmd.Wait()
			return receipt{}, e
		}
		line, e = reader.ReadString('\n')
		if e != nil {
			cmd.Wait()
			return receipt{}, e
		}
	}
	input.Close()
	io.Copy(io.Discard, reader)
	if e = cmd.Wait(); e != nil {
		return receipt{}, e
	}
	var r receipt
	e = json.Unmarshal([]byte(line), &r)
	return r, e
}
