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
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"gopkg.in/go-playground/validator.v9"
)

// BBsAPIsHandler handles the rest BBs APIs implementation
type BBsAPIsHandler struct {
	app *core.Application
}

// NewBBsAPIsHandler creates new Building Block API handler instance
func NewBBsAPIsHandler(app *core.Application) BBsAPIsHandler {
	return BBsAPIsHandler{app: app}
}

// GetUserGroupMemberships gets the user groups memberships
// @Description Gives the user groups memberships
// @ID GetUserGroupMemberships
// @Tags BBs
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} userGroupShortDetail
// @Security Bearer
// @Router /bbs/api/user/{identifier}/groups [get]
func (h *BBsAPIsHandler) IntGetUserGroupMemberships(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	identifier := params["identifier"]
	if len(identifier) <= 0 {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "gets the user groups memberships", nil, nil, http.StatusInternalServerError, false)

	}
	externalID := identifier

	groups, err := h.app.Services.FindGroupsV3(clientID, &model.GroupsFilter{
		MemberExternalID: &externalID,
	})
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "gets the user groups memberships", nil, nil, http.StatusInternalServerError, false)
	}

	userGroups := make([]userGroupShortDetail, len(groups))
	for i, group := range groups {

		status := ""
		if group.CurrentMember != nil {
			status = group.CurrentMember.Status
		}

		ugm := userGroupShortDetail{
			ID:               group.ID,
			Title:            group.Title,
			Privacy:          group.Privacy,
			MembershipStatus: status,
		}

		userGroups[i] = ugm
	}

	data, err := json.Marshal(userGroups)
	if err != nil {
		log.Println("Error on marshal the user group membership")
		return l.HTTPResponseErrorAction(logutils.ActionGet, "gets the user groups memberships", nil, nil, http.StatusInternalServerError, false)

	}
	return l.HTTPResponseSuccessJSON(data)
}

// GetGroup Retrieves group details and members
// @Description Retrieves group details and members
// @ID GetGroup
// @Tags BBs
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} model.Group
// @Security Bearer
// @Router /bbs/api/group/{identifier} [get]
func (h *BBsAPIsHandler) IntGetGroup(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	identifier := params["identifier"]
	if len(identifier) <= 0 {
		log.Println("Identifier is required")
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Identifier is required", nil, nil, http.StatusInternalServerError, false)

	}

	group, err := h.app.Services.GetGroupEntity(clientID, identifier)
	if err != nil {
		log.Printf("Unable to retrieve group with ID '%s': %s", identifier, err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Unable to retrieve group with ID", nil, err, http.StatusInternalServerError, false)

	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
	})
	if err != nil {
		log.Printf("Unable to retrieve memberships: %s", err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Unable to retrieve memberships", nil, err, http.StatusInternalServerError, false)

	}

	group.ApplyLegacyMembership(membershipCollection)

	data, err := json.Marshal(group)
	if err != nil {
		log.Printf("Error on marshal the user group: %s", err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Error on marshal the user group", nil, err, http.StatusInternalServerError, false)

	}
	return l.HTTPResponseSuccessJSON(data)
}

// GetGroupMembersByGroupTitle Retrieves group members by  title
// @Description Retrieves group members by  title
// @ID GetGroupMembersByGroupTitle
// @Tags BBs
// @Accept json
// @Param identifier path string true "Title"
// @Param offset query string false "Offsetting result"
// @Param limit query string false "Limiting the result"
// @Success 200 {array} model.ShortMemberRecord
// @Security Bearer
// @Router /bbs/api/group/title/{title}/members [get]
func (h *BBsAPIsHandler) IntGetGroupMembersByGroupTitle(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	title := params["title"]

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

	group, err := h.app.Services.GetGroupEntityByTitle(clientID, title)
	if err != nil {
		log.Printf("Unable to retrieve group with title '%s': %s", title, err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Unable to retrieve group with title", nil, err, http.StatusInternalServerError, false)

	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		log.Printf("Unable to retrieve memberships: %s", err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Unable to retrieve memberships", nil, err, http.StatusInternalServerError, false)

	}

	shortMembers := []model.ShortMemberRecord{}
	for _, membership := range membershipCollection.Items {
		shortMembers = append(shortMembers, membership.ToShortMemberRecord())
	}

	data, err := json.Marshal(shortMembers)
	if err != nil {
		log.Printf("Error on marshal the short member list: %s", err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Error on marshal the short member list", nil, err, http.StatusInternalServerError, false)

	}
	return l.HTTPResponseSuccessJSON(data)
}

// SynchronizeAuthman Synchronizes Authman groups membership
// @Description Synchronizes Authman groups membership
// @ID SynchronizeAuthmanBBs
// @Tags BBs
// @Accept json
// @Success 200
// @Security Bearer
// @Router /bbs/api/authman/synchronize [post]
func (h *BBsAPIsHandler) SynchronizeAuthman(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	err := h.app.Services.SynchronizeAuthman(clientID)
	if err != nil {
		log.Printf("Error during Authman synchronization: %s", err)
		return l.HTTPResponseErrorAction(logutils.ActionApply, "Error during Authman synchronization", nil, err, http.StatusInternalServerError, false)

	}
	return l.HTTPResponseSuccess()
}

// GroupStats Retrieve group stats
// @Description Retrieve group stats
// @ID GroupStats
// @Tags BBs
// @Accept json
// @Success 200 {object} GroupsStats
// @Security Bearer
// @Router /bbs/api/stats [get]
func (h *BBsAPIsHandler) GroupStats(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	groups, err := h.app.Services.GetAllGroups(clientID)
	if err != nil {
		log.Printf("Error GroupStats(%s): %s", clientID, err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Error GroupStats", nil, err, http.StatusInternalServerError, false)

	}
	groupsCount := len(groups)
	groupStatList := []GroupStat{}
	if groupsCount > 0 {
		for _, group := range groups {

			groupStatList = append(groupStatList, GroupStat{
				Title:          group.Title,
				Privacy:        group.Privacy,
				AuthmanEnabled: group.AuthmanEnabled,
				Stats:          group.Stats,
			})
		}
	}

	groupsStats := GroupsStats{
		GroupsCount: groupsCount,
		GroupsList:  groupStatList,
	}

	data, err := json.Marshal(groupsStats)
	if err != nil {
		log.Printf("Error GroupStats(%s): %s", clientID, err)
		return l.HTTPResponseErrorAction(logutils.ActionGet, "Error GroupStats", nil, err, http.StatusInternalServerError, false)

	}
	return l.HTTPResponseSuccessJSON(data)
}

// CreateGroupEvent creates a group event
// @Description Creates a group event
// @ID CreateGroupEvent
// @Tags BBs
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body CreateGroupEventRequestBody true "body data"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security Bearer
// @Router /bbs/api/group/{group-id}/events [post]
func (h *BBsAPIsHandler) CreateGroupEvent(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "group-id is required", nil, nil, http.StatusInternalServerError, false)

	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create group event - %s\n", err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on marshal the create group event", nil, err, http.StatusInternalServerError, false)

	}

	var requestData intCreateGroupEventRequestBody
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create event request data - %s\n", err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on unmarshal the create event request data", nil, err, http.StatusBadRequest, false)

	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create event data - %s\n", err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on validating create event data", nil, err, http.StatusBadRequest, false)

	}

	//check if allowed to create
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
	if err != nil {
		log.Println(err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error", nil, err, http.StatusBadRequest, false)

	}
	if group == nil {
		log.Printf("there is no a group for the provided group id - %s", groupID)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error", nil, err, http.StatusBadRequest, false)

	}

	grEvent, err := h.app.Services.CreateEvent(clientID, nil, requestData.EventID, group, requestData.ToMembersList, requestData.Creator)
	if err != nil {
		log.Printf("Error on creating an event - %s\n", err)
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on validating create event data", nil, err, http.StatusInternalServerError, false)

	}

	responseData, err := json.Marshal(grEvent)
	if err != nil {
		log.Printf("Error on marshaling an event - %s\n", err)
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on marshaling an event", nil, err, http.StatusInternalServerError, false)

	}

	return l.HTTPResponseSuccessJSON(responseData)
}

// DeleteGroupEvent deletes a group event
// @Description Deletes a group event
// @ID IntDeleteGroupEvent
// @Tags BBs
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Param event-id path string true "Event ID"
// @Success 200
// @Security Bearer
// @Router /api/int/group/{group-id}/events/{event-id} [delete]
func (h *BBsAPIsHandler) DeleteGroupEvent(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "group-id is required", nil, nil, http.StatusInternalServerError, false)

	}
	eventID := params["event-id"]
	if len(eventID) <= 0 {
		log.Println("Event id is required")
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "event-id is required", nil, nil, http.StatusInternalServerError, false)

	}

	err := h.app.Services.DeleteEvent(clientID, nil, eventID, groupID)
	if err != nil {
		log.Printf("Error on deleting an event - %s\n", err)
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on deleting an event", nil, nil, http.StatusInternalServerError, false)

	}

	return l.HTTPResponseSuccess()
}

// SendGroupNotification Sends a notification to members of a group
// @Description Sends a notification to members of a group
// @ID SendGroupNotification
// @Tags BBs
// @Accept json
// @Param APP header string true "APP"
// @Param data body sendGroupNotificationRequestBody true "body data"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security Bearer
// @Router /bbs/api/int/group/{group-id}/notification [post]
func (h *BBsAPIsHandler) SendGroupNotification(clientID string, l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "group-id is required", nil, nil, http.StatusInternalServerError, false)

	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on read the sendGroupNotificationRequestBody - %s\n", err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on read the sendGroupNotificationRequestBody", nil, err, http.StatusBadRequest, false)

	}

	var requestData sendGroupNotificationRequestBody
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the sendGroupNotificationRequestBody - %s\n", err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on unmarshal the sendGroupNotificationRequestBody", nil, err, http.StatusBadRequest, false)

	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating sendGroupNotificationRequestBody - %s\n", err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error on validating sendGroupNotificationRequestBody", nil, err, http.StatusBadRequest, false)

	}

	notification := model.GroupNotification{
		GroupID:        groupID,
		Members:        requestData.Members,
		Sender:         requestData.Sender,
		MemberStatuses: requestData.MemberStatuses,
		Subject:        requestData.Subject,
		Body:           requestData.Body,
		Topic:          requestData.Topic,
		Data:           requestData.Data,
	}
	err = h.app.Services.SendGroupNotification(clientID, notification)
	if err != nil {
		log.Println(err.Error())
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "Error", nil, err, http.StatusBadRequest, false)

	}
	return l.HTTPResponseSuccess()
}
