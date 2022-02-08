package model

// AuthmanSubject contains user name and user email
type AuthmanSubject struct {
	SourceID        string   `json:"sourceId"`
	Success         string   `json:"success"`
	AttributeValues []string `json:"attributeValues"`
	Name            string   `json:"name"`
	ResultCode      string   `json:"resultCode"`
	ID              string   `json:"id"`
} // @name AuthmanSubject
