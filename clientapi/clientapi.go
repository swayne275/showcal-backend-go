// Provides the client API

package clientapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/swayne275/showcal-backend-go/gcalwrapper"
	"github.com/swayne275/showcal-backend-go/tvshowdata"
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
	getEpisodesEndpoint = prefix + "getepisodes"
	showSearchEndpoint  = prefix + "showsearch"
	createEventEndpoint = prefix + "createevent"
)

func sayHello(w http.ResponseWriter, r *http.Request) {
	setupCors(w)
	if (*r).Method == "OPTIONS" {
		return
	}

	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	_, err := w.Write([]byte(message))
	if err != nil {
		// TODO handle errors better
		fmt.Println("sayHello()", err)
	}
}

// Get the body from an http request to this API
func getRequestBody(r http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Error in getRequestBody() for URI %s", r.RequestURI)
		err = errors.Wrapf(err, msg)
		return body, err
	}

	return body, nil
}

// Return the requested key, or error if not present/invalid
func getQueryParam(key string, r *http.Request) (string, error) {
	keys, ok := r.URL.Query()[key]
	if !ok || len(keys[0]) < 1 {
		err := errors.New(fmt.Sprintf("URL param '%s' is missing", key))
		return "", err
	}

	return keys[0], nil
}

func handleGetEpisodes(w http.ResponseWriter, r *http.Request) {
	setupCors(w)
	if (*r).Method == "OPTIONS" {
		return
	}

	idStr, err := getQueryParam("id", r)
	if err != nil {
		fmt.Println("error in handleGetEpisodes()", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Println("error in handleGetEpisodes()", err)
		http.Error(w, "Invalid value for param 'id'", http.StatusBadRequest)
		return
	}

	haveEpisodes, episodes := tvshowdata.GetShowData(id)
	if haveEpisodes {
		output, err := json.Marshal(episodes)
		if err != nil {
			msg := fmt.Sprintf("Unable to process upcoming shows in %s", getEpisodesEndpoint)
			err = errors.Wrapf(err, msg)
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")
		_, err = w.Write(output)
		if err != nil {
			// TODO handle errors better
			fmt.Println("handleGetEpisodes()", err)
		}
	} else {
		http.Error(w, "No upcoming episodes", http.StatusNotFound)
	}
}

func handleShowSearch(w http.ResponseWriter, r *http.Request) {
	setupCors(w)
	if (*r).Method == "OPTIONS" {
		return
	}

	query, err := getQueryParam("query", r)
	if err != nil {
		fmt.Println("Error in handleShowSearch():", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO get candidate episodes, write back
	haveCandidates, candidateShows := tvshowdata.GetCandidateShows(query)
	if haveCandidates {
		output, err := json.Marshal(candidateShows)
		if err != nil {
			msg := fmt.Sprintf("Unable to process candidate shows in %s", showSearchEndpoint)
			err = errors.Wrapf(err, msg)
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")
		_, err = w.Write(output)
		if err != nil {
			// TODO handle errors better
			fmt.Println("searchUpcomingEpisodes()", err)
		}
	} else {
		http.Error(w, "No shows matching that query", http.StatusNotFound)
	}
}

func handleCalendarAdd(w http.ResponseWriter, r *http.Request) {
	body, err := getRequestBody(*r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid episode", http.StatusBadRequest)
		return
	}

	var episodes tvshowdata.Episodes
	err = json.Unmarshal(body, &episodes)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid 'episodes' data", http.StatusBadRequest)
		return
	}
	if len(episodes.Episodes) == 0 {
		fmt.Println("No episodes")
		http.Error(w, "No episodes provided", http.StatusBadRequest)
		return
	}

	gcalwrapper.AddEpisodesToCalendar(episodes)
	msg := "Created calendar event"
	_, err = w.Write([]byte(msg))
	if err != nil {
		// TODO handle errors better
		fmt.Println("handleCalendarAdd():", err)
	}
}

// StartClientAPI starts the web server hosting the client API
func StartClientAPI(port string) error {
	http.HandleFunc("/", sayHello)
	http.HandleFunc("/login", gcalwrapper.HandleLogin)
	http.HandleFunc("/GoogleLogin", gcalwrapper.HandleGoogleLogin)
	http.HandleFunc("/GoogleCallback", gcalwrapper.HandleGoogleCallback)
	http.HandleFunc(getEpisodesEndpoint, handleGetEpisodes)
	http.HandleFunc(showSearchEndpoint, handleShowSearch)
	http.HandleFunc(createEventEndpoint, handleCalendarAdd)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		msg := fmt.Sprintf("Could not start client API server on port %s", port)
		err = errors.Wrapf(err, msg)
		return err
	}

	return nil
}

// setup Cross-Origin Resource Sharing in handler for browser clients
func setupCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding")
}
