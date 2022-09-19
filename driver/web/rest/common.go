package rest

import (
	"groups/core"
	"groups/core/model"
	"log"
	"net/http"
	"strconv"
)

func getStringQueryParam(r *http.Request, paramName string) *string {
	params, ok := r.URL.Query()[paramName]
	if ok && len(params[0]) > 0 {
		value := params[0]
		return &value
	}
	return nil
}

func getInt64QueryParam(r *http.Request, paramName string) *int64 {
	params, ok := r.URL.Query()[paramName]
	if ok && len(params[0]) > 0 {
		val, err := strconv.ParseInt(params[0], 0, 64)
		if err == nil {
			return &val
		}
	}
	return nil
}

func hasGroupMembershipPermission(service core.Services, w http.ResponseWriter, current *model.User, clientID string, group *model.Group) bool {
	if group.Privacy == "private" {
		if current == nil || current.IsAnonymous {
			log.Println("hasGroupMembershipPermission() error - Anonymous user cannot see the events for a private group")

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return false
		}
		if !group.UsesGroupMemberships && !group.IsGroupAdminOrMember(current.ID) && group.HiddenForSearch && !group.CanJoinAutomatically { // NB: group detail panel needs it for user not belonging to the group
			log.Printf("hasGroupMembershipPermission() error - %s cannot see %s private group as he/she is not a member or admin", current.Email, group.Title)

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return false
		} else if group.UsesGroupMemberships {
			membership, _ := service.FindGroupMembership(clientID, group.ID, current.ID)
			if membership == nil || (!membership.IsAdminOrMember() && group.HiddenForSearch && !group.CanJoinAutomatically) {
				log.Printf("hasGroupMembershipPermission() error - %s cannot see  %s private group as he/she is not a member or admin", current.Email, group.Title)

				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Forbidden"))
				return false
			}
		}
	}
	return true
}
