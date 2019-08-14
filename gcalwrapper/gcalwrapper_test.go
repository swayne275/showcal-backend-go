package gcalwrapper

import (
	"testing"
	"time"

	"github.com/swayne275/showcal-backend-go/tvshowdata"
)

func TestFormatEpisodeForCalendar(t *testing.T) {
	// TODO need to decouple gcalwrapper from tvshowdata
	cases := []struct {
		in          tvshowdata.Episode
		expectedOut BasicEvent
	}{
		{tvshowdata.Episode{
			Season:         1,
			Episode:        1,
			Title:          "B",
			AirDate:        tvshowdata.Time{time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			RuntimeMinutes: 30,
			ShowName:       "A",
		}, BasicEvent{
			Summary:     "A: \"B\"",
			Description: "A: \"B\"\nSeason 1, Episode 1",
			Start:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:         time.Date(2019, 1, 1, 0, 30, 0, 0, time.UTC),
		}},
	}

	for _, c := range cases {
		out := formatEpisodeForCalendar(c.in)

		if out != c.expectedOut {
			t.Errorf("incorrect output for '%+v': expected '%+v', got ''%+v'",
				c.in, c.expectedOut, out)
		}
	}
}
