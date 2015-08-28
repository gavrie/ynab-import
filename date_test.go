package main

import "testing"

func TestDateParse(t *testing.T) {
	dates := []string{
		"2014-09-30T00:00:00",
		"30/09/14",
	}

	for _, d := range dates {
		date, err := parseDate(d)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(date)
	}
}
