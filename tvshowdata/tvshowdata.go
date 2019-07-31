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

// Countdown represents an upcoming episode of a TV show
type Countdown struct {
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

// UpcomingEpisodes is the list of future episodes for the show
type UpcomingEpisodes struct {
	Episodes []Countdown
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
func unmarshallCountdownJSON(countdownJSON gjson.Result) (Countdown, error) {
	countdownStruct := Countdown{}
	errMsg := fmt.Sprintf("Could not get episodate countdown: %s", countdownJSON.String())
	err := json.Unmarshal([]byte(countdownJSON.String()), &countdownStruct)
	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return Countdown{}, err
	}

	countdownStruct.AirDate, err = reformatShowDate(countdownJSON)
	if err != nil {
		err = gerrors.Wrapf(err, errMsg)
		return Countdown{}, err
	}

	return countdownStruct, nil
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

// Parse API response into a countdown struct and return it (default if error)
func getUpcomingShowData(queryID int) (Countdown, error) {
	url := fmt.Sprintf("https://episodate.com/api/show-details?q=%d", queryID)
	respStr, err := httpGet(url)
	if err != nil {
		msg := "error calling httpGet wrapper"
		err = gerrors.Wrapf(err, msg)
		return Countdown{}, err
	}

	countdownJSON := gjson.Get(respStr, "tvShow.countdown")
	if !countdownJSON.Exists() {
		fmt.Println("!!! SW err:", "nil countdown")
		msg := fmt.Sprintf("no countdown data for queryID: %d", queryID)
		err := gerrors.Wrapf(gerrors.New("missint tvShow.countdown"), msg)
		return Countdown{}, err
	}
	return unmarshallCountdownJSON(countdownJSON)
}

/*
TODO check if the show has a null value for "countdown". If so, there's
not a known next episode, and it cannot be added to the calendar. If there
is one, we need to look through the episode data to find the next episode,
and save everything from then on to add to the calendar
*/

// GetThe100Data gets the air times of upcoming "The 100" episodes
func GetThe100Data() {
	const arrowID = 33514

	countdownStruct, err := getUpcomingShowData(arrowID)
	if err != nil {
		fmt.Println("Error getting the 100 data:", err)
		return
	}

	fmt.Println(countdownStruct.Season)
	fmt.Println(countdownStruct.Episode)
	fmt.Println(countdownStruct.Name)
	fmt.Println(countdownStruct.AirDate)
}
