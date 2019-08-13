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
		outputHasError := (err != nil)
		if outputHasError != c.shouldHaveError {
			t.Errorf("incorrect output for '%s': expected '%t', got '%t'",
				c.input, c.shouldHaveError, outputHasError)
		}
	}
}
