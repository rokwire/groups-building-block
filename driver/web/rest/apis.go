package rest

import (
	"encoding/json"
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
// @Description Gives the service version.
// @ID Version
// @Produce plain
// @Success 200 {string} v1.1.0
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

//GetGroupCategories gives all group categories
// @Description Gives all group categories.
// @ID GetGroupCategories
// @Accept  json
// @Success 200 {array} string
// @Security APIKeyAuth
// @Router /api/group-categories [get]
func (h *ApisHandler) GetGroupCategories(w http.ResponseWriter, r *http.Request) {
	groupCategories, err := h.app.Services.GetGroupCategories()
	if err != nil {
		log.Println("Error on getting the group categories")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(groupCategories) == 0 {
		groupCategories = make([]string, 0)
	}

	data, err := json.Marshal(groupCategories)
	if err != nil {
		log.Println("Error on marshal the group categories")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//CreateGroup creates a group
func (h *ApisHandler) CreateGroup(current *model.User, w http.ResponseWriter, r *http.Request) {
	if !current.IsMemberOfGroup("urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access") {
		log.Printf("%s is not allowed to create a group - %s", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}

	//TODO
}

//JustMixed test TODO
func (h *ApisHandler) JustMixed(current *model.User, w http.ResponseWriter, r *http.Request) {
	log.Println("JustMixed")
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}
