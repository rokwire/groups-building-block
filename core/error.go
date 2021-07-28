package core

import (
	"encoding/json"
	"fmt"
)

// GroupError group error
type GroupError struct {
	Code    int
	Message string
}

// Error returns the error message
func (err *GroupError) Error() string {
	return err.Message
}

// JSONErrorString constructs json representation of the error
func (err *GroupError) JSONErrorString() string {
	errorData := map[string]interface{}{
		"error": map[string]interface{}{
			"code": err.Code,
			"text": err.Message,
		},
	}
	jsonString, _ := json.Marshal(errorData)
	return string(jsonString)
}

// NewForbiddenError new forbidden error
func NewForbiddenError() *GroupError {
	return &GroupError{Code: 1, Message: "forbidden operation"}
}

// NewBadJSONError new bad json error
func NewBadJSONError() *GroupError {
	return &GroupError{Code: 2, Message: "bad json"}
}

// NewValidationError new validation error
func NewValidationError(err error) *GroupError {
	return &GroupError{Code: 3, Message: fmt.Sprintf("validation error: %s", err)}
}

// NewServerError new generic abstract error
func NewServerError() *GroupError {
	return &GroupError{Code: 4, Message: "server error"}
}

// NewGropDuplicationError duplicate group name error
func NewGropDuplicationError() *GroupError {
	return &GroupError{Code: 5, Message: "group name already in use"}
}

// NewMissingParamError missing param error
func NewMissingParamError(message string) *GroupError {
	return &GroupError{Code: 6, Message: message}
}

// NewNotFoundError not found error
func NewNotFoundError() *GroupError {
	return &GroupError{Code: 7, Message: "group not found"}
}
