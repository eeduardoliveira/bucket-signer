package bucket

import (
	"testing"
	"time"
)

func TestParseExpiration(t *testing.T) {
	cases := map[string]time.Duration{
		"":      DefaultExpiration,
		"5":     5 * time.Minute,
		"60":    60 * time.Minute,
		"61":    MaxExpiration,
		"10080": MaxExpiration,
		"0":     DefaultExpiration,
		"-3":    DefaultExpiration,
		"abc":   DefaultExpiration,
		"1h":    DefaultExpiration,
	}
	for in, want := range cases {
		if got := ParseExpiration(in); got != want {
			t.Errorf("ParseExpiration(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestParseKeyPattern(t *testing.T) {
	if p, err := ParseKeyPattern(""); err != nil || p != DefaultKeyPattern {
		t.Fatalf("empty pattern: %q %v", p, err)
	}
	for _, ok := range []string{"%s/%s-prompt.json", "prompts/%s.json"} {
		if _, err := ParseKeyPattern(ok); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{"static.json", "%s/%s/%s", "%d/%s", "../%s", "/%s", "%s/%v"} {
		if _, err := ParseKeyPattern(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestBuildKey(t *testing.T) {
	if got := BuildKey("%s/%s-prompt.json", "c1"); got != "c1/c1-prompt.json" {
		t.Fatal(got)
	}
	if got := BuildKey("prompts/%s.json", "c1"); got != "prompts/c1.json" {
		t.Fatal(got)
	}
}
