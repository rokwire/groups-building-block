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
	"encoding/json"
	"fmt"
	"net/http"
)

// GroupError group error
type groupError struct {
	ErrorCode int
	HttpCode  int
	Message   string
	SubError  Error
}

type Error interface {
	GetErrorCode() int
	GetHttpCode() int
	GetMessage() string
	GetSubError() Error
	JSONErrorString() string
}

func (err *groupError) GetErrorCode() int {
	return err.ErrorCode
}

func (err *groupError) GetHttpCode() int {
	return err.HttpCode
}

func (err *groupError) GetMessage() string {
	return err.Message
}

func (err *groupError) GetSubError() Error {
	return err.SubError
}

// Error returns the error message
func (err *groupError) Error() string {
	return err.Message
}

// JSONErrorString constructs json representation of the error
func (err *groupError) JSONErrorString() string {
	errorMap := map[string]interface{}{
		"code":      err.ErrorCode,
		"http_code": err.HttpCode,
		"text":      err.Message,
	}
	if err.SubError != nil {

	}

	errorData := map[string]interface{}{
		"error": errorMap,
	}
	jsonString, _ := json.Marshal(errorData)
	return string(jsonString)
}

// NewForbiddenError new forbidden error
func NewForbiddenError() Error {
	err := &groupError{ErrorCode: 1, HttpCode: http.StatusForbidden, Message: "forbidden operation"}
	return err
}

// NewBadJSONError new bad json error
func NewBadJSONError() Error {
	return &groupError{ErrorCode: 2, HttpCode: http.StatusInternalServerError, Message: "bad json"}
}

// NewValidationError new validation error
func NewValidationError(err error) Error {
	return &groupError{ErrorCode: 3, HttpCode: http.StatusInternalServerError, Message: fmt.Sprintf("validation error: %s", err)}
}

// NewDefaultServerError returns default server error
func NewDefaultServerError() Error {
	return NewServerError("server error")
}

// NewServerError new generic abstract error
func NewServerError(message string) Error {
	if message == "" {
		message = "server error"
	}
	return &groupError{ErrorCode: 4, HttpCode: http.StatusInternalServerError, Message: message}
}

// NewGroupDuplicationError duplicate group name error
func NewGroupDuplicationError() Error {
	return &groupError{ErrorCode: 5, HttpCode: http.StatusInternalServerError, Message: "group name already in use"}
}

// NewMissingParamError missing param error
func NewMissingParamError(message string) Error {
	return &groupError{ErrorCode: 6, HttpCode: http.StatusInternalServerError, Message: message}
}

// NewNotFoundError not found error
func NewNotFoundError() Error {
	return &groupError{ErrorCode: 7, HttpCode: http.StatusNotFound, Message: "group not found"}
}
