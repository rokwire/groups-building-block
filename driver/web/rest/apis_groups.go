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
	"groups/utils"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

// GetGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID GetGroups
// @Tags Client
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
// @Router /api/groups [get]
func (h *ApisHandler) GetGroups(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	_, groups, err := h.app.Services.GetGroups(orgID, current, groupsFilter)
	if err != nil {
		log.Printf("apis.GetGroups() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupIDs := []string{}
	for _, grouop := range groups {
		groupIDs = append(groupIDs, grouop.ID)
	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(orgID, model.MembershipFilter{
		GroupIDs: groupIDs,
	})
	if err != nil {
		log.Printf("apis.GetGroups() unable to retrieve memberships: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for index, group := range groups {
		group.ApplyLegacyMembership(membershipCollection)
		groups[index] = group
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("apis.GetGroups() error on marshal the groups items")
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
// @Tags Client
// @Accept  json
// @Param APP header string true "APP"
// @Param title query string false "Deprecated - instead use request body filter! Filtering by group's title (case-insensitive)"
// @Param category query string false "Deprecated - instead use request body filter! category - filter by category"
// @Param privacy query string false "Deprecated - instead use request body filter! privacy - filter by privacy"
// @Param offset query string false "Deprecated - instead use request body filter! offset - skip number of records"
// @Param limit query string false "Deprecated - instead use request body filter! limit - limit the result"
// @Param include_hidden query string false "Deprecated - instead use request body filter! include_hidden - Includes hidden groups if a search by title is performed. Possible value is true. Default false."
// @Param data body model.GroupsFilter true "body data"
// @Success 200 {array} getUserGroupsResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user/groups [get]
func (h *ApisHandler) GetUserGroups(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {

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

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("apis.GetUserGroups() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &groupsFilter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("apis.GetUserGroups() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	if groupsFilter.ResearchGroup == nil {
		b := false
		groupsFilter.ResearchGroup = &b
	}

	groups, err := h.app.Services.GetUserGroups(orgID, current, groupsFilter)
	if err != nil {
		log.Printf("apis.GetUserGroups() error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupIDs := []string{}
	for _, grouop := range groups {
		groupIDs = append(groupIDs, grouop.ID)
	}

	membershipCollection, err := h.app.Services.FindGroupMemberships(orgID, model.MembershipFilter{
		GroupIDs: groupIDs,
	})
	if err != nil {
		log.Printf("apis.GetUserGroups() unable to retrieve memberships: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for index, group := range groups {
		group.ApplyLegacyMembership(membershipCollection)
		groups[index] = group
	}

	data, err := json.Marshal(groups)
	if err != nil {
		log.Println("apis.GetUserGroups() error on marshal the user groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupsV2 gets groups. It can be filtered by category, title and privacy. V2
// @Description Gives the groups list. It can be filtered by category, title and privacy. V2
// @ID GetGroupsV2
// @Tags Client
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
// @Router /api/v2/groups [get]
// @Router /api/v2/groups [post]
func (h *ApisHandler) GetGroupsV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {

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
		log.Printf("apis.GetGroupsV2() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &groupsFilter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("apis.GetGroupsV2() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	if groupsFilter.ResearchGroup == nil {
		b := false
		groupsFilter.ResearchGroup = &b
	}

	_, groups, err := h.app.Services.GetGroups(clientID, current, groupsFilter)
	if err != nil {
		log.Printf("apis.GetGroupsV2() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []model.Group{}
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

// GetGroupsV3 gets groups. It can be filtered by category, title and privacy. V3
// @Description Gives the groups list. It can be filtered by category, title and privacy. V3
// @ID GetGroupsV3
// @Tags Client
// @Accept json
// @Param data body model.GroupsFilter true "body data"
// @Success 200 {object} getGroupsResponseV3
// @Security AppUserAuth
// @Router /api/v3/groups/load [post]
func (h *ApisHandler) GetGroupsV3(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	var groupsFilter model.GroupsFilter

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("adminapis.GetGroupsV3() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &groupsFilter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("adminapis.GetGroupsV3() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	if groupsFilter.ResearchGroup == nil {
		b := false
		groupsFilter.ResearchGroup = &b
	}

	count, groups, err := h.app.Services.GetGroups(clientID, current, groupsFilter)
	if err != nil {
		log.Printf("adminapis.GetGroupsV3() error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []model.Group{}
	}

	result := getGroupsResponseV3{
		Groups: groups,
		Total:  count,
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Println("adminapis.GetGroupsV3() error on marshal the groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetGroupsFilterStatsV3 Gets groups filter stats
// @Description Gets groups filter stats
// @ID GetGroupsFilterStatsV3
// @Tags Client
// @Accept  json
// @Param APP header string true "APP"
// @Param data body model.StatsFilter true "body data"
// @Success 200
// @Security AppUserAuth
// @Router /api/v3/groups/stats [post]
func (h *ApisHandler) GetGroupsFilterStatsV3(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on read GetGroupsFilterStats - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var filter model.StatsFilter
	err = json.Unmarshal(data, &filter)
	if err != nil {
		log.Printf("error on unmarshal GetGroupsFilterStats() - %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats, err := h.app.Services.GetGroupFilterStats(orgID, current, filter, false)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result map[string]int64
	if stats != nil {
		result = stats.Stats
	} else {
		result = make(map[string]int64)
		log.Println("no stats found for the provided filter")
		http.Error(w, "no stats found for the provided filter", http.StatusNotFound)
		return
	}

	data, err = json.Marshal(result)
	if err != nil {
		log.Printf("error on marshal GetGroupsFilterStats - %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetUserGroupsV2 gets the user groups. It can be filtered by category, title and privacy. V2.
// @Description Gives the user groups. It can be filtered by category, title and privacy. V2.
// @ID GetUserGroupsV2
// @Tags Client
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
// @Router /api/v2/user/groups [get]
// @Router /api/v2/user/groups [post]
func (h *ApisHandler) GetUserGroupsV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	requestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("apis.GetUserGroupsV2() error on marshal model.GroupsFilter request body - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestData) > 0 {
		err = json.Unmarshal(requestData, &groupsFilter)
		if err != nil {
			// just log an error and proceed and assume an empty filter
			log.Printf("apis.GetUserGroupsV2() error on unmarshal model.GroupsFilter request body - %s\n", err.Error())
		}
	}

	if groupsFilter.ResearchGroup == nil {
		b := false
		groupsFilter.ResearchGroup = &b
	}

	groups, err := h.app.Services.GetUserGroups(clientID, current, groupsFilter)
	if err != nil {
		log.Printf("apis.GetUserGroupsV2() error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []model.Group{}
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
// @ID GetGroupV2
// @Tags Client
// @Accept json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {object} model.Group
// @Security AppUserAuth
// @Router /api/v2/groups/{id} [get]
func (h *ApisHandler) GetGroupV2(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("apis.GetGroupV2() id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(clientID, current, id)
	if err != nil {
		log.Printf("apis.GetGroupV2() error on getting group %s - %s", id, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if group == nil {
		log.Printf("apis.GetGroupV2() group %s not found", id)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
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

type createGroupRequest struct {
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
	ResearchProfile          map[string]map[string]any      `json:"research_profile"`
	Settings                 *model.GroupSettings           `json:"settings"`
	Attributes               map[string]interface{}         `json:"attributes"`
	MembersConfig            *model.DefaultMembershipConfig `json:"members,omitempty"`
	Administrative           *bool                          `json:"administrative"`
} //@name createGroupRequest

type userGroupShortDetail struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Privacy          string `json:"privacy"`
	MembershipStatus string `json:"membership_status"`
	ResearchOpen     bool   `json:"research_open"`
	ResearchGroup    bool   `json:"research_group"`
}

// CreateGroup creates a group
// @Description Creates a group. The user must be part of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID CreateGroup
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createGroupRequest true "body data"
// @Success 200 {object} createResponse
// @Security AppUserAuth
// @Router /api/groups [post]
func (h *ApisHandler) CreateGroup(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
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

	if requestData.ResearchGroup && !current.HasPermission("research_group_admin") {
		log.Printf("'%s' is not allowed to create research group '%s'. Only user with research_group_admin permission can create research group", current.Email, requestData.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	insertedID, groupErr := h.app.Services.CreateGroup(orgID, current, &model.Group{
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
		Administrative:           requestData.Administrative,
	}, requestData.MembersConfig)
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

type createGroupV3Request struct {
	Title                    string                    `json:"title" validate:"required"`
	Description              *string                   `json:"description"`
	Category                 string                    `json:"category"`
	Tags                     []string                  `json:"tags"`
	Privacy                  string                    `json:"privacy" validate:"required,oneof=public private"`
	Hidden                   bool                      `json:"hidden_for_search"`
	CreatorName              string                    `json:"creator_name"`
	CreatorEmail             string                    `json:"creator_email"`
	CreatorPhotoURL          string                    `json:"creator_photo_url"`
	ImageURL                 *string                   `json:"image_url"`
	WebURL                   *string                   `json:"web_url"`
	MembershipQuestions      []string                  `json:"membership_questions"`
	AuthmanEnabled           bool                      `json:"authman_enabled"`
	AuthmanGroup             *string                   `json:"authman_group"`
	OnlyAdminsCanCreatePolls bool                      `json:"only_admins_can_create_polls" `
	CanJoinAutomatically     bool                      `json:"can_join_automatically"`
	AttendanceGroup          bool                      `json:"attendance_group" `
	ResearchOpen             bool                      `json:"research_open"`
	ResearchGroup            bool                      `json:"research_group"`
	ResearchConsentStatement string                    `json:"research_consent_statement"`
	ResearchConsentDetails   string                    `json:"research_consent_details"`
	ResearchDescription      string                    `json:"research_description"`
	ResearchProfile          map[string]map[string]any `json:"research_profile"`
	Settings                 *model.GroupSettings      `json:"settings"`
	Attributes               map[string]interface{}    `json:"attributes"`
	MembershipStatuses       model.MembershipStatuses  `json:"members,omitempty"`
	Administrative           *bool                     `json:"administrative"`
} //@name createGroupRequest

// CreateGroupV3 Creates a group
// @Description Creates a group. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID CreateGroupV3
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createGroupV3Request true "body data"
// @Success 200 {object} createResponse
// @Security AppUserAuth
// @Router /api/v3/groups [post]
func (h *ApisHandler) CreateGroupV3(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a group - %s\n", err.Error())
		http.Error(w, utils.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	var requestData createGroupV3Request
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

	if requestData.ResearchGroup && !current.HasPermission("research_group_admin") {
		log.Printf("'%s' is not allowed to create research group '%s'. Only user with research_group_admin permission can create research group", current.Email, requestData.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	insertedID, groupErr := h.app.Services.CreateGroupV3(orgID, current, &model.Group{
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
		Administrative:           requestData.Administrative,
	}, requestData.MembershipStatuses)
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

type updateGroupRequest struct {
	Title                      string                    `json:"title" validate:"required"`
	Description                *string                   `json:"description"`
	Category                   string                    `json:"category"`
	Tags                       []string                  `json:"tags"`
	Privacy                    string                    `json:"privacy" validate:"required,oneof=public private"`
	Hidden                     bool                      `json:"hidden_for_search"`
	ImageURL                   *string                   `json:"image_url"`
	WebURL                     *string                   `json:"web_url"`
	MembershipQuestions        []string                  `json:"membership_questions"`
	AuthmanEnabled             bool                      `json:"authman_enabled"`
	AuthmanGroup               *string                   `json:"authman_group"`
	OnlyAdminsCanCreatePolls   bool                      `json:"only_admins_can_create_polls"`
	CanJoinAutomatically       bool                      `json:"can_join_automatically"`
	BlockNewMembershipRequests bool                      `json:"block_new_membership_requests"`
	AttendanceGroup            bool                      `json:"attendance_group" `
	ResearchOpen               bool                      `json:"research_open"`
	ResearchGroup              bool                      `json:"research_group"`
	ResearchConsentStatement   string                    `json:"research_consent_statement"`
	ResearchConsentDetails     string                    `json:"research_consent_details"`
	ResearchDescription        string                    `json:"research_description"`
	ResearchProfile            map[string]map[string]any `json:"research_profile"`
	Settings                   *model.GroupSettings      `json:"settings"`
	Attributes                 map[string]interface{}    `json:"attributes"`
} //@name updateGroupRequest

// UpdateGroup updates a group
// @Description Updates a group.
// @ID UpdateGroup
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body updateGroupRequest true "body data"
// @Param id path string true "ID"
// @Success 200 {string} Successfully updated
// @Security AppUserAuth
// @Router /api/groups/{id} [put]
func (h *ApisHandler) UpdateGroup(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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
	group, err := h.app.Services.GetGroup(orgID, current, id)
	if group.CurrentMember == nil || !group.CurrentMember.IsAdmin() {
		log.Printf("%s is not allowed to update group settings '%s'. Only group admin could update a group", current.Email, group.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}
	if (requestData.ResearchGroup || group.ResearchGroup) && !current.HasPermission("research_group_admin") {
		log.Printf("'%s' is not allowed to update research group '%s'. Only user with research_group_admin permission can update research group", current.Email, group.Title)
		http.Error(w, utils.NewForbiddenError().JSONErrorString(), http.StatusForbidden)
		return
	}

	groupErr := h.app.Services.UpdateGroup(orgID, current, &model.Group{
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
	w.Write([]byte("Successfully updated"))
}

// GetGroupStats Retrieves stats for a group by id
// @Description Retrieves stats for a group by id
// @ID GetGroupStats
// @Tags Client
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} model.GroupStats
// @Security AppUserAuth
// @Router /api/group/{group-id}/stats [get]
func (h *ApisHandler) GetGroupStats(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["id"]
	if len(groupID) <= 0 {
		log.Println("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(orgID, current, groupID)
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
// @Tags Client
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/group/{id} [delete]
func (h *ApisHandler) DeleteGroup(orgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Group id is required")
		http.Error(w, "Group id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroup(orgID, current, id)
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

	err = h.app.Services.DeleteGroup(orgID, current, id)
	if err != nil {
		log.Printf("Error on deleting group - %s\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}
