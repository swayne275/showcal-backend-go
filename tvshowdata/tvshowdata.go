// Get relevant data about a TV show using the "episodate" API

/*
TODO error handling as per:
https://hackernoon.com/golang-handling-errors-gracefully-8e27f1db729f
*/

package tvshowdata

import (
	"encoding/json"
	"fmt"
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
func httpGet(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("!!! SW err:", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("!!! SW err:", err)
			return ""
		}
		return string(bodyBytes)
	}

	fmt.Println("!!! SW got http status code:", resp.StatusCode)
	return ""
}

// Custom unmarshal to deal with non-RFC 3339 time format
func unmarshallCountdownJSON(countdownJSON gjson.Result) Countdown {
	countdownStruct := Countdown{}
	err := json.Unmarshal([]byte(countdownJSON.String()), &countdownStruct)
	if err != nil {
		fmt.Println("!!! SW err:", "bad countdown conversion")
		return Countdown{}
	}

	countdownStruct.AirDate = reformatShowDate(countdownJSON)
	return countdownStruct
}

// Properly format time data for go (modify json copy)
func reformatShowDate(json gjson.Result) time.Time {
	const timeStrFormat = "2006-01-02 15:04:05"

	airDate := gjson.Get(json.String(), "air_date")
	if airDate.Exists() {
		formattedAirDate, _ := time.Parse(timeStrFormat, airDate.String())
		return formattedAirDate
	}
	return time.Now() // TODO better error handling
}

// Parse API response into a countdown struct and return it (default if error)
func getUpcomingShowData(queryID int) Countdown {
	url := fmt.Sprintf("https://episodate.com/api/show-details?q=%d", queryID)
	respStr := httpGet(url)
	countdownJSON := gjson.Get(respStr, "tvShow.countdown")
	if !countdownJSON.Exists() {
		fmt.Println("!!! SW err:", "nil countdown")
		return Countdown{}
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

	countdownStruct := getUpcomingShowData(arrowID)

	fmt.Println(countdownStruct.Season)
	fmt.Println(countdownStruct.Episode)
	fmt.Println(countdownStruct.Name)
	fmt.Println(countdownStruct.AirDate)
}
