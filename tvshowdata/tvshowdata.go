package tvshowdata

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetArrowData() {
	resp, err := http.Get("https://episodate.com/api/show-details?q=29560")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println(result)
}
