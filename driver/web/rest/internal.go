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
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/gorilla/mux"
)

// InternalApisHandler handles the rest Internal APIs implementation
type InternalApisHandler struct {
	app *core.Application
}

// IntGetUserGroupMemberships gets the user groups memberships
// @Description Gives the user groups memberships
// @ID IntGetUserGroupMemberships
// @Tags Internal
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} userGroupMembership
// @Security IntAPIKeyAuth
// @Router /api/int/user/{identifier}/groups [get]
func (h *InternalApisHandler) IntGetUserGroupMemberships(clientID string, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	identifier := params["identifier"]
	if len(identifier) <= 0 {
		log.Println("Identifier is required")
		http.Error(w, "identifier is required", http.StatusBadRequest)
		return
	}
	externalID := identifier

	userGroupMemberships, user, err := h.app.Services.GetUserGroupMembershipsByExternalID(externalID)
	if err != nil {
		log.Println("The user has no group memberships")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userGroups := make([]userGroupMembership, len(userGroupMemberships))
	for i, group := range userGroupMemberships {

		memberStatus := ""

		members := group.Members
		for _, member := range members {
			if member.UserID == user.ID {
				memberStatus = member.Status
			}
		}

		ugm := userGroupMembership{
			ID:               group.ID,
			Title:            group.Title,
			Privacy:          group.Privacy,
			MembershipStatus: memberStatus,
		}

		userGroups[i] = ugm
	}

	data, err := json.Marshal(userGroups)
	if err != nil {
		log.Println("Error on marshal the user group membership")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// IntGetGroup Retrieves group details and members
// @Description Retrieves group details and members
// @ID IntGetGroup
// @Tags Internal
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} model.Group
// @Security IntAPIKeyAuth
// @Router /api/int/group/{identifier} [get]
func (h *InternalApisHandler) IntGetGroup(clientID string, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	identifier := params["identifier"]
	if len(identifier) <= 0 {
		log.Println("Identifier is required")
		http.Error(w, "identifier is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroupEntity(clientID, identifier)
	if err != nil {
		log.Printf("Unable to retrieve group with ID '%s': %s", identifier, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(group)
	if err != nil {
		log.Printf("Error on marshal the user group: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// IntGetGroupMembersByGroupTitle Retrieves group members by  title
// @Description Retrieves group members by  title
// @ID IntGetGroupMembersByGroupTitle
// @Tags Internal
// @Accept json
// @Param identifier path string true "Title"
// @Param offset query string false "Offsetting result"
// @Param limit query string false "Limiting the result"
// @Success 200 {object} model.Group
// @Security IntAPIKeyAuth
// @Router /api/int/group/title/{title}/members [get]
func (h *InternalApisHandler) IntGetGroupMembersByGroupTitle(clientID string, w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortMembers := []model.ShortMemberRecord{}
	if group != nil && len(group.Members) > 0 {
		for i, member := range group.Members {
			if offset == nil || (offset != nil && limit != nil && i >= int(*offset)) {
				shortMembers = append(shortMembers, member.ToShortMemberRecord())
			}
			if limit != nil && len(shortMembers) >= int(*limit) {
				break
			}
		}
	}

	data, err := json.Marshal(shortMembers)
	if err != nil {
		log.Printf("Error on marshal the short member list: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// synchronizeAuthmanRequestBody authman sync body struct
type synchronizeAuthmanRequestBody struct {
	GroupAutoCreateStemNames []string `json:"group_auto_create_stem_names"`
} // @name synchronizeAuthmanRequestBody

//SynchronizeAuthman Synchronizes Authman groups membership
// @Description Synchronizes Authman groups membership
// @ID SynchronizeAuthman
// @Tags Internal
// @Accept json
// @Success 200
// @Security IntAPIKeyAuth
// @Router /int/authman/synchronize [post]
func (h *InternalApisHandler) SynchronizeAuthman(clientID string, w http.ResponseWriter, r *http.Request) {
	err := h.app.Services.SynchronizeAuthman(clientID)
	if err != nil {
		log.Printf("Error during Authman synchronization: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GroupsStats wrapper for group stats API
type GroupsStats struct {
	GroupsCount int         `json:"groups_count"`
	GroupsList  []GroupStat `json:"groups_list"`
} // @name GroupsStats

// GroupStat wrapper for single group stat
type GroupStat struct {
	Title              string `json:"title"`
	Privacy            string `json:"privacy"`
	AuthmanEnabled     bool   `json:"authman_enabled"`
	MembersCount       int    `json:"members_count"`
	MembersAddedLast24 int    `json:"members_added_last_24"`
} // @name GroupStat

// GroupStats Retrieve group stats
// @Description Retrieve group stats
// @ID IntGroupStats
// @Tags Internal
// @Accept json
// @Success 200 {object} GroupsStats
// @Security IntAPIKeyAuth
// @Router /int/stats [get]
func (h *InternalApisHandler) GroupStats(clientID string, w http.ResponseWriter, r *http.Request) {

	groups, err := h.app.Services.GetAllGroups(clientID)
	if err != nil {
		log.Printf("Error GroupStats(%s): %s", clientID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	groupsCount := len(groups)
	groupStatList := []GroupStat{}
	if groupsCount > 0 {
		for _, group := range groups {
			addedLast24Count := 0
			for _, member := range group.Members {
				if time.Now().Unix()-member.DateCreated.Unix() < 24*60*60 {
					addedLast24Count++
				}
			}

			groupStatList = append(groupStatList, GroupStat{
				Title:              group.Title,
				Privacy:            group.Privacy,
				AuthmanEnabled:     group.AuthmanEnabled,
				MembersCount:       len(group.Members),
				MembersAddedLast24: addedLast24Count,
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type intCreateGroupEventRequestBody struct {
	EventID       string           `json:"event_id" bson:"event_id" validate:"required"`
	Creator       *model.Creator   `json:"creator" bson:"creator" validate:"required"`
	ToMembersList []model.ToMember `json:"to_members" bson:"to_members"` // nil or empty means everyone; non-empty means visible to those user ids and admins
} // @name intCreateGroupEventRequestBody

// CreateGroupEvent creates a group event
// @Description Creates a group event
// @ID IntCreateGroupEvent
// @Tags Internal
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body intCreateGroupEventRequestBody true "body data"
// @Param group-id path string true "Group ID"
// @Success 200
// @Security IntAPIKeyAuth
// @Router /api/int/group/{group-id}/events [post]
func (h *InternalApisHandler) CreateGroupEvent(clientID string, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create group event - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData intCreateGroupEventRequestBody
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

	grEvent, err := h.app.Services.CreateEventWithCreator(clientID, requestData.EventID, group, requestData.ToMembersList, requestData.Creator)
	if err != nil {
		log.Printf("Error on creating an event - %s\n", err)
		http.Error(w, fmt.Sprintf("Error on creating an event - %s\n", err), http.StatusInternalServerError)
		return
	}

	responseData, err := json.Marshal(grEvent)
	if err != nil {
		log.Printf("Error on marshaling an event - %s\n", err)
		http.Error(w, fmt.Sprintf("Error on marshaling an event - %s\n", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}
