package rest

import (
	"encoding/json"
	"errors"
	"groups/core"
	"groups/core/model"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
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
// @Router /api/v2/user/groups [get]
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
