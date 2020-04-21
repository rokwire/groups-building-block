package rest

import (
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"
)

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

//Version gives the service version
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

//Test test TODO
func (h *ApisHandler) Test(current model.User, w http.ResponseWriter, r *http.Request) {
	log.Println("TODO" + current.ID)
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}
