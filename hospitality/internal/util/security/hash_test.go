package security

import (
	"bytes"
	"testing"
)

func TestLookupHashIsKeyedAndDeterministic(t *testing.T) {
	one := LookupHash("secret", "smith")
	two := LookupHash("secret", "smith")
	other := LookupHash("other", "smith")
	if !bytes.Equal(one, two) {
		t.Fatal("same key/value must produce same hash")
	}
	if bytes.Equal(one, other) {
		t.Fatal("different keys must produce different hashes")
	}
}
