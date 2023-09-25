package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
	"groups/core/model"
	"io"
	"log"
	"net/http"
)

// GetGroupCalendarEventsV3 Gets the group calendar events
// @Description Gets the group calendar events
// @ID GetGroupCalendarEventsV3
// @Tags Client-V1
// @Accept json
// @Param APP header string true "APP"
// @Param group-id path string true "Group ID"
// @Success 200 {array} string
// @Security AppUserAuth
// @Security APIKeyAuth
// @Router /api/group/{group-id}/events [get]
func (h *ApisHandler) GetGroupCalendarEventsV3(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
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

	events, err := h.app.Services.GetGroupCalendarEvents(clientID, current, groupID)
	if err != nil {
		log.Printf("error getting group events - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

type createCalendarEventMultiGroupData struct {
	Event    map[string]interface{} `json:"event"`
	GroupIDs []string               `json:"group_ids"`
}

// CreateCalendarEventMultiGroup Create a calendar event and link it to multiple group ids
// @Description Create a calendar event and link it to multiple group ids
// @ID CreateCalendarEventMultiGroup
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param data body createCalendarEventMultiGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} createCalendarEventMultiGroupData
// @Security AppUserAuth
// @Router /api/group/events/v3 [post]
func (h *ApisHandler) CreateCalendarEventMultiGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createCalendarEventMultiGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event, groupIDs, err := h.app.Services.CreateCalendarEventForGroups(clientID, current, requestData.Event, requestData.GroupIDs)
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createCalendarEventMultiGroupData{
		Event:    event,
		GroupIDs: groupIDs,
	})
	if err != nil {
		log.Printf("api.CreateCalendarEventMultiGroup() Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type createCalendarEventSingleGroupData struct {
	Event     map[string]interface{} `json:"event"`
	ToMembers []model.ToMember       `json:"to_members"`
}

// CreateCalendarEventSingleGroup Create a calendar event and link it to a single group id
// @Description Create a calendar event and link it to a single group id
// @ID CreateCalendarEventSingleGroup
// @Tags Client-V1
// @Accept json
// @Produce json
// @Param APP header string true "APP"
// @Param id path string true "group id"
// @Param data body createCalendarEventSingleGroupData true "body data"
// @Param group-id path string true "Group ID"
// @Success 200 {object} createCalendarEventSingleGroupData
// @Security AppUserAuth
// @Router /api/group/{group-id}/events/v3 [post]
func (h *ApisHandler) CreateCalendarEventSingleGroup(clientID string, current *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID := params["id"]
	if len(groupID) <= 0 {
		log.Println("apis.CreateCalendarEventSingleGroup() id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on marshal the create group item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createCalendarEventSingleGroupData
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on unmarshal the create event request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event, err := h.app.Services.CreateCalendarEventSingleGroup(clientID, current, requestData.Event, groupID, requestData.ToMembers)
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on validating create event data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createCalendarEventSingleGroupData{
		Event:     event,
		ToMembers: requestData.ToMembers,
	})
	if err != nil {
		log.Printf("api.CreateCalendarEventSingleGroup() Error on marshaling response data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
