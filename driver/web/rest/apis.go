package rest

import (
	"log"
	"net/http"
)

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
}

//Test test TODO
func (h *ApisHandler) Test(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO")
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler() *ApisHandler {
	return &ApisHandler{}
}
