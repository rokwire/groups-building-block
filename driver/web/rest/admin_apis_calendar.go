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
// @ID AdminGetGroupCalendarEventsV3
// @Tags Admin
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Param data body model.GroupEventFilter false "body data"
// @Success 200 {object} string
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/group/{group-id}/events/v3/load [post]
func (h *AdminApisHandler) GetGroupCalendarEventsV3(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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
		log.Printf("adminapis.GetGroupCalendarEventsV3() - error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &filter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("adminapis.GetGroupCalendarEventsV3() - error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	//check if allowed to see the events for this group
	group, hasPermission := h.app.Services.CheckUserGroupMembershipPermission(OrgID, current, groupID)
	if group == nil || group.CurrentMember == nil || !hasPermission {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	events, err := h.app.Services.GetGroupCalendarEvents(OrgID, current, groupID, nil, filter)
	if err != nil {
		log.Printf("adminapis.GetGroupCalendarEventsV3() - error getting group events - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(events)
	if err != nil {
		log.Println("adminapis.GetGroupCalendarEventsV3() - Error on marshal the group events")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CreateCalendarEventMultiGroup Create a calendar event and link it to multiple group ids
// @Description Create a calendar event and link it to multiple group ids
// @ID AdminCreateCalendarEventMultiGroup
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createCalendarEventMultiGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} createCalendarEventMultiGroupData
// @Security AppUserAuth
// @Router /api/admin/group/events/v3 [post]
func (h *AdminApisHandler) CreateCalendarEventMultiGroup(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventMultiGroup() - Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createCalendarEventMultiGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventMultiGroup() - Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventMultiGroup() - Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// admins accounts
	var adminIdentifier []model.AccountIdentifiers
	if requestData.AdminsIdentifiers != nil {
		for _, s := range *&requestData.AdminsIdentifiers {
			accountID := ""
			var externalID *string

			//get account id if available
			if s.AccountID != nil && len(*s.AccountID) > 0 {
				value := s.AccountID
				accountID = *value
			}

			//get external id if available
			if s.ExternalID != nil && len(*s.ExternalID) > 0 {
				value := s.ExternalID
				externalID = value
			}
			// validate that we have at least one
			if len(accountID) == 0 && externalID == nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			adminIdentifier = append(adminIdentifier, model.AccountIdentifiers{AccountID: &accountID, ExternalID: externalID})
		}
	}

	event, groupIDs, err := h.app.Services.CreateCalendarEventForGroups(OrgID, adminIdentifier, current, requestData.Event, requestData.GroupIDs)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventMultiGroup() - Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createCalendarEventMultiGroupData{
		AdminsIdentifiers: adminIdentifier,
		Event:             event,
		GroupIDs:          groupIDs,
	})
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventMultiGroup() - Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CreateCalendarEventSingleGroup Create a calendar event and link it to a single group id
// @Description Create a calendar event and link it to a single group id
// @ID AdminCreateCalendarEventSingleGroup
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "group id"
// @Param data body createCalendarEventSingleGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} createCalendarEventSingleGroupData
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/events/v3 [post]
func (h *AdminApisHandler) CreateCalendarEventSingleGroup(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("adminapis.CreateCalendarEventSingleGroup() - id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createCalendarEventSingleGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membship, err := h.app.Services.FindGroupMembership(OrgID, groupID, current.ID)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - Error retrieving user membership for group %s - %s\n", groupID, err.Error())
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if membship == nil || !membship.IsAdmin() {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - User %s is not admin of the group - %s\n", current.ID, groupID)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	event, member, err := h.app.Services.CreateCalendarEventSingleGroup(OrgID, current, requestData.Event, groupID, requestData.ToMembers)
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createCalendarEventSingleGroupData{
		Event:     event,
		ToMembers: member,
	})
	if err != nil {
		log.Printf("adminapis.CreateCalendarEventSingleGroup() - Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateCalendarEventSingleGroup Updates a calendar event for a single group id
// @Description Updates a calendar event and for a single group id
// @ID AdminUpdateCalendarEventSingleGroup
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "group id"
// @Param data body updateCalendarEventSingleGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} updateCalendarEventSingleGroupData
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/events/v3 [put]
func (h *AdminApisHandler) UpdateCalendarEventSingleGroup(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("adminapis.UpdateCalendarEventSingleGroup() - id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData updateCalendarEventSingleGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membship, err := h.app.Services.FindGroupMembership(OrgID, groupID, current.ID)
	if err != nil {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - Error retrieving user membership for group %s - %s\n", groupID, err.Error())
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if membship == nil || !membship.IsAdmin() {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - User %s is not admin of the group - %s\n", current.ID, groupID)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	event, member, err := h.app.Services.UpdateCalendarEventSingleGroup(OrgID, current, requestData.Event, groupID, requestData.ToMembers)
	if err != nil {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(updateCalendarEventSingleGroupData{
		Event:     event,
		ToMembers: member,
	})
	if err != nil {
		log.Printf("adminapis.UpdateCalendarEventSingleGroup() - Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetAdminGroupIDsForEventID Get all group IDs where the current user is an admin
// @Description Get all group IDs where the current user is an admin
// @ID AdminGetAdminGroupIDsForEventID
// @Tags Admin
// @Param APP header string true "APP"
// @Param event-id path string true "Event ID"
// @Success 200 {object} getPutAdminGroupIDsForEventIDRequestAndResponse
// @Security AppUserAuth
// @Router /api/admin/user/event/{event_id}/groups [get]
func (h *AdminApisHandler) GetAdminGroupIDsForEventID(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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
// @ID AdminUpdateGroupMappingsEventID
// @Tags Admin
// @Param APP header string true "APP"
// @Param event-id path string true "Event ID"
// @Param data body getPutAdminGroupIDsForEventIDRequestAndResponse true "body data"
// @Success 200 {object} getPutAdminGroupIDsForEventIDRequestAndResponse
// @Security AppUserAuth
// @Router /api/admin/user/event/{event-id}/groups [get]
func (h *AdminApisHandler) UpdateGroupMappingsEventID(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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
