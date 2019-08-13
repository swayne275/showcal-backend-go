package tvshowdata

import (
	"testing"
)

// Tests Time method UnmarshalJSON()
func TestUnmarshalJSON(t *testing.T) {
	var time Time
	cases := []struct {
		input           []byte
		shouldHaveError bool
	}{
		{[]byte("\"2019-09-03 02:00:00\""), false},
		{[]byte("\"2019-09-03T02:00:00Z\""), false},
		{[]byte("bad json"), true},
		{[]byte("\"not a time\""), true},
		{[]byte("\"03/09/2019, 02:00:00\""), true},
	}

	for _, c := range cases {
		err := time.UnmarshalJSON(c.input)
		gotError := (err != nil)

		if gotError != c.shouldHaveError {
			t.Errorf("incorrect output for '%s': expected '%t', got '%t'",
				c.input, c.shouldHaveError, gotError)
		}
	}
}

// Tests Running method MarshalJSON()
func TestMarshalJSON(t *testing.T) {
	cases := []struct {
		input             Running
		expectedOutputStr string
		shouldHaveError   bool
	}{
		{Running{true}, "true", false},
		{Running{false}, "false", false},
	}

	for _, c := range cases {
		out, err := c.input.MarshalJSON()
		outStr := string(out)
		gotError := (err != nil)

		if outStr != c.expectedOutputStr {
			t.Errorf("incorrect output for '%+v': expected '%s', got '%s'",
				c.input, c.expectedOutputStr, outStr)
		}

		if gotError != c.shouldHaveError {
			t.Errorf("incorrect output for '%+v': expected '%t', got '%t'",
				c.input, c.shouldHaveError, gotError)
		}
	}
}
