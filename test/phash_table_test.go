package tests

import (
	"encoding/json"
	"testing"

	"github.com/ast-jean/audiophash/cmd/audiophash" // adjust module path to your module
	"github.com/ast-jean/audiophash/pkg/config"
)

type TestCase struct {
	ID       string  `json:"id"`
	Base     string  `json:"base"`
	Variant  string  `json:"variant"`
	ExpectOp string  `json:"expectOp"` // "<=" or ">="
	Percent  float64 `json:"percent"`
}

func loadManifest(t *testing.T, path string) []TestCase {
	t.Helper()
	b := loadFile(t, path)
	var cases []TestCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	return cases
}

func TestPHashTable(t *testing.T) {
	cases := loadManifest(t, "test/tests.json")

	// Use default config or customize
	cfg := config.DefaultConfig(44100)
	_ = cfg.ValidateAndFill()

	for _, tc := range cases {
		tc := tc // capture
		t.Run(tc.ID, func(t *testing.T) {
			t.Parallel()

			// load files
			b1 := loadFile(t, tc.Base)
			b2 := loadFile(t, tc.Variant)

			// compute hashes (assumes detectors accept raw file bytes and format detection internally)
			h1, err := audiophash.AudioPHashBytes(b1, &cfg)
			if err != nil {
				t.Fatalf("hash base error: %v", err)
			}
			h2, err := audiophash.AudioPHashBytes(b2, &cfg)
			if err != nil {
				t.Fatalf("hash variant error: %v", err)
			}

			u1, err := HexToUint64(h1)
			if err != nil {
				t.Fatalf("hex decode h1: %v", err)
			}
			u2, err := HexToUint64(h2)
			if err != nil {
				t.Fatalf("hex decode h2: %v", err)
			}

			pct := HammingPercent(u1, u2)
			switch tc.ExpectOp {
			case "<=":
				if pct > tc.Percent {
					t.Fatalf("FAILED %s: percent=%v > allowed %v (h1=%s h2=%s)", tc.ID, pct, tc.Percent, h1, h2)
				}
			case ">=":
				if pct < tc.Percent {
					t.Fatalf("FAILED %s: percent=%v < required %v (h1=%s h2=%s)", tc.ID, pct, tc.Percent, h1, h2)
				}
			default:
				t.Fatalf("invalid expectOp %s", tc.ExpectOp)
			}
		})
	}
}
