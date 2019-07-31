// Get relevant data about a TV show using the "episodate" API

// TODO need to get user's timezone down to here for comparison

package tvshowdata

import (
	"encoding/json"
	"fmt"
	"gerrors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// Episode represents an upcoming episode of a TV show
type Episode struct {
	Season  float64
	Episode float64
	Name    string
	AirDate time.Time
}

// UpcomingEpisodes is the list of Episodes for the show
type UpcomingEpisodes struct {
	Episodes []Episode
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
			// TODO get this error propagated up
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

// Get a list of upcoming shows for a particular Episodate query ID
func getUpcomingShows(queryID int64) (UpcomingEpisodes, error) {
	url := fmt.Sprintf("https://episodate.com/api/show-details?q=%d", queryID)
	resp, err := httpGet(url)
	if err != nil {
		msg := "error calling httpGet wrapper"
		err = gerrors.Wrapf(err, msg)
		return UpcomingEpisodes{}, err
	}

	hasFutureEpisodes, err := checkForFutureEpisodes(resp, queryID)
	if err != nil {
		msg := "Error checking if future episodes exist"
		err = gerrors.Wrapf(err, msg)
		return UpcomingEpisodes{}, err
	}
	if !hasFutureEpisodes {
		// TODO this isn't an error, unsure how I want to propagate besides empty info?
		return UpcomingEpisodes{}, nil
	}

	return parseUpcomingEpisodes(resp)
}

/*
TODO check if the show has a null value for "countdown". If so, there's
not a known next episode, and it cannot be added to the calendar. If there
is one, we need to look through the episode data to find the next episode,
and save everything from then on to add to the calendar
*/

// GetShowData gets the air times of upcoming episodes for the given queryID
func GetShowData(queryID int64) {
	episodeList, err := getUpcomingShows(queryID)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return
	}

	if len(episodeList.Episodes) > 0 {
		for _, episode := range episodeList.Episodes {
			fmt.Printf("%+v\n", episode)
		}
	} else {
		fmt.Println("No future episodes for show ID", queryID)
	}
}
