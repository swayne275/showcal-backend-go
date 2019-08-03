// Get relevant data about a TV show using the "episodate" API

// TODO need to get user's timezone down to here for comparison

package tvshowdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/swayne275/gerrors"
	"github.com/tidwall/gjson"
)

// Episode represents an upcoming episode of a TV show
// Not using struct tags for casing consistency upon json.Marshal
// Would be nice if this was supported: AirDate time.Time `time:"2006-01-02 15:04:05"`
type Episode struct {
	Season  float64   //`json:"season"`
	Episode float64   //`json:"episode"`
	Name    string    //`json:"name"`
	AirDate time.Time // can't put struct tag due to non RFC 3339 format
}

// UpcomingEpisodes is the list of Episodes for the show
type UpcomingEpisodes struct {
	Episodes []Episode
}

// Show is the basic show details, and if it is still running
type Show struct {
	Name         string
	ID           float64
	StillRunning bool
}

// CandidateShows is the list of candidate Shows for the query
type CandidateShows struct {
	Shows []CandidateShows
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

func unmarshallShow(showJSON, gjson.Result) (Show, error) {
	// TODO build custom unmarshal to take "name", "id", "status" == "Running"
}

// Custom unmarshal to deal with non-RFC 3339 time format
func unmarshallEpisode(countdownJSON gjson.Result) (Episode, error) {
	episode := Episode{}
	errMsg := fmt.Sprintf("Could not get episodate countdown: %s", countdownJSON.String())
	err := json.Unmarshal([]byte(countdownJSON.String()), &episode)
	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return Episode{}, err
	}

	episode.AirDate, err = reformatShowDate(countdownJSON)
	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return Episode{}, err
	}

	return episode, nil
}

// Properly format time data for go (modify json copy)
func reformatShowDate(json gjson.Result) (time.Time, error) {
	const timeStrFormat = "2006-01-02 15:04:05"

	airDate := gjson.Get(json.String(), "air_date")
	if !airDate.Exists() {
		msg := fmt.Sprintf("invalid data given to reformatShowData: %s", json.String())
		err := gerrors.Wrapf(gerrors.New("no date to convert"), msg)
		return time.Now(), err
	}

	formattedAirDate, _ := time.Parse(timeStrFormat, airDate.String())

	return formattedAirDate, nil
}

func checkForCandidateShows(queryData, query string) (bool, error) {
	total := gjson.Get(queryData, "total")
	if !total.Exists() || !(total.Type.String() == "Number") {
		msg := fmt.Sprintf("api returned invalid number of matching shows for query %s", query)
		err := gerrors.Wrapf(gerrors.New("missing total"), msg)
		return false, err
	}
	if total.Float() < 1 {
		// no matching shows
		return false, nil
	}
	// TODO check if tv_shows exists and is array

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
	candidateShows := CandidateShows{}
	allCandidates := gjson.Get(queryData, "tv_shows")
	// TODO previous check if shows exists validates type and presense

	// declare error here to preserve any error from the ForEach loop
	allCandidates.ForEach(func(key, value gjson.Result) bool {
		show = Show{}
		show, err = unmarshallShow(value)
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
func getUpcomingShows(queryID int64) (UpcomingEpisodes, error) {
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

}

// GetShowData gets the air times of upcoming episodes for the given queryID
func GetShowData(queryID int64) (bool, UpcomingEpisodes) {
	episodeList, err := getUpcomingShows(queryID)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return false, UpcomingEpisodes{}
	}

	return (len(episodeList.Episodes) > 0), episodeList
}
