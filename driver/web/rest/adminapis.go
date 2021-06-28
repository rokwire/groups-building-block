package rest

import (
	"encoding/json"
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"
)

//AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

//GetGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID GetGroups
// @Tags Admin
// @Accept  json
// @Param APP header string true "APP"
// @Param category query string false "Category"
// @Param title query string false "Filtering by group's title"
// @Success 200 {array} getGroupsResponse
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/groups [get]
func (h *AdminApisHandler) GetGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	var category *string
	catogies, ok := r.URL.Query()["category"]
	if ok && len(catogies[0]) > 0 {
		category = &catogies[0]
	}

	var title *string
	titles, ok := r.URL.Query()["title"]
	if ok && len(titles[0]) > 0 {
		title = &titles[0]
	}

	groups, err := h.app.Services.GetGroups(clientID, current, category, title)
	if err != nil {
		log.Printf("error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("Error on marshal the groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
