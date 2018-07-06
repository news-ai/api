package utils

import (
	"os"
)

var APIURL = ""
var BASEURL = ""

func InitURL() {
	if os.Getenv("RUN_WITH_DEVAPPSERVER") == "1" {
		BASEURL = "http://localhost:8080"
		APIURL = BASEURL + "/api"
		return
	}

	if os.Getenv("BASE_URL") != "" {
		BASEURL = os.Getenv("BASE_URL")
	} else {
		BASEURL = "https://tabulae.newsai.org"
	}
	APIURL = BASEURL + "/api"
}
