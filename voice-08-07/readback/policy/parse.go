package policy

import (
	"fmt"
	"strings"
)

type Outcome int

const (
	Parsed Outcome = iota
	Clarified
	Rejected
)

// ParseResult is either an exact Order, a spoken clarifying question, or a
// spoken rejection. There is no fourth state: ambiguity is never resolved by
// guessing.
type ParseResult struct {
	Outcome Outcome
	Order   Order
	Say     string
}

func clarify(format string, args ...any) ParseResult {
	return ParseResult{Outcome: Clarified, Say: fmt.Sprintf(format, args...)}
}

func reject(format string, args ...any) ParseResult {
	return ParseResult{Outcome: Rejected, Say: fmt.Sprintf(format, args...)}
}

func parsed(o Order) ParseResult {
	return ParseResult{Outcome: Parsed, Order: o}
}

var fillerWords = map[string]bool{"please": true, "the": true, "a": true, "number": true}

// Normalize turns raw heard text into lowercase alphanumeric tokens.
// Punctuation and hyphens become spaces ("PR-14" -> "pr 14"), filler words
// drop, and a spelled-out "p r" collapses to "pr".
func Normalize(text string) []string {
	mapped := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return ' '
		}
	}, text)
	raw := strings.Fields(mapped)
	toks := make([]string, 0, len(raw))
	i := 0
	for i < len(raw) {
		if raw[i] == "p" && i+1 < len(raw) && raw[i+1] == "r" {
			toks = append(toks, "pr")
			i += 2
			continue
		}
		if !fillerWords[raw[i]] {
			toks = append(toks, raw[i])
		}
		i++
	}
	return toks
}

const sayAgain = "say again. commands are merge, deploy, flag, and roll back."

// Parse turns one normalized utterance into a ParseResult against the world.
func Parse(w *World, tokens []string) ParseResult {
	if len(tokens) == 0 {
		return reject(sayAgain)
	}
	rest := tokens[1:]
	switch tokens[0] {
	case "merge":
		return parseMerge(w, rest)
	case "deploy":
		return parseDeploy(w, rest)
	case "flag", "set":
		return parseFlag(w, rest)
	case "rollback":
		return parseRollback(w, rest)
	case "roll":
		if len(rest) > 0 && rest[0] == "back" {
			return parseRollback(w, rest[1:])
		}
	}
	return reject(sayAgain)
}

func parseMerge(w *World, rest []string) ParseResult {
	repoToks, numToks := splitMerge(rest)
	if len(repoToks) == 0 {
		return reject("say merge, then the repo name.")
	}
	names := make([]string, len(w.Repos))
	for i, r := range w.Repos {
		names[i] = r.Name
	}
	spokenRepo := strings.Join(repoToks, " ")
	name, ambiguous, ok := Match(spokenRepo, names)
	if ambiguous {
		return clarify("more than one repo sounds like %s. say the full repo name.", spokenRepo)
	}
	if !ok {
		return reject("no repo called %s. repos are %s.", spokenRepo, strings.Join(names, ", "))
	}
	repo := w.repo(name)
	if len(numToks) == 0 {
		return mergeWithoutNumber(repo)
	}
	n, ok := numberFrom(numToks)
	if !ok {
		return reject("I did not catch the PR number. say merge %s P R, then the number.", repo.Name)
	}
	if !repo.prOpen(n) {
		return reject("%s has no open PR %s. open PRs are %s.", repo.Name, spokenDigits(n), spokenList(repo.OpenPRs))
	}
	return parsed(Order{Verb: VerbMerge, Repo: repo.Name, PR: n})
}

// splitMerge divides "roxiq pr one four" into repo tokens and number tokens,
// with or without an explicit "pr" / "pull request" keyword.
func splitMerge(rest []string) (repoToks, numToks []string) {
	for i, tok := range rest {
		if tok != "pr" && tok != "pull" {
			continue
		}
		num := rest[i+1:]
		if len(num) > 0 && num[0] == "request" {
			num = num[1:]
		}
		return rest[:i], num
	}
	cut := len(rest)
	for cut > 0 && isNumberish(rest[cut-1]) {
		cut--
	}
	return rest[:cut], rest[cut:]
}

func mergeWithoutNumber(repo *Repo) ParseResult {
	switch len(repo.OpenPRs) {
	case 0:
		return reject("%s has no open PRs.", repo.Name)
	case 1:
		return parsed(Order{Verb: VerbMerge, Repo: repo.Name, PR: repo.OpenPRs[0]})
	}
	return clarify("%s has %d open PRs: %s. say merge %s P R, then the number.",
		repo.Name, len(repo.OpenPRs), spokenList(repo.OpenPRs), repo.Name)
}

func parseDeploy(w *World, rest []string) ParseResult {
	svcToks, envToks := rest, []string(nil)
	for i, tok := range rest {
		if tok == "to" {
			svcToks, envToks = rest[:i], rest[i+1:]
			break
		}
	}
	if len(svcToks) == 0 {
		return reject("say deploy, then the service name.")
	}
	names := make([]string, len(w.Services))
	for i, s := range w.Services {
		names[i] = s.Name
	}
	spokenSvc := strings.Join(svcToks, " ")
	name, ambiguous, ok := Match(spokenSvc, names)
	if ambiguous {
		return clarify("more than one service sounds like %s. say the full service name.", spokenSvc)
	}
	if !ok {
		return reject("no service called %s. services are %s.", spokenSvc, strings.Join(names, ", "))
	}
	svc := w.service(name)
	if len(envToks) == 0 {
		return clarify("deploy %s where? say deploy %s to staging, or to production.", svc.Name, svc.Name)
	}
	env, _, ok := Match(strings.Join(envToks, " "), []string{"staging", "production", "prod"})
	if !ok {
		return reject("environments are staging and production.")
	}
	if env == "prod" {
		env = "production"
	}
	version := bumpVersion(svc.Versions["staging"])
	if env == "production" {
		version = svc.Versions["staging"]
	}
	return parsed(Order{Verb: VerbDeploy, Service: svc.Name, Env: env, Version: version})
}

func parseRollback(w *World, rest []string) ParseResult {
	if len(rest) == 0 {
		return reject("say roll back, then the service name.")
	}
	names := make([]string, len(w.Services))
	for i, s := range w.Services {
		names[i] = s.Name
	}
	spokenSvc := strings.Join(rest, " ")
	name, ambiguous, ok := Match(spokenSvc, names)
	if ambiguous {
		return clarify("more than one service sounds like %s. say the full service name.", spokenSvc)
	}
	if !ok {
		return reject("no service called %s. services are %s.", spokenSvc, strings.Join(names, ", "))
	}
	svc := w.service(name)
	if svc.PrevProd == "" {
		return reject("nothing to roll %s back to.", svc.Name)
	}
	return parsed(Order{Verb: VerbRollback, Service: svc.Name, Env: "production", Version: svc.PrevProd})
}

func parseFlag(w *World, rest []string) ParseResult {
	if len(rest) < 2 {
		return clarify("say flag, the flag name, then on or off.")
	}
	last := rest[len(rest)-1]
	if last != "on" && last != "off" {
		return clarify("flag %s on, or off? say the whole order again.", strings.Join(rest, " "))
	}
	names := make([]string, len(w.Flags))
	for i, f := range w.Flags {
		names[i] = f.Name
	}
	spokenFlag := strings.Join(rest[:len(rest)-1], " ")
	name, ambiguous, ok := Match(spokenFlag, names)
	if ambiguous {
		return clarify("more than one flag sounds like %s. say the full flag name.", spokenFlag)
	}
	if !ok {
		return reject("no flag called %s. flags are %s.", spokenFlag, strings.Join(names, ", "))
	}
	want := last == "on"
	if w.flag(name).On == want {
		return reject("%s is already %s.", name, last)
	}
	return parsed(Order{Verb: VerbFlag, Flag: name, On: want})
}
