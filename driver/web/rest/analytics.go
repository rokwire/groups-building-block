package rest

import (
	"encoding/json"
	"groups/core"
	"log"
	"net/http"
	"time"
)

// AnalyticsApisHandler handles the rest Analytics APIs implementation
type AnalyticsApisHandler struct {
	app *core.Application
}

type analyticsGetPostsResponse struct {
	ID           string  `json:"id"`
	ClientID     string  `json:"client_id"`
	GroupID      string  `json:"group_id"`
	MemberUserID string  `json:"member_user_id"`
	DateCreated  string  `json:"date_created"`
	DateUpdated  *string `json:"date_updated"`
}

// AnalyticsGetPosts Gets posts
// @Description Gets posts
// @ID AnalyticsGetPosts
// @Tags Analytics
// @Accept json
// @Param start_date query string false "Start date string - RFC3339 encoded"
// @Param end_date query string false "End date string - RFC3339 encoded"
// @Success 200 {array} analyticsGetPostsResponse
// @Security IntAPIKeyAuth
// @Router /api/analytics/posts [get]
func (h *AnalyticsApisHandler) AnalyticsGetPosts(clientID string, w http.ResponseWriter, r *http.Request) {
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

	posts, err := h.app.Services.AnalyticsFindPosts(startDate, endDate)
	if err != nil {
		log.Printf("unable to retrieve posts: %s", err)
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	reponse := make([]analyticsGetPostsResponse, len(posts))
	for i, post := range posts {
		var dateUpdated *string
		if post.DateUpdated != nil {
			val := post.DateUpdated.Format(time.RFC3339)
			dateUpdated = &val
		}
		reponse[i] = analyticsGetPostsResponse{
			ID:           post.ID,
			ClientID:     post.ClientID,
			GroupID:      post.GroupID,
			MemberUserID: post.Creator.UserID,
			DateCreated:  post.DateCreated.Format(time.RFC3339),
			DateUpdated:  dateUpdated,
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
