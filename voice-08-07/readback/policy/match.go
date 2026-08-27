package policy

import "strings"

// Match resolves a spoken name against a closed candidate set.
// Fuzz tolerates recognizer slop (edit distance scaled to length) but the
// RESULT is always an exact member of the set — never a guess between two
// candidates: a tie for best is reported as ambiguous, not resolved.
func Match(spoken string, candidates []string) (name string, ambiguous, ok bool) {
	target := squash(spoken)
	if target == "" {
		return "", false, false
	}
	best, second := -1, -1
	for _, c := range candidates {
		d := levenshtein(target, squash(c))
		if best == -1 || d < best {
			best, second, name = d, best, c
			continue
		}
		if second == -1 || d < second {
			second = d
		}
	}
	if best == -1 || best > maxDist(len(target)) {
		return "", false, false
	}
	if second == best {
		return "", true, false
	}
	return name, false, true
}

func maxDist(n int) int {
	if n <= 4 {
		return 1
	}
	return 2
}

// squash lowercases and strips everything but letters and digits, so
// "dark mode", "dark-mode" and "Dark Mode" are the same name.
func squash(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
