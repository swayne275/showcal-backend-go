// TODO figure out how to test network code (mock), get json from API
// to run against, and how to test fcns calling other fcns

package tvshowdata

import (
	"testing"
)

// Tests Time method UnmarshalJSON()
func TestTimeUnmarshalJSON(t *testing.T) {
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
		{[]byte{}, true},
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
func TestRunningMarshalJSON(t *testing.T) {
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

// Tests Running method MarshalJSON()
func TestRunningUnmarshalJSON(t *testing.T) {
	var running Running
	cases := []struct {
		input           []byte
		shouldHaveError bool
	}{
		{[]byte("\"Running\""), false},
		{[]byte("\"Ended\""), false},
		{[]byte("bad json"), true},
		{[]byte{}, true},
	}

	for _, c := range cases {
		err := running.UnmarshalJSON(c.input)
		haveError := (err != nil)

		if haveError != c.shouldHaveError {
			t.Errorf("incorrect error status for '%s': expected '%t', got '%t'",
				c.input, c.shouldHaveError, haveError)
		}
	}
}

func TestGetShowURLs(t *testing.T) {
	// test getShowSearchURL
	searchCases := []struct {
		in              string
		expectedOut     string
		expectedHaveErr bool
	}{
		{"American Dad", "https://www.episodate.com/api/search?q=American+Dad", false},
		{"Friends", "https://www.episodate.com/api/search?q=Friends", false},
		{"", "", true},
	}

	for _, c := range searchCases {
		out, err := getShowSearchURL(c.in)
		haveErr := (err != nil)

		if out != c.expectedOut {
			t.Errorf("incorrect output for '%s': expected '%s', got '%s'",
				c.in, c.expectedOut, out)
		}

		if haveErr != c.expectedHaveErr {
			t.Errorf("incorrect error status for '%s': expected '%t', got '%t'",
				c.in, c.expectedHaveErr, haveErr)
		}
	}

	// test getShowDetailsURL()
	detailsCases := []struct {
		in          int64
		expectedOut string
	}{
		{0, "https://episodate.com/api/show-details?q=0"},
		{2550, "https://episodate.com/api/show-details?q=2550"},
	}

	for _, c := range detailsCases {
		out := getShowDetailsURL(c.in)

		if out != c.expectedOut {
			t.Errorf("incorrect output for '%d': expected '%s', got '%s'",
				c.in, c.expectedOut, out)
		}
	}
}
