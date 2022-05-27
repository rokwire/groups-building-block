package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"groups/core"
	"log"
	"net/http"
	"time"
)

// InternalApisHandler handles the rest Internal APIs implementation
type InternalApisHandler struct {
	app *core.Application
}

// IntGetUserGroupMemberships gets the user groups memberships
// @Description Gives the user groups memberships
// @ID IntGetUserGroupMemberships
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

//SynchronizeAuthman Synchronizes Authman groups memberhip
// @Description Synchronizes Authman groups memberhip
// @ID SynchronizeAuthman
// @Accept json
// @Success 200
// @Security IntAPIKeyAuth
// @Router /int/authman/synchronize [get]
func (h *InternalApisHandler) SynchronizeAuthman(clientID string, w http.ResponseWriter, r *http.Request) {

	err := h.app.Services.SynchronizeAuthman(clientID)
	if err != nil {
		log.Printf("Error during Authman synchronization: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
