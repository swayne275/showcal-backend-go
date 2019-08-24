// TODO figure out how to test network code (mock), get json from API
// to run against, and how to test fcns calling other fcns

package tvshowdata

import (
	"testing"
	"time"
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

func TestCheckForCandidateShows(t *testing.T) {
	cases := []struct {
		inData      string
		inQuery     string
		expectedOut bool
		expectErr   bool
	}{
		{
			inData:      "{\"page\":1,\"pages\":0,\"tv_shows\":[]}",
			inQuery:     "nototal",
			expectedOut: false,
			expectErr:   true,
		},
		{
			inData:      "{\"total\":0,\"page\":1,\"pages\":0,\"tv_shows\":[]}",
			inQuery:     "badtotal-type",
			expectedOut: false,
			expectErr:   true,
		},
		{
			inData:      "{\"total\":\"hello\",\"page\":1,\"pages\":0,\"tv_shows\":[]}",
			inQuery:     "badtotal-value",
			expectedOut: false,
			expectErr:   true,
		},
		{
			inData:      "{\"total\":\"1\",\"page\":1,\"pages\":0}",
			inQuery:     "noshows-missing",
			expectedOut: false,
			expectErr:   true,
		},
		{
			inData:      "{\"total\":\"1\",\"page\":1,\"pages\":0,\"tv_shows\":\"hello\"}",
			inQuery:     "noshows-type",
			expectedOut: false,
			expectErr:   true,
		},
		{
			inData:      "{\"total\":\"0\",\"page\":1,\"pages\":0,\"tv_shows\":[]}",
			inQuery:     "somerandomjunk",
			expectedOut: false,
			expectErr:   false,
		},
		{
			inData:      "{\"total\":\"1\",\"page\":1,\"pages\":1,\"tv_shows\":[{\"id\":2550,\"name\":\"American Dad!\",\"permalink\":\"american-dad\",\"start_date\":\"2005-02-06\",\"end_date\":null,\"country\":\"US\",\"network\":\"TBS\",\"status\":\"Running\",\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\"}]}",
			inQuery:     "American Dad",
			expectedOut: true,
			expectErr:   false,
		},
	}

	for _, c := range cases {
		out, err := checkForCandidateShows(c.inData, c.inQuery)
		gotErr := (err != nil)

		if gotErr != c.expectErr {
			t.Errorf("incorrect output error for '%s': expected '%t', got '%t'",
				c.inQuery, c.expectErr, gotErr)
		}

		if out != c.expectedOut {
			t.Errorf("incorrect output for '%s': expected '%t', got '%t'",
				c.inQuery, c.expectedOut, out)
		}
	}
}

func TestCheckForFutureEpisodes(t *testing.T) {
	cases := []struct {
		name        string
		inData      string
		inQuery     int64
		expectedOut bool
		expectedErr bool
	}{
		{
			name:        "2550 good data",
			inData:      "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2119-08-27 02:00:00\"}}}",
			inQuery:     2550,
			expectedOut: true,
			expectedErr: false,
		},
		{
			name:        "2550 missing object",
			inData:      "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\"}}",
			inQuery:     2550,
			expectedOut: false,
			expectedErr: true,
		},
		{
			name:        "2550 null object",
			inData:      "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"countdown\":null}}",
			inQuery:     2550,
			expectedOut: false,
			expectedErr: false,
		},
	}

	for _, c := range cases {
		out, err := checkForFutureEpisodes(c.inData, c.inQuery)
		gotErr := (err != nil)

		if gotErr != c.expectedErr {
			t.Errorf("incorrect output error for '%s': expected '%t', got '%t'",
				c.name, c.expectedErr, gotErr)
		}

		if out != c.expectedOut {
			t.Errorf("incorrect output error for '%s': expected '%t', got '%t'",
				c.name, c.expectedErr, gotErr)
		}
	}
}

func TestParseCandidateShows(t *testing.T) {
	cases := []struct {
		name        string
		inData      string
		expectedOut Shows
		expectErr   bool
	}{
		{
			name:   "good data, two shows",
			inData: "{\"total\":\"1\",\"page\":1,\"pages\":1,\"tv_shows\":[{\"id\":2550,\"name\":\"American Dad!\",\"permalink\":\"american-dad\",\"start_date\":\"2005-02-06\",\"end_date\":null,\"country\":\"US\",\"network\":\"TBS\",\"status\":\"Running\",\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\"},{\"id\":25501,\"name\":\"American Dad1!\",\"permalink\":\"american-dad\",\"start_date\":\"2005-02-06\",\"end_date\":null,\"country\":\"US\",\"network\":\"TBS\",\"status\":\"Running\",\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\"}]}",
			expectedOut: Shows{[]Show{
				Show{
					Name:         "American Dad!",
					ID:           2550,
					StillRunning: Running{true},
				},
				Show{
					Name:         "American Dad1!",
					ID:           25501,
					StillRunning: Running{true},
				},
			}},
			expectErr: false,
		},
		{
			name:        "missing tv_shows",
			inData:      "{\"total\":\"1\",\"page\":1,\"pages\":1}",
			expectedOut: Shows{},
			expectErr:   true,
		},
		{
			name:        "incorrect show format",
			inData:      "{\"total\":\"1\",\"page\":1,\"pages\":1,\"tv_shows\":[{\"s1\":2550,\"s2\":\"American Dad!\",\"s3\":\"Running\"}]}",
			expectedOut: Shows{},
			expectErr:   true,
		},
		{
			name:        "bad show data",
			inData:      "{\"total\":\"1\",\"page\":1,\"pages\":1,\"tv_shows\":[{hello}]}",
			expectedOut: Shows{},
			expectErr:   true,
		},
	}

	for _, c := range cases {
		out, err := parseCandidateShows(c.inData)
		gotErr := (err != nil)

		if gotErr != c.expectErr {
			t.Errorf("incorrect output error for '%s': expected '%t', got '%t'",
				c.name, c.expectErr, gotErr)
		}

		for idx, show := range c.expectedOut.Shows {
			if out.Shows[idx] != show {
				t.Errorf("incorrect output for '%s': expected '%+v', got '%+v'",
					c.name, show, out.Shows[idx])
			}
		}
	}
}

// convert time into episode format, throw away error (convenience)
func timeParseNoErr(format, timeToConvert string) Time {
	convertedTime, _ := time.Parse(timeStrFormat, timeToConvert)
	return Time{convertedTime}
}

func TestParseUpcomingEpisodes(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    Episodes
		wantErr bool
	}{
		{
			name:  "good data, two episodes",
			input: "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"description\":\"cut\",\"status\":\"Running\",\"runtime\":30,\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\",\"rating\":\"9.0625\",\"rating_count\":\"16\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2019-08-27 02:00:00\"},\"episodes\":[{\"season\":15,\"episode\":21,\"name\":\"Downtown\",\"air_date\":\"2119-09-03 02:00:00\"},{\"season\":15,\"episode\":22,\"name\":\"Cheek to Cheek: A Stripper's Story\",\"air_date\":\"2119-09-10 02:00:00\"}]}}",
			want: Episodes{[]Episode{
				Episode{
					Season:         15,
					Episode:        21,
					Title:          "Downtown",
					AirDate:        timeParseNoErr(timeStrFormat, "2119-09-03 02:00:00"),
					RuntimeMinutes: 30,
					ShowName:       "American Dad!",
				},
				Episode{
					Season:         15,
					Episode:        22,
					Title:          "Cheek to Cheek: A Stripper's Story",
					AirDate:        timeParseNoErr(timeStrFormat, "2119-09-10 02:00:00"),
					RuntimeMinutes: 30,
					ShowName:       "American Dad!",
				},
			}},
			wantErr: false,
		},
		{
			name:    "no episodes",
			input:   "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"description\":\"cut\",\"status\":\"Running\",\"runtime\":30,\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\",\"rating\":\"9.0625\",\"rating_count\":\"16\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2019-08-27 02:00:00\"}}}",
			want:    Episodes{},
			wantErr: true,
		},
		{
			name:    "no show name",
			input:   "{\"tvShow\":{\"id\":2550,\"description\":\"cut\",\"status\":\"Running\",\"runtime\":30,\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\",\"rating\":\"9.0625\",\"rating_count\":\"16\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2019-08-27 02:00:00\"},\"episodes\":[{\"season\":15,\"episode\":21,\"name\":\"Downtown\",\"air_date\":\"2119-09-03 02:00:00\"},{\"season\":15,\"episode\":22,\"name\":\"Cheek to Cheek: A Stripper's Story\",\"air_date\":\"2119-09-10 02:00:00\"}]}}",
			want:    Episodes{},
			wantErr: true,
		},
		{
			name:    "no run time",
			input:   "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"description\":\"cut\",\"status\":\"Running\",\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\",\"rating\":\"9.0625\",\"rating_count\":\"16\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2019-08-27 02:00:00\"},\"episodes\":[{\"season\":15,\"episode\":21,\"name\":\"Downtown\",\"air_date\":\"2119-09-03 02:00:00\"},{\"season\":15,\"episode\":22,\"name\":\"Cheek to Cheek: A Stripper's Story\",\"air_date\":\"2119-09-10 02:00:00\"}]}}",
			want:    Episodes{},
			wantErr: true,
		},
		{
			name:    "bad episode data",
			input:   "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"description\":\"cut\",\"status\":\"Running\",\"runtime\":30,\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\",\"rating\":\"9.0625\",\"rating_count\":\"16\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2019-08-27 02:00:00\"},\"episodes\":[{5},{\"season\":15,\"episode\":22,\"name\":\"Cheek to Cheek: A Stripper's Story\",\"air_date\":\"2119-09-10 02:00:00\"}]}}",
			want:    Episodes{},
			wantErr: true,
		},
		{
			name:    "bad episode data",
			input:   "{\"tvShow\":{\"id\":2550,\"name\":\"American Dad!\",\"description\":\"cut\",\"status\":\"Running\",\"runtime\":30,\"image_thumbnail_path\":\"https://static.episodate.com/images/tv-show/thumbnail/2550.jpg\",\"rating\":\"9.0625\",\"rating_count\":\"16\",\"countdown\":{\"season\":15,\"episode\":20,\"name\":\"The Hand that Rocks the Rogu\",\"air_date\":\"2019-08-27 02:00:00\"},\"episodes\":[{\"s1\":15,\"s2\":22,\"s3\":\"Cheek to Cheek: A Stripper's Story\",\"s4\":\"2119-09-10 02:00:00\"}]}}",
			want:    Episodes{},
			wantErr: true,
		},
	}

	for _, c := range cases {
		got, err := parseUpcomingEpisodes(c.input)
		gotErr := (err != nil)

		if gotErr != c.wantErr {
			t.Errorf("incorrect output error for '%s': expected '%t', got '%t'",
				c.name, c.wantErr, gotErr)
		}

		for idx, episode := range c.want.Episodes {
			if got.Episodes[idx] != episode {
				t.Errorf("incorrect output for '%s': expected '%+v', got '%+v'",
					c.name, episode, got.Episodes[idx])
			}
		}
	}
}
