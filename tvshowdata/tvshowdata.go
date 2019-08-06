// Get relevant data about a TV show using the "episodate" API

// TODO need to get user's timezone down to here for comparison

// TODO continue by unmarshaliing the Show type and finishing some
// other TODOs

package tvshowdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/swayne275/gerrors"
	"github.com/tidwall/gjson"
)

// Episode represents an upcoming episode of a TV show
type Episode struct {
	Season  float64 `json:"season"`
	Episode float64 `json:"episode"`
	Name    string  `json:"name"`
	AirDate Time    `json:"air_date"`
}

// Time is a custom time to properly unmarshal non-RFC 3339 time from API
type Time struct {
	time.Time
}

// UnmarshalJSON reformats API given time as RFC 3339, when Time struct used
func (t *Time) UnmarshalJSON(data []byte) error {
	const timeStrFormat = "2006-01-02 15:04:05"
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return gerrors.Wrapf(err, "Unable to unmarshal time from API")
	}

	var err error
	t.Time, err = time.Parse(timeStrFormat, s)
	if err != nil {
		return gerrors.Wrapf(err, "unable to reformat time from API")
	}
	return nil
}

// UpcomingEpisodes is the list of Episodes for the show
type UpcomingEpisodes struct {
	Episodes []Episode
}

// Show is the basic show details, and if it is still running
type Show struct {
	Name         string  `json:"name"`
	ID           float64 `json:"id"`
	StillRunning string  `json:"status"`
}

// CandidateShows is the list of candidate Shows for the query
type CandidateShows struct {
	Shows []Show
}

// Simple HTTP Get that returns the response body as a string ("" if error)
func httpGet(url string) (string, error) {
	errMsg := fmt.Sprintf("error fetching data from episodate api for url: %s", url)
	resp, err := http.Get(url)

	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		newErr := fmt.Sprintf("Got HTTP StatusCode: %d", resp.StatusCode)
		err = gerrors.Wrapf(gerrors.New(newErr), errMsg)
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return "", err
	}

	return string(bodyBytes), nil
}

/*
func unmarshallShow(showJSON, gjson.Result) (Show, error) {
	// TODO build custom unmarshal to take "name", "id", "status" == "Running"
}*/

// Custom unmarshal to deal with non-RFC 3339 time format
func unmarshallEpisode(countdownJSON gjson.Result) (Episode, error) {
	episode := Episode{}
	errMsg := fmt.Sprintf("Could not get episodate countdown: %s", countdownJSON.String())
	err := json.Unmarshal([]byte(countdownJSON.String()), &episode)
	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return Episode{}, err
	}

	return episode, nil
}

func checkForCandidateShows(queryData, query string) (bool, error) {
	total := gjson.Get(queryData, "total")
	msg := fmt.Sprintf("error getting total shows for query '%s'", query)
	if !total.Exists() || !(total.Type.String() == "String") {
		err := gerrors.Wrapf(gerrors.New("missing 'total'"), msg)
		return false, err
	}
	if !(total.Type.String() == "String") {
		// For some reason the total field is a string
		err := gerrors.Wrapf(gerrors.New("incorrect type for 'total'"), msg)
		return false, err
	}

	numShows, err := strconv.Atoi(total.String())
	if err != nil {
		err := gerrors.Wrapf(err, "could not convert 'total' to int")
		return false, err
	}
	if numShows < 1 {
		// no matching shows
		return false, nil
	}

	// verify matching show data exists
	tvShows := gjson.Get(queryData, "tv_shows")
	if !tvShows.Exists() {
		err := gerrors.Wrapf(gerrors.New("no 'tv_shows'"), msg)
		return false, err
	}
	if !(tvShows.Type.String() == "JSON") {
		err := gerrors.Wrapf(gerrors.New("invalid 'tv_shows' type"), msg)
		return false, err
	}

	return true, nil
}

// Determine if there are likely future episodes of a show or not
func checkForFutureEpisodes(showData string, ID int64) (bool, error) {
	countdown := gjson.Get(showData, "tvShow.countdown")
	if !countdown.Exists() {
		msg := fmt.Sprintf("api returned invalid countdown data for queryID: %d", ID)
		err := gerrors.Wrapf(gerrors.New("missing tvShow.countdown"), msg)
		return false, err
	}
	if countdown.Type.String() == "Null" {
		// no known future episodes
		return false, nil
	}

	return true, nil
}

func parseCandidateShows(queryData string) (CandidateShows, error) {
	allCandidates := gjson.Get(queryData, "tv_shows")
	// TODO previous check if shows exists validates type and presense

	// declare error here to preserve any error from the ForEach loop
	var err error
	candidateShows := CandidateShows{}

	allCandidates.ForEach(func(key, value gjson.Result) bool {
		show := Show{}
		//show, err = unmarshallShow(value)
		err = json.Unmarshal([]byte(value.String()), &show)
		fmt.Println("!!! SW result")
		fmt.Println(show)
		if err != nil {
			msg := "Could not unmarshal show from API"
			err = gerrors.Wrapf(err, msg)
			// stop iterating
			return false
		}

		candidateShows.Shows = append(candidateShows.Shows, show)

		// keep iterating
		return true
	})

	return candidateShows, err
}

func parseUpcomingEpisodes(showData string) (UpcomingEpisodes, error) {
	upcomingEpisodes := UpcomingEpisodes{}
	allEpisodes := gjson.Get(showData, "tvShow.episodes")
	if !allEpisodes.Exists() || !allEpisodes.IsArray() {
		msg := fmt.Sprintf("invalid data given to parseUpcomingEpisodes: %s", showData)
		err := gerrors.Wrapf(gerrors.New("no episode list in api response"), msg)
		return UpcomingEpisodes{}, err
	}

	// declare error here to preserve any error from the ForEach loop
	var err error
	now := time.Now()

	allEpisodes.ForEach(func(key, value gjson.Result) bool {
		episode := Episode{}
		episode, err = unmarshallEpisode(value)
		if err != nil {
			msg := "Could not unmarshal episode from API"
			err = gerrors.Wrapf(err, msg)
			// stop iterating
			return false
		}

		if episode.AirDate.After(now) {
			// TODO add to list
			upcomingEpisodes.Episodes = append(upcomingEpisodes.Episodes, episode)
		}

		return true // keep iterating
	})

	return upcomingEpisodes, err
}

// TODO use runtime package to get function names for errors
// Get a list of potential shows matching the query
func getUpcomingShows(query string) (CandidateShows, error) {
	url := fmt.Sprintf("https://www.episodate.com/api/search?q=%s", url.QueryEscape(query))
	resp, err := httpGet(url)
	if err != nil {
		msg := "error calling httpGet wrapper in getUpcomingShows"
		err = gerrors.Wrapf(err, msg)
		return CandidateShows{}, err
	}

	haveCandidates, err := checkForCandidateShows(resp, query)
	if err != nil {
		msg := "error checking if candidates exist"
		err = gerrors.Wrapf(err, msg)
		return CandidateShows{}, err
	}
	if !haveCandidates {
		return CandidateShows{}, nil
	}

	return parseCandidateShows(resp)
}

// Get a list of upcoming shows for a particular Episodate query ID
func getUpcomingEpisodes(queryID int64) (UpcomingEpisodes, error) {
	url := fmt.Sprintf("https://episodate.com/api/show-details?q=%d", queryID)
	resp, err := httpGet(url)
	if err != nil {
		msg := "error calling httpGet wrapper"
		err = gerrors.Wrapf(err, msg)
		return UpcomingEpisodes{}, err
	}

	haveFutureEpisodes, err := checkForFutureEpisodes(resp, queryID)
	if err != nil {
		msg := "Error checking if future episodes exist"
		err = gerrors.Wrapf(err, msg)
		return UpcomingEpisodes{}, err
	}
	if !haveFutureEpisodes {
		// TODO this isn't an error, unsure how I want to propagate besides empty info?
		return UpcomingEpisodes{}, nil
	}

	return parseUpcomingEpisodes(resp)
}

// GetCandidateShows gets a list of TV shows for the given queryShow
func GetCandidateShows(queryShow string) (bool, CandidateShows) {
	showList, err := getUpcomingShows(queryShow)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return false, CandidateShows{}
	}

	return (len(showList.Shows) > 0), showList
}

// GetShowData gets the air times of upcoming episodes for the given queryID
func GetShowData(queryID int64) (bool, UpcomingEpisodes) {
	episodeList, err := getUpcomingEpisodes(queryID)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return false, UpcomingEpisodes{}
	}

	return (len(episodeList.Episodes) > 0), episodeList
}
