package fontconfig

import (
	"os"
	"testing"
)

// ported from fontconfig/test/test-bz1744377.c: 2000 Keith Packard
func TestParse(t *testing.T) {
	doc := []byte(`
	<fontconfig>
  		<include ignore_missing="yes">blahblahblah</include>
	</fontconfig>
	`)
	doc2 := []byte(`
	<fontconfig>
  		<include ignore_missing="no">blahblahblah</include>
	</fontconfig>
	`)
	cfg := NewFcConfig()

	if err := cfg.ParseAndLoadFromMemory(doc, os.Stdout); err != nil {
		t.Errorf("expected no error since 'ignore_missing' is true, got %s", err)
	}
	if err := cfg.ParseAndLoadFromMemory(doc2, os.Stdout); err == nil {
		t.Error("expected error on invalid include")
	}
}
