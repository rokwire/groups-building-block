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

package core

import (
	"fmt"
	"groups/core/model"
	"groups/driven/notifications"
	"groups/driven/storage"
	"log"
	"strings"
	"sync"
)

func (app *Application) findAdminGroupsForEvent(clientID string, current *model.User, eventID string) ([]string, error) {
	return app.storage.FindAdminGroupsForEvent(nil, clientID, current, eventID)
}

func (app *Application) updateGroupMappingsForEvent(clientID string, current *model.User, eventID string, groupIDs []string) ([]string, error) {
	return app.storage.UpdateGroupMappingsForEvent(nil, clientID, current, eventID, groupIDs)
}

func (app *Application) findEventUserIDs(eventID string) ([]string, error) {
	return app.storage.FindEventUserIDs(nil, eventID)
}

func (app *Application) findGroupMembershipsStatusAndGroupsTitle(userID string) ([]model.GetGroupMembershipsResponse, error) {
	return app.storage.FindGroupMembershipStatusAndGroupTitle(nil, userID)
}

func (app *Application) findGroupMembershipsByGroupID(groupID string) ([]string, error) {
	return app.storage.FindGroupMembershipByGroupID(nil, groupID)
}
func (app *Application) findGroupsEvents(eventIDs []string) ([]model.GetGroupsEvents, error) {
	return app.storage.FindGroupsEvents(nil, eventIDs)
}

func (app *Application) getUserData(userID string) (*model.UserDataResponse, error) {
	var wg sync.WaitGroup
	var events []model.Event
	var groupMemberships []model.GroupMembership
	var groups []model.Group
	var posts []model.Post
	var eventsErr, membershipsErr, groupsErr, postsErr error

	// Fetch events asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		events, eventsErr = app.storage.GetEventByUserID(userID)
	}()

	// Fetch group memberships asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		groupMemberships, membershipsErr = app.storage.GetGroupMembershipByUserID(userID)
	}()

	// Fetch posts asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		posts, postsErr = app.storage.GetPostsByUserID(userID)
	}()

	// Wait for group memberships to be fetched, then fetch groups
	wg.Add(1)
	go func() {
		defer wg.Done()
		var groupIDs []string
		if groupMemberships != nil {
			for _, membership := range groupMemberships {
				groupIDs = append(groupIDs, membership.GroupID)
			}
		}
		groups, groupsErr = app.storage.FindGroupsByGroupIDs(groupIDs)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors from any of the goroutines
	if eventsErr != nil {
		return nil, eventsErr
	}
	if membershipsErr != nil {
		return nil, membershipsErr
	}
	if groupsErr != nil {
		return nil, groupsErr
	}
	if postsErr != nil {
		return nil, postsErr
	}

	// Prepare the response
	userData := &model.UserDataResponse{
		EventResponse:            events,
		GroupMembershipsResponse: groupMemberships,
		GroupResponse:            groups,
		PostResponse:             posts,
	}

	return userData, nil
}

func (app *Application) findGroupsByGroupIDs(groupIDs []string) ([]model.Group, error) {
	return app.storage.FindGroupsByGroupIDs(groupIDs)
}

func (app *Application) getEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	events, err := app.storage.FindEvents(clientID, current, groupID, filterByToMembers)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (app *Application) getGroupCalendarEvents(clientID string, current *model.User, groupID string, published *bool, filter model.GroupEventFilter) (map[string]interface{}, error) {
	mappings, err := app.storage.FindEvents(clientID, current, groupID, true)
	if err != nil {
		return nil, err
	}

	if len(mappings) > 0 {
		var eventIDs []string
		for _, eventMapping := range mappings {
			eventIDs = append(eventIDs, eventMapping.EventID)
		}
		currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.NetID}
		return app.calendar.GetGroupCalendarEvents(currentAccount, eventIDs, current.AppID, current.OrgID, published, filter)
	}

	return nil, err
}

func (app *Application) createEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	var skipUserID *string

	if current != nil && creator == nil {
		creator = &model.Creator{
			UserID: current.ID,
			Name:   current.Name,
			Email:  current.Email,
		}
	}
	if creator != nil {
		skipUserID = &creator.UserID
	}

	var event *model.Event
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		_, err := app.storage.CreateEvent(context, clientID, eventID, group.ID, toMemberList, creator)
		if err != nil {
			return err
		}

		app.notifyGroupMembersForNewEvent(context, clientID, current, group, nil, skipUserID)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (app *Application) createCalendarEventForGroups(clientID string, adminIdentifiers []model.AccountIdentifiers, current *model.User, event map[string]interface{}, groupIDs []string) (map[string]interface{}, []string, error) {
	var mappedGroupIDs []string
	var createdEvent map[string]interface{}

	app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		memberships, err := app.findGroupMemberships(context, clientID, model.MembershipFilter{
			GroupIDs: groupIDs,
			UserID:   &current.ID,
			Statuses: []string{"admin"},
		})
		if err != nil {
			return err
		}

		if memberships.GetMembershipByAccountID(current.ID) != nil {
			currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.NetID}
			createdEvent, err = app.calendar.CreateCalendarEvent(adminIdentifiers, currentAccount, event, current.OrgID, current.AppID, groupIDs)
			if err != nil {
				return err
			}

			if createdEvent != nil {

				eventID := createdEvent["id"].(string)

				for _, membership := range memberships.Items {
					mapping, err := app.storage.CreateEvent(context, clientID, eventID, membership.GroupID, nil, &model.Creator{
						UserID: current.ID,
						Name:   current.Name,
						Email:  current.Email,
					})
					if err != nil {
						log.Printf("Error create goup mapping: %s", err)
					}
					if mapping != nil {
						mappedGroupIDs = append(mappedGroupIDs, mapping.GroupID)
					}

					group, grErr := app.storage.FindGroup(context, clientID, membership.GroupID, &current.ID)
					if grErr != nil {
						return grErr
					}

					app.notifyGroupMembersForNewEvent(context, clientID, current, group, mapping, &current.ID)
				}
				return nil
			}
		}
		return nil
	})

	return createdEvent, mappedGroupIDs, nil
}

func (app *Application) createCalendarEventForGroupsMembers(clientID string, orgID string, appID string, eventID string, groupIDs []string, members []model.ToMember) error {
	for _, groupID := range groupIDs {

		var userIDs []string
		memberships, err := app.findGroupMemberships(nil, clientID, model.MembershipFilter{
			GroupIDs: []string{groupID},
			Statuses: []string{"admin", "member"},
		})
		if err != nil {
			return err
		}

		for _, membership := range memberships.Items {
			if len(members) > 0 {
				for _, toMember := range members {
					if toMember.UserID == membership.UserID {
						userIDs = append(userIDs, membership.UserID)
						break
					}
				}
			} else {
				userIDs = append(userIDs, membership.UserID)
			}
		}
		err = app.calendar.AddPeopleToCalendarEvent(userIDs, eventID, orgID, appID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) createCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	var createdEvent map[string]interface{}

	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		memberships, err := app.findGroupMemberships(context, clientID, model.MembershipFilter{
			GroupIDs: []string{groupID},
			UserID:   &current.ID,
			Statuses: []string{"admin"},
		})
		if err != nil {
			return err
		}

		var groupIDs []string
		if groupID != "" {
			groupIDs = append(groupIDs, groupID)
		}

		if memberships.GetMembershipByAccountID(current.ID) != nil {
			currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.ExternalID}
			createdEvent, err = app.calendar.CreateCalendarEvent([]model.AccountIdentifiers{}, currentAccount, event, current.OrgID, current.AppID, groupIDs)
			if err != nil {
				return err
			}

			if createdEvent != nil {

				eventID := createdEvent["id"].(string)

				mapping, err := app.storage.CreateEvent(context, clientID, eventID, groupID, members, &model.Creator{
					UserID: current.ID,
					Name:   current.Name,
					Email:  current.Email,
				})
				if err != nil {
					log.Printf("Error create goup mapping: %s", err)
				}

				group, grErr := app.storage.FindGroup(context, clientID, groupID, &current.ID)
				if grErr != nil {
					return grErr
				}

				app.notifyGroupMembersForNewEvent(context, clientID, current, group, mapping, &current.ID)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return createdEvent, members, nil
}

func (app *Application) updateCalendarEventSingleGroup(clientID string, current *model.User, event map[string]interface{}, groupID string, members []model.ToMember) (map[string]interface{}, []model.ToMember, error) {
	memberships, err := app.findGroupMemberships(nil, clientID, model.MembershipFilter{
		GroupIDs: []string{groupID},
		UserID:   &current.ID,
		Statuses: []string{"admin"},
	})
	if err != nil {
		return nil, nil, err
	}

	if len(memberships.Items) > 0 {
		var groupIDs []string
		for _, membership := range memberships.Items {
			groupIDs = append(groupIDs, membership.GroupID)
		}

		eventID := event["id"].(string)
		currentAccount := model.AccountIdentifiers{AccountID: &current.ID, ExternalID: &current.NetID}
		createdEvent, err := app.calendar.UpdateCalendarEvent(currentAccount, eventID, event, current.OrgID, current.AppID)
		if err != nil {
			return nil, nil, err
		}

		if createdEvent != nil {
			var mappedGroupIDs []string
			eventID := createdEvent["id"].(string)

			err := app.storage.UpdateEvent(clientID, eventID, groupID, members)
			if err != nil {
				return nil, nil, err
			}

			for _, groupID := range groupIDs {
				mapping, err := app.storage.CreateEvent(nil, clientID, eventID, groupID, members, &model.Creator{
					UserID: current.ID,
					Name:   current.Name,
					Email:  current.Email,
				})
				if err != nil {
					log.Printf("app.updateCalendarEventSingleGroup() Error create goup mapping: %s", err)
				}
				if mapping != nil {
					mappedGroupIDs = append(mappedGroupIDs, mapping.GroupID)
				}
			}
			return createdEvent, members, nil
		}
	}

	return nil, nil, nil
}

func (app *Application) updateEvent(clientID string, _ *model.User, eventID string, groupID string, toMemberList []model.ToMember) error {
	return app.storage.UpdateEvent(clientID, eventID, groupID, toMemberList)
}

func (app *Application) deleteEvent(clientID string, _ *model.User, eventID string, groupID string) error {
	err := app.storage.DeleteEvent(clientID, eventID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) notifyGroupMembersForNewEvent(context storage.TransactionContext, clientID string, current *model.User, group *model.Group, event *model.Event, skipUserID *string) {
	var userIDs []string
	var recipients []notifications.Recipient
	if len(event.ToMembersList) > 0 {
		userIDs = event.GetMembersAsUserIDs(skipUserID)
	}

	result, err := app.storage.FindGroupMembershipsWithContext(context, clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
		UserIDs:  userIDs,
		Statuses: []string{"member", "admin"},
	})
	if err != nil {
		app.logger.Errorf("notifyGroupMembersForNewEvent() Error finding group memberships: %s", err)
	}

	recipients = result.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
		return member.IsAdminOrMember() && (skipUserID == nil || *skipUserID != member.UserID),
			member.NotificationsPreferences.OverridePreferences &&
				(member.NotificationsPreferences.EventsMuted || member.NotificationsPreferences.AllMute)
	})

	if len(recipients) > 0 {
		topic := "group.events"
		appID := app.config.AppID
		orgID := app.config.OrgID
		if current != nil {
			appID = current.AppID
			orgID = current.OrgID
		}
		groupStr := "Group"
		if group.ResearchGroup {
			groupStr = "Research Project"
		}

		err = app.notifications.SendNotification(
			recipients,
			&topic,
			fmt.Sprintf("%s - %s", groupStr, group.Title),
			fmt.Sprintf("New event has been published in '%s' %s", group.Title, strings.ToLower(groupStr)),
			map[string]string{
				"type":        "group",
				"operation":   "event_created",
				"entity_type": "group",
				"entity_id":   group.ID,
				"entity_name": group.Title,
			},
			appID,
			orgID,
			nil,
		)
		if err != nil {
			app.logger.Errorf("notifyGroupMembersForNewEvent() Error sending notification group memberships: %s", err)
		}
	}
}
