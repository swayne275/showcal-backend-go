package main

import (
	webserver "showcal-backend-go/clientapi"
)

func main() {
	//const queryID = 33514 // The 100
	//const queryID = 2550 // American Dad
	//const queryID = 3564 // Friends

	webserver.StartClientAPI()
}
