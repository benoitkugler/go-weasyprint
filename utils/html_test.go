package utils

import (
	"testing"
)

var dates = [7]string{
	"1997",
	"1997-07",
	"1997-07-16",
	"1997-07-16T19:20+01:00",
	"1997-07-16T19:20:30+01:00",
	"1997-07-16T19:20:30.45+01:00",
	"1997-07-16T19:20:30.45-01:15",
}

func TestReW3CDate(t *testing.T) {
	for _, date := range dates {
		if !W3CDateRe.MatchString(date) {
			t.Errorf("date %s not matching", date)
		}

		if parseW3cDate("test", date).IsZero() {
			t.Errorf("date %s not parsed", date)
		}
	}

}
