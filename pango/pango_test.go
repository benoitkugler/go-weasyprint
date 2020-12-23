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

func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("expected same values, got %v and %v", a, b)
	}
}
