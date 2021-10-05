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

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

//Version gives the service version
// @Description Gives the service version.
// @ID Version
// @Produce plain
// @Success 200 {string} v1.4.9
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

//GetGroupCategories gives all group categories
// @Description Gives all group categories.
// @ID GetGroupCategories
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {array} string
// @Security APIKeyAuth
// @Router /api/group-categories [get]
func (h *ApisHandler) GetGroupCategories(clientID string, w http.ResponseWriter, r *http.Request) {
	groupCategories, err := h.app.Services.GetGroupCategories()
	if err != nil {
		log.Println("Error on getting the group categories")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(groupCategories) == 0 {
		groupCategories = make([]string, 0)
	}

	data, err := json.Marshal(groupCategories)
	if err != nil {
		log.Println("Error on marshal the group categories")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type createGroupRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     *string  `json:"description"`
	Category        string   `json:"category" validate:"required"`
	Tags            []string `json:"tags"`
	Privacy         string   `json:"privacy" validate:"required,oneof=public private"`
	CreatorName     string   `json:"creator_name"`
	CreatorEmail    string   `json:"creator_email"`
	CreatorPhotoURL string   `json:"creator_photo_url"`
	ImageURL        *string  `json:"image_url"`
	WebURL          *string  `json:"web_url"`
} //@name createGroupRequest

//GetUserGroupMemberships gets the user groups memberships
// @Description Gives the user groups memberships
// @ID GetUserGroupMemberships
// @Accept json
// @Param identifier path string true "Identifier"
// @Success 200 {object} userGroupMembership
// @Security IntAPIKeyAuth
// @Router /api/int/user/{identifier}/groups [get]
func (h *ApisHandler) GetUserGroupMemberships(clientID string, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	identifier := params["identifier"]
	if len(identifier) <= 0 {
		log.Println("Identifier is required")
		http.Error(w, "identifier is required", http.StatusBadRequest)
		return
	}
	externalID := identifier

	userGroupMemberships, user, err := h.app.Services.GetUserGroupMemberships(externalID)
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
			if member.User.ID == user.ID {
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

type userGroupMembership struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Privacy          string `json:"privacy"`
	MembershipStatus string `json:"membership_status"`
}

//CreateGroup creates a group
// @Description Creates a group. The user must be part of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID CreateGroup
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

	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte("Forbidden"))
	// 	return
	// }

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a group - %s\n", err.Error())
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	var requestData createGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create group data - %s\n", err.Error())
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	//validate
	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create group data - %s\n", err.Error())
		http.Error(w, core.NewValidationError(err).JSONErrorString(), http.StatusBadRequest)
		return
	}

	title := requestData.Title
	description := requestData.Description
	category := requestData.Category
	tags := requestData.Tags
	privacy := requestData.Privacy
	creatorName := requestData.CreatorName
	creatorEmail := requestData.CreatorEmail
	creatorPhotoURL := requestData.CreatorPhotoURL
	imageURL := requestData.ImageURL
	webURL := requestData.WebURL

	insertedID, groupErr := h.app.Services.CreateGroup(clientID, *current, title, description, category, tags, privacy,
		creatorName, creatorEmail, creatorPhotoURL, imageURL, webURL)
	if groupErr != nil {
		log.Println(groupErr.Error())
		http.Error(w, groupErr.JSONErrorString(), http.StatusBadRequest)
		return
	}

	data, err = json.Marshal(createResponse{InsertedID: *insertedID})
	if err != nil {
		log.Println("Error on marshal create group response")
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type updateGroupRequest struct {
	Category            string   `json:"category" validate:"required"`
	Title               string   `json:"title" validate:"required"`
	Privacy             string   `json:"privacy" validate:"required,oneof=public private"`
	Description         *string  `json:"description"`
	ImageURL            *string  `json:"image_url"`
	WebURL              *string  `json:"web_url"`
	Tags                []string `json:"tags"`
	MembershipQuestions []string `json:"membership_questions"`
} //@name updateGroupRequest

//UpdateGroup updates a group
// @Description Updates a group.
// @ID UpdateGroup
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
		http.Error(w, core.NewMissingParamError("Group id is required").JSONErrorString(), http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the update group item - %s\n", err.Error())
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	var requestData updateGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the update group request data - %s\n", err.Error())
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating update group data - %s\n", err.Error())
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroupEntity(clientID, id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, core.NewBadJSONError().JSONErrorString(), http.StatusBadRequest)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided id - %s", id)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, core.NewNotFoundError().JSONErrorString(), http.StatusBadRequest)
		return
	}
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to update a group", current.Email)
		http.Error(w, core.NewForbiddenError().JSONErrorString(), http.StatusBadRequest)
		return
	}

	category := requestData.Category
	title := requestData.Title
	privacy := requestData.Privacy
	description := requestData.Description
	imageURL := requestData.ImageURL
	webURL := requestData.WebURL
	tags := requestData.Tags
	membershipQuestions := requestData.MembershipQuestions

	groupErr := h.app.Services.UpdateGroup(clientID, current, id, category, title, privacy, description, imageURL, webURL, tags, membershipQuestions)
	if groupErr != nil {
		log.Printf("Error on updating group - %s\n", err)
		http.Error(w, groupErr.JSONErrorString(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated"))
}

//DeleteGroup deletes a group
// @Description Deletes a group.
// @ID DeleteGroup
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

	//check if allowed to update
	group, err := h.app.Services.GetGroupEntity(clientID, id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		log.Printf("there is no a group for the provided id - %s", id)
		//do not say to much to the user as we do not know if he/she is an admin for the group yet
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to update a group", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
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

type getGroupsResponse struct {
	ID                  string   `json:"id"`
	Category            string   `json:"category"`
	Title               string   `json:"title"`
	Privacy             string   `json:"privacy"`
	Description         *string  `json:"description"`
	ImageURL            *string  `json:"image_url"`
	WebURL              *string  `json:"web_url"`
	MembersCount        int      `json:"members_count"`
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
} // @name getGroupsResponse

//GetGroups gets groups. It can be filtered by category
// @Description Gives the groups list. It can be filtered by category
// @ID GetGroups
// @Accept  json
// @Param APP header string true "APP"
// @Param category query string false "Category"
// @Param title query string false "Filtering by group's title (case-insensitive)"
// @Success 200 {array} getGroupsResponse
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/groups [get]
func (h *ApisHandler) GetGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	groups, err := h.app.Services.GetGroups(clientID, current, category, privacy, title, offset, limit, order)
	if err != nil {
		log.Printf("error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	MembersCount        int      `json:"members_count"`
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

//GetUserGroups gets the user groups.
// @Description Gives the user groups.
// @ID GetUserGroups
// @Accept  json
// @Param APP header string true "APP"
// @Success 200 {array} getUserGroupsResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user/groups [get]
func (h *ApisHandler) GetUserGroups(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	groups, err := h.app.Services.GetUserGroups(clientID, current)
	if err != nil {
		log.Printf("error getting user groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

type getGroupResponse struct {
	ID                  string   `json:"id"`
	Category            string   `json:"category"`
	Title               string   `json:"title"`
	Privacy             string   `json:"privacy"`
	Description         *string  `json:"description"`
	ImageURL            *string  `json:"image_url"`
	WebURL              *string  `json:"web_url"`
	MembersCount        int      `json:"members_count"`
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

//GetGroup gets a group
// @Description Gives a group
// @ID GetGroup
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
		log.Printf("error getting a group - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

type createMemberRequest struct {
	Name          string `json:"name"`
	Email         string `json:"email" validate:"required"`
	PhotoURL      string `json:"photo_url"`
	MemberAnswers []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"member_answers"`
} // @name createMemberRequest

//CreatePendingMember creates a group pending member
// @Description Creates a group pending member
// @ID CreatePendingMember
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createMemberRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {string} Successfully created
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

	var requestData createMemberRequest
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

	name := requestData.Name
	email := requestData.Email
	photoURL := requestData.PhotoURL
	memberAnswers := requestData.MemberAnswers
	mAnswers := make([]model.MemberAnswer, len(memberAnswers))
	if memberAnswers != nil {
		for i, current := range memberAnswers {
			mAnswers[i] = model.MemberAnswer{Question: current.Question, Answer: current.Answer}
		}
	}

	err = h.app.Services.CreatePendingMember(clientID, *current, groupID, name, email, photoURL, mAnswers)
	if err != nil {
		log.Printf("Error on creating a pending member - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully created"))
}

//DeletePendingMember deletes a group pending member
// @Description Deletes a group pending member
// @ID DeletePendingMember
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

	err := h.app.Services.DeletePendingMember(clientID, *current, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

//DeleteMember deletes a member membership from a group
// @Description Deletes a member membership from a group
// @ID DeleteMember
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

	err := h.app.Services.DeleteMember(clientID, *current, groupID)
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

//MembershipApproval approve/deny a membership
// @Description Аpprove/Deny a membership
// @ID MembershipApproval
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

	//check if allowed to update
	group, err := h.app.Services.GetGroupEntityByMembership(clientID, membershipID)
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
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to make approval", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	approve := *requestData.Approve
	rejectedReason := requestData.RejectedReason

	err = h.app.Services.ApplyMembershipApproval(clientID, *current, membershipID, approve, rejectedReason)
	if err != nil {
		log.Printf("Error on applying membership approval - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully processed"))
}

//DeleteMembership deletes membership
// @Description Deletes a membership
// @ID DeleteMembership
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

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntityByMembership(clientID, membershipID)
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
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to delete membership", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.DeleteMembership(clientID, *current, membershipID)
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
	Status string `json:"status" validate:"required,oneof=member admin"`
} // @name updateMembershipRequest

//UpdateMembership updates a membership
// @Description Updates a membership
// @ID UpdateMembership
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

	//check if allowed to update
	group, err := h.app.Services.GetGroupEntityByMembership(clientID, membershipID)
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
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to make update", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	status := requestData.Status

	err = h.app.Services.UpdateMembership(clientID, *current, membershipID, status)
	if err != nil {
		log.Printf("Error on updating membership - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated"))
}

//GetGroupEvents gives the group events
// @Description Gives the group events.
// @ID GetGroupEvents
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
	if group.Privacy == "private" {
		if current == nil {
			log.Println("Anonymous user cannot see the events for a private group")

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}
		if !group.IsGroupAdminOrMember(current.ID) {
			log.Printf("%s cannot see the events for the %s private group as he/she is not a member or admin", current.Email, group.Title)

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}
	}

	events, err := h.app.Services.GetEvents(clientID, groupID)
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

type createGroupEventRequest struct {
	EventID string `json:"event_id" validate:"required"`
} // @name createGroupEventRequest

//CreateGroupEvent creates a group event
// @Description Creates a group event
// @ID CreateGroupEvent
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createGroupEventRequest true "body data"
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

	var requestData createGroupEventRequest
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
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to create event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	eventID := requestData.EventID

	err = h.app.Services.CreateEvent(clientID, *current, eventID, groupID)
	if err != nil {
		log.Printf("Error on creating an event - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully created"))
}

//DeleteGroupEvent deletes a group event
// @Description Deletes a group event
// @ID DeleteGroupEvent
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
	if !group.IsGroupAdmin(current.ID) {
		log.Printf("%s is not allowed to delete event for %s", current.Email, group.Title)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	err = h.app.Services.DeleteEvent(clientID, *current, eventID, groupID)
	if err != nil {
		log.Printf("Error on deleting an event - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted"))
}

//GetGroupPosts gets all posts for the desired group.
// @Description gets all posts for the desired group.
// @ID GetGroupPosts
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

	posts, err := h.app.Services.GetPosts(clientID, current, id, offset, limit, order)
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

//CreateGroupPost creates a post within the desired group.
// @Description creates a post within the desired group.
// @ID CreateGroupPost
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

	post.GroupID = id // Set group id from the query param

	post, err = h.app.Services.CreatePost(clientID, current, post)
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

//UpdateGroupPost Updates a post within the desired group.
// @Description Updates a post within the desired group.
// @ID UpdateGroupPost
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

//DeleteGroupPost Updates a post within the desired group.
// @Description Updates a post within the desired group.
// @ID DeleteGroupPost
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

	err := h.app.Services.DeletePost(clientID, current, groupID, postID)
	if err != nil {
		log.Printf("error deleting posts for post (%s) - %s", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}

//NewAdminApisHandler creates new rest Handler instance
func NewAdminApisHandler(app *core.Application) *AdminApisHandler {
	return &AdminApisHandler{app: app}
}
