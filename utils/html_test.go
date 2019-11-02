package utils

import (
	"testing"
)

var dates = [6]string{
	"1997",
	"1997-07",
	"1997-07-16",
	"1997-07-16T19:20+01:00",
	"1997-07-16T19:20:30+01:00",
	"1997-07-16T19:20:30.45+01:00",
}

func TestReW3CDate(t *testing.T) {
	for _, date := range dates {
		if !W3CDateRe.MatchString(date) {
			t.Errorf("date %s not matching", date)
		}
	}
}
