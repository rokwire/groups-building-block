package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"groups/core/model"
	"log"
	"net/http"
	"strconv"
)

// GetGroupsV2 gets groups. It can be filtered by category, title and privacy. V2
// @Description Gives the groups list. It can be filtered by category, title and privacy. V2
// @ID AdminGetGroupsV2
// @Tags Admin-V2
// @Accept  json
// @Param APP header string true "APP"
// @Param category query string false "Category"
// @Param title query string false "Filtering by group's title (case-insensitive)"
// @Param category query string false "category - filter by category"
// @Param privacy query string false "privacy - filter by privacy"
// @Param offset query string false "offset - skip number of records"
// @Param limit query string false "limit - limit the result"
// @Success 200 {array} model.Group
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

	groups, err := h.app.Services.GetGroups(clientID, current, category, privacy, title, offset, limit, order)
	if err != nil {
		log.Printf("apis.GetGroupsV2() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(groups)
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
// @Success 200 {array} model.Group
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

	data, err := json.Marshal(groups)
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
// @Success 200 {object} model.Group
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
	group, err := h.app.Services.GetGroup(clientID, current, id)
	if err != nil {
		log.Printf("apis.GetGroupV2() error getting a group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if group.Privacy == "private" {
		if current == nil || current.IsAnonymous {
			log.Println("apis.GetGroupV2() error - Anonymous user cannot see the events for a private group")

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}
		membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
		if (membership == nil || !membership.IsAdminOrMember()) && group.HiddenForSearch { // NB: group detail panel needs it for user not belonging to the group
			log.Printf("apis.GetGroupV2() error - %s cannot see the events for the %s private group as he/she is not a member or admin", current.Email, group.Title)

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}
	}

	data, err := json.Marshal(group)
	if err != nil {
		log.Println("apis.GetGroupV2() error on marshal the group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
