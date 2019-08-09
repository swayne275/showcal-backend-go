// Provides the client API

package clientapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"showcal-backend-go/gcalwrapper"
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
	apiVersion = "v1"
	// prefix for all main API endpoints
	prefix = "/api/" + apiVersion + "/"

	// endpoints
	epIDEndpoint        = prefix + "upcomingepisodes"
	epSearchEndpoint    = prefix + "episodesearch"
	createEventEndpoint = prefix + "createevent"
)

type showID struct {
	ID int64 `json:"id"`
}

type showQuery struct {
	Query string `json:"query"`
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
}

// Get the body from an http request to this API
func getRequestBody(r http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Error in getRequestBody() for URI %s", r.RequestURI)
		err = gerrors.Wrapf(err, msg)
		return body, err
	}

	return body, nil
}

func getUpcomingEpisodes(w http.ResponseWriter, r *http.Request) {
	body, err := getRequestBody(*r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid show ID", http.StatusBadRequest)
		return
	}

	var id showID
	err = json.Unmarshal(body, &id)
	if err != nil {
		msg := fmt.Sprintf("Unable to parse show id for %s", epIDEndpoint)
		err = gerrors.Wrapf(err, msg)
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	haveEpisodes, episodes := tvshowdata.GetShowData(id.ID)
	if haveEpisodes {
		output, err := json.Marshal(episodes)
		if err != nil {
			msg := fmt.Sprintf("Unable to process upcoming shows in %s", epIDEndpoint)
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

func searchUpcomingEpisodes(w http.ResponseWriter, r *http.Request) {
	body, err := getRequestBody(*r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid show query", http.StatusBadRequest)
		return
	}

	var query showQuery
	err = json.Unmarshal(body, &query)
	if err != nil {
		msg := fmt.Sprintf("Unable to parse show query for %s", epSearchEndpoint)
		err = gerrors.Wrapf(err, msg)
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO get candidate episodes, write back
	haveCandidates, candidateShows := tvshowdata.GetCandidateShows(query.Query)
	if haveCandidates {
		output, err := json.Marshal(candidateShows)
		if err != nil {
			msg := fmt.Sprintf("Unable to process candidate shows in %s", epSearchEndpoint)
			err = gerrors.Wrapf(err, msg)
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)
	} else {
		http.Error(w, "No shows matching that query", http.StatusNotFound)
	}
}

// StartClientAPI starts the web server hosting the client API
func StartClientAPI(port string) error {
	http.HandleFunc("/", sayHello)
	http.HandleFunc("/login", gcalwrapper.HandleLogin)
	http.HandleFunc("/GoogleLogin", gcalwrapper.HandleGoogleLogin)
	http.HandleFunc("/GoogleCallback", gcalwrapper.HandleGoogleCallback)
	http.HandleFunc(epIDEndpoint, getUpcomingEpisodes)
	http.HandleFunc(epSearchEndpoint, searchUpcomingEpisodes)
	http.HandleFunc(createEventEndpoint, calendarAddHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		msg := fmt.Sprintf("Could not start client API server on port %s", port)
		err = gerrors.Wrapf(err, msg)
		return err
	}

	return nil
}

func calendarAddHandler(w http.ResponseWriter, r *http.Request) {

}
