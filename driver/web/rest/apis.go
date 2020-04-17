package rest

import (
	"groups/core/model"
	"log"
	"net/http"
)

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
}

//Test test TODO
func (h *ApisHandler) Test(current *model.User, w http.ResponseWriter, r *http.Request) {
	log.Println("TODO" + current.ID)
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler() *ApisHandler {
	return &ApisHandler{}
}
