package fribidi

import "testing"

func assert(t *testing.T, cond bool) {
	if !cond {
		t.Error("assertion failed")
	}
}

func (runList *Run) validate2(t *testing.T) {
	assert(t, runList != nil)
	assert(t, runList.next != nil)
	assert(t, runList.next.prev == runList)
	assert(t, runList.type_ == maskSENTINEL)
	var q *Run
	for q = runList.next; q.type_ != maskSENTINEL; q = q.next {
		assert(t, q.next != nil)
		assert(t, q.next.prev == q)
	}
	assert(t, q == runList)
}
