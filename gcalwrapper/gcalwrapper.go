// Convenience wrapper for Google Calendar API V3

/*
TODO
This will start off as part of the showCal project, but should probably
be spun into its own package down the road. It will also initially
include the OAuth2 and GCal bits in this one file, but those will
likely be split into two files (in the same package) when I'm closer
to completing this component
*/

package gcalwrapper

//package main // uncomment to run just this part

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/swayne275/gerrors"
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
)

const htmlIndex = `<html><body>
<a href="/GoogleLogin">Log in with Google</a>
</body></html>
`

func main() {
	// home page, where user initiates the process
	http.HandleFunc("/", HandleLogin)
	// handle redirect to google services
	http.HandleFunc("/GoogleLogin", HandleGoogleLogin)
	// handle the oauth2 info given back from google
	http.HandleFunc("/GoogleCallback", HandleGoogleCallback)
	fmt.Println(http.ListenAndServe(":"+serverPort, nil))
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
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

	calendarService, err := calendar.New(client)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	event := BasicEvent{
		Summary:     "The 100 Test Episode",
		Description: "The 100: Season X, Episode Y",
		Start:       time.Date(2019, 8, 8, 9, 24, 0, 0, time.UTC),
		End:         time.Date(2019, 8, 8, 10, 24, 0, 0, time.UTC),
	}

	err = createSingleEvent(event, calendarService)
	if err != nil {
		fmt.Fprintln(w, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

// Creates a single event in the user's primary calendar
func createSingleEvent(event BasicEvent, service *calendar.Service) error {
	gcalEvent, err := buildCalendarEvent(event)
	if err != nil {
		err = gerrors.Wrapf(err, "Error in createSingleEvent()")
		return err
	}

	createdEvent, err := service.Events.Insert("primary", &gcalEvent).Do()
	if err != nil {
		gerrors.Wrapf(err, "Error in createSingleEvent()")
		return err
	}

	fmt.Println("Calendar event created:", createdEvent.HtmlLink)
	return nil
}
