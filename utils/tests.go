package utils

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

type capturedLogs struct {
	stack []string
}

func (c *capturedLogs) Write(p []byte) (int, error) {
	b := new(bytes.Buffer)
	i, err := b.Write(p)
	if err != nil {
		return i, err
	}
	c.stack = append(c.stack, strings.TrimSuffix(b.String(), "\n"))
	return i, nil
}

func CaptureLogs() *capturedLogs {
	out := capturedLogs{}
	log.SetOutput(&out)
	return &out
}

func (c capturedLogs) Logs() []string {
	return c.stack
}

// CheckEqual compares logs ignoring date time in logged.
func (c capturedLogs) CheckEqual(refs []string, t *testing.T) {
	gots := c.Logs()
	for i, ref := range refs {
		g := gots[i][20:]
		if g != ref {
			t.Fatalf("expected \n%s\n got \n%s", ref, g)
		}
	}
}

func (c capturedLogs) AssertNoLogs(t *testing.T) {
	l := c.Logs()
	if len(l) > 0 {
		t.Fatalf("expected no logs, got %v (%d)", l, len(l))
	}
}
