// Get relevant data about a TV show using the "episodate" API

package tvshowdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// countdown represents an upcoming episode of a TV show
type countdown struct {
	Season  float64
	Episode float64
	Name    string
	AirDate time.Time
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

// Properly format time data for go (modify json copy)
func reformatShowDate(json gjson.Result) gjson.Result {
	const timeStrFormat = "2006-01-02 15:04:05"

	airDate := gjson.Get(json.String(), "air_date")
	if airDate.Exists() {
		formattedAirDate, _ := time.Parse(timeStrFormat, airDate.String())
		fmt.Println(formattedAirDate)
		sjson.Set(json.String(), "air_date", formattedAirDate)
		fmt.Println(gjson.Get(json.String(), "air_date"))
	}
	return json
}

// Parse API response into a countdown struct and return it (default if error)
func getUpcomingShowData(queryID int) countdown {
	url := fmt.Sprintf("https://episodate.com/api/show-details?q=%d", queryID)
	respStr := httpGet(url)
	countdownJSON := gjson.Get(respStr, "tvShow.countdown")
	if !countdownJSON.Exists() {
		fmt.Println("!!! SW err:", "nil countdown")
		return countdown{}
	}

	formattedCountdownJSON := reformatShowDate(countdownJSON)
	fmt.Println(formattedCountdownJSON)
	countdownStruct := countdown{}
	err := json.Unmarshal([]byte(formattedCountdownJSON.String()), &countdownStruct)
	if err != nil {
		fmt.Println("!!! SW err:", "bad countdown conversion")
		return countdown{}
	}
	return countdownStruct
}

/*
TODO I need to format into a time.Time as it's going into the struct for this to
work properly. Not sure how to do that. Should look into JSON unmarshall to see
how the json -> struct conversion actually works
*/

// GetArrowData gets the air times of upcoming "Arrow" episodes
func GetArrowData() {
	const timeStrFormat = "2006-01-02 15:04:05"
	const arrowID = 29560

	countdownStruct := getUpcomingShowData(arrowID)

	fmt.Println(countdownStruct.Season)
	fmt.Println(countdownStruct.Episode)
	fmt.Println(countdownStruct.Name)
	fmt.Println(countdownStruct.AirDate)
	// !!! SW above tells time how to interpret, now need to convert to user local timezone
	t, _ := time.Parse(timeStrFormat, "2019-10-16 01:00:00")
	fmt.Println(t)
}
