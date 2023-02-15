package rest

import (
	"encoding/json"
	"groups/core/model"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// GetGroupsV2 gets groups. It can be filtered by category, title and privacy. V2
// @Description Gives the groups list. It can be filtered by category, title and privacy. V2
// @ID AdminGetGroupsV2
// @Tags Admin-V2
// @Accept  json
// @Param APP header string true "APP"
// @Param title query string false "Deprecated - instead use request body filter! Filtering by group's title (case-insensitive)"
// @Param category query string false "Deprecated - instead use request body filter! category - filter by category"
// @Param privacy query string false "Deprecated - instead use request body filter! privacy - filter by privacy"
// @Param offset query string false "Deprecated - instead use request body filter! offset - skip number of records"
// @Param limit query string false "Deprecated - instead use request body filter! limit - limit the result"
// @Param include_hidden query string false "Deprecated - instead use request body filter! include_hidden - Includes hidden groups if a search by title is performed. Possible value is true. Default false."
// @Param data body model.GroupsFilter true "body data"
// @Success 200 {array} model.Group
// @Security AppUserAuth
// @Router /api/admin/v2/groups [get]
func (h *AdminApisHandler) GetGroupsV2(current *model.User, w http.ResponseWriter, r *http.Request) {
	var groupsFilter model.GroupsFilter

	catogies, ok := r.URL.Query()["category"]
	if ok && len(catogies[0]) > 0 {
		groupsFilter.Category = &catogies[0]
	}

	privacyParam, ok := r.URL.Query()["privacy"]
	if ok && len(privacyParam[0]) > 0 {
		groupsFilter.Privacy = &privacyParam[0]
	}

	titles, ok := r.URL.Query()["title"]
	if ok && len(titles[0]) > 0 {
		groupsFilter.Title = &titles[0]
	}

	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			groupsFilter.Offset = &val
		}
	}

	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			groupsFilter.Limit = &val
		}
	}

	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		groupsFilter.Order = &orders[0]
	}

	hiddens, ok := r.URL.Query()["include_hidden"]
	if ok && len(hiddens[0]) > 0 {
		if strings.ToLower(hiddens[0]) == "true" {
			val := true
			groupsFilter.IncludeHidden = &val
		}
	}

	err := json.NewDecoder(r.Body).Decode(&groupsFilter)
	if err != nil {
		// just log an error and proceed and assume an empty filter
		log.Printf("adminapis.GetGroupsV2() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
	}

	groups, err := h.app.Services.GetGroups(nil, groupsFilter)
	if err != nil {
		log.Printf("adminapis.GetGroupsV2() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []model.Group{}
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("adminapis.GetGroupsV2() error on marshal the groups items")
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
// @Param title query string false "Deprecated - instead use request body filter! Filtering by group's title (case-insensitive)"
// @Param category query string false "Deprecated - instead use request body filter! category - filter by category"
// @Param privacy query string false "Deprecated - instead use request body filter! privacy - filter by privacy"
// @Param offset query string false "Deprecated - instead use request body filter! offset - skip number of records"
// @Param limit query string false "Deprecated - instead use request body filter! limit - limit the result"
// @Param include_hidden query string false "Deprecated - instead use request body filter! include_hidden - Includes hidden groups if a search by title is performed. Possible value is true. Default false."
// @Param data body model.GroupsFilter true "body data"
// @Success 200 {array} model.Group
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/v2/user/groups [get]
func (h *AdminApisHandler) GetUserGroupsV2(current *model.User, w http.ResponseWriter, r *http.Request) {
	var groupsFilter model.GroupsFilter

	catogies, ok := r.URL.Query()["category"]
	if ok && len(catogies[0]) > 0 {
		groupsFilter.Category = &catogies[0]
	}

	privacyParam, ok := r.URL.Query()["privacy"]
	if ok && len(privacyParam[0]) > 0 {
		groupsFilter.Privacy = &privacyParam[0]
	}

	titles, ok := r.URL.Query()["title"]
	if ok && len(titles[0]) > 0 {
		groupsFilter.Title = &titles[0]
	}

	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			groupsFilter.Offset = &val
		}
	}

	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			groupsFilter.Limit = &val
		}
	}

	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		groupsFilter.Order = &orders[0]
	}

	hiddens, ok := r.URL.Query()["include_hidden"]
	if ok && len(hiddens[0]) > 0 {
		if strings.ToLower(hiddens[0]) == "true" {
			val := true
			groupsFilter.IncludeHidden = &val
		}
	}

	err := json.NewDecoder(r.Body).Decode(&groupsFilter)
	if err != nil {
		// just log an error and proceed and assume an empty filter
		log.Printf("adminapis.GetGroupsV2() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
	}

	groups, err := h.app.Services.GetUserGroups(current.ID, groupsFilter)
	if err != nil {
		log.Printf("adminapis.GetUserGroupsV2() error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []model.Group{}
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("adminapis.GetUserGroupsV2() error on marshal the user groups items")
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
// @Success 200 {object} model.Group
// @Security AppUserAuth
// @Router /api/admin/v2/groups/{id} [get]
func (h *AdminApisHandler) GetGroupV2(current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("adminapis.GetGroupV2() id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(current, id)
	if err != nil {
		log.Printf("adminapis.GetGroupV2() error on getting group %s - %s", id, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if group == nil {
		log.Printf("adminapis.GetGroupV2() group %s not found", id)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	data, err := json.Marshal(group)
	if err != nil {
		log.Println("adminapis.GetGroupV2() error on marshal the group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
