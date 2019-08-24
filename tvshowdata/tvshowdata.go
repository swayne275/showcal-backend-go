// Get relevant data about a TV show using the "episodate" API

// TODO need to get user's timezone down to here for comparison
// TODO use runtime package to get function names for errors
// TODO figure out how to organize this (utilities, biz logic, etc)
// TODO summary {show}: {title}
// TODO description {show} Season {season}, Episode {episode}
// TODO validate Episodes, maybe also clean old ones from here
// TODO propagate runtime, title from Show into Episode
// TODO if status is 'running' but null countdown, have error or somethin
// TODO get url fcns should error on blank input

package tvshowdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/swayne275/gerrors"
	"github.com/tidwall/gjson"
)

const (
	// episode airdate format provided by the API
	timeStrFormat = "2006-01-02 15:04:05"

	// gjson variable types (<>.Type.String()
	gjsonString = "String"
	gjsonJSON   = "JSON"
	gjsonNull   = "Null"

	// episodate unpopulated endpoints used
	upShowSearch  = "https://www.episodate.com/api/search?q=%s"
	upShowDetails = "https://episodate.com/api/show-details?q=%d"
)

// Episode represents an upcoming episode of a TV show
type Episode struct {
	Season         int64  `json:"season"`
	Episode        int64  `json:"episode"`
	Title          string `json:"name"`
	AirDate        Time   `json:"air_date"`
	RuntimeMinutes int64  `json:"runtime"`
	ShowName       string `json:"show_name"`
}

// Episodes is the list of Episodes for the show
type Episodes struct {
	Episodes []Episode `json:"episodes"`
}

// Show is the basic show details, and if it is still running
type Show struct {
	Name         string  `json:"name"`
	ID           int64   `json:"id"`
	StillRunning Running `json:"status"`
}

// Shows is the list of candidate Shows for the query
type Shows struct {
	Shows []Show `json:"shows"`
}

// Time is a custom time to properly unmarshal non-RFC 3339 time from API
type Time struct {
	time.Time
}

// UnmarshalJSON reformats API given time as RFC 3339, when Time struct used
func (t *Time) UnmarshalJSON(data []byte) error {
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return gerrors.Wrapf(err, "Unable to unmarshal time from API")
	}

	// first try parsing as RFC3339 in case it's in the proper format
	var err error
	t.Time, err = time.Parse(time.RFC3339, s)
	if err == nil {
		// the time is already properly formatted
		return nil
	}

	t.Time, err = time.Parse(timeStrFormat, s)
	if err != nil {
		return gerrors.Wrapf(err, fmt.Sprintf("unable to reformat time: %s", s))
	}

	return nil
}

// Running is used to convert string running status to bool (true if running)
type Running struct {
	bool
}

// MarshalJSON marshals the Running struct into a simple bool
func (r Running) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.bool)
}

// UnmarshalJSON reformats string "show running" status to a bool
func (r *Running) UnmarshalJSON(data []byte) error {
	var running string

	if err := json.Unmarshal(data, &running); err != nil {
		return gerrors.Wrapf(err, "Unable to unmarshal show running status from API")
	}

	r.bool = (strings.ToLower(running) == "running")
	return nil
}

// GetCandidateShows returns a list of TV shows for the given queryShow
func GetCandidateShows(queryShow string) (bool, Shows) {
	showList, err := getUpcomingShows(queryShow)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return false, Shows{}
	}

	return (len(showList.Shows) > 0), showList
}

// GetShowData gets the air times of upcoming episodes for the given queryID
func GetShowData(queryID int64) (bool, Episodes) {
	episodeList, err := getUpcomingEpisodes(queryID)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return false, Episodes{}
	}

	return (len(episodeList.Episodes) > 0), episodeList
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

// Determines if there are any shows matching query from the API
func checkForCandidateShows(queryData, query string) (bool, error) {
	msg := fmt.Sprintf("error getting total shows for query '%s'", query)
	total := gjson.Get(queryData, "total")

	if !total.Exists() {
		err := gerrors.Wrapf(gerrors.New("missing 'total'"), msg)
		return false, err
	}
	if !(total.Type.String() == gjsonString) {
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
	if !(tvShows.Type.String() == gjsonJSON) {
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
	if countdown.Type.String() == gjsonNull {
		// no known future episodes
		return false, nil
	}

	return true, nil
}

// Unmarshals any shows matching the query to appropriate format
func parseCandidateShows(queryData string) (Shows, error) {
	allCandidates := gjson.Get(queryData, "tv_shows")
	if !allCandidates.Exists() {
		return Shows{}, gerrors.New("No 'tv_shows' field in API response")
	}

	// declare error here to preserve any error from the ForEach loop
	var err error
	candidateShows := Shows{}

	allCandidates.ForEach(func(key, value gjson.Result) bool {
		show := Show{}
		err = json.Unmarshal([]byte(value.String()), &show)
		if err != nil {
			msg := "Could not unmarshal show from API"
			err = gerrors.Wrapf(err, msg)
			// stop iterating
			return false
		}
		if show == (Show{}) {
			err = gerrors.New(fmt.Sprintf("Couldn't parse show data for: '%s'", value.String()))
			return false
		}
		candidateShows.Shows = append(candidateShows.Shows, show)

		// keep iterating
		return true
	})

	return candidateShows, err
}

// Unmarshals any upcoming episodes to the appropriate format
func parseUpcomingEpisodes(showData string) (Episodes, error) {
	errMsg := fmt.Sprintf("invalid data given to parseUpcomingEpisodes: %s", showData)

	showName := gjson.Get(showData, "tvShow.name")
	runtimeMin := gjson.Get(showData, "tvShow.runtime")
	allEpisodes := gjson.Get(showData, "tvShow.episodes")
	if !allEpisodes.Exists() || !allEpisodes.IsArray() {
		err := gerrors.Wrapf(gerrors.New("no episode list in api response"), errMsg)
		return Episodes{}, err
	}
	if !showName.Exists() {
		err := gerrors.Wrapf(gerrors.New("No name' in api response"), errMsg)
		return Episodes{}, err
	}
	if !runtimeMin.Exists() || runtimeMin.Int() == 0 {
		err := gerrors.Wrapf(gerrors.New("Missing/invalid 'runtime' in API response"), errMsg)
		return Episodes{}, err
	}

	// declare error here to preserve any error from the ForEach loop
	var err error
	now := time.Now()

	upcomingEpisodes := Episodes{}
	allEpisodes.ForEach(func(key, value gjson.Result) bool {
		episode := Episode{}
		err = json.Unmarshal([]byte(value.String()), &episode)
		if err != nil {
			msg := "Could not unmarshal episode from API"
			err = gerrors.Wrapf(err, msg)
			// stop iterating
			return false
		}
		if episode == (Episode{}) {
			err = gerrors.New(fmt.Sprintf("Couldn't parse episode data for: '%s'", value.String()))
			return false
		}

		episode.RuntimeMinutes = runtimeMin.Int()
		episode.ShowName = showName.String()

		if episode.AirDate.After(now) {
			upcomingEpisodes.Episodes = append(upcomingEpisodes.Episodes, episode)
		}

		return true // keep iterating
	})

	return upcomingEpisodes, err
}

// Get a list of potential shows matching the query
func getUpcomingShows(query string) (Shows, error) {
	url, err := getShowSearchURL(query)
	if err != nil {
		err := gerrors.Wrap(err, "error in getUpcomingShows()")
		return Shows{}, err
	}

	resp, err := httpGet(url)
	if err != nil {
		msg := "error calling httpGet wrapper in getUpcomingShows"
		err = gerrors.Wrapf(err, msg)
		return Shows{}, err
	}

	haveCandidates, err := checkForCandidateShows(resp, query)
	if err != nil {
		msg := "error checking if candidates exist"
		err = gerrors.Wrapf(err, msg)
		return Shows{}, err
	}
	if !haveCandidates {
		err := gerrors.Wrapf(gerrors.New("No matching shows"),
			fmt.Sprintf("No shows matching query %s", query))
		return Shows{}, err
	}

	return parseCandidateShows(resp)
}

// Get a list of upcoming shows for a particular Episodate query ID
func getUpcomingEpisodes(queryID int64) (Episodes, error) {
	url := getShowDetailsURL(queryID)
	resp, err := httpGet(url)
	if err != nil {
		msg := "error calling httpGet wrapper"
		err = gerrors.Wrapf(err, msg)
		return Episodes{}, err
	}

	haveFutureEpisodes, err := checkForFutureEpisodes(resp, queryID)
	if err != nil {
		msg := "Error checking if future episodes exist"
		err = gerrors.Wrapf(err, msg)
		return Episodes{}, err
	}
	if !haveFutureEpisodes {
		err := gerrors.Wrapf(gerrors.New("No upcoming episodes"),
			fmt.Sprintf("No upcoming episodes found for queryID %d", queryID))
		return Episodes{}, err
	}

	return parseUpcomingEpisodes(resp)
}

// getShowSearchURL returns the endpoint to search for shows matching query
func getShowSearchURL(query string) (string, error) {
	if query == "" {
		err := gerrors.Wrapf(gerrors.New("Empty 'query' given"),
			"Invalid data passed to getShowSearchURL()")
		return "", err
	}

	htmlQuery := url.QueryEscape(query)
	return fmt.Sprintf(upShowSearch, htmlQuery), nil
}

// getShowDetailsURL returns the endpoint to get show details for id
func getShowDetailsURL(id int64) string {
	return fmt.Sprintf(upShowDetails, id)
}
