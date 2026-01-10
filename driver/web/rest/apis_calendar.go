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
	"groups/core/model"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

// GetGroupCalendarEventsV3 Gets the group calendar events
// @Description Gets the group calendar events
// @ID GetGroupCalendarEventsV3
// @Tags Client
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Param data body model.GroupEventFilter false "body data"
// @Success 200 {object} string
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{group-id}/events/v3/load [post]
func (h *ApisHandler) GetGroupCalendarEventsV3(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	var filter model.GroupEventFilter
	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.GetGroupsV2() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &filter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("adminapis.GetGroupsV2() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	//check if allowed to see the events for this group
	group, hasPermission := h.app.Services.CheckUserGroupMembershipPermission(OrgID, current, groupID)
	if group == nil || group.CurrentMember == nil || !hasPermission {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	published := true
	events, err := h.app.Services.GetGroupCalendarEvents(OrgID, current, groupID, &published, filter)
	if err != nil {
		log.Printf("error getting group events - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(events)
	if err != nil {
		log.Println("Error on marshal the group events")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type createCalendarEventMultiGroupData struct {
	AdminsIdentifiers []model.AccountIdentifiers `json:"admins_identifiers"`
	Event             map[string]interface{}     `json:"event"`
	GroupIDs          []string                   `json:"group_ids"`
}

// CreateCalendarEventMultiGroup Create a calendar event and link it to multiple group ids
// @Description Create a calendar event and link it to multiple group ids
// @ID CreateCalendarEventMultiGroup
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createCalendarEventMultiGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} createCalendarEventMultiGroupData
// @Security AppUserAuth
// @Router /api/group/events/v3 [post]
func (h *ApisHandler) CreateCalendarEventMultiGroup(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createCalendarEventMultiGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event, groupIDs, err := h.app.Services.CreateCalendarEventForGroups(OrgID, nil, current, requestData.Event, requestData.GroupIDs)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createCalendarEventMultiGroupData{
		Event:    event,
		GroupIDs: groupIDs,
	})
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type createCalendarEventSingleGroupData struct {
	Event     map[string]interface{} `json:"event"`
	ToMembers []model.ToMember       `json:"to_members"`
}

// CreateCalendarEventSingleGroup Create a calendar event and link it to a single group id
// @Description Create a calendar event and link it to a single group id
// @ID CreateCalendarEventSingleGroup
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "group id"
// @Param data body createCalendarEventSingleGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} createCalendarEventSingleGroupData
// @Security AppUserAuth
// @Router /api/group/{group-id}/events/v3 [post]
func (h *ApisHandler) CreateCalendarEventSingleGroup(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("apis.CreateCalendarEventSingleGroup() id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createCalendarEventSingleGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membship, err := h.app.Services.FindGroupMembership(OrgID, groupID, current.ID)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() - Error retrieving user membership for group %s - %s\n", groupID, err.Error())
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if membship == nil || !membship.IsAdmin() {
		log.Printf("aapi.CreateCalendarEventSingleGroup() - User %s is not admin of the group - %s\n", current.ID, groupID)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	event, member, err := h.app.Services.CreateCalendarEventSingleGroup(OrgID, current, requestData.Event, groupID, requestData.ToMembers)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createCalendarEventSingleGroupData{
		Event:     event,
		ToMembers: member,
	})
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type updateCalendarEventSingleGroupData struct {
	Event     map[string]interface{} `json:"event"`
	ToMembers []model.ToMember       `json:"to_members"`
}

// UpdateCalendarEventSingleGroup Updates a calendar event for a single group id
// @Description Updates a calendar event and for a single group id
// @ID UpdateCalendarEventSingleGroup
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "group id"
// @Param data body updateCalendarEventSingleGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} updateCalendarEventSingleGroupData
// @Security AppUserAuth
// @Router /api/group/{group-id}/events/v3 [put]
func (h *ApisHandler) UpdateCalendarEventSingleGroup(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("apis.UpdateCalendarEventSingleGroup() id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("api.UpdateCalendarEventSingleGroup() Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData updateCalendarEventSingleGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("api.UpdateCalendarEventSingleGroup() Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("api.UpdateCalendarEventSingleGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membship, err := h.app.Services.FindGroupMembership(OrgID, groupID, current.ID)
	if err != nil {
		log.Printf("api.UpdateCalendarEventSingleGroup() - Error retrieving user membership for group %s - %s\n", groupID, err.Error())
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if membship == nil || !membship.IsAdmin() {
		log.Printf("aapi.UpdateCalendarEventSingleGroup() - User %s is not admin of the group - %s\n", current.ID, groupID)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	event, member, err := h.app.Services.UpdateCalendarEventSingleGroup(OrgID, current, requestData.Event, groupID, requestData.ToMembers)
	if err != nil {
		log.Printf("api.UpdateCalendarEventSingleGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(updateCalendarEventSingleGroupData{
		Event:     event,
		ToMembers: member,
	})
	if err != nil {
		log.Printf("api.UpdateCalendarEventSingleGroup() Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// getPutAdminGroupIDsForEventIDRequestAndResponse
type getPutAdminGroupIDsForEventIDRequestAndResponse struct {
	GroupIDs []string `json:"group_ids"`
} // @name getPutAdminGroupIDsForEventIDRequestAndResponse

// GetAdminGroupIDsForEventID Get all group IDs where the current user is an admin
// @Description Get all group IDs where the current user is an admin
// @ID GetAdminGroupIDsForEventID
// @Tags Client
// @Param APP header string true "APP"
// @Param event-id path string true "Event ID"
// @Param data body getPutAdminGroupIDsForEventIDRequestAndResponse true "body data"
// @Success 200 {object} getPutAdminGroupIDsForEventIDRequestAndResponse
// @Security AppUserAuth
// @Router /api/user/event/{event-id}/groups [get]
func (h *ApisHandler) GetAdminGroupIDsForEventID(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	eventID := params["event-id"]
	if len(eventID) <= 0 {
		log.Println("Event ID is required")
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	//check if allowed to see the events for this group
	groupIDs, err := h.app.Services.FindAdminGroupsForEvent(OrgID, current, eventID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if groupIDs == nil {
		groupIDs = []string{}
	}

	data, err := json.Marshal(getPutAdminGroupIDsForEventIDRequestAndResponse{GroupIDs: groupIDs})
	if err != nil {
		log.Println("adminapis.GetAdminGroupIDsForEventID() - Error on marshal the group events")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateGroupMappingsEventID Updates the group mappings for an event with id
// @Description Updates the group mappings for an event with id
// @ID UpdateGroupMappingsEventID
// @Tags Client
// @Param APP header string true "APP"
// @Param event-id path string true "Event ID"
// @Success 200 {object} getPutAdminGroupIDsForEventIDRequestAndResponse
// @Security AppUserAuth
// @Router /api/user/event/{event-id}/groups [put]
func (h *ApisHandler) UpdateGroupMappingsEventID(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	eventID := params["event-id"]
	if len(eventID) <= 0 {
		log.Println("Event ID is required")
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("api.UpdateGroupMappingsEventID() Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData getPutAdminGroupIDsForEventIDRequestAndResponse
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("api.UpdateGroupMappingsEventID() Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to see the events for this group
	groupIDs, err := h.app.Services.UpdateGroupMappingsForEvent(OrgID, current, eventID, requestData.GroupIDs)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if groupIDs == nil {
		groupIDs = []string{}
	}

	data, err = json.Marshal(getPutAdminGroupIDsForEventIDRequestAndResponse{GroupIDs: groupIDs})
	if err != nil {
		log.Println("apis.UpdateGroupMappingsEventID() - Error on marshal the group events")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
