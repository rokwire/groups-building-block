package rest

import (
	"encoding/json"
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
func (h *BBSApisHandler) GetEventUserIDs(user *model.User, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	eventID := params["event_id"]
	if len(eventID) <= 0 {
		log.Println("event_id is required")
		http.Error(w, "eventID is required", http.StatusBadRequest)
		return
	}

	userIDs, err := h.app.Services.GetEventUserIDs(eventID)
	if err != nil {
		log.Printf("bbs.GetEventUserIDs() error: %s", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(getEventUserIDsResponse{UserIDs: userIDs})
	if err != nil {
		log.Printf("bbs.GetEventUserIDs() error: %s", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
