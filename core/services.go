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
	"groups/driven/rewards"
	"groups/driven/storage"
	"groups/utils"
	"sort"
	"time"

	"github.com/google/uuid"

	"groups/core/model"
	"groups/driven/notifications"
	"log"

	"strings"
)

const defaultConfigSyncTimeout = 60

func (app *Application) applyDataProtection(current *model.User, group model.Group) model.Group {
	//1 apply data protection for "anonymous"
	if current == nil || current.IsAnonymous {
		group.Members = []model.Member{}
	} else {

		member := group.GetMemberByUserID(current.ID)
		if member != nil && (member.IsRejected() || member.IsPendingMember()) {
			group.Members = []model.Member{}
			group.Members = append(group.Members, *member)
		}
	}
	return group
}

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getGroupEntity(clientID string, id string) (*model.Group, error) {
	group, err := app.storage.FindGroup(clientID, id)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) getGroupEntityByMembership(clientID string, membershipID string) (*model.Group, error) {
	group, err := app.storage.FindGroupByMembership(clientID, membershipID)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) getGroupEntityByTitle(clientID string, title string) (*model.Group, error) {
	group, err := app.storage.FindGroupByTitle(clientID, title)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) getGroupStats(clientID string, id string) (*model.GroupStats, error) {
	return app.storage.GetGroupStats(clientID, id)
}

func (app *Application) getGroupCategories() ([]string, error) {
	groupCategories, err := app.storage.ReadAllGroupCategories()
	if err != nil {
		return nil, err
	}
	return groupCategories, nil
}
func (app *Application) getUserGroupMemberships(id string, external bool) ([]*model.Group, *model.User, error) {
	getUserGroupMemberships, user, err := app.storage.FindUserGroupsMemberships(id, external)
	if err != nil {
		return nil, nil, err
	}
	return getUserGroupMemberships, user, nil
}

func (app *Application) createGroup(clientID string, current *model.User, group *model.Group) (*string, *utils.GroupError) {
	insertedID, err := app.storage.CreateGroup(clientID, current, group)
	if err != nil {
		return nil, err
	}

	handleRewardsAsync := func(clientID, userID string) {
		count, grErr := app.storage.FindUserGroupsCount(clientID, current.ID)
		if grErr != nil {
			log.Printf("Error createGroup(): %s", grErr)
		} else {
			if count != nil && *count == 1 {
				app.rewards.CreateUserReward(current.ID, rewards.GroupsUserCreatedFirstGroup, "")
			}
		}
	}
	go handleRewardsAsync(clientID, current.ID)

	return insertedID, nil
}

func (app *Application) updateGroup(clientID string, current *model.User, group *model.Group) *utils.GroupError {

	err := app.storage.UpdateGroupWithoutMembers(clientID, current, group)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) deleteGroup(clientID string, current *model.User, id string) error {
	err := app.storage.DeleteGroup(clientID, id)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getGroups(clientID string, current *model.User, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, category, privacy, title, offset, limit, order)
	if err != nil {
		return nil, err
	}

	visibleGroups := make([]model.Group, 0)
	for _, group := range groups {

		if group.Privacy != "private" ||
			group.IsGroupAdminOrMember(current.ID) ||
			(title != nil && strings.EqualFold(group.Title, *title) && !group.HiddenForSearch) {
			visibleGroups = append(visibleGroups, group)
		}
	}

	//apply data protection
	groupsList := make([]model.Group, len(visibleGroups))
	for i := range visibleGroups {
		groupsList[i] = app.applyDataProtection(current, visibleGroups[i])
	}

	return groupsList, nil
}

func (app *Application) getAllGroups(clientID string) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getUserGroups(clientID string, current *model.User, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]model.Group, error) {
	// find the user groups
	groups, err := app.storage.FindUserGroups(clientID, current.ID, category, privacy, title, offset, limit, order)
	if err != nil {
		return nil, err
	}

	//apply data protection
	groupsList := make([]model.Group, len(groups))
	for i, item := range groups {
		groupsList[i] = app.applyDataProtection(current, item)
	}

	return groupsList, nil
}

func (app *Application) loginUser(clientID string, current *model.User) error {
	return app.storage.LoginUser(clientID, current)
}

func (app *Application) deleteUser(clientID string, current *model.User) error {
	return app.storage.DeleteUser(clientID, current.ID)
}

func (app *Application) getGroup(clientID string, current *model.User, id string) (*model.Group, error) {
	// find the group
	group, err := app.storage.FindGroup(clientID, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	//apply data protection
	res := app.applyDataProtection(current, *group)

	return &res, nil
}

func (app *Application) getGroupMembers(clientID string, _ *model.User, groupID string, filter *model.GroupMembersFilter) ([]model.Member, error) {
	return app.storage.GetGroupMembers(clientID, groupID, filter)
}

func (app *Application) createPendingMember(clientID string, current *model.User, group *model.Group, member *model.Member) error {

	if group.CanJoinAutomatically {
		member.Status = "member"
	} else {
		member.Status = "pending"
	}

	err := app.storage.CreatePendingMember(clientID, current, group, member)
	if err != nil {
		return err
	}

	group, err = app.storage.FindGroup(clientID, group.ID)
	if err == nil && group != nil {
		members := group.Members
		if len(members) > 0 {
			recipients := []notifications.Recipient{}
			for _, member := range members {
				if member.Status == "admin" {
					recipients = append(recipients, notifications.Recipient{
						UserID: member.UserID,
						Name:   member.Name,
					})
				}
			}
			if len(recipients) > 0 {
				topic := "group.invitations"

				message := fmt.Sprintf("New membership request for '%s' group has been submitted", group.Title)
				if group.CanJoinAutomatically {
					message = fmt.Sprintf("%s joined '%s' group", member.GetDisplayName(), group.Title)
				}

				app.notifications.SendNotification(
					recipients,
					&topic,
					fmt.Sprintf("Group - %s", group.Title),
					message,
					map[string]string{
						"type":        "group",
						"operation":   "pending_member",
						"entity_type": "group",
						"entity_id":   group.ID,
						"entity_name": group.Title,
					},
				)
			}
		}
	} else {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}

	if group.CanJoinAutomatically && group.AuthmanEnabled {
		err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, member.ExternalID)
		if err != nil {
			log.Printf("err app.createPendingMember() - error storing member in Authman: %s", err)
		}
	}

	return nil
}

func (app *Application) deletePendingMember(clientID string, current *model.User, groupID string) error {
	err := app.storage.DeletePendingMember(clientID, groupID, current.ID)
	if err != nil {
		return err
	}

	group, err := app.storage.FindGroup(clientID, groupID)
	if err == nil && group != nil {
		if group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.RemoveAuthmanMemberFromGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createPendingMember() - error storing member in Authman: %s", err)
			}
		}
	}

	return nil
}

func (app *Application) createMember(clientID string, current *model.User, group *model.Group, member *model.Member) error {

	if (member.UserID == "" && member.ExternalID != "") ||
		(member.UserID != "" && member.ExternalID == "") {
		if member.ExternalID == "" {
			user, err := app.storage.FindUser(clientID, member.UserID, false)
			if err == nil && user != nil {
				member.ApplyFromUserIfEmpty(user)
			} else {
				log.Printf("error app.createMember() - unable to find user: %s", err)
			}
		}
		if member.UserID == "" {
			user, err := app.storage.FindUser(clientID, member.ExternalID, true)
			if err == nil && user != nil {
				member.ApplyFromUserIfEmpty(user)
			} else {
				log.Printf("error app.createMember() - unable to find user: %s", err)
			}
		}
	}

	err := app.storage.CreateMemberUnchecked(clientID, current, group, member)
	if err != nil {
		return err
	}

	group, err = app.storage.FindGroup(clientID, group.ID)
	if err == nil && group != nil {
		members := group.Members
		if len(members) > 0 {
			recipients := []notifications.Recipient{}
			for _, adminMember := range members {
				if adminMember.Status == "admin" && adminMember.UserID != current.ID {
					recipients = append(recipients, notifications.Recipient{
						UserID: adminMember.UserID,
						Name:   adminMember.Name,
					})
				}
			}

			var message string
			if member.Status == "member" || member.Status == "admin" {
				message = fmt.Sprintf("New member joined '%s' group", group.Title)
			} else {
				message = fmt.Sprintf("New membership request for '%s' group has been submitted", group.Title)
			}

			if len(recipients) > 0 {
				topic := "group.invitations"
				app.notifications.SendNotification(
					recipients,
					&topic,
					fmt.Sprintf("Group - %s", group.Title),
					message,
					map[string]string{
						"type":        "group",
						"operation":   "pending_member",
						"entity_type": "group",
						"entity_id":   group.ID,
						"entity_name": group.Title,
					},
				)
			}
		}

		if group.AuthmanEnabled && group.AuthmanGroup != nil {
			err = app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, member.ExternalID)
			if err != nil {
				return err
			}
		}

	} else {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}
	if err == nil && group != nil {
		if group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createMember() - error storing member in Authman: %s", err)
			}
		}
	}

	return nil
}

func (app *Application) deleteMember(clientID string, current *model.User, groupID string) error {
	err := app.storage.DeleteMember(clientID, groupID, current.ID, false)
	if err != nil {
		return err
	}

	group, err := app.storage.FindGroup(clientID, groupID)
	if err == nil && group != nil {
		if group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.RemoveAuthmanMemberFromGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createPendingMember() - error storing member in Authman: %s", err)
			}
		}
	}

	return nil
}

func (app *Application) applyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	err := app.storage.ApplyMembershipApproval(clientID, membershipID, approve, rejectReason)
	if err != nil {
		return fmt.Errorf("error applying membership approval: %s", err)
	}

	group, err := app.storage.FindGroupByMembership(clientID, membershipID)
	if err == nil && group != nil {
		topic := "group.invitations"
		member := group.GetMemberByID(membershipID)
		if member != nil {
			if approve {
				app.notifications.SendNotification(
					[]notifications.Recipient{
						notifications.Recipient{
							UserID: member.UserID,
							Name:   member.Name,
						},
					},
					&topic,
					fmt.Sprintf("Group - %s", group.Title),
					fmt.Sprintf("Your membership in '%s' group has been approved", group.Title),
					map[string]string{
						"type":        "group",
						"operation":   "membership_approve",
						"entity_type": "group",
						"entity_id":   group.ID,
						"entity_name": group.Title,
					},
				)
			} else {
				app.notifications.SendNotification(
					[]notifications.Recipient{
						notifications.Recipient{
							UserID: member.UserID,
							Name:   member.Name,
						},
					},
					&topic,
					fmt.Sprintf("Group - %s", group.Title),
					fmt.Sprintf("Your membership in '%s' group has been rejected with a reason: %s", group.Title, rejectReason),
					map[string]string{
						"type":        "group",
						"operation":   "membership_reject",
						"entity_type": "group",
						"entity_id":   group.ID,
						"entity_name": group.Title,
					},
				)
			}
		}
	} else {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}

	if err == nil && group != nil {
		member := group.GetMemberByID(membershipID)
		if member != nil && group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.applyMembershipApproval() - error storing member in Authman: %s", err)
			}
		}
	}

	return nil
}

func (app *Application) deleteMembership(clientID string, current *model.User, membershipID string) error {
	err := app.storage.DeleteMembership(clientID, current, membershipID)
	if err != nil {
		return err
	}

	group, err := app.storage.FindGroupByMembership(clientID, membershipID)
	if err == nil && group != nil {
		member := group.GetMemberByID(membershipID)
		if member != nil && group.CanJoinAutomatically && group.AuthmanEnabled {
			err := app.authman.RemoveAuthmanMemberFromGroup(*group.AuthmanGroup, current.ExternalID)
			if err != nil {
				log.Printf("err app.createPendingMember() - error storing member in Authman: %s", err)
			}
		}
	}
	return nil
}

func (app *Application) updateMembership(clientID string, current *model.User, membershipID string, status string, dateAttended *time.Time) error {
	err := app.storage.UpdateMembership(clientID, current, membershipID, status, dateAttended)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getEvents(clientID string, current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	events, err := app.storage.FindEvents(clientID, current, groupID, filterByToMembers)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (app *Application) createEvent(clientID string, current *model.User, eventID string, group *model.Group, toMemberList []model.ToMember) (*model.Event, error) {
	event, err := app.storage.CreateEvent(clientID, current, eventID, group.ID, toMemberList)
	if err != nil {
		return nil, err
	}

	var recipients []notifications.Recipient
	if len(event.ToMembersList) > 0 {
		recipients = event.GetMembersAsNotificationRecipients(&current.ID)
	} else {
		recipients = group.GetMembersAsNotificationRecipients(&current.ID)
	}
	topic := "group.events"
	app.notifications.SendNotification(
		recipients,
		&topic,
		fmt.Sprintf("Group - %s", group.Title),
		fmt.Sprintf("New event has been published in '%s' group", group.Title),
		map[string]string{
			"type":        "group",
			"operation":   "event_created",
			"entity_type": "group",
			"entity_id":   group.ID,
			"entity_name": group.Title,
		},
	)

	return event, nil
}

func (app *Application) createEventWithCreator(clientID string, eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	event, err := app.storage.CreateEventWithCreator(clientID, eventID, group.ID, toMemberList, creator)
	if err != nil {
		return nil, err
	}

	var recipients []notifications.Recipient
	if len(event.ToMembersList) > 0 {
		recipients = event.GetMembersAsNotificationRecipients(&creator.UserID)
	} else {
		recipients = group.GetMembersAsNotificationRecipients(&creator.UserID)
	}
	topic := "group.events"
	app.notifications.SendNotification(
		recipients,
		&topic,
		fmt.Sprintf("Group - %s", group.Title),
		fmt.Sprintf("New event has been published in '%s' group", group.Title),
		map[string]string{
			"type":        "group",
			"operation":   "event_created",
			"entity_type": "group",
			"entity_id":   group.ID,
			"entity_name": group.Title,
		},
	)

	return event, nil
}

func (app *Application) updateEvent(clientID string, current *model.User, eventID string, groupID string, toMemberList []model.ToMember) error {
	return app.storage.UpdateEvent(clientID, current, eventID, groupID, toMemberList)
}

func (app *Application) deleteEvent(clientID string, current *model.User, eventID string, groupID string) error {
	err := app.storage.DeleteEvent(clientID, current, eventID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {
	return app.storage.FindPosts(clientID, current, groupID, filterPrivatePostsValue, filterByToMembers, offset, limit, order)
}

func (app *Application) getPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return app.storage.FindPost(clientID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
}

func (app *Application) getUserPostCount(clientID string, userID string) (*int64, error) {
	return app.storage.GetUserPostCount(clientID, userID)
}

func (app *Application) createPost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {
	post, err := app.storage.CreatePost(clientID, current, post)
	if err != nil {
		return nil, err
	}

	handleRewardsAsync := func(clientID, userID string) {
		count, grErr := app.storage.GetUserPostCount(clientID, current.ID)
		if grErr != nil {
			log.Printf("Error createPost(): %s", grErr)
		} else if count != nil {
			if *count > 1 {
				app.rewards.CreateUserReward(current.ID, rewards.GroupsUserSubmittedPost, "")
			} else if *count == 1 {
				app.rewards.CreateUserReward(current.ID, rewards.GroupsUserSubmittedFirstPost, "")
			}
		}
	}
	go handleRewardsAsync(clientID, current.ID)

	handleNotification := func() {

		recipients, _ := app.getPostNotificationRecipients(clientID, post, &current.ID)

		if len(recipients) == 0 {
			recipients = group.GetMembersAsNotificationRecipients(&current.ID)
		}
		if len(recipients) > 0 {
			title := fmt.Sprintf("Group - %s", group.Title)
			body := fmt.Sprintf("New post has been published in '%s' group", group.Title)
			if post.UseAsNotification {
				title = post.Subject
				body = post.Body
			}

			topic := "group.posts"
			app.notifications.SendNotification(
				recipients,
				&topic,
				title,
				body,
				map[string]string{
					"type":         "group",
					"operation":    "post_created",
					"entity_type":  "group",
					"entity_id":    group.ID,
					"entity_name":  group.Title,
					"post_id":      *post.ID,
					"post_subject": post.Subject,
					"post_body":    post.Body,
				},
			)
		}
	}
	go handleNotification()

	return post, nil
}

func (app *Application) getPostNotificationRecipients(clientID string, post *model.Post, skipUserID *string) ([]notifications.Recipient, error) {
	if post == nil {
		return nil, nil
	}

	if len(post.ToMembersList) > 0 {
		return post.GetMembersAsNotificationRecipients(skipUserID), nil
	}

	var err error
	for {
		if post.ParentID == nil {
			break
		}

		post, err = app.storage.FindPost(clientID, nil, post.GroupID, *post.ParentID, true, false)
		if err != nil {
			log.Printf("error app.getPostToMemberList() - %s", err)
			return nil, fmt.Errorf("error app.getPostToMemberList() - %s", err)
		}

		if post != nil && len(post.ToMembersList) > 0 {
			return post.GetMembersAsNotificationRecipients(skipUserID), nil
		}
	}

	return nil, nil
}

func (app *Application) updatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error) {
	return app.storage.UpdatePost(clientID, current.ID, post)
}

func (app *Application) reportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {

	if !sendToDean && !sendToGroupAdmins {
		sendToDean = true
	}

	var creatorExternalID string
	creator, err := app.storage.FindUser(clientID, post.Creator.UserID, false)
	if err != nil {
		log.Printf("error retrieving user: %s", err)
	} else if creator != nil {
		creatorExternalID = creator.ExternalID
	}

	err = app.storage.ReportPostAsAbuse(clientID, current.ID, group, post)
	if err != nil {
		log.Printf("error while reporting an abuse post: %s", err)
		return fmt.Errorf("error while reporting an abuse post: %s", err)
	}

	subject := ""
	if sendToDean && !sendToGroupAdmins {
		subject = "Report violation of Student Code to Dean of Students"
	} else if !sendToDean && sendToGroupAdmins {
		subject = "Report obscene, threatening, or harassing content to Group Administrators"
	} else {
		subject = "Report violation of Student Code to Dean of Students and obscene, threatening, or harassing content to Group Administrators"
	}

	subject = fmt.Sprintf("%s %s", subject, post.DateCreated.Format(time.RFC850))

	if sendToDean {
		body := fmt.Sprintf(`
<div>Violation by: %s %s\n</div>
<div>Group title: %s\n</div>
<div>Post Title: %s\n</div>
<div>Post Body: %s\n</div>
<div>Reported by: %s %s\n</div>
<div>Reported comment: %s\n</div>
	`, creatorExternalID, post.Creator.Name, group.Title, post.Subject, post.Body,
			current.ExternalID, current.Name, comment)
		body = strings.ReplaceAll(body, `\n`, "\n")
		app.notifications.SendMail(app.config.ReportAbuseRecipientEmail, subject, body)
	}
	if sendToGroupAdmins {
		toMembers := group.GetAllAdminsAsRecipients()

		body := fmt.Sprintf(`
Violation by: %s %s
Group title: %s
Post Title: %s
Post Body: %s
Reported by: %s %s
Reported comment: %s
	`, creatorExternalID, post.Creator.Name, group.Title, post.Subject, post.Body,
			current.ExternalID, current.Name, comment)

		app.notifications.SendNotification(toMembers, nil, subject, body, map[string]string{
			"type":         "group",
			"operation":    "report_abuse_post",
			"entity_type":  "group",
			"entity_id":    group.ID,
			"entity_name":  group.Title,
			"post_id":      *post.ID,
			"post_subject": post.Subject,
			"post_body":    post.Body,
		})
	}

	return nil
}

func (app *Application) deletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	return app.storage.DeletePost(clientID, userID, groupID, postID, force)
}

// TODO this logic needs to be refactored because it's over complicated!
func (app *Application) synchronizeAuthman(clientID string, checkThreshold bool) error {
	startTime := time.Now()
	transaction := func(context storage.TransactionContext) error {
		times, err := app.storage.FindSyncTimes(context, clientID)
		if err != nil {
			return err
		}
		if times != nil && times.StartTime != nil {
			config, err := app.storage.FindSyncConfig(clientID)
			if err != nil {
				log.Printf("error finding sync configs for clientID %s: %v", clientID, err)
			}
			timeout := defaultConfigSyncTimeout
			if config != nil && config.Timeout > 0 {
				timeout = config.Timeout
			}
			if times.EndTime == nil {
				if !startTime.After(times.StartTime.Add(time.Minute * time.Duration(timeout))) {
					log.Println("Another Authman sync process is running for clientID " + clientID)
					return fmt.Errorf("another Authman sync process is running" + clientID)
				}
				log.Printf("Authman sync past timeout threshold %d mins\n", timeout)
			}
			if checkThreshold {
				if config == nil {
					log.Printf("missing sync configs for clientID %s", clientID)
					return fmt.Errorf("missing sync configs for clientID %s: %v", clientID, err)
				}
				if !startTime.After(times.StartTime.Add(time.Minute * time.Duration(config.TimeThreshold))) {
					log.Println("Authman has already been synced for clientID " + clientID)
					return fmt.Errorf("Authman has already been synced for clientID %s", clientID)
				}
			}
		}

		return app.storage.SaveSyncTimes(context, model.SyncTimes{StartTime: &startTime, EndTime: nil, ClientID: clientID})
	}

	err := app.storage.PerformTransaction(transaction)
	if err != nil {
		return err
	}

	log.Printf("Global Authman synchronization started for clientID: %s\n", clientID)

	app.authmanSyncInProgress = true
	finishAuthmanSync := func() {
		endTime := time.Now()
		err := app.storage.SaveSyncTimes(nil, model.SyncTimes{StartTime: &startTime, EndTime: &endTime, ClientID: clientID})
		if err != nil {
			log.Printf("Error saving sync configs to end sync: %s\n", err)
			return
		}
		log.Printf("Global Authman synchronization finished for clientID: %s\n", clientID)
	}
	defer finishAuthmanSync()

	configs, err := app.storage.FindManagedGroupConfigs(clientID)
	if err != nil {
		return fmt.Errorf("error finding managed group configs for clientID %s", clientID)
	}

	for _, config := range configs {
		for _, stemName := range config.AuthmanStems {
			stemGroups, err := app.authman.RetrieveAuthmanStemGroups(stemName)
			if err != nil {
				return fmt.Errorf("error on requesting Authman for stem groups: %s", err)
			}

			if stemGroups != nil && len(stemGroups.WsFindGroupsResults.GroupResults) > 0 {
				for _, stemGroup := range stemGroups.WsFindGroupsResults.GroupResults {
					storedStemGroup, err := app.storage.FindAuthmanGroupByKey(clientID, stemGroup.Name)
					if err != nil {
						return fmt.Errorf("error on requesting Authman for stem groups: %s", err)
					}

					title, adminUINs := stemGroup.GetGroupPrettyTitleAndAdmins()

					defaultAdminsMapping := map[string]bool{}
					for _, externalID := range adminUINs {
						defaultAdminsMapping[externalID] = true
					}
					for _, externalID := range app.config.AuthmanAdminUINList {
						defaultAdminsMapping[externalID] = true
					}
					for _, externalID := range config.AdminUINs {
						defaultAdminsMapping[externalID] = true
					}

					constructedAdminUINs := []string{}
					if len(defaultAdminsMapping) > 0 {
						for key := range defaultAdminsMapping {
							constructedAdminUINs = append(constructedAdminUINs, key)
						}
					}

					if storedStemGroup == nil {
						var members []model.Member
						if len(constructedAdminUINs) > 0 {
							members = app.buildMembersByExternalIDs(clientID, constructedAdminUINs, "admin")
						}

						emptyText := ""
						_, err := app.storage.CreateGroup(clientID, nil, &model.Group{
							Title:                title,
							Description:          &emptyText,
							Category:             "Academic", // Hardcoded.
							Privacy:              "private",
							HiddenForSearch:      true,
							CanJoinAutomatically: true,
							AuthmanEnabled:       true,
							AuthmanGroup:         &stemGroup.Name,
							Members:              members,
						})
						if err != nil {
							return fmt.Errorf("error on create Authman stem group: '%s' - %s", stemGroup.Name, err)
						}

						log.Printf("Created new `%s` group", title)
					} else {
						missedUINs := []string{}
						groupUpdated := false
						for _, uin := range adminUINs {
							found := false
							for index, member := range storedStemGroup.Members {
								if member.ExternalID == uin {
									if member.Status != "admin" {
										now := time.Now()
										storedStemGroup.Members[index].Status = "admin"
										storedStemGroup.Members[index].DateUpdated = &now
										groupUpdated = true
										break
									}
									found = true
								}
							}
							if !found {
								missedUINs = append(missedUINs, uin)
							}
						}

						if len(missedUINs) > 0 {
							missedMembers := app.buildMembersByExternalIDs(clientID, missedUINs, "admin")
							if len(missedMembers) > 0 {
								storedStemGroup.Members = append(storedStemGroup.Members, missedMembers...)
								groupUpdated = true
							}
						}

						if storedStemGroup.Title != title {
							storedStemGroup.Title = title
							groupUpdated = true
						}

						if storedStemGroup.Category == "" {
							storedStemGroup.Category = "Academic" // Hardcoded.
							groupUpdated = true
						}

						if groupUpdated {
							err := app.storage.UpdateGroupWithMembers(clientID, nil, storedStemGroup)
							if err != nil {
								fmt.Errorf("error app.synchronizeAuthmanGroup() - unable to update group admins of '%s' - %s", storedStemGroup.Title, err)
							}
						}
					}
				}
			}
		}
	}

	authmanGroups, err := app.storage.FindAuthmanGroups(clientID)
	if err != nil {
		return err
	}

	if len(authmanGroups) > 0 {
		for _, authmanGroup := range authmanGroups {
			err := app.synchronizeAuthmanGroup(clientID, &authmanGroup)
			if err != nil {
				fmt.Errorf("error app.synchronizeAuthmanGroup() '%s' - %s", authmanGroup.Title, err)
			}
		}
	}

	return nil
}

func (app *Application) buildMembersByExternalIDs(clientID string, externalIDs []string, memberStatus string) []model.Member {
	if len(externalIDs) > 0 {
		users, _ := app.storage.FindUsers(clientID, externalIDs, true)
		members := []model.Member{}
		userExternalIDmapping := map[string]model.User{}
		for _, user := range users {
			userExternalIDmapping[user.ExternalID] = user
		}

		for _, externalID := range externalIDs {
			if value, ok := userExternalIDmapping[externalID]; ok {
				members = append(members, model.Member{
					ID:          uuid.NewString(),
					UserID:      value.ID,
					ExternalID:  externalID,
					Name:        value.Name,
					Email:       value.Email,
					Status:      memberStatus,
					DateCreated: time.Now(),
				})
			} else {
				members = append(members, model.Member{
					ID:          uuid.NewString(),
					ExternalID:  externalID,
					Status:      memberStatus,
					DateCreated: time.Now(),
				})
			}
		}
		return members
	}
	return nil
}

// TODO this logic needs to be refactored because it's over complicated!
func (app *Application) synchronizeAuthmanGroup(clientID string, authmanGroup *model.Group) error {
	if !authmanGroup.IsAuthmanSyncEligible() {
		log.Printf("Authman synchronization failed for group '%s' due to bad settings", authmanGroup.Title)
		return nil // Pass only Authman enabled groups
	}

	log.Printf("Authman synchronization for group %s started", authmanGroup.Title)
	defer log.Printf("Authman synchronization for group %s finished", authmanGroup.Title)

	defaultAdminsMapping := map[string]bool{}
	admins := authmanGroup.GetAllAdminMembers()
	if len(admins) > 0 {
		for _, admin := range admins {
			defaultAdminsMapping[admin.ExternalID] = true
		}
	}

	now := time.Now().UTC()

	if authmanGroup.AuthmanGroup != nil {
		authmanExternalIDs, authmanErr := app.authman.RetrieveAuthmanGroupMembers(*authmanGroup.AuthmanGroup)
		if authmanErr != nil {
			return fmt.Errorf("error on requesting Authman for %s: %s", *authmanGroup.AuthmanGroup, authmanErr)
		}
		externalIDMapping := map[string]model.Member{}
		for _, member := range authmanGroup.Members {
			if _, ok := externalIDMapping[member.ExternalID]; !ok {
				externalIDMapping[member.ExternalID] = member
			}
		}

		localUsersMapping := map[string]model.User{}
		localUsers, userErr := app.storage.FindUsers(clientID, authmanExternalIDs, true)
		if authmanErr != nil {
			return fmt.Errorf("error on getting users(%+v) for Authman %s: %s", authmanExternalIDs, *authmanGroup.AuthmanGroup, userErr)
		} else if len(localUsers) > 0 {
			for _, user := range localUsers {
				localUsersMapping[user.ExternalID] = user
			}
		}

		members := []model.Member{}
		userIDMapping := map[string]interface{}{}
		missingInfoExternalIDs := []string{}
		for _, externalID := range authmanExternalIDs {
			if mappedMember, ok := externalIDMapping[externalID]; ok {
				members = append(members, mappedMember)
				if mappedMember.UserID != "" {
					userIDMapping[mappedMember.UserID] = true
				}
				if mappedMember.Name == "" || mappedMember.Email == "" {
					missingInfoExternalIDs = append(missingInfoExternalIDs, externalID)
				}
				continue //SH: This was changed from "break" to fix missing members. This flow should still be optimized // TBD: Why? This flow looks complicated and needs to be revised and redesign.
			}

			if user, ok := localUsersMapping[externalID]; ok {
				// Add missed members
				member := authmanGroup.GetMemberByUserID(user.ID)
				if member != nil {
					if member.IsPendingMember() || member.IsRejected() || member.IsMember() {
						member.Status = "member"
						member.DateUpdated = &now
						members = append(members, *member)
						log.Printf("User(%s, %s, %s) is set as member '%s'", user.ID, user.ExternalID, user.Email, authmanGroup.Title)
					} else if member.IsAdmin() {
						members = append(members, *member)
					} else {
						log.Printf("User(%s, %s, %s) is already a member or admin of '%s'", user.ID, user.ExternalID, user.Email, authmanGroup.Title)
					}
				} else {
					members = append(members, model.Member{
						ID:            uuid.NewString(),
						UserID:        user.ID,
						Status:        "member",
						ExternalID:    externalID,
						Name:          user.Name,
						Email:         user.Email,
						MemberAnswers: authmanGroup.CreateMembershipEmptyAnswers(),
						DateCreated:   now,
						DateUpdated:   &now,
					})
					log.Printf("User(%s, %s, %s) has been created as regular member of '%s'", externalID, user.Name, user.Email, authmanGroup.Title)
				}
				userIDMapping[user.ID] = true
			} else {
				members = append(members, model.Member{
					ID:            uuid.NewString(),
					Status:        "member",
					ExternalID:    externalID,
					MemberAnswers: authmanGroup.CreateMembershipEmptyAnswers(),
					DateCreated:   now,
					DateUpdated:   &now,
				})
				missingInfoExternalIDs = append(missingInfoExternalIDs, externalID)
				log.Printf("Empty User(ExternalID: %s) has been created as regular member of '%s'", externalID, authmanGroup.Title)
			}
		}

		// Fetch user info for the required users
		if len(missingInfoExternalIDs) > 0 {
			userMapping, err := app.authman.RetrieveAuthmanUsers(missingInfoExternalIDs)
			if err != nil {
				log.Printf("error on retriving missing user info for(%+v): %s", missingInfoExternalIDs, err)
			} else if len(userMapping) > 0 {
				for i, member := range members {
					updatedInfo := false
					if mappedUser, ok := userMapping[member.ExternalID]; ok {
						if member.Name == "" && mappedUser.Name != "" {
							member.Name = mappedUser.Name
							updatedInfo = true
						}
						if member.Email == "" && len(mappedUser.AttributeValues) > 0 {
							member.Email = mappedUser.AttributeValues[0]
							updatedInfo = true
						}
						if !updatedInfo {
							log.Printf("The user has missing info: %+v Group: '%s' Authman Group: '%s'", mappedUser, authmanGroup.Title, *authmanGroup.AuthmanGroup)
						}
					}
					if updatedInfo {
						members[i] = member
					}
				}
			}
		}

		newExternalMembersMapping := map[string]interface{}{}
		for _, member := range members {
			newExternalMembersMapping[member.ExternalID] = true
		}

		// Add remaining admins
		if len(authmanGroup.Members) > 0 {
			for _, member := range authmanGroup.Members {
				val := userIDMapping[member.UserID]
				if val == nil && member.IsAdmin() {
					found := false
					for i, innerMember := range members {
						if member.ExternalID == innerMember.ExternalID {
							innerMember.Status = "admin"
							innerMember.DateUpdated = &now
							members[i] = innerMember
							found = true
							log.Printf("set user(%s, %s, %s) to 'admin' in '%s'", innerMember.UserID, innerMember.Name, innerMember.Email, authmanGroup.Title)
							break
						}
					}
					if !found {
						members = append(members, member)
						log.Printf("add remaining admin user(%s, %s, %s, %s) to '%s'", member.UserID, member.ExternalID, member.Name, member.Email, authmanGroup.Title)
					}
				} else if _, ok := newExternalMembersMapping[member.ExternalID]; !ok {
					log.Printf("User(%s, %s) will be removed as a member of '%s', because it's not defined in Authman group", member.ExternalID, member.Name, authmanGroup.Title)
				}
			}
		}

		// Setup default admins - (check user exists, member exists or not exists and cover all possible scenarios) - separately
		membersMapping := map[string]bool{}
		for _, member := range members {
			membersMapping[member.ExternalID] = true
		}
		if len(defaultAdminsMapping) > 0 {
			for key := range defaultAdminsMapping {
				if _, ok := membersMapping[key]; ok {
					for index, innerMember := range members {
						if innerMember.ExternalID == key {
							innerMember.Status = "admin"
							innerMember.DateUpdated = &now
							members[index] = innerMember
						}
					}
				} else {
					user, err := app.storage.FindUser(clientID, key, true)
					if err != nil {
						return fmt.Errorf("error on retrieving  authman admin user(%s) for group(%s, %s): %s", key, authmanGroup.ID, authmanGroup.Title, err)
					}
					if user != nil {
						members = append(members, model.Member{
							ID:            uuid.NewString(),
							UserID:        user.ID,
							Status:        "admin",
							ExternalID:    user.ExternalID,
							Name:          user.Name,
							Email:         user.Email,
							MemberAnswers: authmanGroup.CreateMembershipEmptyAnswers(),
							DateCreated:   now,
							DateUpdated:   &now,
						})
					} else {
						members = append(members, model.Member{
							ID:            uuid.NewString(),
							Status:        "admin",
							ExternalID:    key,
							MemberAnswers: authmanGroup.CreateMembershipEmptyAnswers(),
							DateCreated:   now,
							DateUpdated:   &now,
						})
					}
				}
			}
		}

		// Sort
		if len(members) > 1 {
			sort.SliceStable(members, func(i, j int) bool {
				if members[j].Status != "admin" && members[i].Status == "admin" {
					return true
				}
				return false
			})
		}

		err := app.storage.UpdateGroupMembers(clientID, authmanGroup.ID, members)
		if err != nil {
			return fmt.Errorf("error on updating authman group(%s, %s): %s", authmanGroup.ID, authmanGroup.Title, err)
		}
	}

	return nil
}

func (app *Application) sendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string) {
	app.notifications.SendNotification(recipients, topic, title, text, data)
}

func (app *Application) getManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfig, error) {
	return app.storage.FindManagedGroupConfigs(clientID)
}

func (app *Application) createManagedGroupConfig(config model.ManagedGroupConfig) (*model.ManagedGroupConfig, error) {
	config.ID = uuid.NewString()
	config.DateCreated = time.Now()
	config.DateUpdated = nil
	err := app.storage.InsertManagedGroupConfig(config)
	return &config, err
}

func (app *Application) updateManagedGroupConfig(config model.ManagedGroupConfig) error {
	return app.storage.UpdateManagedGroupConfig(config)
}

func (app *Application) deleteManagedGroupConfig(id string, clientID string) error {
	return app.storage.DeleteManagedGroupConfig(id, clientID)
}

func (app *Application) getSyncConfig(clientID string) (*model.SyncConfig, error) {
	return app.storage.FindSyncConfig(clientID)
}

func (app *Application) updateSyncConfig(config model.SyncConfig) error {
	return app.storage.SaveSyncConfig(nil, config)
}
