// test/phash_table_test.go
package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ast-jean/audiophash/cmd/audiophash"
	"github.com/ast-jean/audiophash/pkg/config"
)

type TestCase struct {
	ID       string  `json:"id"`
	Base     string  `json:"base"`
	Variant  string  `json:"variant"`
	ExpectOp string  `json:"expectOp"` // "<=" or ">="
	Percent  float64 `json:"percent"`
}

// TestMain optionally generates variants if needed, then runs tests.
func TestMain(m *testing.M) {
	// If variants dir missing, try to run generator script
	variantsDir := "test/fixtures/variants"
	genScript := "test/scripts/gen_variants.sh"

	if _, err := os.Stat(variantsDir); os.IsNotExist(err) {
		if _, err2 := os.Stat(genScript); err2 == nil {
			fmt.Println("variants directory missing; running generator script:", genScript)
			cmd := exec.Command("bash", genScript)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("failed to run generator script:", err)
				// continue anyway; tests will fail with missing files
			}
		}
	}

	os.Exit(m.Run())
}

// findManifest attempts to locate tests.json in a few likely locations and returns the bytes.
func findManifest() ([]byte, string, error) {
	candidates := []string{
		"test/tests.json",
		"tests.json",
		"../tests.json",
		"./tests.json",
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			b, err := ioutil.ReadFile(p)
			return b, p, err
		}
	}
	return nil, "", fmt.Errorf("manifest not found (tried: %v)", candidates)
}

// simple ones count (avoid extra imports)
func bitsOnesCount64(x uint64) int {
	// builtin popcount algorithm
	var cnt int
	for x != 0 {
		x &= x - 1
		cnt++
	}
	return cnt
}

func TestPHashTable(t *testing.T) {
	manifestBytes, manifestPath, err := findManifest()
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	t.Logf("using manifest: %s", manifestPath)

	var cases []TestCase
	if err := json.Unmarshal(manifestBytes, &cases); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	if len(cases) == 0 {
		t.Fatalf("no test cases found in manifest")
	}

	// Use default config and validate
	cfg := config.DefaultConfig(44100)
	if err := cfg.ValidateAndFill(); err != nil {
		t.Fatalf("invalid default config: %v", err)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			// do not parallelize here to avoid ffmpeg / disk contention in CI
			// t.Parallel()

			// Resolve base and variant path relative to manifest (helps when running tests from different cwd)
			manifestDir := filepath.Dir(manifestPath)
			basePath := tc.Base
			variantPath := tc.Variant
			if !filepath.IsAbs(basePath) {
				basePath = filepath.Join(manifestDir, tc.Base)
			}
			if !filepath.IsAbs(variantPath) {
				variantPath = filepath.Join(manifestDir, tc.Variant)
			}

			// read files
			b1, err := ioutil.ReadFile(basePath)
			if err != nil {
				t.Fatalf("read base %s: %v", basePath, err)
			}
			b2, err := ioutil.ReadFile(variantPath)
			if err != nil {
				t.Fatalf("read variant %s: %v", variantPath, err)
			}

			// determine format from extension (simple)
			formatFromExt := func(p string) string {
				ext := strings.ToLower(filepath.Ext(p))
				switch ext {
				case ".wav":
					return "wav"
				case ".mp3":
					return "mp3"
				case ".raw", ".pcm":
					return "pcm16le"
				default:
					return "wav" // safe default; decoders must handle or error
				}
			}

			f1 := formatFromExt(basePath)
			f2 := formatFromExt(variantPath)
			if f1 != f2 {
				// that's okay — our AudioPHashBytes will resample/handle formats individually
			}

			h1, err := audiophash.AudioPHashBytes(b1, &cfg, f1)
			if err != nil {
				t.Fatalf("hash base error: %v", err)
			}
			h2, err := audiophash.AudioPHashBytes(b2, &cfg, f2)
			if err != nil {
				t.Fatalf("hash variant error: %v", err)
			}

			u1, err := HexToUint64(h1)
			if err != nil {
				t.Fatalf("hex decode h1: %v (h1=%s)", err, h1)
			}
			u2, err := HexToUint64(h2)
			if err != nil {
				t.Fatalf("hex decode h2: %v (h2=%s)", err, h2)
			}

			d := HammingDistance(u1, u2)
			percent := float64(d) / 64.0 * 100.0

			t.Logf("%s: %s vs %s → Hamming=%d (%.2f%%)", tc.ID, filepath.Base(basePath), filepath.Base(variantPath), d, percent)

			switch tc.ExpectOp {
			case "<=":
				if percent > tc.Percent {
					t.Fatalf("FAILED %s: percent=%.2f > allowed %.2f (h1=%s h2=%s)", tc.ID, percent, tc.Percent, h1, h2)
				}
			case ">=":
				if percent < tc.Percent {
					t.Fatalf("FAILED %s: percent=%.2f < required %.2f (h1=%s h2=%s)", tc.ID, percent, tc.Percent, h1, h2)
				}
			default:
				t.Fatalf("invalid expectOp %q for test %s", tc.ExpectOp, tc.ID)
			}
		})
	}
}
