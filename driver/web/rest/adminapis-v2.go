package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"groups/core/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GetGroupsV2 gets groups. It can be filtered by category, title and privacy. V2
// @Description Gives the groups list. It can be filtered by category, title and privacy. V2
// @ID AdminGetGroupsV2
// @Tags Admin-V2
// @Accept  json
// @Param APP header string true "APP"
// @Param title query string false "Filtering by group's title (case-insensitive)"
// @Param category query string false "category - filter by category"
// @Param privacy query string false "privacy - filter by privacy"
// @Param offset query string false "offset - skip number of records"
// @Param limit query string false "limit - limit the result"
// @Param include_hidden query string false "include_hidden - Includes hidden groups if a search by title is performed. Possible value is true. Default false."
// @Success 200 {array} model.GroupV2
// @Security AppUserAuth
// @Router /api/admin/v2/groups [get]
func (h *AdminApisHandler) GetGroupsV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	var includeHidden *bool
	hiddens, ok := r.URL.Query()["include_hidden"]
	if ok && len(hiddens[0]) > 0 {
		if strings.ToLower(hiddens[0]) == "true" {
			val := true
			includeHidden = &val
		}
	}

	groups, err := h.app.Services.GetGroups(clientID, current, category, privacy, title, offset, limit, order, includeHidden)
	if err != nil {
		log.Printf("apis.GetGroupsV2() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupsV2 := make([]model.GroupV2, len(groups))
	for i, group := range groups {
		groupsV2[i] = group.ToGroupV2(&current.ID)
	}

	data, err := json.Marshal(groupsV2)
	if err != nil {
		log.Println("apis.GetGroupsV2() error on marshal the groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetUserGroupsV2 gets the user groups. It can be filtered by category, title and privacy. V2.
// @Description Gives the user groups. It can be filtered by category, title and privacy. V2.
// @ID AdminGetUserGroupsV2
// @Tags Admin-V2
// @Accept  json
// @Param APP header string true "APP"
// @Param category query string false "Category"
// @Param title query string false "Filtering by group's title (case-insensitive)"
// @Param category query string false "category - filter by category"
// @Param privacy query string false "privacy - filter by privacy"
// @Param offset query string false "offset - skip number of records"
// @Param limit query string false "limit - limit the result"
// @Success 200 {array} model.GroupV2
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/v2/user/groups [get]
func (h *AdminApisHandler) GetUserGroupsV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {

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

	groups, err := h.app.Services.GetUserGroups(clientID, current, category, privacy, title, offset, limit, order)
	if err != nil {
		log.Printf("apis.GetUserGroupsV2() error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupsV2 := make([]model.GroupV2, len(groups))
	for i, group := range groups {
		groupsV2[i] = group.ToGroupV2(&current.ID)
	}

	data, err := json.Marshal(groupsV2)
	if err != nil {
		log.Println("apis.GetUserGroupsV2() error on marshal the user groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupV2 gets a group. V2
// @Description Gives a group. V2
// @ID AdminGetGroup
// @Tags Admin-V2
// @Accept json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {object} model.GroupV2
// @Security AppUserAuth
// @Router /api/admin/v2/groups/{id} [get]
func (h *AdminApisHandler) GetGroupV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("apis.GetGroupV2() id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to see the events for this group
	group, err := h.app.Services.GetGroupEntity(clientID, nil, id)
	if err != nil {
		log.Printf("apis.GetGroupV2() error getting a group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !hasGroupMembershipPermission(h.app.Services, w, current, clientID, group) {
		return
	}

	groupV2 := group.ToGroupV2(&current.ID)

	data, err := json.Marshal(groupV2)
	if err != nil {
		log.Println("apis.GetGroupV2() error on marshal the group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
