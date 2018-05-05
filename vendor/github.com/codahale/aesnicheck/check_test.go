package aesnicheck

import "testing"

func TestHasAESNI(t *testing.T) {
	t.Logf("Has AES-NI support: %v", HasAESNI()) // shouldn't explode, right?
}
