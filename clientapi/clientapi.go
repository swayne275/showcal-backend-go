// Provides the client API

package clientapi

import (
	"net/http"
	"strings"
)

var serverPort = "8080"

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
}

// StartClientAPI starts the web server hosting the client API
func StartClientAPI() {
	http.HandleFunc("/", sayHello)
	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		panic(err)
	}
}
