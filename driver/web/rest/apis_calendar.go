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

	event, groupIDs, err := h.app.Services.CreateCalendarEvent(requestData.Event, requestData.GroupIDs)
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

	event, err := h.app.Services.CreateCalendarEventSingleGroup(requestData.Event, groupID, requestData.ToMembers)
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
