package main

import "github.com/swayne275/showcal-backend-go/clientapi"

const (
	// ServerPort is where the web server is hosted
	ServerPort = "8080"
)

func main() {
	//const queryID = 33514 // The 100
	//const queryID = 2550 // American Dad
	//const queryID = 3564 // Friends

	err := clientapi.StartClientAPI(ServerPort)
	if err != nil {
		panic(err)
	}
}
