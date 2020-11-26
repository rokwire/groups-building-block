package rest

import (
	"encoding/json"
	"groups/core"
	"groups/core/model"
	"io/ioutil"
	"log"
	"net/http"
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
// @Success 200 {string} v1.1.0
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

//GetGroupCategories gives all group categories
// @Description Gives all group categories.
// @ID GetGroupCategories
// @Accept  json
// @Success 200 {array} string
// @Security APIKeyAuth
// @Router /api/group-categories [get]
func (h *ApisHandler) GetGroupCategories(w http.ResponseWriter, r *http.Request) {
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
} //@name createGroupRequest

//CreateGroup creates a group
// @Description Creates a group. The user must be part of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID CreateGroup
// @Accept json
// @Produce json
// @Param data body createGroupRequest true "body data"
// @Success 200 {object} createResponse
// @Security AppUserAuth
// @Router /api/groups [post]
func (h *ApisHandler) CreateGroup(current *model.User, w http.ResponseWriter, r *http.Request) {
	if !current.IsMemberOfGroup("urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access") {
		log.Printf("%s is not allowed to create a group", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a group - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create group data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//validate
	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create group data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	insertedID, err := h.app.Services.CreateGroup(*current, title, description, category, tags, privacy,
		creatorName, creatorEmail, creatorPhotoURL)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createResponse{InsertedID: *insertedID})
	if err != nil {
		log.Println("Error on marshal create group response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
// @Param data body updateGroupRequest true "body data"
// @Param id path string true "ID"
// @Success 200 {string} Successfully updated
// @Security AppUserAuth
// @Router /api/groups/{id} [put]
func (h *ApisHandler) UpdateGroup(current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
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

	var requestData updateGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the update group request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating update group data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check if allowed to update
	group, err := h.app.Services.GetGroupEntity(id)
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
	}

	category := requestData.Category
	title := requestData.Title
	privacy := requestData.Privacy
	description := requestData.Description
	imageURL := requestData.ImageURL
	webURL := requestData.WebURL
	tags := requestData.Tags
	membershipQuestions := requestData.MembershipQuestions

	err = h.app.Services.UpdateGroup(current, id, category, title, privacy, description, imageURL, webURL, tags, membershipQuestions)
	if err != nil {
		log.Printf("Error on updating group - %s\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated"))
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
// @Param category query string false "Category"
// @Success 200 {array} getGroupsResponse
// @Security APIKeyAuth
// @Security AppUserAuth
// @Router /api/groups [get]
func (h *ApisHandler) GetGroups(current *model.User, w http.ResponseWriter, r *http.Request) {
	var category *string
	catogies, ok := r.URL.Query()["category"]
	if ok && len(catogies[0]) > 0 {
		category = &catogies[0]
	}

	groups, err := h.app.Services.GetGroups(current, category)
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
// @Success 200 {array} getUserGroupsResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/user/groups [get]
func (h *ApisHandler) GetUserGroups(current *model.User, w http.ResponseWriter, r *http.Request) {
	groups, err := h.app.Services.GetUserGroups(current)
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
// @Param id path string true "ID"
// @Success 200 {object} getGroupResponse
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/groups/{id} [get]
func (h *ApisHandler) GetGroup(current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	group, err := h.app.Services.GetGroup(current, id)
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
// @Param data body createMemberRequest true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {string} Successfully created
// @Security AppUserAuth
// @Router /api/group/{group-id}/pending-members [post]
func (h *ApisHandler) CreatePendingMember(current *model.User, w http.ResponseWriter, r *http.Request) {
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

	err = h.app.Services.CreatePendingMember(*current, groupID, name, email, photoURL, mAnswers)
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
// @Param group-id path string true "Group ID"
// @Success 200 {string} string "Successfuly deleted"
// @Security AppUserAuth
// @Router /api/group/{group-id}/pending-members [delete]
func (h *ApisHandler) DeletePendingMember(current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeletePendingMember(*current, groupID)
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
// @Param group-id path string true "Group ID"
// @Success 200 {string} string "Successfuly deleted"
// @Security AppUserAuth
// @Router /api/group/{group-id}/members [delete]
func (h *ApisHandler) DeleteMember(current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["group-id"]
	if len(groupID) <= 0 {
		log.Println("group-id is required")
		http.Error(w, "group-id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteMember(*current, groupID)
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
// @Param data body membershipApprovalRequest true "body data"
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully processed
// @Security AppUserAuth
// @Router /api/memberships/{membership-id}/approval [put]
func (h *ApisHandler) MembershipApproval(current *model.User, w http.ResponseWriter, r *http.Request) {
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
	group, err := h.app.Services.GetGroupEntityByMembership(membershipID)
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
	}

	approve := *requestData.Approve
	rejectedReason := requestData.RejectedReason

	err = h.app.Services.ApplyMembershipApproval(*current, membershipID, approve, rejectedReason)
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
// @Param membership-id path string true "Membership ID"
// @Success 200 {string} Successfully deleted
// @Security AppUserAuth
// @Router /api/memberships/{membership-id} [delete]
func (h *ApisHandler) DeleteMembership(current *model.User, w http.ResponseWriter, r *http.Request) {
	//validate input
	params := mux.Vars(r)
	membershipID := params["membership-id"]
	if len(membershipID) <= 0 {
		log.Println("Membership id is required")
		http.Error(w, "Membership id is required", http.StatusBadRequest)
		return
	}

	//check if allowed to delete
	group, err := h.app.Services.GetGroupEntityByMembership(membershipID)
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
	}

	err = h.app.Services.DeleteMembership(*current, membershipID)
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

//UpdateMembership updates membership
func (h *ApisHandler) UpdateMembership(current *model.User, w http.ResponseWriter, r *http.Request) {
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
	group, err := h.app.Services.GetGroupEntityByMembership(membershipID)
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
	}

	status := requestData.Status

	err = h.app.Services.UpdateMembership(*current, membershipID, status)
	if err != nil {
		log.Printf("Error on updating membership - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated"))
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}
