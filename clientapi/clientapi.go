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

/*
TODO
- standardize error format as JSON
- endpoint for searching by string, returning basic show structs as response
*/

const (
	serverPort = "8080"
	apiVersion = "v1"
	prefix     = apiVersion + "/api/"

	// endpoints
	epIDEndpoint     = prefix + "upcomingepisodes"
	epSearchEndpoint = prefix + "episodesearch"
)

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

	haveEpisodes, episodes := tvshowdata.GetShowData(id.ID)
	if haveEpisodes {
		output, err := json.Marshal(episodes)
		if err != nil {
			msg := "Unable to process upcoming shows"
			err = gerrors.Wrapf(err, msg)
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)
	} else {
		http.Error(w, "No upcoming episodes", http.StatusNotFound)
	}
}

// StartClientAPI starts the web server hosting the client API
func StartClientAPI() error {
	http.HandleFunc("/", sayHello)
	http.HandleFunc(epIDEndpoint, getUpcomingEpisodes)
	http.HandleFunc(epSearchEndpoint, searchUpcomingEpisodes)

	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		msg := fmt.Sprintf("Could not start client API server on port %s", serverPort)
		err = gerrors.Wrapf(err, msg)
		return err
	}

	return nil
}
