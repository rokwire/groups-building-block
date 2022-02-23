package model

// CoreService wrapper record for the corresponding service record response
type CoreService struct {
	Host      string `json:"host"`
	ServiceID string `json:"service_id"`
}
