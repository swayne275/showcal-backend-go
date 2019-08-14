// Convenience wrapper for Google Calendar API V3

/*
TODO
This will start off as part of the showCal project, but should probably
be spun into its own package down the road. It will also initially
include the OAuth2 and GCal bits in this one file, but those will
likely be split into two files (in the same package) when I'm closer
to completing this component
*/

// TODO look up best practices for handling token/client oauth2 stuff

package gcalwrapper

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/swayne275/gerrors"
	"github.com/swayne275/showcal-backend-go/tvshowdata"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// BasicEvent is a simple calendar event with name, description, start, end
type BasicEvent struct {
	Summary     string
	Description string
	Start       time.Time
	End         time.Time
}

// TODO pass this down from main
// change to something else if run as main
const serverPort = "8080"

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL: "http://localhost:" + serverPort + "/GoogleCallback",
		// from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		// TODO stored in environment variables for now, fix this
		ClientID:     os.Getenv("googlekey"),
		ClientSecret: os.Getenv("googlesecret"),
		Scopes:       []string{"https://www.googleapis.com/auth/calendar"},
		Endpoint:     google.Endpoint,
	}
	// Get a random string for each user login
	oauthStateString = "random"
	testToken        oauth2.Token
	hasRun           bool
)

const htmlIndex = `<html><body>
<a href="/GoogleLogin">Log in with Google</a>
</body></html>
`

// AddEpisodesToCalendar concurrently adds one more more events to the calendar
func AddEpisodesToCalendar(episodes tvshowdata.Episodes) {
	if !hasValidToken() {
		return
	}

	service, err := getCalendarService(testToken)
	if err != nil {
		fmt.Println(err) // TODO
	}

	for idx := range episodes.Episodes {
		// TODO error handling from createSingleEvent()
		// TODO verify: can't pass episode since it changes each loop
		go func(ep tvshowdata.Episode) {
			err := createSingleEvent(formatEpisodeForCalendar(ep), service)
			if err != nil {
				// TODO better error handling
				fmt.Println("AddEpsidoesToCalendar err:", err, "episode:",
					ep)
			}
		}(episodes.Episodes[idx])
	}
}

// TODO this only checks if it's been init
func hasValidToken() bool {
	if !hasRun {
		msg := fmt.Sprintf("Go to http://localhost:%s/login to auth with google services",
			serverPort)
		fmt.Println(msg)
		return false
	}

	return true
}

// convert a valid OAuth2 token into a calendar service with background context
func getCalendarService(token oauth2.Token) (*calendar.Service, error) {
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&token))

	calendarService, err := calendar.New(client)
	if err != nil {
		fmt.Println(err)
		return calendarService, err
	}

	return calendarService, nil
}

// HandleLogin directs a user to auth their google account with showCal
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlIndex)
}

// HandleGoogleLogin redirects to google services to auth with gcal
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback processes oauth2 data from google services
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n",
			oauthStateString, state)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// TODO remove test code after building real solution
	hasRun = true
	testToken = *token
}

// Creates a single event in the user's primary calendar
func createSingleEvent(event BasicEvent, service *calendar.Service) error {
	gcalEvent, err := buildCalendarEvent(event)
	if err != nil {
		err = gerrors.Wrapf(err, "Error in createSingleEvent()")
		return err
	}

	createdEvent, err := service.Events.Insert("primary", &gcalEvent).Do()
	if err != nil {
		err = gerrors.Wrapf(err, "Error in createSingleEvent()")
		return err
	}

	fmt.Println("Calendar event created:", createdEvent.HtmlLink)
	return nil
}

// Converts standard struct into google calendar event format
func buildCalendarEvent(event BasicEvent) (calendar.Event, error) {
	if event.Summary == "" {
		err := gerrors.Wrapf(gerrors.New("No event summary"),
			"Error in buildCalendarEvent()")
		return calendar.Event{}, err
	}

	gcalEvent := calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Start:       &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339)},
		End:         &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339)},
	}

	return gcalEvent, nil
}

// todo need to check input data is valid somewhere
// todo check that runtime isn't greater than 3 hours
// formatEpisodeForCalendar converts a TV show episode into a calendar event
func formatEpisodeForCalendar(episode tvshowdata.Episode) BasicEvent {
	summary := fmt.Sprintf("%s: \"%s\"", episode.ShowName, episode.Title)
	description := fmt.Sprintf("%s: \"%s\"\nSeason %d, Episode %d",
		episode.ShowName, episode.Title, episode.Season, episode.Episode)
	event := BasicEvent{
		Summary:     summary,
		Description: description,
		Start:       episode.AirDate.Time,
		End:         episode.AirDate.Time.Add(time.Minute * time.Duration(episode.RuntimeMinutes)),
	}

	return event
}
