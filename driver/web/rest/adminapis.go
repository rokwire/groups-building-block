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
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

// GetUserGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID AdminGetUserGroups
// @Tags Admin
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

	if groupsFilter.ResearchGroup == nil {
		b := false
		groupsFilter.ResearchGroup = &b
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
// @Tags Admin
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
		log.Printf("adminapis.GetAllGroups() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &groupsFilter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("adminapis.GetAllGroups() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	if groupsFilter.ResearchGroup == nil {
		b := false
		groupsFilter.ResearchGroup = &b
	}

	groups, err := h.app.Services.GetGroups(clientID, nil, groupsFilter)
	if err != nil {
		log.Printf("adminapis.GetAllGroups() error getting groups - %s", err.Error())
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
		log.Printf("adminapis.GetAllGroups() unable to retrieve memberships: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for index, group := range groups {
		group.ApplyLegacyMembership(membershipCollection)
		groups[index] = group
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("adminapis.GetAllGroups() error on marshal the groups items")
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
// @Tags Admin
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

type adminCreateGroupRequest struct {
	Title                    string                         `json:"title" validate:"required"`
	Description              *string                        `json:"description"`
	Category                 string                         `json:"category"`
	Tags                     []string                       `json:"tags"`
	Privacy                  string                         `json:"privacy" validate:"required,oneof=public private"`
	Hidden                   bool                           `json:"hidden_for_search"`
	CreatorName              string                         `json:"creator_name"`
	CreatorEmail             string                         `json:"creator_email"`
	CreatorPhotoURL          string                         `json:"creator_photo_url"`
	ImageURL                 *string                        `json:"image_url"`
	WebURL                   *string                        `json:"web_url"`
	MembershipQuestions      []string                       `json:"membership_questions"`
	AuthmanEnabled           bool                           `json:"authman_enabled"`
	AuthmanGroup             *string                        `json:"authman_group"`
	OnlyAdminsCanCreatePolls bool                           `json:"only_admins_can_create_polls" `
	CanJoinAutomatically     bool                           `json:"can_join_automatically"`
	AttendanceGroup          bool                           `json:"attendance_group" `
	ResearchOpen             bool                           `json:"research_open"`
	ResearchGroup            bool                           `json:"research_group"`
	ResearchConsentStatement string                         `json:"research_consent_statement"`
	ResearchConsentDetails   string                         `json:"research_consent_details"`
	ResearchDescription      string                         `json:"research_description"`
	ResearchProfile          map[string]map[string][]string `json:"research_profile"`
	Settings                 *model.GroupSettings           `json:"settings"`
	Attributes               map[string]interface{}         `json:"attributes"`
	MembersConfig            *model.DefaultMembershipConfig `json:"members,omitempty"`
} //@name adminCreateGroupRequest

// CreateGroup creates a group
// @Description Creates a group. The user must be part of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID AdminCreateGroup
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createGroupRequest true "body data"
// @Success 200 {object} createResponse
// @Security AppUserAuth
// @Router /api/groups [post]
func (h *AdminApisHandler) CreateGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a group - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	var requestData adminCreateGroupRequest
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

	if requestData.ResearchGroup && !current.HasPermission("research_group_admin") {
		log.Printf("'%s' is not allowed to create research group '%s'. Only user with research_group_admin permission can create research group", current.Email, requestData.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	groupData := &model.Group{
		Title:                    requestData.Title,
		Description:              requestData.Description,
		Category:                 requestData.Category,
		Tags:                     requestData.Tags,
		Privacy:                  requestData.Privacy,
		HiddenForSearch:          requestData.Hidden,
		ImageURL:                 requestData.ImageURL,
		WebURL:                   requestData.WebURL,
		MembershipQuestions:      requestData.MembershipQuestions,
		AuthmanGroup:             requestData.AuthmanGroup,
		AuthmanEnabled:           requestData.AuthmanEnabled,
		OnlyAdminsCanCreatePolls: requestData.OnlyAdminsCanCreatePolls,
		CanJoinAutomatically:     requestData.CanJoinAutomatically,
		AttendanceGroup:          requestData.AttendanceGroup,
		ResearchGroup:            requestData.ResearchGroup,
		ResearchOpen:             requestData.ResearchOpen,
		ResearchConsentStatement: requestData.ResearchConsentStatement,
		ResearchConsentDetails:   requestData.ResearchConsentDetails,
		ResearchDescription:      requestData.ResearchDescription,
		ResearchProfile:          requestData.ResearchProfile,
		Settings:                 requestData.Settings,
		Attributes:               requestData.Attributes,
	}

	insertedID, groupErr := h.app.Services.CreateGroup(clientID, current, groupData, requestData.MembersConfig)
	if groupErr != nil {
		log.Println(groupErr.Error())
		http.Error(w, groupErr.JSONErrorString(), http.StatusBadRequest)
		return
	}

	if insertedID != nil {
		data, err = json.Marshal(createResponse{InsertedID: *insertedID})
		if err != nil {
			log.Println("Error on marshal create group response")
			http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateGroup updates a group
// @Description Updates a group.
// @ID AdminUpdateGroup
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body updateGroupRequest true "body data"
// @Param id path string true "ID"
// @Success 200 {string} Successfully updated
// @Security AppUserAuth
// @Router /api/groups/{id} [put]
func (h *AdminApisHandler) UpdateGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Group id is required")
		http.Error(w, utils.NewMissingParamError("Group id is required").JSONErrorString(), http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
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
	if (requestData.ResearchGroup || group.ResearchGroup) && !current.HasPermission("research_group_admin") {
		log.Printf("'%s' is not allowed to update research group '%s'. Only user with research_group_admin permission can update research group", current.Email, group.Title)
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
		WebURL:                   requestData.WebURL,
		MembershipQuestions:      requestData.MembershipQuestions,
		AuthmanGroup:             requestData.AuthmanGroup,
		AuthmanEnabled:           requestData.AuthmanEnabled,
		OnlyAdminsCanCreatePolls: requestData.OnlyAdminsCanCreatePolls,
		CanJoinAutomatically:     requestData.CanJoinAutomatically,
		AttendanceGroup:          requestData.AttendanceGroup,

		ResearchGroup:            requestData.ResearchGroup,
		ResearchOpen:             requestData.ResearchOpen,
		ResearchConsentStatement: requestData.ResearchConsentStatement,
		ResearchConsentDetails:   requestData.ResearchConsentDetails,
		ResearchDescription:      requestData.ResearchDescription,
		ResearchProfile:          requestData.ResearchProfile,
		Settings:                 requestData.Settings,
		Attributes:               requestData.Attributes,
	})
	if groupErr != nil {
		log.Printf("Error on updating group - %s\n", err)
		http.Error(w, groupErr.JSONErrorString(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetGroupMembers Gets the list of group members.
// @Description Gets the list of group members.
// @ID AdminGetGroupMembers
// @Tags Admin
// @Accept plain
// @Param data body model.MembershipFilter true "body data"
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.GroupMembership
// @Security AppUserAuth
// @Router /api/admin/group/{group-id}/members [get]
func (h *AdminApisHandler) GetGroupMembers(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("adminapis.GetGroupMembers() Error: group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.GetGroupMembers() Error on marshal model.MembershipFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var request model.MembershipFilter
	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &request)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("adminapis.GetGroupMembers() Error on unmarshal model.MembershipFilter request body - %s\n", err.Error())
		}
	}

	request.GroupIDs = append(request.GroupIDs, groupID)

	//check if allowed to update
	members, err := h.app.Services.FindGroupMemberships(clientID, request)
	if err != nil {
		log.Printf("adminapis.GetGroupMembers()  error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if members.Items == nil {
		members.Items = []model.GroupMembership{}
	}

	data, err := json.Marshal(members.Items)
	if err != nil {
		log.Printf("adminapis.GetGroupMembers() error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type adminUpdateMembershipRequest struct {
	Status *string `json:"status" validate:"required,oneof=pending member admin rejected"`
} // @name adminUpdateMembershipRequest

// UpdateMembership updates a membership. Only the status can be changed.
// @Description Updates a membership. Only the status can be changed.
// @ID AdminUpdateMembership
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body updateMembershipRequest true "body data"
// @Param membership-id path string true "Membership ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/admin/memberships/{membership-id} [put]
func (h *AdminApisHandler) UpdateMembership(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("adminapis.UpdateMembership() Error on Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.UpdateMembership() Error on marshal the membership update item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData adminUpdateMembershipRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("adminapis.UpdateMembership() Error on unmarshal the membership request update data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("adminapis.UpdateMembership() Error on validating membership update data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	membership, err := h.app.Services.FindGroupMembershipByID(clientID, membershipID)
	if err != nil || membership == nil {
		log.Printf("adminapis.UpdateMembership() Error: Membership %s not found - %s\n", membershipID, err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var status *string
	status = requestData.Status

	err = h.app.Services.UpdateMembership(clientID, current, membershipID, status, nil, nil)
	if err != nil {
		log.Printf("adminapis.UpdateMembership() Error on updating membership - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// DeleteMembership deletes membership
// @Description Deletes a membership
// @ID AdminDeleteMembership
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param membership-id path string true "Membership ID"
// @Success 200
// @Security AppUserAuth
// @Router /api/admin/memberships/{membership-id} [delete]
func (h *AdminApisHandler) DeleteMembership(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("adminapis.DeleteMembership() Error on Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	membership, err := h.app.Services.FindGroupMembershipByID(clientID, membershipID)
	if err != nil {
		log.Printf("adminapis.DeleteMembership() Error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if membership == nil {
		log.Printf("adminapis.DeleteMembership() Error on Membership %s not found", membershipID)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	err = h.app.Services.DeleteMembershipByID(clientID, current, membershipID)
	if err != nil {
		log.Printf("adminapis.DeleteMembership() Error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetGroupEvents gives the group events
// @Description Gives the group events.
// @ID AdminGetGroupEvents
// @Tags Admin
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
// @Tags Admin
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
// @Tags Admin
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
// @Tags Admin
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
	data, err := io.ReadAll(r.Body)
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
	data, err := io.ReadAll(r.Body)
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
	data, err := io.ReadAll(r.Body)
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
