package main

import (
	webserver "showcal-backend-go/clientapi"
	tvshowdata "showcal-backend-go/tvshowdata"
)

func main() {
	tvshowdata.GetThe100Data()
	webserver.StartClientAPI()
}
