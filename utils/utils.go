// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	val, ok := req.Header["User-Agent"]
	if ok && len(val) != 0 && val[0] == "ELB-HealthChecker/2.0" {
		return
	}

	header := make(map[string][]string)
	for key, value := range req.Header {
		var logValue []string
		//do not log api key, cookies and Authorization
		if key == "Rokwire-Api-Key" || key == "Cookie" || key == "Authorization" || key == "Rokwire_gs_api_key" {
			logValue = append(logValue, "---")
		} else {
			logValue = value
		}
		header[key] = logValue
	}
	log.Printf("%s %s %s", method, path, header)
}

//Equal compares two slices
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

//EqualPointers compares two pointers slices
func EqualPointers(a, b *[]string) bool {
	if a == nil && b == nil {
		return true //equals
	}
	if a != nil && b == nil {
		return false // not equals
	}
	if a == nil && b != nil {
		return false // not equals
	}

	//both are not nil
	return Equal(*a, *b)
}
