// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"encoding/json"
	"groups/core"
	"groups/core/model"
	"groups/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

// GetUserGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID AdminGetUserGroups
// @Tags Admin-V1
// @Accept  json
// @Param APP header string true "APP"
// @Param title query string false "Filtering by group's title (case-insensitive)"
// @Param category query string false "category - filter by category"
// @Param privacy query string false "privacy - filter by privacy"
// @Param offset query string false "offset - skip number of records"
// @Param limit query string false "limit - limit the result"
// @Param include_hidden query string false "include_hidden - Includes hidden groups if a search by title is performed. Possible value is true. Default false."
// @Success 200 {array} model.Group
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/admin/user/groups [get]
func (h *AdminApisHandler) GetUserGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	groups, err := h.app.Services.GetGroups(clientID, current, groupsFilter)
	if err != nil {
		log.Printf("error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupIDs := []string{}
	for _, grouop := range groups {
		groupIDs = append(groupIDs, grouop.ID)
	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: groupIDs,
	})
	if err != nil {
		log.Printf("Unable to retrieve memberships: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for index, group := range groups {
		group.ApplyLegacyMembership(membershipCollection)
		groups[index] = group
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

// GetAllGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID AdminGetAllGroups
// @Tags Admin-V1
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
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/admin/groups [get]
func (h *AdminApisHandler) GetAllGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("apis.GetAllGroups() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &groupsFilter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("apis.GetAllGroups() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	groups, err := h.app.Administration.GetGroups(clientID, groupsFilter)
	if err != nil {
		log.Printf("apis.GetAllGroups() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupIDs := []string{}
	for _, grouop := range groups {
		groupIDs = append(groupIDs, grouop.ID)
	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: groupIDs,
	})
	if err != nil {
		log.Printf("apis.GetAllGroups() unable to retrieve memberships: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for index, group := range groups {
		group.ApplyLegacyMembership(membershipCollection)
		groups[index] = group
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("apis.GetAllGroups() error on marshal the groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupStats Retrieves stats for a group by id
// @Description Retrieves stats for a group by id
// @ID AdminGetGroupStats
// @Tags Admin-V1
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.GroupStats
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/stats [get]
func (h *AdminApisHandler) GetGroupStats(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(clientID, current, groupID)
	if err != nil {
		log.Printf("error getting group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if group == nil {
		log.Printf("error getting group stats")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(group.Stats)
	if err != nil {
		log.Println("Error on marshal the group stats")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupPosts gets all posts for the desired group.
// @Description gets all posts for the desired group.
// @ID AdminGetGroupPosts
// @Tags Admin-V1
// @Param APP header string true "APP"
// @Success 200 {array} model.Post
// @Security AppUserAuth
// @Router /api/admin/group/{groupID}/posts [get]
func (h *AdminApisHandler) GetGroupPosts(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["group-id"]
	if len(id) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
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

	posts, err := h.app.Services.GetPosts(clientID, current, id, nil, false, offset, limit, order)
	if err != nil {
		log.Printf("error getting posts for group (%s) - %s", id, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(posts)
	if err != nil {
		log.Printf("error on marshal posts for group (%s) - %s", id, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupEvents gives the group events
// @Description Gives the group events.
// @ID AdminGetGroupEvents
// @Tags Admin-V1
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} string
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/events [get]
func (h *AdminApisHandler) GetGroupEvents(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	events, err := h.app.Services.GetEvents(clientID, current, groupID, false)
	if err != nil {
		log.Printf("error getting group events - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]string, len(events))
	for i, e := range events {
		result[i] = e.EventID
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Println("Error on marshal the group events")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// DeleteGroup deletes a group
// @Description Deletes a group.
// @ID AdminDeleteGroup
// @Tags Admin-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/admin/group/{id} [delete]
func (h *AdminApisHandler) DeleteGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroupEntity(clientID, id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, utils.NewServerError().JSONErrorString(), http.StatusInternalServerError)
		return
	}
	if group.AuthmanEnabled && !current.HasPermission("managed_group_admin") {
		log.Printf("%s is not allowed to update group settings '%s'. Only user with managed_group_admin permission could delete a managed group", current.Email, group.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	err = h.app.Services.DeleteGroup(clientID, current, id)
	if err != nil {
		log.Printf("Error on deleting group - %s\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

// DeleteGroupEvent deletes a group event
// @Description Deletes a group event
// @ID AdminDeleteGroupEvent
// @Tags Admin-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Param event-id path string true "Event ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/event/{event-id} [delete]
func (h *AdminApisHandler) DeleteGroupEvent(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}
	eventID := params["event-id"]
	if len(eventID) <= 0 {
		log.Println("Event id is required")
		http.Error(w, "Event id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteEvent(clientID, current, eventID, groupID)
	if err != nil {
		log.Printf("Error on deleting an event - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

// DeleteGroupPost Updates a post within the desired group.
// @Description Updates a post within the desired group.
// @ID AdminDeleteGroupPost
// @Tags Admin-V1
// @Accept  json
// @Param APP header string true "APP"
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/group/{groupId}/posts/{postId} [delete]
func (h *AdminApisHandler) DeleteGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	postID := params["postID"]
	if len(postID) <= 0 {
		log.Println("postID is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeletePost(clientID, current, groupID, postID, true)
	if err != nil {
		log.Printf("error deleting posts for post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// GetManagedGroupConfigs gets managed group configs
// @Description Gets managed group configs
// @ID AdminGetManagedGroupConfigs
// @Tags Admin
// @Accept json
// @Param APP header string true "APP"
// @Success 200 {array}  model.ManagedGroupConfig
// @Security AppUserAuth
// @Router /api/admin/managed-group-configs [get]
func (h *AdminApisHandler) GetManagedGroupConfigs(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	configs, err := h.app.Services.GetManagedGroupConfigs(clientID)
	if err != nil {
		log.Printf("error getting managed group configs events - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(configs)
	if err != nil {
		log.Println("Error on marshal managed group configs")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CreateManagedGroupConfig creates a new managed group config
// @Description Creates a new managed group config
// @ID AdminCreateManagedGroupConfig
// @Tags Admin
// @Accept plain
// @Param data body  model.ManagedGroupConfig true "body data"
// @Param APP header string true "APP"
// @Success 200 {object} model.ManagedGroupConfig
// @Security AppUserAuth
// @Router /api/admin/managed-group-configs [post]
func (h *AdminApisHandler) CreateManagedGroupConfig(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body on create managed group config - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var config model.ManagedGroupConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error on unmarshal the managed group config data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config.ClientID = clientID
	newConfig, err := h.app.Services.CreateManagedGroupConfig(config)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(newConfig)
	if err != nil {
		log.Println("Error on marshal created managed group config")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// UpdateManagedGroupConfig updates an existing managed group config
// @Description Updates an existing managed group config
// @ID AdminUpdateManagedGroupConfig
// @Tags Admin
// @Accept plain
// @Param data body  model.ManagedGroupConfig true "body data"
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/admin/managed-group-configs [put]
func (h *AdminApisHandler) UpdateManagedGroupConfig(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body on create managed group config - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var config model.ManagedGroupConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error on unmarshal the managed group config data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config.ClientID = clientID
	err = h.app.Services.UpdateManagedGroupConfig(config)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// DeleteManagedGroupConfig Deletes a managed group config
// @Description Deletes a managed group config
// @ID AdminDeleteManagedGroupConfig
// @Tags Admin
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/admin/managed-group-configs/{id} [delete]
func (h *AdminApisHandler) DeleteManagedGroupConfig(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]
	if len(id) <= 0 {
		log.Println("id param is required")
		http.Error(w, "id param is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteManagedGroupConfig(id, clientID)
	if err != nil {
		log.Printf("error deleting managed group config for id (%s) - %s", id, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// GetSyncConfig gets sync config
// @Description Gets sync config
// @ID AdminGetSyncConfigs
// @Tags Admin
// @Accept json
// @Param APP header string true "APP"
// @Success 200 {array}  model.SyncConfig
// @Security AppUserAuth
// @Router /api/admin/sync-configs [get]
func (h *AdminApisHandler) GetSyncConfig(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	configs, err := h.app.Services.GetSyncConfig(clientID)
	if err != nil {
		log.Printf("error getting sync config - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(configs)
	if err != nil {
		log.Println("Error on marshal sync config")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// SaveSyncConfig saves sync config
// @Description Saves sync config
// @ID AdminSaveSyncConfig
// @Tags Admin
// @Accept plain
// @Param data body model.SyncConfig true "body data"
// @Param APP header string true "APP"
// @Success 200
// @Security AppUserAuth
// @Router /api/admin/sync-configs [put]
func (h *AdminApisHandler) SaveSyncConfig(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body on create sync config - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var config model.SyncConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error on unmarshal the sync config data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config.ClientID = clientID
	err = h.app.Services.UpdateSyncConfig(config)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// SynchronizeAuthman Synchronizes Authman groups membership
// @Description Synchronizes Authman groups membership
// @Tags Admin
// @ID AdminSynchronizeAuthman
// @Accept json
// @Success 200
// @Security AppUserAuth
// @Router /admin/authman/synchronize [post]
func (h *AdminApisHandler) SynchronizeAuthman(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.SynchronizeAuthman(clientID)
	if err != nil {
		log.Printf("Error during Authman synchronization: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
