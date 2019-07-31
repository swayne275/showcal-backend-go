// Provides the client API

package clientapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"showcal-backend-go/tvshowdata"
	"strings"

	"github.com/swayne275/gerrors"
)

const serverPort = "8080"

type showID struct {
	ID int64 `json:"id"`
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
}

func getUpcomingEpisodes(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		msg := "Unable to get query ID from getUpcomingEpisodes body"
		err = gerrors.Wrapf(err, msg)
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var id showID
	err = json.Unmarshal(body, &id)
	if err != nil {
		msg := "Unable to parse show ID from request body"
		err = gerrors.Wrapf(err, msg)
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tvshowdata.GetShowData(id.ID)
}

// StartClientAPI starts the web server hosting the client API
func StartClientAPI() error {
	http.HandleFunc("/", sayHello)
	http.HandleFunc("/upcomingepisodes", getUpcomingEpisodes)
	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		msg := fmt.Sprintf("Could not start client API server on port %s", serverPort)
		err = gerrors.Wrapf(err, msg)
		return err
	}

	return nil
}
