// Get relevant data about a TV show using the "episodate" API

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

// BasicShowData tells the show name, and if it's still running
type BasicShowData struct {
	Name         string
	StillRunning bool
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
func checkForFutureEpisodes(showData string, ID int) (bool, error) {
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

// Parse API response into a countdown struct and return it (default if error)
func getUpcomingShowData(queryID int) (UpcomingEpisodes, error) {
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

	// we know the countdown object exists at this point
	trialEp := UpcomingEpisodes{}
	episode, err := unmarshallEpisode(gjson.Get(resp, "tvShow.countdown"))
	trialEp.Episodes = append(trialEp.Episodes, episode)

	return trialEp, err
}

/*
TODO check if the show has a null value for "countdown". If so, there's
not a known next episode, and it cannot be added to the calendar. If there
is one, we need to look through the episode data to find the next episode,
and save everything from then on to add to the calendar
*/

// GetThe100Data gets the air times of upcoming "The 100" episodes
func GetThe100Data() {
	//const showD = 33514 // The 100
	const showID = 3564 // Friends

	episodeList, err := getUpcomingShowData(showID)
	if err != nil {
		fmt.Println("Error getting the show data:", err)
		return
	}

	if len(episodeList.Episodes) > 0 {
		for _, episode := range episodeList.Episodes {
			fmt.Printf("%+v\n", episode)
		}
	} else {
		fmt.Println("No future episodes for show ID", showID)
	}
}
