package names

import "testing"

func TestNormalize(t *testing.T) {
	for input, expected := range map[string]string{
		"  O'CONNOR  ":     "o'connor",
		"van   Der  Waals": "van der waals",
		"Müller<script>":   "müllerscript",
	} {
		if got := Normalize(input); got != expected {
			t.Errorf("Normalize(%q) = %q, want %q", input, got, expected)
		}
	}
}
