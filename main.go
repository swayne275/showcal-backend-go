package main

import (
	webserver "showcal-backend-go/clientapi"
	tvshowdata "showcal-backend-go/tvshowdata"
)

func main() {
	//const queryID = 33514 // The 100
	//const queryID = 2550 // American Dad
	const queryID = 3564 // Friends

	tvshowdata.GetShowData(queryID)
	webserver.StartClientAPI()
}
