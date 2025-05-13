package rest

import (
	"encoding/json"
	"errors"
	"groups/core"
	"groups/core/model"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"gopkg.in/go-playground/validator.v9"
)

// BBSApisHandler handles the rest BBS APIs implementation
type BBSApisHandler struct {
	app *core.Application
}

// getEventUserIDsResponse response
type getEventUserIDsResponse struct {
	UserIDs []string `json:"user_ids"`
} // @name getEventUserIDsResponse

// GetEventUserIDs Gets all related group users linked for the described event id
// @Description  Gets all related group users linked for the described event id
// @ID BBSGetEventUsers
// @Tags BBS
// @Param event_id path string true "Event ID"
// @Success 200 {array} getEventUserIDsResponse
// @Security AppUserAuth
// @Router /groups/{group_id}/updated [put]
func (h *BBSApisHandler) GetEventUserIDs(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	params := mux.Vars(req)
	eventID := params["event_id"]
	if len(eventID) <= 0 {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypePathParam, nil, errors.New("missing event_id"), http.StatusBadRequest, false)
	}

	userIDs, err := h.app.Services.GetEventUserIDs(eventID)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	data, err := json.Marshal(getEventUserIDsResponse{UserIDs: userIDs})
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	return log.HTTPResponseSuccessJSON(data)
}

// GetGroupMemberships Gets all related group memberships status and group title using userID
// @Description  Gets all related group memberships status and group title using userID
// @ID GetGroupMemberships
// @Tags BBS
// @Param user_id path string true "User ID"
// @Success 200 {array} []model.GetGroupMembershipsResponse
// @Security AppUserAuth
// @Router /api/bbs/groups/{user_id}/memberships [get]
func (h *BBSApisHandler) GetGroupMemberships(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	params := mux.Vars(req)
	userID := params["user_id"]
	if len(userID) <= 0 {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypePathParam, nil, errors.New("missing user_id"), http.StatusBadRequest, false)
	}

	groupMembershipsStatusAndGroupsTitle, err := h.app.Services.GetGroupMembershipsStatusAndGroupTitle(userID)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	data, err := json.Marshal(groupMembershipsStatusAndGroupsTitle)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	return log.HTTPResponseSuccessJSON(data)

}

// GetGroupMembershipsByGroupID Gets all related group memberships status and group title using groupID
// @Description  Gets all related group memberships status and group title using groupID
// @ID GetGroupMembershipsByGroupID
// @Tags BBS
// @Param group_id path string true "Group ID"
// @Success 200 {array} []string
// @Security AppUserAuth
// @Router /api/bbs/groups/{group_id}/group-memberships [get]
func (h *BBSApisHandler) GetGroupMembershipsByGroupID(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	params := mux.Vars(req)
	groupID := params["group_id"]
	if len(groupID) <= 0 {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypePathParam, nil, errors.New("missing group_id"), http.StatusBadRequest, false)
	}
	groupMembershipsIDs, err := h.app.Services.GetGroupMembershipsByGroupID(groupID)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	data, err := json.Marshal(groupMembershipsIDs)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	return log.HTTPResponseSuccessJSON(data)

}

// GetGroupsEvents Gets all related eventID and groupID using eventIDs
// @Description  Gets all related eventID and groupID using eventIDs
// @ID GetGroupsEvents
// @Tags BBS
// @Param events-ids query string false "comma separated eventIDs query"
// @Success 200 {array} []model.GetGroupsEvents
// @Security AppUserAuth
// @Router /api/bbs/groups/events [get]
func (h *BBSApisHandler) GetGroupsEvents(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	var eventsIDs []string
	eventsArg := req.URL.Query().Get("events-ids")
	if eventsArg != "" {
		eventsIDs = strings.Split(eventsArg, ",")
	}

	groupevents, err := h.app.Services.GetGroupsEvents(eventsIDs)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}
	data, err := json.Marshal(groupevents)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	return log.HTTPResponseSuccessJSON(data)
}

// GetGroupsByGroupIDs Gets all related groups by groupIDs
// @Description  Gets all related groups by groupIDs
// @ID GetGroupsbyGroupsIDs
// @Tags BBS
// @Param group-ids query string true "comma separated groupIDs query"
// @Success 200 {array} []model.Group
// @Security AppUserAuth
// @Router /api/bbs/groups [get]
func (h *BBSApisHandler) GetGroupsByGroupIDs(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	var groupIDs []string
	groupArg := req.URL.Query().Get("group-ids")
	if len(groupArg) <= 0 {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeQueryParam, nil, errors.New("missing group_ids"), http.StatusBadRequest, false)
	}
	if groupArg != "" {
		groupIDs = strings.Split(groupArg, ",")
	}
	groups, err := h.app.Services.GetGroupsByGroupIDs(groupIDs)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}
	data, err := json.Marshal(groups)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	return log.HTTPResponseSuccessJSON(data)
}

// onGroupUpdatedRequestBody response
type onGroupUpdatedRequestBody struct {
	Operation string `json:"operation" validate:"required,oneof=event_update poll_update social_update"`
} // @name onGroupUpdatedRequestBody

// OnGroupUpdated Recieves a callback notification from other BBS that group related resource has been updated
// @Description Recieves a callback notification from other BBS that group related resource has been updated
// @ID OnGroupUpdated
// @Tags BBS
// @Param group-id path string true "group id"
// @Param group-id body onGroupUpdatedRequestBody true "Event type. Supported values: event_update, poll_update, social_update"
// @Success 200
// @Security AppUserAuth
// @Router /api/bbs/groups [get]
func (h *BBSApisHandler) OnGroupUpdated(log *logs.Log, req *http.Request, user *model.User) logs.HTTPResponse {
	params := mux.Vars(req)
	groupID := params["group_id"]
	if len(groupID) <= 0 {
		return log.HTTPResponseErrorAction(logutils.ActionUpdate, logutils.TypePathParam, nil, errors.New("missing group_id"), http.StatusBadRequest, false)
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionUpdate, logutils.TypePathParam, nil, errors.New("unable to read request body"), http.StatusBadRequest, false)
	}

	var requestData onGroupUpdatedRequestBody
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionUpdate, logutils.TypePathParam, nil, errors.New("unable to unmarshal request body"), http.StatusBadRequest, false)
	}

	//validate
	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionUpdate, logutils.TypePathParam, nil, errors.New("unable to validate request body"), http.StatusBadRequest, false)
	}

	var operationType model.ExternalOperation
	switch requestData.Operation {
	case "event_update":
		operationType = model.ExternalOperationEventUpdate
	case "poll_update":
		operationType = model.ExternalOperationPollUpdate
	case "social_update":
		operationType = model.ExternalOperationSocialUpdate
	default:
		return log.HTTPResponseErrorAction(logutils.ActionUpdate, logutils.TypePathParam, nil, errors.New("unsupported operation"), http.StatusBadRequest, false)
	}

	err = h.app.BBS.OnUpdatedGroupExternalEntity(groupID, operationType)
	if err != nil {
		return log.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeError, nil, err, http.StatusBadRequest, false)
	}

	return log.HTTPResponseSuccess()
}
