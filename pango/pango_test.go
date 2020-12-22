package pango

import "testing"

func assertFalse(t *testing.T, b bool, message string) {
	if b {
		t.Error(message + ": expected false, got true")
	}
}
func assertTrue(t *testing.T, b bool, message string) {
	if !b {
		t.Error(message + ": expected true, got false")
	}
}

func assertEqualInt(t *testing.T, a, b int) {
	if a != b {
		t.Errorf("expected same values, got %d and %d", a, b)
	}
}

func assertEqualUInt(t *testing.T, a, b uint32) {
	if a != b {
		t.Errorf("expected same values, got %d and %d", a, b)
	}
}
