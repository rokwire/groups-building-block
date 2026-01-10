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
	"groups/core/model"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetGroupPosts gets all posts for the desired group.
// @Description gets all posts for the desired group.
// @ID AdminGetGroupPosts
// @Tags Admin
// @Param APP header string true "APP"
// @Param groupID query string true "groupID"
// @Param type query string false "Values: message|post"
// @Param offset query string false "offset"
// @Param limit query integer false "limit"
// @Param order query string false "asc|desc"
// @Success 200 {array} model.Post
// @Security AppUserAuth
// @Router /api/admin/group/{groupID}/posts [get]
func (h *AdminApisHandler) GetGroupPosts(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var filter model.PostsFilter
	id := params["group-id"]
	if len(id) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}
	filter.GroupID = id

	postTypesQuery, ok := r.URL.Query()["type"]
	if ok && len(postTypesQuery) > 0 {
		if postTypesQuery[0] != "message" && postTypesQuery[0] != "post" {
			log.Println("the 'type' query param can be 'message' or 'post'")
			http.Error(w, "the 'type' query param can be 'message' or 'post'", http.StatusBadRequest)
			return
		}
		filter.PostType = &postTypesQuery[0]
	}

	scheduleOnlyQuery, ok := r.URL.Query()["scheduled_only"]
	if ok && len(scheduleOnlyQuery) > 0 {
		if scheduleOnlyQuery[0] != "true" && scheduleOnlyQuery[0] != "false" {
			log.Println("the 'scheduled_only' query param can be 'true', 'false', or missing")
			http.Error(w, "the 'scheduled_only' query param can be 'true', 'false', or missing", http.StatusBadRequest)
			return
		}
		if scheduleOnlyQuery[0] == "true" {
			val := true
			filter.ScheduledOnly = &val
		}
		if scheduleOnlyQuery[0] == "false" {
			val := false
			filter.ScheduledOnly = &val
		}
	}

	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			filter.Offset = &val
		}
	}

	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			filter.Limit = &val
		}
	}

	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		filter.Order = &orders[0]
	}

	posts, err := h.app.Services.GetPosts(OrgID, current, filter, nil, false)
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

// GetGroupPost Gets a post within the desired group.
// @Description Gets a post within the desired group.
// @ID AdminGetGroupPost
// @Tags Admin
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {object} model.Post
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/group/{groupId}/posts/{postId} [get]
func (h *AdminApisHandler) GetGroupPost(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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
	group, err := h.app.Services.GetGroup(OrgID, current, groupID)
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

	post, err := h.app.Services.GetPost(OrgID, &current.ID, groupID, postID, true, true)
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

// CreateGroupPost creates a post within the desired group.
// @Description creates a post within the desired group.
// @ID AdminCreateGroupPost
// @Tags Admin
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Success 200 {object} model.Post
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/group/{groupId}/posts [post]
func (h *AdminApisHandler) CreateGroupPost(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["groupID"]
	if len(id) <= 0 {
		log.Println("groupID is required")
		http.Error(w, "groupID is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
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
	group, err := h.app.Services.GetGroup(OrgID, current, id)
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
	if group.CurrentMember == nil || !group.CurrentMember.IsAdminOrMember() {
		log.Printf("%s is not a member of %s", current.Email, group.Title)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	if group.CurrentMember.IsMember() && group.Settings != nil && !group.Settings.PostPreferences.AllowSendPost {
		log.Printf("posts are not allowed for group '%s'", group.Title)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if group.CurrentMember.IsMember() && post.ParentID != nil && group.Settings != nil && !group.Settings.PostPreferences.CanSendPostReplies {
		log.Printf("replies are not allowed for group '%s'", group.Title)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	post.GroupID = id // Set group id from the query param

	post, err = h.app.Services.CreatePost(OrgID, current, post, group)
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

// UpdateGroupPost Updates a post within the desired group.
// @Description Updates a post within the desired group.
// @ID AdminUpdateGroupPost
// @Tags Admin
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {object} model.Post
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/admin/group/{groupId}/posts/{postId} [put]
func (h *AdminApisHandler) UpdateGroupPost(OrgID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	data, err := io.ReadAll(r.Body)
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

	if post.ID == "" || postID != post.ID {
		log.Printf("unexpected post id query param (%s) and post json data", postID)
		http.Error(w, fmt.Sprintf("inconsistent post id query param (%s) and post json data", postID), http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroup(OrgID, current, groupID)
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
		log.Printf("%s is not a member of %s", current.Email, group.Title)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	post, err = h.app.Services.UpdatePost(OrgID, current, group, post)
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
