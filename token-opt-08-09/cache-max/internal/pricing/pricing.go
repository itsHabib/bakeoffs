// Package pricing loads the single editable $/MTok table. It holds no analysis
// logic — just the rates and the TODO/example fallback.
package pricing

import (
	"encoding/json"
	"os"
)

// Rates is the price of each billable token class, in USD per million tokens.
type Rates struct {
	FreshInput   float64 `json:"fresh_input"`
	Output       float64 `json:"output"`
	CacheRead    float64 `json:"cache_read"`
	CacheWrite5m float64 `json:"cache_write_5m"`
	CacheWrite1h float64 `json:"cache_write_1h"`
}

// filled reports whether every rate has been given a real (non-TODO) value.
func (r Rates) filled() bool {
	return r.FreshInput > 0 && r.Output > 0 && r.CacheRead > 0 &&
		r.CacheWrite5m > 0 && r.CacheWrite1h > 0
}

type file struct {
	Rates   Rates `json:"rates_usd_per_mtok"`
	Example Rates `json:"example_rates_usd_per_mtok"`
}

// Load reads the table. If the real rates are still TODO (any zero), it returns
// the example rates with usingExample=true so callers can flag the output.
func Load(path string) (rates Rates, usingExample bool, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Rates{}, false, err
	}
	var f file
	if err := json.Unmarshal(b, &f); err != nil {
		return Rates{}, false, err
	}
	if f.Rates.filled() {
		return f.Rates, false, nil
	}
	return f.Example, true, nil
}

// USD converts a token count at a per-MTok rate to dollars.
func USD(tokens int, ratePerMTok float64) float64 {
	return float64(tokens) / 1_000_000 * ratePerMTok
}
