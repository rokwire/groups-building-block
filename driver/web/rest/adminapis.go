package rest

import (
	"encoding/json"
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"
	"strconv"
)

//AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

//GetUserGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID GetUserGroups
// @Tags Admin
// @Accept  json
// @Param APP header string true "APP"
// @Param category query string false "Category"
// @Param title query string false "Filtering by group's title - case insensitive"
// @Success 200 {array} getGroupsResponse
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/admin/groups [get]
func (h *AdminApisHandler) GetUserGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	var category *string
	catogies, ok := r.URL.Query()["category"]
	if ok && len(catogies[0]) > 0 {
		category = &catogies[0]
	}

	var privacy *string
	privacyParam, ok := r.URL.Query()["privacy"]
	if ok && len(privacyParam[0]) > 0 {
		privacy = &privacyParam[0]
	}

	var title *string
	titles, ok := r.URL.Query()["title"]
	if ok && len(titles[0]) > 0 {
		title = &titles[0]
	}

	var offset *int64
	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			offset = &val
		}
	}

	var limit *int64
	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			limit = &val
		}
	}

	var order *string
	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		order = &orders[0]
	}

	groups, err := h.app.Services.GetGroups(clientID, current, category, privacy, title, offset, limit, order)
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

//GetAllGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID GetAllGroups
// @Tags Admin
// @Accept  json
// @Param APP header string true "APP"
// @Param category query string false "Category"
// @Param title query string false "Filtering by group's title - case insensitive"
// @Success 200 {array} getGroupsResponse
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/admin/groups [get]
func (h *AdminApisHandler) GetAllGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	var category *string
	catogies, ok := r.URL.Query()["category"]
	if ok && len(catogies[0]) > 0 {
		category = &catogies[0]
	}

	var privacy *string
	privacyParam, ok := r.URL.Query()["privacy"]
	if ok && len(privacyParam[0]) > 0 {
		privacy = &privacyParam[0]
	}

	var title *string
	titles, ok := r.URL.Query()["title"]
	if ok && len(titles[0]) > 0 {
		title = &titles[0]
	}

	var offset *int64
	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			offset = &val
		}
	}

	var limit *int64
	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			limit = &val
		}
	}

	var order *string
	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		order = &orders[0]
	}

	groups, err := h.app.Administration.GetGroups(clientID, category, privacy, title, offset, limit, order)
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

// LoginUser Logs in the user and refactor the user record and linked data if need
// @Description Logs in the user and refactor the user record and linked data if need
// @ID LoginAdminUser
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/user/login [get]
func (h *AdminApisHandler) LoginUser(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.LoginUser(clientID, current)
	if err != nil {
		log.Printf("error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
