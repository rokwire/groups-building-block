package utils

import (
	"log"
	"net/http"
)

//LogRequest logs the request as hide some header fields because of security reasons
func LogRequest(req *http.Request) {
	if req == nil {
		return
	}

	method := req.Method
	path := req.URL.Path

	header := make(map[string][]string)
	for key, value := range req.Header {
		var logValue []string
		//do not log api key, cookies and Authorization
		if key == "Rokwire-Api-Key" || key == "Cookie" || key == "Authorization" {
			logValue = append(logValue, "---")
		} else {
			logValue = value
		}
		header[key] = logValue
	}
	log.Printf("%s %s %s", method, path, header)
}
