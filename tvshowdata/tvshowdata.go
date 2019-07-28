// Get relevant data about a TV show using the "episodate" API

package tvshowdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// countdown represents an upcoming episode of a TV show
type countdown struct {
	Season  float64
	Episode float64
	Name    string
	AirDate time.Time
}

// GetArrowData gets the air times of upcoming "Arrow" episodes
func GetArrowData() {
	resp, err := http.Get("https://episodate.com/api/show-details?q=29560")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		bodyString := string(bodyBytes)
		countdownJSON := gjson.Get(bodyString, "tvShow.countdown")
		if !countdownJSON.Exists() {
			fmt.Println("nil countdown")
		} else {
			countdownStruct := countdown{}
			err := json.Unmarshal([]byte(countdownJSON.String()), &countdownStruct)
			if err != nil {
				fmt.Println("bad countdown conversion")
			}
			fmt.Println(countdownStruct.Season)
			fmt.Println(countdownStruct.Episode)
			fmt.Println(countdownStruct.Name)
			fmt.Println(countdownStruct.AirDate)
		}
	}
	const shortForm = "2006-01-02 15:04:05"
	// !!! SW above tells time how to interpret, now need to convert to user local timezone
	t, _ := time.Parse(shortForm, "2019-10-16 01:00:00")
	fmt.Println(t)
}
