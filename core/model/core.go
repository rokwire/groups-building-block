package model

// CoreService wrapper record for the corresponding service record response
type CoreService struct {
	Host      string `json:"host"`
	ServiceID string `json:"service_id"`
}

//Creator represents group member entity
type Creator struct {
	UserID string `json:"user_id" bson:"user_id"`
	Name   string `json:"name" bson:"name"`
	Email  string `json:"email" bson:"email"`
} //@name Creator
