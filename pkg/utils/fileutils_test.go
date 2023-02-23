package utils

import "testing"

func TestGetFnameOnly(t *testing.T) {
	fname := "test.txt"
	fnameOnly := getFnameOnly(fname)
	if fnameOnly != "test" {
		t.Errorf("Expected fnameOnly to be 'test', got '%s'", fnameOnly)
	}

	fname = "test.testext.txt"
	fnameOnly = getFnameOnly(fname)
	if fnameOnly != "test.testext" {
		t.Errorf("Expected fnameOnly to be 'test.testext', got '%s'", fnameOnly)
	}

	fname = "/path/to/test.txt"
	fnameOnly = getFnameOnly(fname)
	if fnameOnly != "test" {
		t.Errorf("Expected fnameOnly to be 'test', got '%s'", fnameOnly)
	}
}

