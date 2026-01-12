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

	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"gopkg.in/go-playground/validator.v9"
)

// ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

// SynchronizeAuthman Synchronizes Authman groups membership
// @Description Synchronizes Authman groups membership
// @Tags Client
// @ID InternalSynchronizeAuthman
// @Accept json
// @Success 200
// @Security AppUserAuth
// @Router /authman/synchronize [post]
func (h *ApisHandler) SynchronizeAuthman(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.SynchronizeAuthman(orgID)
	if err != nil {
		log.Printf("Error during Authman synchronization: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// Version gives the service version
// @Description Gives the service version.
// @ID Version
// @Tags Client
// @Produce plain
// @Success 200 {string} v1.4.9
// @Router /version [get]
func (h ApisHandler) Version(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	return log.HTTPResponseSuccessMessage(h.app.Services.GetVersion())
}

// LoginUser Deprecated: Don't use it! Logs in the user and refactor the user record and linked data if need
// @Description Deprecated: Don't use it! Logs in the user and refactor the user record and linked data if need
// @ID LoginUser
// @Tags Client
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user/login [get]
func (h *ApisHandler) LoginUser(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// DeleteUser Deletes a user with all the involved information from the Notifications BB (this includes - group membership & posts (and child posts - no matter of the creator))
// @Description Deletes a user with all the involved information from the Notifications BB (this includes - group membership & posts (and child posts - no matter of the creator))
// @ID DeleteUser
// @Tags Client
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user [delete]
func (h *ApisHandler) DeleteUser(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.DeleteUser(orgID, current)
	if err != nil {
		log.Printf("error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetUserGroupMemberships gets the user groups memberships
// @Description Gives the user groups memberships
// @ID GetUserGroupMemberships
// @Tags Client
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} userGroupShortDetail
// @Security AppUserAuth
// @Router /api/user/group-memberships [get]
func (h *ApisHandler) GetUserGroupMemberships(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	userGroups, err := h.app.Services.GetUserGroups(orgID, current, model.GroupsFilter{})
	if err != nil {
		log.Println("The user has no group memberships")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userGroupShortDetailList := make([]userGroupShortDetail, len(userGroups))
	for i, group := range userGroups {

		ugm := userGroupShortDetail{
			ID:               group.ID,
			Title:            group.Title,
			Privacy:          group.Privacy,
			MembershipStatus: group.CurrentMember.Status,
		}

		userGroupShortDetailList[i] = ugm
	}

	data, err := json.Marshal(userGroupShortDetailList)
	if err != nil {
		log.Println("Error on marshal the user group membership")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type getGroupResponse struct {
	ID                  string   `json:"id"`
	Category            string   `json:"category"`
	Title               string   `json:"title"`
	Privacy             string   `json:"privacy"`
	Description         *string  `json:"description"`
	ImageURL            *string  `json:"image_url"`
	WebURL              *string  `json:"web_url"`
	Tags                []string `json:"tags"`
	MembershipQuestions []string `json:"membership_questions"`

	Members []struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		PhotoURL       string `json:"photo_url"`
		Status         string `json:"status"`
		RejectedReason string `json:"rejected_reason"`

		MemberAnswers []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		} `json:"member_answers"`

		DateCreated time.Time  `json:"date_created"`
		DateUpdated *time.Time `json:"date_updated"`
	} `json:"members"`

	DateCreated time.Time  `json:"date_created"`
	DateUpdated *time.Time `json:"date_updated"`
} // @name getGroupResponse

// GetGroup gets a group
// @Description Gives a group
// @ID GetGroup
// @Tags Client
// @Accept json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {object} getGroupResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/groups/{id} [get]
func (h *ApisHandler) GetGroup(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(orgID, current, id)
	if err != nil {
		log.Printf("adminapis.GetGroupV2() error on getting group %s", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(orgID, model.MembershipFilter{
		GroupIDs: []string{id},
	})
	if err != nil {
		log.Printf("Unable to retrieve memberships: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	group.ApplyLegacyMembership(membershipCollection)

	data, err := json.Marshal(group)
	if err != nil {
		log.Println("Error on marshal the group")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type createPendingMemberRequest struct {
	MemberAnswers []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"member_answers"`
	NotificationsPreferences *model.NotificationsPreferences `json:"notifications_preferences"`
} // @name createPendingMemberRequest

// CreatePendingMember creates a group pending member
// @Description Creates a group pending member
// @ID CreatePendingMember
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createPendingMemberRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {string} Successfully created
// @Failure 423 {string} block_new_membership_requests flag is true
// @Security AppUserAuth
// @Router /api/group/{group-id}/pending-members [post]
func (h *ApisHandler) CreatePendingMember(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a pending member - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createPendingMemberRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create pending member data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//validate
	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create pending member data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroupEntity(orgID, groupID)
	if err != nil {
		log.Printf("error getting a group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if group == nil || group.BlockNewMembershipRequests {
		log.Printf("error on create pending member - block_new_membership_requests is true")
		http.Error(w, "block_new_membership_requests flag is true", http.StatusLocked)
		return
	}

	memberAnswers := requestData.MemberAnswers
	mAnswers := make([]model.MemberAnswer, len(memberAnswers))
	if memberAnswers != nil {
		for i, current := range memberAnswers {
			mAnswers[i] = model.MemberAnswer{Question: current.Question, Answer: current.Answer}
		}
	}

	member := &model.GroupMembership{
		UserID:        current.ID,
		ExternalID:    current.ExternalID,
		Name:          current.Name,
		NetID:         current.NetID,
		Email:         current.Email,
		MemberAnswers: mAnswers,
	}

	if requestData.NotificationsPreferences != nil {
		member.NotificationsPreferences = *requestData.NotificationsPreferences
	}

	err = h.app.Services.CreatePendingMembership(orgID, current, group, member)
	if err != nil {
		log.Printf("Error on creating a pending member - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully created"))
}

// DeletePendingMember deletes a group pending member
// @Description Deletes a group pending member
// @ID DeletePendingMember
// @Tags Client
// @Accept plain
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {string} string "Successfuly deleted"
// @Security AppUserAuth
// @Router /api/group/{group-id}/pending-members [delete]
func (h *ApisHandler) DeletePendingMember(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeletePendingMembership(orgID, current, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

// GetGroupMembers Gets the list of group members. The result would be empty if the current user doesn't belong to the requested group.
// @Description Gets the list of group members. The result would be empty if the current user doesn't belong to the requested group.
// @ID GetGroupMembers
// @Tags Client
// @Accept plain
// @Param data body model.MembershipFilter true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.GroupMembership
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [get]
func (h *ApisHandler) GetGroupMembers(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal model.MembershipFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var request model.MembershipFilter
	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &request)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("Error on unmarshal model.MembershipFilter request body - %s\n", err.Error())
		}
	}

	request.GroupIDs = append(request.GroupIDs, groupID)

	//check if allowed to update
	members, err := h.app.Services.FindGroupMemberships(orgID, request)
	if err != nil {
		log.Printf("api.GetGroupMembers error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if members.Items == nil {
		members.Items = []model.GroupMembership{}
	}

	data, err := json.Marshal(members.Items)
	if err != nil {
		log.Printf("api.GetGroupMembers error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupMembersV2 Gets the list of group members. The result would be empty if the current user doesn't belong to the requested group.
// @Description Gets the list of group members. The result would be empty if the current user doesn't belong to the requested group.
// @ID GetGroupMembersV2
// @Tags Client
// @Accept plain
// @Param data body model.MembershipFilter true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.GroupMembership
// @Security AppUserAuth
// @Router /api/group/{group-id}/members/v2 [post]
func (h *ApisHandler) GetGroupMembersV2(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal model.MembershipFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var request model.MembershipFilter
	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &request)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("Error on unmarshal model.MembershipFilter request body - %s\n", err.Error())
		}
	}

	request.GroupIDs = append(request.GroupIDs, groupID)

	//check if allowed to update
	members, err := h.app.Services.FindGroupMemberships(orgID, request)
	if err != nil {
		log.Printf("api.GetGroupMembersV2 error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if members.Items == nil {
		members.Items = []model.GroupMembership{}
	}

	data, err := json.Marshal(members.Items)
	if err != nil {
		log.Printf("api.GetGroupMembersV2 error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// createMemberRequest
type createMemberRequest struct {
	UserID        string     `json:"user_id" bson:"user_id"`
	ExternalID    string     `json:"external_id" bson:"external_id"`
	Name          string     `json:"name" bson:"name"`
	NetID         string     `json:"net_id" bson:"net_id"`
	Email         string     `json:"email" bson:"email"`
	PhotoURL      string     `json:"photo_url" bson:"photo_url"`
	Status        string     `json:"status" bson:"status"` //pending, member, admin, rejected
	DateAttended  *time.Time `json:"date_attended" bson:"date_attended"`
	MemberAnswers []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"member_answers"`
} //@name createMemberRequest

// CreateMember Adds a member to a group. The current user is required to be an admin of the group
// @Description Adds a member to a group. The current user is required to be an admin of the group
// @ID CreateMember
// @Tags Client
// @Accept plain
// @Param data body createMemberRequest true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [post]
func (h *ApisHandler) CreateMember(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a pending member - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createMemberRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create pending member data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(requestData.UserID) == 0 && len(requestData.ExternalID) == 0 {
		log.Printf("error: api.CreateMember() - expected user_id or external_id")
		http.Error(w, "expected user_id or external_id", http.StatusBadRequest)
		return
	}

	if requestData.Status != "" &&
		!(requestData.Status == "member" ||
			requestData.Status == "admin" ||
			requestData.Status == "rejected" ||
			requestData.Status == "pending") {
		log.Printf("error: api.CreateMember() - expected status with possible value (member, admin, rejected, pending)")
		http.Error(w, "expected status with possible value (member, admin, rejected, pending)", http.StatusBadRequest)
		return
	} else if requestData.Status == "" {
		requestData.Status = "pending"
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroupEntity(orgID, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("error: api.CreateMember() - there is no a group for the provided id - %s", groupID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	membership, _ := h.app.Services.FindGroupMembership(orgID, group.ID, current.ID)
	if membership == nil || !membership.IsAdmin() {
		log.Printf("error: api.CreateMember() - %s is not allowed to create group member", current.Email)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	member := model.GroupMembership{
		GroupID:      groupID,
		UserID:       requestData.UserID,
		ExternalID:   requestData.ExternalID,
		Email:        requestData.Email,
		Name:         requestData.Name,
		NetID:        requestData.NetID,
		PhotoURL:     requestData.PhotoURL,
		Status:       requestData.Status,
		DateAttended: requestData.DateAttended,
	}

	err = h.app.Services.CreateMembership(orgID, current, group, &member)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// createMembershipsRequest is the request body for creating multiple group memberships
type createMembershipsRequest struct {
	Members []model.MembershipStatus `json:"members"`
}

//@name createMembershipsRequest

// MultiCreateMembers create multiple group memberships.
// @Description Create multiple members in group with desired status
// @ID MultiCreateMembers
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createMembershipsRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security AppUserAuth
// @Router /group/{group-id}/members [post]
func (h *ApisHandler) MultiCreateMembers(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("apis.MultiUpdateMembers() Error on Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("apis.MultiUpdateMembers() Error on read request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createMembershipsRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("apis.MultiUpdateMembers() Error on unmarshal request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	for _, item := range requestData.Members {
		err = validate.Struct(item)
		if err != nil {
			log.Printf("apis.MultiUpdateMembers() Error on validating request data - %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	//check if allowed to see the events for this group
	group, hasPermission := h.app.Services.CheckUserGroupMembershipPermission(orgID, current, groupID)
	if group == nil || group.CurrentMember == nil || !hasPermission {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	err = h.app.Services.CreateMembershipsStatuses(orgID, current, groupID, model.MembershipStatuses(requestData.Members))
	if err != nil {
		log.Printf("apis.MultiUpdateMembers() Error - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// MultiUpdateMembers Updates multiple members in a group at once with status and other details
// @Description Updates multiple members in a group at once with status and other details
// @ID MultiUpdateMembers
// @Tags Client
// @Accept plain
// @Param data body model.MembershipMultiUpdate true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{group-id}/members/multi-update [put]
func (h *ApisHandler) MultiUpdateMembers(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error: api.MultiUpdateMembers() - Error on marshal create a pending member - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var operation model.MembershipMultiUpdate
	err = json.Unmarshal(data, &operation)
	if err != nil {
		log.Printf("error: api.MultiUpdateMembers() - Error on unmarshal the create pending member data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(operation.UserIDs) == 0 {
		log.Printf("error: api.MultiUpdateMembers() - expected user_id or external_id")
		http.Error(w, "expected user_id or external_id", http.StatusBadRequest)
		return
	}

	if !operation.IsStatusValid() {
		log.Printf("error: api.MultiUpdateMembers() - expected status with possible value (null, member, admin, rejected, pending)")
		http.Error(w, "expected status with possible value (member, admin, rejected, pending)", http.StatusBadRequest)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroup(orgID, current, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("error: api.MultiUpdateMembers() - there is no a group for the provided id - %s", groupID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if group.CurrentMember == nil || !group.CurrentMember.IsAdmin() {
		log.Printf("error: api.MultiUpdateMembers() - %s is not allowed to create group member", current.Email)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.UpdateMemberships(orgID, current, group, operation)
	if err != nil {
		log.Printf("error: api.MultiUpdateMembers() - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// DeleteMember deletes a member membership from a group
// @Description Deletes a member membership from a group
// @ID DeleteMember
// @Tags Client
// @Accept plain
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {string} string "Successfuly deleted"
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [delete]
func (h *ApisHandler) DeleteMember(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteMembership(orgID, current, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

type membershipApprovalRequest struct {
	Approve        *bool  `json:"approve" validate:"required"`
	RejectedReason string `json:"reject_reason"`
} // @name membershipApprovalRequest

// MembershipApproval approve/deny a membership
// @Description Аpprove/Deny a membership
// @ID MembershipApproval
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body membershipApprovalRequest true "body data"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully processed
// @Security AppUserAuth
// @Router /api/memberships/{membership-id}/approval [put]
func (h *ApisHandler) MembershipApproval(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the membership item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData membershipApprovalRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the membership request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating membership data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membership, err := h.app.Services.FindGroupMembershipByID(orgID, membershipID)
	if err != nil || membership == nil {
		log.Printf("Membership %s not found - %s\n", membershipID, err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroup(orgID, current, membership.GroupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided membership id - %s", membershipID)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if group.CurrentMember == nil || !group.CurrentMember.IsAdmin() {
		log.Printf("%s is not allowed to make approval", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	approve := *requestData.Approve
	rejectedReason := requestData.RejectedReason

	err = h.app.Services.ApplyMembershipApproval(orgID, current, membershipID, approve, rejectedReason)
	if err != nil {
		log.Printf("Error on applying membership approval - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully processed"))
}

// DeleteMembership deletes membership
// @Description Deletes a membership
// @ID DeleteMembership
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/memberships/{membership-id} [delete]
func (h *ApisHandler) DeleteMembership(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	membership, err := h.app.Services.FindGroupMembershipByID(orgID, membershipID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if membership == nil {
		log.Printf("Membership %s not found", membershipID)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroup(orgID, current, membership.GroupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided membership id - %s", membershipID)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if group.CurrentMember == nil || !group.CurrentMember.IsAdmin() {
		log.Printf("%s is not allowed to delete membership", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.DeleteMembershipByID(orgID, current, membershipID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

type updateMembershipRequest struct {
	Status                   *string                         `json:"status" validate:"required,oneof=member admin"`
	DateAttended             *time.Time                      `json:"date_attended"`
	NotificationsPreferences *model.NotificationsPreferences `json:"notifications_preferences"`
} // @name updateMembershipRequest

// UpdateMembership updates a membership. Only admin can update the status and date_attended fields of a membership record. Member is allowed to update only his/her notification preferences.
// @Description Updates a membership. Only admin can update the status and date_attended fields of a membership record. Member is allowed to update only his/her notification preferences.
// @ID UpdateMembership
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body updateMembershipRequest true "body data"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully updated
// @Security AppUserAuth
// @Router /api/memberships/{membership-id} [put]
func (h *ApisHandler) UpdateMembership(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the membership update item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData updateMembershipRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the membership request update data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating membership update data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membership, err := h.app.Services.FindGroupMembershipByID(orgID, membershipID)
	if err != nil || membership == nil {
		log.Printf("Membership %s not found - %s\n", membershipID, err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroup(orgID, current, membership.GroupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided membership id - %s", membershipID)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if group.CurrentMember == nil || (!group.CurrentMember.IsAdmin() && group.CurrentMember.UserID != membership.UserID) {
		log.Printf("%s is not allowed to make update on membership record %s", current.Email, membershipID)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	var status *string
	var dateAttended *time.Time
	var notificationsPreferences *model.NotificationsPreferences
	if group.CurrentMember.IsAdmin() {
		status = requestData.Status
		dateAttended = requestData.DateAttended
	}
	if group.CurrentMember.UserID == membership.UserID {
		notificationsPreferences = requestData.NotificationsPreferences
	}

	err = h.app.Services.UpdateMembership(orgID, current, membershipID, status, dateAttended, notificationsPreferences)
	if err != nil {
		log.Printf("Error on updating membership - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated"))
}

// SynchAuthmanGroup Synchronizes Authman group. Only admin of the group could initiate the operation
// @Description Synchronizes Authman group. Only admin of the group could initiate the operation
// @ID SynchAuthmanGroup
// @Tags Client
// @Accept plain
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{group-id}/authman/synchronize [post]
func (h *ApisHandler) SynchAuthmanGroup(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to update
	isAdmin, err := h.app.Services.IsGroupAdmin(orgID, groupID, current.ID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		log.Printf("%s is not allowed to make Authman Synch for group '%s'", current.Email, groupID)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.SynchronizeAuthmanGroup(orgID, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetResearchProfileUserCount Retrieves the user count matching the provided research profile
// @Description Retrieves the user count matching the provided research profile
// @ID GetResearchProfileUserCount
// @Tags Client
// @Accept json
// @Param APP header string true "APP"
// @Param data body map[string]map[string][]string true "Research profile"
// @Success 200 {integer} 0
// @Security AppUserAuth
// @Router /api/research-profile/user-count [post]
func (h *ApisHandler) GetResearchProfileUserCount(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	var researchProfile map[string]map[string]any
	err := json.NewDecoder(r.Body).Decode(&researchProfile)
	if err != nil {
		log.Printf("error decoding body - %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := h.app.Services.GetResearchProfileUserCount(orgID, current, researchProfile)
	if err != nil {
		log.Printf("error getting user count - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(count)
	if err != nil {
		log.Println("Error on marshal response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// reportAbuseGroupRequestBody request body for report abuse group API call
type reportAbuseGroupRequestBody struct {
	Comment string `json:"comment"`
} // @name reportAbuseGroupRequestBody

// ReportAbuseGroup Reports an abusive group
// @Description Reports an abusive group
// @ID ReportAbuseGroup
// @Tags Client
// @Accept  json
// @Param APP header string true "APP"
// @Param data body reportAbuseGroupRequestBody true "body data"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{id}/report/abuse [put]
func (h *ApisHandler) ReportAbuseGroup(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["id"]
	if len(groupID) <= 0 {
		log.Println("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on read reportAbuseGroupRequestBody - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var body reportAbuseGroupRequestBody
	err = json.Unmarshal(data, &body)
	if err != nil {
		log.Printf("error on unmarshal reportAbuseGroupRequestBody (%s) - %s", groupID, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(orgID, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided group id - %s", groupID)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.app.Services.ReportGroupAsAbuse(orgID, current, group, body.Comment)
	if err != nil {
		log.Printf("error on report group as abuse (%s) - %s", groupID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// GetUserData Gets all related user data
// @Description  Gets all related user data
// @ID GetUserData
// @Tags Client
// @Success 200 {object} model.UserDataResponse
// @Security AppUserAuth
// @Router /api/user-data [get]
func (h *ApisHandler) GetUserData(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	userData, err := h.app.Services.GetUserData(current.ID)
	if err != nil {
		log.Printf("error getting user data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	data, err := json.Marshal(userData)
	if err != nil {
		log.Printf("Error on read user data - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// NewApisHandler creates new rest Client APIs Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}

// NewAdminApisHandler creates new rest Asmin APIs Handler instance
func NewAdminApisHandler(app *core.Application) *AdminApisHandler {
	return &AdminApisHandler{app: app}
}

// NewInternalApisHandler creates new rest Internal APIs Handler instance
func NewInternalApisHandler(app *core.Application) *InternalApisHandler {
	return &InternalApisHandler{app: app}
}

// NewAnalyticsApisHandler creates new rest Analytics Handler instance
func NewAnalyticsApisHandler(app *core.Application) *AnalyticsApisHandler {
	return &AnalyticsApisHandler{app: app}
}

// NewBBApisHandler creates new rest BB Api Handler instance
func NewBBApisHandler(app *core.Application) *BBSApisHandler {
	return &BBSApisHandler{app: app}
}
