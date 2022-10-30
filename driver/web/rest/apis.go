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
	"fmt"
	"groups/core"
	"groups/core/model"
	"groups/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

// ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

// Version gives the service version
// @Description Gives the service version.
// @ID Version
// @Tags Client-V1
// @Produce plain
// @Success 200 {string} v1.4.9
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

type createGroupRequest struct {
	Title                    string   `json:"title" validate:"required"`
	Description              *string  `json:"description"`
	Category                 string   `json:"category" validate:"required"`
	Tags                     []string `json:"tags"`
	Privacy                  string   `json:"privacy" validate:"required,oneof=public private"`
	Hidden                   bool     `json:"hidden_for_search"`
	CreatorName              string   `json:"creator_name"`
	CreatorEmail             string   `json:"creator_email"`
	CreatorPhotoURL          string   `json:"creator_photo_url"`
	ImageURL                 *string  `json:"image_url"`
	WebURL                   *string  `json:"web_url"`
	MembershipQuestions      []string `json:"membership_questions"`
	AuthmanEnabled           bool     `json:"authman_enabled"`
	AuthmanGroup             *string  `json:"authman_group"`
	OnlyAdminsCanCreatePolls bool     `json:"only_admins_can_create_polls" `
	CanJoinAutomatically     bool     `json:"can_join_automatically"`
	AttendanceGroup          bool     `json:"attendance_group" `
} //@name createGroupRequest

type userGroupShortDetail struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Privacy          string `json:"privacy"`
	MembershipStatus string `json:"membership_status"`
}

// CreateGroup creates a group
// @Description Creates a group. The user must be part of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID CreateGroup
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createGroupRequest true "body data"
// @Success 200 {object} createResponse
// @Security AppUserAuth
// @Router /api/groups [post]
func (h *ApisHandler) CreateGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	// NOTE: Permissions should be handled using the admin auth wrap function and the associated authorization policy
	// if !current.IsMemberOfGroup("urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access") {
	// 	log.Printf("%s is not allowed to create a group", current.Email)
	// 	http.Error(w, corebb.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
	// 	return
	// }

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a group - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	var requestData createGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create group data - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	//validate
	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create group data - %s\n", err.Error())
		http.Error(w, utils.NewValidationError(err).JSONErrorString(), http.StatusBadRequest)
		return
	}

	if requestData.AuthmanEnabled && !current.HasPermission("managed_group_admin") {
		log.Printf("Only managed_group_admin could create a managed group")
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	insertedID, groupErr := h.app.Services.CreateGroup(clientID, current, &model.Group{
		Title:                    requestData.Title,
		Description:              requestData.Description,
		Category:                 requestData.Category,
		Tags:                     requestData.Tags,
		Privacy:                  requestData.Privacy,
		HiddenForSearch:          requestData.Hidden,
		ImageURL:                 requestData.ImageURL,
		MembershipQuestions:      requestData.MembershipQuestions,
		AuthmanGroup:             requestData.AuthmanGroup,
		AuthmanEnabled:           requestData.AuthmanEnabled,
		OnlyAdminsCanCreatePolls: requestData.OnlyAdminsCanCreatePolls,
		CanJoinAutomatically:     requestData.CanJoinAutomatically,
		AttendanceGroup:          requestData.AttendanceGroup,
	})
	if groupErr != nil {
		log.Println(groupErr.Error())
		http.Error(w, groupErr.JSONErrorString(), http.StatusBadRequest)
		return
	}

	data, err = json.Marshal(createResponse{InsertedID: *insertedID})
	if err != nil {
		log.Println("Error on marshal create group response")
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type updateGroupRequest struct {
	Title                      string   `json:"title" validate:"required"`
	Description                *string  `json:"description"`
	Category                   string   `json:"category" validate:"required"`
	Tags                       []string `json:"tags"`
	Privacy                    string   `json:"privacy" validate:"required,oneof=public private"`
	Hidden                     bool     `json:"hidden_for_search"`
	ImageURL                   *string  `json:"image_url"`
	WebURL                     *string  `json:"web_url"`
	MembershipQuestions        []string `json:"membership_questions"`
	AuthmanEnabled             bool     `json:"authman_enabled"`
	AuthmanGroup               *string  `json:"authman_group"`
	OnlyAdminsCanCreatePolls   bool     `json:"only_admins_can_create_polls"`
	CanJoinAutomatically       bool     `json:"can_join_automatically"`
	BlockNewMembershipRequests bool     `json:"block_new_membership_requests"`
	AttendanceGroup            bool     `json:"attendance_group" `
} //@name updateGroupRequest

// UpdateGroup updates a group
// @Description Updates a group.
// @ID UpdateGroup
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body updateGroupRequest true "body data"
// @Param id path string true "ID"
// @Success 200 {string} Successfully updated
// @Security AppUserAuth
// @Router /api/groups/{id} [put]
func (h *ApisHandler) UpdateGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Group id is required")
		http.Error(w, utils.NewMissingParamError("Group id is required").JSONErrorString(), http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the update group item - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	var requestData updateGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the update group request data - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating update group data - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroup(clientID, current, id)
	if group.CurrentMember == nil || !group.CurrentMember.IsAdmin() {
		log.Printf("%s is not allowed to update group settings '%s'. Only group admin could update a group", current.Email, group.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}
	if (requestData.AuthmanEnabled || group.AuthmanEnabled) && !current.HasPermission("managed_group_admin") {
		log.Printf("%s is not allowed to update group settings '%s'. Only group admin with managed_group_admin permission could update a managed group", current.Email, group.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	groupErr := h.app.Services.UpdateGroup(clientID, current, &model.Group{
		ID:                       id,
		Title:                    requestData.Title,
		Description:              requestData.Description,
		Category:                 requestData.Category,
		Tags:                     requestData.Tags,
		Privacy:                  requestData.Privacy,
		HiddenForSearch:          requestData.Hidden,
		ImageURL:                 requestData.ImageURL,
		MembershipQuestions:      requestData.MembershipQuestions,
		AuthmanGroup:             requestData.AuthmanGroup,
		AuthmanEnabled:           requestData.AuthmanEnabled,
		OnlyAdminsCanCreatePolls: requestData.OnlyAdminsCanCreatePolls,
		CanJoinAutomatically:     requestData.CanJoinAutomatically,
		AttendanceGroup:          requestData.AttendanceGroup,
	})
	if groupErr != nil {
		log.Printf("Error on updating group - %s\n", err)
		http.Error(w, groupErr.JSONErrorString(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated"))
}

// GetGroupStats Retrieves stats for a group by id
// @Description Retrieves stats for a group by id
// @ID GetGroupStats
// @Tags Client-V1
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.GroupStats
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/stats [get]
func (h *ApisHandler) GetGroupStats(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["id"]
	if len(groupID) <= 0 {
		log.Println("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(clientID, current, groupID)
	if err != nil {
		log.Printf("error getting group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if group == nil {
		log.Printf("error getting group stats - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// DeleteGroup deletes a group
// @Description Deletes a group.
// @ID DeleteGroup
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/group/{id} [delete]
func (h *ApisHandler) DeleteGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroup(clientID, current, id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, utils.NewServerError().JSONErrorString(), http.StatusInternalServerError)
		return
	}
	if group.CurrentMember == nil || !group.CurrentMember.IsAdmin() {
		log.Printf("%s is not allowed to update group settings '%s'. Only group admin could delete group", current.Email, group.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}
	if group.AuthmanEnabled && !current.HasPermission("managed_group_admin") {
		log.Printf("%s is not allowed to update group settings '%s'. Only group admin with managed_group_admin permission could delete a managed group", current.Email, group.Title)
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

// GetGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID GetGroups
// @Tags Client-V1
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
// @Router /api/groups [get]
func (h *ApisHandler) GetGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

type getUserGroupsResponse struct {
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
} // @name getUserGroupsResponse

// GetUserGroups gets the user groups.
// @Description Gives the user groups.
// @ID GetUserGroups
// @Tags Client-V1
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {array} getUserGroupsResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user/groups [get]
func (h *ApisHandler) GetUserGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {

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
		log.Printf("error getting user groups - %s", err.Error())
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
		log.Println("Error on marshal the user groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// LoginUser Logs in the user and refactor the user record and linked data if need
// @Description Logs in the user and refactor the user record and linked data if need
// @ID LoginUser
// @Tags Client-V1
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user/login [get]
func (h *ApisHandler) LoginUser(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.LoginUser(clientID, current)
	if err != nil {
		log.Printf("error login user - %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type getUserStatsResponse struct {
	Count int64 `json:"posts_count" bson:"posts_count"`
} // @name getUserStatsResponse

// GetUserStats Gets user stat information. Responds with {"posts_count": xxx}
// @Description Gets user stat information. Responds with {"posts_count": xxx}
// @ID GetUserStats
// @Tags Client-V1
// @Param APP header string true "APP"
// @Success 200 {object} getUserStatsResponse
// @Security AppUserAuth
// @Router /api/user/stats [get]
func (h *ApisHandler) GetUserStats(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	stats, err := h.app.Services.GetUserPostCount(clientID, current.ID)
	if err != nil {
		log.Printf("error getting user(%s) post count:  %s", current.ID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := getUserStatsResponse{
		Count: 0,
	}
	if stats != nil {
		response.Count = *stats
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("error on marshal user(%s) stats: %s", current.ID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// DeleteUser Deletes a user with all the involved information from the Notifications BB (this includes - group membership & posts (and child posts - no matter of the creator))
// @Description Deletes a user with all the involved information from the Notifications BB (this includes - group membership & posts (and child posts - no matter of the creator))
// @ID DeleteUser
// @Tags Client-V1
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user [delete]
func (h *ApisHandler) DeleteUser(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.DeleteUser(clientID, current)
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
// @Tags Client-V1
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} userGroupShortDetail
// @Security AppUserAuth
// @Router /api/user/group-memberships [get]
func (h *ApisHandler) GetUserGroupMemberships(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	userGroups, err := h.app.Services.GetUserGroups(clientID, current, nil, nil, nil, nil, nil, nil)
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
// @Tags Client-V1
// @Accept json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {object} getGroupResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/groups/{id} [get]
func (h *ApisHandler) GetGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(clientID, current, id)
	if err != nil {
		log.Printf("adminapis.GetGroupV2() error on getting group %s", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
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
	NotificationsPreferences *model.NotificationsPreferences `json:"notification_preferences"`
} // @name createPendingMemberRequest

// CreatePendingMember creates a group pending member
// @Description Creates a group pending member
// @ID CreatePendingMember
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createPendingMemberRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {string} Successfully created
// @Failure 423 {string} block_new_membership_requests flag is true
// @Security AppUserAuth
// @Router /api/group/{group-id}/pending-members [post]
func (h *ApisHandler) CreatePendingMember(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
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

	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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

	err = h.app.Services.CreatePendingMembership(clientID, current, group, member)
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
// @Tags Client-V1
// @Accept plain
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {string} string "Successfuly deleted"
// @Security AppUserAuth
// @Router /api/group/{group-id}/pending-members [delete]
func (h *ApisHandler) DeletePendingMember(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeletePendingMembership(clientID, current, groupID)
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
// @ID CreateMember
// @Tags Client-V1
// @Accept plain
// @Param data body model.MembershipFilter true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.Member
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [get]
func (h *ApisHandler) GetGroupMembers(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	requestData, err := ioutil.ReadAll(r.Body)
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
	members, err := h.app.Services.FindGroupMemberships(clientID, request)
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
// @Tags Client-V1
// @Accept plain
// @Param data body createMemberRequest true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [post]
func (h *ApisHandler) CreateMember(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
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
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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

	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
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

	err = h.app.Services.CreateMembership(clientID, current, group, &member)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// DeleteMember deletes a member membership from a group
// @Description Deletes a member membership from a group
// @ID DeleteMember
// @Tags Client-V1
// @Accept plain
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {string} string "Successfuly deleted"
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [delete]
func (h *ApisHandler) DeleteMember(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteMembership(clientID, current, groupID)
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
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body membershipApprovalRequest true "body data"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully processed
// @Security AppUserAuth
// @Router /api/memberships/{membership-id}/approval [put]
func (h *ApisHandler) MembershipApproval(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
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

	membership, err := h.app.Services.FindGroupMembershipByID(clientID, membershipID)
	if err != nil || membership == nil {
		log.Printf("Membership %s not found - %s\n", membershipID, err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroup(clientID, current, membership.GroupID)
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

	err = h.app.Services.ApplyMembershipApproval(clientID, current, membershipID, approve, rejectedReason)
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
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/memberships/{membership-id} [delete]
func (h *ApisHandler) DeleteMembership(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	membership, err := h.app.Services.FindGroupMembershipByID(clientID, membershipID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if membership == nil {
		log.Printf("Membership %s not found", membershipID)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroup(clientID, current, membership.GroupID)
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

	err = h.app.Services.DeleteMembershipByID(clientID, current, membershipID)
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
	NotificationsPreferences *model.NotificationsPreferences `json:"notification_preferences"`
} // @name updateMembershipRequest

// UpdateMembership updates a membership. Only admin can update the status and date_attended fields of a membership record. Member is allowed to update only his/her notification preferences.
// @Description Updates a membership. Only admin can update the status and date_attended fields of a membership record. Member is allowed to update only his/her notification preferences.
// @ID UpdateMembership
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body updateMembershipRequest true "body data"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully updated
// @Security AppUserAuth
// @Router /api/memberships/{membership-id} [put]
func (h *ApisHandler) UpdateMembership(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
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

	membership, err := h.app.Services.FindGroupMembershipByID(clientID, membershipID)
	if err != nil || membership == nil {
		log.Printf("Membership %s not found - %s\n", membershipID, err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroup(clientID, current, membership.GroupID)
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

	err = h.app.Services.UpdateMembership(clientID, current, membershipID, status, dateAttended, notificationsPreferences)
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
// @Tags Client-V1
// @Accept plain
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{group-id}/authman/synchronize [post]
func (h *ApisHandler) SynchAuthmanGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to update
	isAdmin, err := h.app.Services.IsGroupAdmin(clientID, groupID, current.ID)
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

	err = h.app.Services.SynchronizeAuthmanGroup(clientID, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetGroupEvents gives the group events
// @Description Gives the group events.
// @ID GetGroupEvents
// @Tags Client-V1
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} string
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{group-id}/events [get]
func (h *ApisHandler) GetGroupEvents(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to see the events for this group
	group, hasPermission := h.app.Services.CheckUserGroupMembershipPermission(clientID, current, groupID)
	if group == nil || group.CurrentMember == nil || !hasPermission {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	events, err := h.app.Services.GetEvents(clientID, current, groupID, group.CurrentMember == nil || !group.CurrentMember.IsAdminOrMember())
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

// GetGroupEventsV2 gives the group events V2
// @Description Gives the group events.
// @ID GetGroupEventsV2
// @Tags Client-V1
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.Event
// @Security AppUserAuth
// @Router /api/group/{group-id}/events/v2 [get]
func (h *ApisHandler) GetGroupEventsV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to see the events for this group
	group, hasPermission := h.app.Services.CheckUserGroupMembershipPermission(clientID, current, groupID)
	if group == nil || group.CurrentMember == nil || !hasPermission {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	events, err := h.app.Services.GetEvents(clientID, current, groupID, group.CurrentMember == nil || !group.CurrentMember.IsAdminOrMember())
	if err != nil {
		log.Printf("error getting group events - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove  ToMembersList for non-admins
	if len(events) > 0 && !group.CurrentMember.IsAdmin() {
		for i, event := range events {
			event.ToMembersList = nil
			events[i] = event
		}
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

type groupEventRequest struct {
	EventID       string           `json:"event_id" validate:"required"`
	ToMembersList []model.ToMember `json:"to_members" bson:"to_members"` // nil or empty means everyone; non-empty means visible to those user ids and admins
} // @name groupEventRequest

// CreateGroupEvent creates a group event
// @Description Creates a group event
// @ID CreateGroupEvent
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body groupEventRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {string} Successfully created
// @Security AppUserAuth
// @Router /api/group/{group-id}/events [post]
func (h *ApisHandler) CreateGroupEvent(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData groupEventRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to create
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdmin() {
		log.Printf("%s is not allowed to create event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	_, err = h.app.Services.CreateEvent(clientID, current, requestData.EventID, group, requestData.ToMembersList, &model.Creator{
		UserID: current.ID,
		Name:   current.Name,
		Email:  current.Email,
	})
	if err != nil {
		log.Printf("Error on creating an event - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully created"))
}

// UpdateGroupEvent updates a group event
// @Description Updates a group event
// @ID UpdateGroupEvent
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body groupEventRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {string} Successfully created
// @Security AppUserAuth
// @Router /api/group/{group-id}/events [put]
func (h *ApisHandler) UpdateGroupEvent(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the update group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData groupEventRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the update event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating update event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to create
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdmin() {
		log.Printf("%s is not allowed to create event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.UpdateEvent(clientID, current, requestData.EventID, group.ID, requestData.ToMembersList)
	if err != nil {
		log.Printf("Error on updating a group event - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// DeleteGroupEvent deletes a group event
// @Description Deletes a group event
// @ID DeleteGroupEvent
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Param event-id path string true "Event ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/group/{group-id}/event/{event-id} [delete]
func (h *ApisHandler) DeleteGroupEvent(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdmin() {
		log.Printf("%s is not allowed to delete event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.DeleteEvent(clientID, current, eventID, groupID)
	if err != nil {
		log.Printf("Error on deleting an event - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

// GetGroupPosts gets all posts for the desired group.
// @Description gets all posts for the desired group.
// @ID GetGroupPosts
// @Tags Client-V1
// @Param APP header string true "APP"
// @Success 200 {array} postResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{groupID}/posts [get]
func (h *ApisHandler) GetGroupPosts(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["groupID"]
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

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(clientID, id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided group id - %s", id)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)

	if membership == nil || !membership.IsAdminOrMember() {
		log.Printf("%s is not allowed to get posts for group %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	var filterPrivatePostsValue *bool
	if group == nil || err != nil || membership == nil || !membership.IsAdminOrMember() {
		filter := false
		filterPrivatePostsValue = &filter
	}

	filterByToMembers := true
	if membership != nil && membership.IsAdmin() {
		filterByToMembers = false
	}

	posts, err := h.app.Services.GetPosts(clientID, current, id, filterPrivatePostsValue, filterByToMembers, offset, limit, order)
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

// CreateGroupPost creates a post within the desired group.
// @Description creates a post within the desired group.
// @ID CreateGroupPost
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Success 200 {object} postResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{groupId}/posts [post]
func (h *ApisHandler) CreateGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["groupID"]
	if len(id) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var post *model.Post
	err = json.Unmarshal(data, &post)
	if err != nil {
		log.Printf("error on unmarshal posts for group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if allowed to create
	group, err := h.app.Services.GetGroupEntity(clientID, id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdminOrMember() {
		log.Printf("the user is not member of the group - %s", id)
		// do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	post.GroupID = id // Set group id from the query param

	post, err = h.app.Services.CreatePost(clientID, current, post, group)
	if err != nil {
		log.Printf("error getting posts for group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(post)
	if err != nil {
		log.Printf("error on marshal posts for group - %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type postResponse struct {
	ID       string `json:"id"`
	GroupID  string `json:"group_id"`
	ParentID string `json:"parent_id"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Private  bool   `json:"private"`
}

// GetGroupPost Gets a post within the desired group.
// @Description Gets a post within the desired group.
// @ID GetGroupPost
// @Tags Client-V1
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {object} postResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{groupId}/posts/{postId} [get]
func (h *ApisHandler) GetGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["groupID"]
	if len(groupID) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "group id is required", http.StatusBadRequest)
		return
	}

	postID := params["postID"]
	if len(postID) <= 0 {
		log.Println("postID is required")
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	if group.CurrentMember == nil || !group.CurrentMember.IsAdminOrMember() {
		log.Printf("%s is not allowed to delete event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	post, err := h.app.Services.GetPost(clientID, &current.ID, groupID, postID, true, false)
	if err != nil {
		log.Printf("error getting post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(post)
	if err != nil {
		log.Printf("error on marshal post (%s) - %s", postID, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateGroupPost Updates a post within the desired group.
// @Description Updates a post within the desired group.
// @ID UpdateGroupPost
// @Tags Client-V1
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {object} postResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{groupId}/posts/{postId} [put]
func (h *ApisHandler) UpdateGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["groupID"]
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

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var post *model.Post
	err = json.Unmarshal(data, &post)
	if err != nil {
		log.Printf("error on unmarshal post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if post.ID == nil || postID != *post.ID {
		log.Printf("unexpected post id query param (%s) and post json data", postID)
		http.Error(w, fmt.Sprintf("inconsistent post id query param (%s) and post json data", postID), http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdminOrMember() {
		log.Printf("%s is not allowed to delete event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	post, err = h.app.Services.UpdatePost(clientID, current, post)
	if err != nil {
		log.Printf("error update post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(post)
	if err != nil {
		log.Printf("error on marshal post (%s) - %s", postID, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// reactToGroupPostRequestBody request body for reaction API call
type reactToGroupPostRequestBody struct {
	Reaction string `json:"reaction"`
} // @name reactToGroupPostRequestBody

// ReactToGroupPost Reacts to a post within the desired group.
// @Description Reacts to a post within the desired group.
// @ID ReactToGroupPost
// @Tags Client-V1
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {string} Success
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{groupId}/posts/{postId}/reactions [put]
func (h *ApisHandler) ReactToGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["groupID"]
	if len(groupID) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "group id is required", http.StatusBadRequest)
		return
	}

	postID := params["postID"]
	if len(postID) <= 0 {
		log.Println("postID is required")
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on read reactToGroupPostRequestBody - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var body reactToGroupPostRequestBody
	err = json.Unmarshal(data, &body)
	if err != nil {
		log.Printf("error on unmarshal reactToGroupPostRequestBody (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	membership, err := h.app.Services.FindGroupMembership(clientID, groupID, current.ID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if membership == nil {
		log.Printf("there is no a group for the provided group id - %s", groupID)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !membership.IsAdminOrMember() {
		log.Printf("%s is not allowed to react to posts for group %s", current.Email, groupID)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.ReactToPost(clientID, current, groupID, postID, body.Reaction)
	if err != nil {
		log.Printf("error reacting to post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

// reportAbuseGroupPostRequestBody request body for report abuse API call
type reportAbuseGroupPostRequestBody struct {
	Comment           string `json:"comment"`
	SendToGroupAdmins bool   `json:"send_to_group_admins" bson:"send_to_group_admins"`
	SendToDean        bool   `json:"send_to_dean" bson:"send_to_dean"`
} // @name reportAbuseGroupPostRequestBody

// ReportAbuseGroupPost Reports an abusive group post
// @Description Reports an abusive group post
// @ID ReportAbuseGroupPost
// @Tags Client-V1
// @Accept  json
// @Param APP header string true "APP"
// @Param data body reportAbuseGroupPostRequestBody true "body data"
// @Success 200
// @Security AppUserAuth
// @Router /api/group/{groupId}/posts/{postId}/report/abuse [put]
func (h *ApisHandler) ReportAbuseGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["groupID"]
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

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on read reportAbuseGroupPostRequestBody - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var body *reportAbuseGroupPostRequestBody
	err = json.Unmarshal(data, &body)
	if err != nil {
		log.Printf("error on unmarshal reportAbuseGroupPostRequestBody (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdminOrMember() {
		log.Printf("%s is not allowed to delete event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	post, err := h.app.Services.GetPost(clientID, &current.ID, group.ID, postID, true, false)
	if err != nil {
		log.Printf("error retrieve post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.app.Services.ReportPostAsAbuse(clientID, current, group, post, body.Comment, body.SendToDean, body.SendToGroupAdmins)
	if err != nil {
		log.Printf("error update post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// DeleteGroupPost Updates a post within the desired group.
// @Description Updates a post within the desired group.
// @ID DeleteGroupPost
// @Tags Client-V1
// @Accept  json
// @Param APP header string true "APP"
// @Success 200
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{groupId}/posts/{postId} [delete]
func (h *ApisHandler) DeleteGroupPost(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["groupID"]
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

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntity(clientID, groupID)
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
	membership, _ := h.app.Services.FindGroupMembership(clientID, group.ID, current.ID)
	if membership == nil || !membership.IsAdminOrMember() {
		log.Printf("%s is not allowed to delete event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.DeletePost(clientID, current, groupID, postID, false)
	if err != nil {
		log.Printf("error deleting posts for post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}

// NewAdminApisHandler creates new rest Handler instance
func NewAdminApisHandler(app *core.Application) *AdminApisHandler {
	return &AdminApisHandler{app: app}
}

// NewInternalApisHandler creates new rest Handler instance
func NewInternalApisHandler(app *core.Application) *InternalApisHandler {
	return &InternalApisHandler{app: app}
}
