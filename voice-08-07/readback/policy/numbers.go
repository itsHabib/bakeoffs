package policy

import (
	"strconv"
	"strings"
)

var unitWords = map[string]int{
	"zero": 0, "oh": 0, "one": 1, "two": 2, "three": 3, "four": 4,
	"five": 5, "six": 6, "seven": 7, "eight": 8, "nine": 9,
}

var teenWords = map[string]int{
	"ten": 10, "eleven": 11, "twelve": 12, "thirteen": 13, "fourteen": 14,
	"fifteen": 15, "sixteen": 16, "seventeen": 17, "eighteen": 18, "nineteen": 19,
}

var tensWords = map[string]int{
	"twenty": 20, "thirty": 30, "forty": 40, "fifty": 50,
	"sixty": 60, "seventy": 70, "eighty": 80, "ninety": 90,
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isNumberish(tok string) bool {
	if isDigits(tok) {
		return true
	}
	_, u := unitWords[tok]
	_, t := teenWords[tok]
	_, x := tensWords[tok]
	return u || t || x
}

// numberFrom converts a run of spoken-number tokens to one integer.
// Single digits concatenate ("one four" -> 14, matching how a readback
// spells digits out); a tens word absorbs a following unit ("forty five"
// -> 45); teens and literal digit strings stand alone ("fourteen" -> 14,
// "40" -> 40). Any non-number token fails the whole run: no guessing.
func numberFrom(toks []string) (int, bool) {
	var b strings.Builder
	i := 0
	for i < len(toks) {
		tok := toks[i]
		if d, ok := unitWords[tok]; ok {
			b.WriteString(strconv.Itoa(d))
			i++
			continue
		}
		if v, ok := teenWords[tok]; ok {
			b.WriteString(strconv.Itoa(v))
			i++
			continue
		}
		if v, ok := tensWords[tok]; ok {
			if i+1 < len(toks) {
				if d, ok := unitWords[toks[i+1]]; ok && d > 0 {
					b.WriteString(strconv.Itoa(v + d))
					i += 2
					continue
				}
			}
			b.WriteString(strconv.Itoa(v))
			i++
			continue
		}
		if isDigits(tok) {
			b.WriteString(tok)
			i++
			continue
		}
		return 0, false
	}
	if b.Len() == 0 || b.Len() > 4 {
		return 0, false
	}
	n, err := strconv.Atoi(b.String())
	if err != nil {
		return 0, false
	}
	return n, true
}

var digitNames = [10]string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

// spokenDigits renders a number the way a readback speaks it: each digit
// individually, so 14 is "one four" and can never be misheard as forty.
func spokenDigits(n int) string {
	s := strconv.Itoa(n)
	parts := make([]string, 0, len(s))
	for _, r := range s {
		parts = append(parts, digitNames[r-'0'])
	}
	return strings.Join(parts, " ")
}

// spokenList joins spoken numbers for a clarify line: "one four and four zero".
func spokenList(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = spokenDigits(n)
	}
	if len(parts) <= 1 {
		return strings.Join(parts, "")
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}
