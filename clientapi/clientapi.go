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
	prefix     = "/api/" + apiVersion + "/"

	// endpoints
	epIDEndpoint     = prefix + "upcomingepisodes"
	epSearchEndpoint = prefix + "episodesearch"
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

func getUpcomingEpisodes(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Unable to get json body from %s", epIDEndpoint)
		err = gerrors.Wrapf(err, msg)
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Unable to get json body from %s", epSearchEndpoint)
		err = gerrors.Wrapf(err, msg)
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
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
