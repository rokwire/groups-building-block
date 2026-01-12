// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"log"
	"net/http"
	"time"
)

// AnalyticsApisHandler handles the rest Analytics APIs implementation
type AnalyticsApisHandler struct {
	app *core.Application
}

type analyticsGetGroupsResponse struct {
	ID              string           `json:"id"`
	OrgID           string           `json:"org_id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Privacy         string           `json:"privacy"`
	Category        string           `json:"category"`
	HiddenForSearch bool             `json:"hidden_for_search"`
	AuthmanEnabled  bool             `json:"authman_enabled"`
	AuthmanGroup    *string          `json:"authman_group"`
	DateCreated     string           `json:"date_created"`
	DateUpdated     *string          `json:"date_updated"`
	ResearchOpen    bool             `json:"research_open"`
	ResearchGroup   bool             `json:"research_group"`
	Stats           model.GroupStats `json:"stats"`
}

// AnalyticsGetGroups Gets groups
// @Description Gets groups
// @ID AnalyticsGetGroups
// @Tags Analytics
// @Accept json
// @Param start_date query string false "Start date string - RFC3339 encoded"
// @Param end_date query string false "End date string - RFC3339 encoded"
// @Success 200 {array} analyticsGetGroupsResponse
// @Security IntAPIKeyAuth
// @Router /api/analytics/groups [get]
func (h *AnalyticsApisHandler) AnalyticsGetGroups(OrgID string, w http.ResponseWriter, r *http.Request) {
	var startDate *time.Time
	startDateStr, ok := r.URL.Query()["start_date"]
	if ok && len(startDateStr) > 0 && len(startDateStr[0]) > 0 {
		date, err := time.Parse(time.RFC3339, startDateStr[0])
		if err != nil {
			log.Println("unable to parse start_date")
			http.Error(w, "unable to parse start_date", http.StatusInternalServerError)
			return
		}
		startDate = &date
	}

	var endDate *time.Time
	endDateStr, ok := r.URL.Query()["end_date"]
	if ok && len(endDateStr) > 0 && len(endDateStr[0]) > 0 {
		date, err := time.Parse(time.RFC3339, endDateStr[0])
		if err != nil {
			log.Println("unable to parse end_date")
			http.Error(w, "unable to parse end_date", http.StatusInternalServerError)
			return
		}
		endDate = &date
	}

	groups, err := h.app.Services.AnalyticsFindGroups(startDate, endDate)
	if err != nil {
		log.Printf("unable to retrieve posts: %s", err)
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	reponse := make([]analyticsGetGroupsResponse, len(groups))
	for i, group := range groups {
		var dateUpdated *string
		if group.DateUpdated != nil {
			val := group.DateUpdated.Format(time.RFC3339)
			dateUpdated = &val
		}
		category := ""
		if group.GetNewCategory() != nil {
			val := group.GetNewCategory()
			category = *val
		}
		var description string
		if group.Description != nil {
			description = *group.Description
		}
		reponse[i] = analyticsGetGroupsResponse{
			ID:              group.ID,
			OrgID:           group.OrgID,
			Title:           group.Title,
			Privacy:         group.Privacy,
			Description:     description,
			Category:        category,
			HiddenForSearch: group.HiddenForSearch,
			AuthmanEnabled:  group.AuthmanEnabled,
			AuthmanGroup:    group.AuthmanGroup,
			ResearchGroup:   group.ResearchGroup,
			ResearchOpen:    group.ResearchOpen,
			Stats:           group.Stats,
			DateCreated:     group.DateCreated.Format(time.RFC3339),
			DateUpdated:     dateUpdated,
		}
	}

	data, err := json.Marshal(reponse)
	if err != nil {
		log.Println("Error on marshal the user group membership")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type analyticsGetPostsResponse struct {
	ID           string  `json:"id"`
	OrgID        string  `json:"org_id"`
	GroupID      string  `json:"group_id"`
	MemberUserID string  `json:"member_user_id"`
	DateCreated  string  `json:"date_created"`
	DateUpdated  *string `json:"date_updated"`
}

type analyticsGetGroupsMembersResponse struct {
	ID          string  `json:"id"`
	OrgID       string  `json:"org_id"`
	GroupID     string  `json:"group_id"`
	DateCreated string  `json:"date_created"`
	DateUpdated *string `json:"date_updated"`
}

// AnalyticsGetGroupsMembers Gets groups members
// @Description Gets groups members
// @ID AnalyticsGetGroupsMembers
// @Tags Analytics
// @Accept json
// @Param group_id query string false "Group ID"
// @Param start_date query string false "Start date string - RFC3339 encoded"
// @Param end_date query string false "End date string - RFC3339 encoded"
// @Success 200 {array} analyticsGetGroupsMembersResponse
// @Security IntAPIKeyAuth
// @Router /api/analytics/members [get]
func (h *AnalyticsApisHandler) AnalyticsGetGroupsMembers(OrgID string, w http.ResponseWriter, r *http.Request) {
	var groupID *string
	groupIDStr, ok := r.URL.Query()["group_id"]
	if ok && len(groupIDStr) > 0 && len(groupIDStr[0]) > 0 {
		val := groupIDStr[0]
		groupID = &val
	}

	var startDate *time.Time
	startDateStr, ok := r.URL.Query()["start_date"]
	if ok && len(startDateStr) > 0 && len(startDateStr[0]) > 0 {
		date, err := time.Parse(time.RFC3339, startDateStr[0])
		if err != nil {
			log.Println("unable to parse start_date")
			http.Error(w, "unable to parse start_date", http.StatusInternalServerError)
			return
		}
		startDate = &date
	}

	var endDate *time.Time
	endDateStr, ok := r.URL.Query()["end_date"]
	if ok && len(endDateStr) > 0 && len(endDateStr[0]) > 0 {
		date, err := time.Parse(time.RFC3339, endDateStr[0])
		if err != nil {
			log.Println("unable to parse end_date")
			http.Error(w, "unable to parse end_date", http.StatusInternalServerError)
			return
		}
		endDate = &date
	}

	members, err := h.app.Services.AnalyticsFindMembers(groupID, startDate, endDate)
	if err != nil {
		log.Printf("unable to retrieve posts: %s", err)
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	reponse := make([]analyticsGetGroupsMembersResponse, len(members))
	for i, member := range members {
		var dateUpdated *string
		if member.DateUpdated != nil {
			val := member.DateUpdated.Format(time.RFC3339)
			dateUpdated = &val
		}
		reponse[i] = analyticsGetGroupsMembersResponse{
			ID:          member.ID,
			GroupID:     member.GroupID,
			OrgID:       member.OrgID,
			DateCreated: member.DateCreated.Format(time.RFC3339),
			DateUpdated: dateUpdated,
		}
	}

	data, err := json.Marshal(reponse)
	if err != nil {
		log.Println("Error on marshal the user group membership")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
