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
	"errors"
	"fmt"
	"groups/driven/rewards"
	"groups/driven/storage"
	"groups/utils"
	"time"

	"github.com/google/uuid"

	"groups/core/model"
	"groups/driven/notifications"
	"log"

	"strings"
)

const (
	defaultConfigSyncTimeout   = 60
	maxEmbeddedMemberGroupSize = 10000
	authmanUserBatchSize       = 5000
)

/*
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
}*/

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getGroupEntity(appID string, orgID string, id string) (*model.Group, error) {
	group, err := app.storage.FindGroupWithContext(nil, appID, orgID, id, nil)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) getGroupEntityByTitle(appID string, orgID string, title string) (*model.Group, error) {
	group, err := app.storage.FindGroupByTitle(appID, orgID, title)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (app *Application) isGroupAdmin(appID string, orgID string, groupID string, userID string) (bool, error) {
	membership, err := app.storage.FindGroupMembershipWithContext(nil, appID, orgID, groupID, userID)
	if err != nil {
		return false, err
	}
	if membership == nil || membership.Status != "admin" {
		return false, nil
	}

	return true, nil
}

func (app *Application) createGroup(current *model.User, group *model.Group) (*string, *utils.GroupError) {
	insertedID, err := app.storage.CreateGroup(current, group, nil)
	if err != nil {
		return nil, err
	}

	if group.ResearchGroup {
		searchParams := app.formatCoreAccountSearchParams(group.ResearchProfile)
		//TODO: verify this verbage
		app.notifications.SendNotification(nil, nil, "A new research project is available", fmt.Sprintf("%s by %s", group.Title, current.Name),
			map[string]string{
				"type":        "group",
				"operation":   "research_group",
				"entity_type": "group",
				"entity_id":   group.ID,
				"entity_name": group.Title,
			},
			searchParams,
			current.AppID,
			current.OrgID)

	}

	handleRewardsAsync := func(appID, orgID, userID string) {
		count, grErr := app.storage.FindUserGroupsCount(appID, orgID, current.ID)
		if grErr != nil {
			log.Printf("Error createGroup(): %s", grErr)
		} else {
			if count != nil && *count == 1 {
				app.rewards.CreateUserReward(current.ID, rewards.GroupsUserCreatedFirstGroup, "")
			}
		}
	}
	go handleRewardsAsync(current.AppID, current.OrgID, current.ID)

	return insertedID, nil
}

func (app *Application) updateGroup(userID *string, group *model.Group) *utils.GroupError {

	err := app.storage.UpdateGroup(userID, group, nil)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) updateGroupDateUpdated(appID string, orgID string, groupID string) error {
	err := app.storage.UpdateGroupDateUpdated(appID, orgID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) deleteGroup(appID string, orgID string, id string) error {
	err := app.storage.DeleteGroup(appID, orgID, id)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getGroups(userID *string, filter model.GroupsFilter) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(userID, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getAllGroups(appID string, orgID string) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(nil, model.GroupsFilter{AppID: appID, OrgID: orgID})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getUserGroups(userID string, filter model.GroupsFilter) ([]model.Group, error) {
	// find the user groups
	groups, err := app.storage.FindUserGroups(userID, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) deleteUser(current *model.User) error {
	return app.storage.PerformTransaction(func(sessionContext storage.TransactionContext) error {
		posts, err := app.storage.FindAllUserPosts(sessionContext, current.AppID, current.OrgID, current.ID)
		if err != nil {
			log.Printf("error on find all posts for user (%s) - %s", current.ID, err.Error())
			return err
		}
		for _, post := range posts {
			err = app.deletePost(sessionContext, current.AppID, current.OrgID, current.ID, post.GroupID, post.ID, true)
			if err != nil {
				log.Printf("error on delete all posts for user (%s) - %s", current.ID, err.Error())
				return err
			}
		}

		memberships, err := app.storage.FindGroupMembershipsWithContext(sessionContext, model.MembershipFilter{AppID: current.AppID, OrgID: current.OrgID, UserID: &current.ID})
		if err != nil {
			log.Printf("error getting user memberships - %s", err.Error())
			return err
		}
		for _, membership := range memberships.Items {
			err = app.storage.DeleteMembershipWithContext(sessionContext, membership.AppID, membership.OrgID, membership.GroupID, membership.UserID)
			if err != nil {
				log.Printf("error deleting user membership - %s", err.Error())
				return err
			}
		}

		return app.storage.DeleteUserWithContext(sessionContext, current.AppID, current.OrgID, current.ID)
	})

}

func (app *Application) getGroup(current *model.User, id string) (*model.Group, error) {
	// find the group
	var userID *string
	if current != nil {
		userID = &current.ID
	}

	group, err := app.storage.FindGroupWithContext(nil, current.AppID, current.OrgID, id, userID)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (app *Application) applyMembershipApproval(appID string, orgID string, membershipID string, approve bool, rejectReason string) error {
	err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		membership, err := app.storage.ApplyMembershipApproval(context, appID, orgID, membershipID, approve, rejectReason)
		if err != nil {
			return fmt.Errorf("error applying membership approval: %s", err)
		}

		return app.storage.UpdateGroupStats(context, membership.AppID, membership.OrgID, membership.GroupID, false, true, false, true)
	})
	if err != nil {
		return err
	}

	membership, err := app.storage.FindGroupMembershipByID(appID, orgID, membershipID)
	if err == nil && membership != nil {
		group, _ := app.storage.FindGroupWithContext(nil, appID, orgID, membership.GroupID, nil)
		topic := "group.invitations"
		if approve {
			app.notifications.SendNotification(
				[]notifications.Recipient{
					membership.ToNotificationRecipient(membership.NotificationsPreferences.OverridePreferences &&
						(membership.NotificationsPreferences.InvitationsMuted || membership.NotificationsPreferences.AllMute)),
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
				nil,
				appID,
				orgID,
			)
		} else {
			app.notifications.SendNotification(
				[]notifications.Recipient{
					membership.ToNotificationRecipient(membership.NotificationsPreferences.OverridePreferences &&
						(membership.NotificationsPreferences.InvitationsMuted || membership.NotificationsPreferences.AllMute)),
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
				nil,
				appID,
				orgID,
			)
		}

		if approve && group.CanJoinAutomatically && group.AuthmanEnabled && membership.ExternalID != "" {
			err := app.authman.AddAuthmanMemberToGroup(*group.AuthmanGroup, membership.ExternalID)
			if err != nil {
				log.Printf("err app.applyMembershipApproval() - error storing member in Authman: %s", err)
			}
		}
	} else {
		log.Printf("Unable to retrieve group by membership id: %s\n", err)
		// return err // No reason to fail if the main part succeeds
	}

	return nil
}

func (app *Application) updateMembership(membership *model.GroupMembership, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	if membership != nil {
		if status != nil && membership.Status != *status {
			membership.Status = *status
		}
		if dateAttended != nil && membership.DateAttended == nil {
			membership.DateAttended = dateAttended
		}
		if notificationsPreferences != nil {
			membership.NotificationsPreferences = *notificationsPreferences
		}

		return app.storage.PerformTransaction(func(context storage.TransactionContext) error {
			err := app.storage.UpdateMembership(context, membership)
			if err != nil {
				return err
			}

			return app.storage.UpdateGroupStats(context, membership.AppID, membership.OrgID, membership.GroupID, false, true, false, true)
		})
	}

	return nil
}

func (app *Application) getEvents(current *model.User, groupID string, filterByToMembers bool) ([]model.Event, error) {
	events, err := app.storage.FindEvents(current, groupID, filterByToMembers)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (app *Application) createEvent(eventID string, group *model.Group, toMemberList []model.ToMember, creator *model.Creator) (*model.Event, error) {
	var skipUserID *string
	if creator != nil {
		skipUserID = &creator.UserID
	}

	event, err := app.storage.CreateEvent(group.AppID, group.OrgID, eventID, group.ID, toMemberList, creator)
	if err != nil {
		return nil, err
	}

	var userIDs []string
	var recipients []notifications.Recipient
	if len(event.ToMembersList) > 0 {
		userIDs = event.GetMembersAsUserIDs(skipUserID)
	}

	result, _ := app.storage.FindGroupMembershipsWithContext(nil, model.MembershipFilter{
		GroupIDs: []string{group.ID},
		AppID:    group.AppID,
		OrgID:    group.OrgID,
		UserIDs:  userIDs,
		Statuses: []string{"member", "admin"},
	})
	recipients = result.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
		return member.IsAdminOrMember() && (skipUserID == nil || *skipUserID != member.UserID),
			member.NotificationsPreferences.OverridePreferences &&
				(member.NotificationsPreferences.EventsMuted || member.NotificationsPreferences.AllMute)
	})

	if len(recipients) > 0 {
		topic := "group.events"
		go app.notifications.SendNotification(
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
			nil,
			group.AppID,
			group.OrgID,
		)
	}

	return event, nil
}

func (app *Application) updateEvent(appID string, orgID string, eventID string, groupID string, toMemberList []model.ToMember) error {
	return app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		err := app.storage.UpdateEventWithContext(context, appID, orgID, eventID, groupID, toMemberList)
		if err == nil {
			return err
		}

		return app.storage.UpdateGroupStats(context, appID, orgID, groupID, true, false, false, false)
	})
}

func (app *Application) deleteEvent(appID string, orgID string, eventID string, groupID string) error {
	return app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		err := app.storage.DeleteEventWithContext(context, appID, orgID, eventID, groupID)
		if err != nil {
			return err
		}

		return app.storage.UpdateGroupStats(context, appID, orgID, groupID, true, false, false, false)
	})
}

func (app *Application) createPost(current *model.User, post *model.Post, group *model.Group) error {
	if group.Settings != nil && !group.Settings.PostPreferences.CanSendPostToAdmins {
		userIDs := post.GetMembersAsUserIDs(&current.ID)
		memberships, err := app.Services.FindGroupMemberships(model.MembershipFilter{
			GroupIDs: []string{post.GroupID},
			AppID:    post.AppID,
			OrgID:    post.OrgID,
			UserIDs:  userIDs,
			Statuses: []string{"admin"},
		})
		if err != nil {
			return err
		}

		var toMembers []model.ToMember
		for _, membership := range memberships.Items {
			toMembers = append(toMembers, model.ToMember{
				UserID: membership.UserID,
				Name:   membership.Name,
			})
		}
		post.ToMembersList = toMembers
	}

	if current != nil {
		membership, err := app.storage.FindGroupMembershipWithContext(nil, post.AppID, post.OrgID, post.GroupID, current.ID)
		if membership == nil || err != nil || !membership.IsAdminOrMember() {
			return fmt.Errorf("the user is not member or admin of the group")
		}

		if post.ID == "" { // Always required
			post.ID = uuid.NewString()
		}

		if post.Replies != nil { // This is constructed only for GET all for group
			post.Replies = nil
		}

		if post.ParentID != nil {
			topPost, _ := app.storage.FindTopPostByParentID(current, post.GroupID, *post.ParentID, false)
			if topPost != nil && topPost.ParentID == nil {
				post.TopParentID = &topPost.ID
			}
		}

		now := time.Now()
		post.DateCreated = &now
		post.DateUpdated = &now
		post.Creator = model.Creator{
			UserID: current.ID,
			Email:  current.Email,
			Name:   current.Name,
		}

		err = app.storage.PerformTransaction(func(context storage.TransactionContext) error {
			err := app.storage.CreatePost(context, post)
			if err != nil {
				return err
			}

			return app.storage.UpdateGroupStats(context, post.AppID, post.OrgID, post.GroupID, true, false, false, false)
		})
		if err != nil {
			return err
		}
	}

	handleRewardsAsync := func(appID, orgID, userID string) {
		count, grErr := app.storage.GetUserPostCount(appID, orgID, userID)
		if grErr != nil {
			log.Printf("Error createPost(): %s", grErr)
		} else if count != nil {
			if *count > 1 {
				app.rewards.CreateUserReward(userID, rewards.GroupsUserSubmittedPost, "")
			} else if *count == 1 {
				app.rewards.CreateUserReward(userID, rewards.GroupsUserSubmittedFirstPost, "")
			}
		}
	}
	go handleRewardsAsync(current.AppID, current.OrgID, current.ID)

	handleNotification := func() {

		recipientsUserIDs, _ := app.getPostNotificationRecipientsAsUserIDs(post, &current.ID)

		result, _ := app.storage.FindGroupMembershipsWithContext(nil, model.MembershipFilter{
			GroupIDs: []string{group.ID},
			AppID:    group.AppID,
			OrgID:    group.OrgID,
			UserIDs:  recipientsUserIDs,
			Statuses: []string{"member", "admin"},
		})
		recipients := result.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
			return member.IsAdminOrMember() && (current.ID != member.UserID),
				member.NotificationsPreferences.OverridePreferences &&
					(member.NotificationsPreferences.PostsMuted || member.NotificationsPreferences.AllMute)
		})

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
					"post_id":      post.ID,
					"post_subject": post.Subject,
					"post_body":    post.Body,
				},
				nil,
				current.AppID,
				current.OrgID,
			)
		}
	}
	go handleNotification()

	return nil
}

func (app *Application) getPostNotificationRecipientsAsUserIDs(post *model.Post, skipUserID *string) ([]string, error) {
	if post == nil {
		return nil, nil
	}

	if len(post.ToMembersList) > 0 {
		return post.GetMembersAsUserIDs(skipUserID), nil
	}

	var err error
	for {
		if post.ParentID == nil {
			break
		}

		post, err = app.storage.FindPost(nil, post.AppID, post.OrgID, nil, post.GroupID, *post.ParentID, true, false)
		if err != nil {
			log.Printf("error app.getPostToMemberList() - %s", err)
			return nil, fmt.Errorf("error app.getPostToMemberList() - %s", err)
		}

		if post != nil && len(post.ToMembersList) > 0 {
			return post.GetMembersAsUserIDs(skipUserID), nil
		}
	}

	return nil, nil
}

func (app *Application) updatePost(userID string, group *model.Group, post *model.Post) error {
	if group.Settings != nil && !group.Settings.PostPreferences.CanSendPostToAdmins {
		userIDs := post.GetMembersAsUserIDs(&userID)
		memberships, err := app.Services.FindGroupMemberships(model.MembershipFilter{
			GroupIDs: []string{post.GroupID},
			AppID:    post.AppID,
			OrgID:    post.OrgID,
			UserIDs:  userIDs,
			Statuses: []string{"admin"},
		})
		if err != nil {
			return err
		}

		if len(memberships.Items) > 0 {
			var toMembers []model.ToMember
			for _, membership := range memberships.Items {
				toMembers = append(toMembers, model.ToMember{
					UserID: membership.UserID,
					Name:   membership.Name,
				})
			}
			post.ToMembersList = toMembers
		}
	}

	originalPost, _ := app.storage.FindPost(nil, post.AppID, post.OrgID, &userID, post.GroupID, post.ID, true, true)
	if originalPost == nil {
		return fmt.Errorf("unable to find post with id (%s) ", post.ID)
	}
	if originalPost.Creator.UserID != userID {
		return fmt.Errorf("only creator of the post can update it")
	}

	if post.ID == "" { // Always required
		return fmt.Errorf("Missing id")
	}

	now := time.Now()
	post.DateUpdated = &now

	return app.storage.PerformTransaction(func(context storage.TransactionContext) error {
		err := app.storage.UpdatePost(context, userID, post)
		if err != nil {
			return err
		}

		return app.storage.UpdateGroupStats(context, post.AppID, post.OrgID, post.GroupID, true, false, false, false)
	})
}

func (app *Application) reactToPost(current *model.User, groupID string, postID string, reaction string) error {
	transaction := func(context storage.TransactionContext) error {
		post, err := app.storage.FindPost(context, current.AppID, current.OrgID, &current.ID, groupID, postID, true, false)
		if err != nil {
			return fmt.Errorf("error finding post: %v", err)
		}
		if post == nil {
			return fmt.Errorf("missing post for id %s", postID)
		}

		for _, userID := range post.Reactions[reaction] {
			if current.ID == userID {
				err = app.storage.ReactToPost(context, current.ID, postID, reaction, false)
				if err != nil {
					return fmt.Errorf("error removing reaction: %v", err)
				}

				return nil
			}
		}

		err = app.storage.ReactToPost(context, current.ID, postID, reaction, true)
		if err != nil {
			return fmt.Errorf("error adding reaction: %v", err)
		}

		return nil
	}

	return app.storage.PerformTransaction(transaction)
}

func (app *Application) reportPostAsAbuse(current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {
	if post == nil {
		return errors.New("post is missing")
	}
	if !sendToDean && !sendToGroupAdmins {
		sendToDean = true
	}

	var creatorExternalID string
	creator, err := app.storage.FindUser(post.AppID, post.OrgID, post.Creator.UserID, false)
	if err != nil {
		log.Printf("error retrieving user: %s", err)
	} else if creator != nil {
		creatorExternalID = creator.ExternalID
	}

	err = app.storage.ReportPostAsAbuse(post)
	if err != nil {
		log.Printf("error while reporting an abuse post: %s", err)
		return fmt.Errorf("error while reporting an abuse post: %s", err)
	}

	subject := ""
	if sendToDean && !sendToGroupAdmins {
		subject = "Report violation of Student Code to Dean of Students"
	} else if !sendToDean && sendToGroupAdmins {
		subject = "Report of Obscene, Harassing, or Threatening Content to Group Administrators"
	} else {
		subject = "Report violation of Student Code to Dean of Students and obscene, threatening, or harassing content to Group Administrators"
	}

	subject = fmt.Sprintf("%s %s", subject, post.DateCreated.Format(time.RFC850))

	if sendToDean {
		config, err := app.storage.FindConfig(model.ConfigTypeApplication, post.AppID, post.OrgID)
		if err != nil || config == nil {
			log.Printf("error finding application config for appID %s, orgID %s: %v", post.AppID, post.OrgID, err)
			return fmt.Errorf("error finding application config for appID %s, orgID %s: %v", post.AppID, post.OrgID, err)
		}
		appConfig, err := model.GetConfigData[model.ApplicationConfigData](*config)
		if err != nil {
			log.Printf("error asserting as application config for appID %s, orgID %s: %v", post.AppID, post.OrgID, err)
			return fmt.Errorf("error asserting as application config for appID %s, orgID %s: %v", post.AppID, post.OrgID, err)
		}

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
		app.notifications.SendMail(appConfig.ReportAbuseRecipientEmail, subject, body)
	}
	if sendToGroupAdmins {
		result, _ := app.storage.FindGroupMembershipsWithContext(nil, model.MembershipFilter{
			GroupIDs: []string{group.ID},
			AppID:    post.AppID,
			OrgID:    post.OrgID,
			Statuses: []string{"admin"},
		})
		toMembers := result.GetMembersAsRecipients(func(membership model.GroupMembership) (bool, bool) {
			return membership.UserID != current.ID, false
		})

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
			"post_id":      post.ID,
			"post_subject": post.Subject,
			"post_body":    post.Body,
		},
			nil,
			current.AppID,
			current.OrgID)
	}

	return nil
}

func (app *Application) deletePost(context storage.TransactionContext, appID string, orgID string, userID string, groupID string, postID string, force bool) error {
	deleteWrapper := func(transactionContext storage.TransactionContext) error {
		membership, _ := app.storage.FindGroupMembershipWithContext(transactionContext, appID, orgID, groupID, userID)
		filterToMembers := true
		if membership != nil && membership.IsAdmin() {
			filterToMembers = false
		}

		originalPost, _ := app.storage.FindPost(transactionContext, appID, orgID, &userID, groupID, postID, true, filterToMembers)
		if originalPost == nil {
			return fmt.Errorf("unable to find post with id (%s) ", postID)
		}

		if !force {
			if originalPost == nil || membership == nil || (!membership.IsAdmin() && originalPost.Creator.UserID != userID) {
				return fmt.Errorf("only creator of the post or group admin can delete it")
			}
		}

		childPosts, err := app.storage.FindPostsByParentID(transactionContext, appID, orgID, userID, groupID, postID, true, false, false, nil)
		if len(childPosts) > 0 && err == nil {
			for _, post := range childPosts {
				app.deletePost(transactionContext, appID, orgID, userID, groupID, post.ID, true)
			}
		}

		err = app.storage.DeletePost(transactionContext, appID, orgID, userID, groupID, postID, force)
		if err != nil {
			return err
		}

		return app.storage.UpdateGroupStats(transactionContext, appID, orgID, groupID, true, false, false, false)
	}

	if context != nil {
		return deleteWrapper(context)
	}
	return app.storage.PerformTransaction(func(transactionContext storage.TransactionContext) error {
		return deleteWrapper(transactionContext)
	})
}

// TODO this logic needs to be refactored because it's over complicated!
func (app *Application) synchronizeAuthman(appID string, orgID string, checkThreshold bool) error {
	startTime := time.Now()
	transaction := func(context storage.TransactionContext) error {
		times, err := app.storage.FindSyncTimes(context, appID, orgID)
		if err != nil {
			return err
		}
		if times != nil && times.StartTime != nil {
			config, err := app.storage.FindConfig(model.ConfigTypeSync, appID, orgID)
			if err != nil {
				log.Printf("error finding sync config for appID %s, orgID %s: %v", appID, orgID, err)
			}

			timeout := defaultConfigSyncTimeout
			var syncConfig *model.SyncConfigData
			if config != nil {
				syncConfig, err = model.GetConfigData[model.SyncConfigData](*config)
				if err != nil {
					log.Printf("error asserting as sync config for appID %s, orgID %s: %v", appID, orgID, err)
				}
				if syncConfig != nil && syncConfig.Timeout > 0 {
					timeout = syncConfig.Timeout
				}
			}

			if times.EndTime == nil {
				if !startTime.After(times.StartTime.Add(time.Minute * time.Duration(timeout))) {
					log.Printf("Another Authman sync process is running for appID %s, orgID %s", appID, orgID)
					return fmt.Errorf("another Authman sync process is running for appID %s, orgID %s", appID, orgID)
				}
				log.Printf("Authman sync past timeout threshold %d mins for appID %s, orgID %s\n", timeout, appID, orgID)
			}
			if checkThreshold {
				if syncConfig == nil {
					log.Printf("missing sync configs for appID %s, orgID %s", appID, orgID)
					return fmt.Errorf("missing sync configs for appID %s, orgID %s: %v", appID, orgID, err)
				}
				if !startTime.After(times.StartTime.Add(time.Minute * time.Duration(syncConfig.TimeThreshold))) {
					log.Printf("Authman has already been synced for appID %s, orgID %s", appID, orgID)
					return fmt.Errorf("Authman has already been synced for appID %s, orgID %s", appID, orgID)
				}
			}
		}

		return app.storage.SaveSyncTimes(context, model.SyncTimes{StartTime: &startTime, EndTime: nil, AppID: appID, OrgID: orgID})
	}

	err := app.storage.PerformTransaction(transaction)
	if err != nil {
		return err
	}

	log.Printf("Global Authman synchronization started for appID %s, orgID %s\n", appID, orgID)

	app.authmanSyncInProgress = true
	finishAuthmanSync := func() {
		endTime := time.Now()
		err := app.storage.SaveSyncTimes(nil, model.SyncTimes{StartTime: &startTime, EndTime: &endTime, AppID: appID, OrgID: orgID})
		if err != nil {
			log.Printf("Error saving sync configs to end sync: %s\n", err)
			return
		}
		log.Printf("Global Authman synchronization finished for appID %s, orgID %s\n", appID, orgID)
	}
	defer finishAuthmanSync()

	appConfig, err := app.storage.FindConfig(model.ConfigTypeApplication, appID, orgID)
	if err != nil {
		return fmt.Errorf("error finding application config for appID %s, orgID %s: %v", appID, orgID, err)
	}
	if appConfig == nil {
		return fmt.Errorf("missing application config for appID %s, orgID %s", appID, orgID)
	}
	applicationConfig, err := model.GetConfigData[model.ApplicationConfigData](*appConfig)
	if err != nil {
		return fmt.Errorf("error asserting as application config for appID %s, orgID %s: %v", appID, orgID, err)
	}

	configTypeManagedGroup := model.ConfigTypeManagedGroup
	mgConfigs, err := app.storage.FindConfigs(&configTypeManagedGroup, &appID, &orgID)
	if err != nil {
		return fmt.Errorf("error finding managed group configs for appID %s, orgID %s", appID, orgID)
	}

	for _, config := range mgConfigs {
		mgConfig, err := model.GetConfigData[model.ManagedGroupConfigData](config)
		if err != nil {
			log.Printf("Error asserting as managed group config for appID %s, orgID %s: %v\n", appID, orgID, err)
			continue
		}
		for _, stemName := range mgConfig.AuthmanStems {
			stemGroups, err := app.authman.RetrieveAuthmanStemGroups(stemName)
			if err != nil {
				return fmt.Errorf("error on requesting Authman for stem groups: %s", err)
			}

			if stemGroups != nil && len(stemGroups.WsFindGroupsResults.GroupResults) > 0 {
				for _, stemGroup := range stemGroups.WsFindGroupsResults.GroupResults {
					storedStemGroup, err := app.storage.FindAuthmanGroupByKey(appID, orgID, stemGroup.Name)
					if err != nil {
						return fmt.Errorf("error on requesting Authman for stem groups: %s", err)
					}

					title, adminUINs := stemGroup.GetGroupPrettyTitleAndAdmins()

					defaultAdminsMapping := map[string]bool{}
					for _, externalID := range adminUINs {
						defaultAdminsMapping[externalID] = true
					}
					for _, externalID := range applicationConfig.AuthmanAdminUINList {
						defaultAdminsMapping[externalID] = true
					}
					for _, externalID := range mgConfig.AdminUINs {
						defaultAdminsMapping[externalID] = true
					}

					constructedAdminUINs := []string{}
					if len(defaultAdminsMapping) > 0 {
						for key := range defaultAdminsMapping {
							constructedAdminUINs = append(constructedAdminUINs, key)
						}
					}

					if storedStemGroup == nil {
						var memberships []model.GroupMembership
						if len(constructedAdminUINs) > 0 {
							memberships = app.buildMembersByExternalIDs(appID, orgID, constructedAdminUINs, "admin")
						}

						emptyText := ""
						_, err := app.storage.CreateGroup(nil, &model.Group{
							AppID:                appID,
							OrgID:                orgID,
							Title:                title,
							Description:          &emptyText,
							Category:             "Academic", // Hardcoded.
							Privacy:              "private",
							HiddenForSearch:      true,
							CanJoinAutomatically: true,
							AuthmanEnabled:       true,
							AuthmanGroup:         &stemGroup.Name,
						}, memberships)
						if err != nil {
							return fmt.Errorf("error on create Authman stem group: '%s' - %s", stemGroup.Name, err)
						}

						log.Printf("Created new `%s` group", title)
					} else {
						missedUINs := []string{}
						groupUpdated := false

						existingAdmins, err := app.storage.FindGroupMembershipsWithContext(nil, model.MembershipFilter{
							GroupIDs: []string{storedStemGroup.ID},
							AppID:    appID,
							OrgID:    orgID,
							Statuses: []string{"admin"},
						})

						membershipsForUpdate := []model.GroupMembership{}
						if len(existingAdmins.Items) > 0 {
							for _, uin := range adminUINs {
								found := false
								for _, member := range existingAdmins.Items {
									if member.ExternalID == uin {
										if member.Status != "admin" {
											now := time.Now()
											member.Status = "admin"
											member.DateUpdated = &now
											membershipsForUpdate = append(membershipsForUpdate, member)
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
						} else if err != nil {
							log.Printf("error retreiving admins for group: %s - %s", stemGroup.Name, err)
						}

						if len(missedUINs) > 0 {
							missedMembers := app.buildMembersByExternalIDs(appID, orgID, missedUINs, "admin")
							if len(missedMembers) > 0 {
								membershipsForUpdate = append(membershipsForUpdate, missedMembers...)
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
							err := app.storage.UpdateGroup(nil, storedStemGroup, membershipsForUpdate)
							if err != nil {
								log.Printf("error app.synchronizeAuthmanGroup() - unable to update group admins of '%s' - %s", storedStemGroup.Title, err)
							}
						}
					}
				}
			}
		}
	}

	authmanGroups, err := app.storage.FindAuthmanGroups(appID, orgID)
	if err != nil {
		return err
	}

	for _, authmanGroup := range authmanGroups {
		err := app.synchronizeAuthmanGroup(authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID)
		if err != nil {
			log.Printf("error app.synchronizeAuthmanGroup() '%s' - %s", authmanGroup.Title, err)
		}
	}

	return nil
}

func (app *Application) buildMembersByExternalIDs(appID string, orgID string, externalIDs []string, memberStatus string) []model.GroupMembership {
	if len(externalIDs) > 0 {
		users, _ := app.storage.FindUsers(appID, orgID, externalIDs, true)
		members := []model.GroupMembership{}
		userExternalIDmapping := map[string]model.User{}
		for _, user := range users {
			userExternalIDmapping[user.ExternalID] = user
		}

		for _, externalID := range externalIDs {
			if value, ok := userExternalIDmapping[externalID]; ok {
				members = append(members, model.GroupMembership{
					ID:          uuid.NewString(),
					UserID:      value.ID,
					ExternalID:  externalID,
					Name:        value.Name,
					Email:       value.Email,
					Status:      memberStatus,
					DateCreated: time.Now(),
				})
			} else {
				members = append(members, model.GroupMembership{
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
func (app *Application) synchronizeAuthmanGroup(appID string, orgID string, groupID string) error {
	if groupID == "" {
		return errors.New("Missing group ID")
	}
	var group *model.Group
	var err error
	group, err = app.checkGroupSyncTimes(appID, orgID, groupID)
	if err != nil {
		return err
	}

	log.Printf("Authman synchronization for group %s started", *group.AuthmanGroup)

	authmanExternalIDs, authmanErr := app.authman.RetrieveAuthmanGroupMembers(*group.AuthmanGroup)
	if authmanErr != nil {
		return fmt.Errorf("error on requesting Authman for %s: %s", *group.AuthmanGroup, authmanErr)
	}

	app.authmanSyncInProgress = true
	finishAuthmanSync := func() {
		endTime := time.Now()
		group.SyncEndTime = &endTime
		err = app.storage.UpdateGroupSyncTimes(nil, group)
		if err != nil {
			log.Printf("Error saving group to end sync for Authman %s: %s\n", *group.AuthmanGroup, err)
			return
		}
		log.Printf("Authman synchronization for group %s finished", *group.AuthmanGroup)
	}
	defer finishAuthmanSync()

	err = app.syncAuthmanGroupMemberships(group, authmanExternalIDs)
	if err != nil {
		return fmt.Errorf("error updating group memberships for Authman %s: %s", *group.AuthmanGroup, err)
	}

	return nil
}

func (app *Application) checkGroupSyncTimes(appID string, orgID string, groupID string) (*model.Group, error) {
	var group *model.Group
	var err error
	startTime := time.Now()
	transaction := func(context storage.TransactionContext) error {
		group, err = app.storage.FindGroupWithContext(context, appID, orgID, groupID, nil)
		if err != nil {
			return fmt.Errorf("error finding group for ID %s: %s", groupID, err)
		}
		if group == nil {
			return fmt.Errorf("missing group for ID %s", groupID)
		}
		if !group.IsAuthmanSyncEligible() {
			return fmt.Errorf("Authman synchronization failed for group '%s' due to bad settings", group.Title)
		}

		if group.SyncStartTime != nil {
			config, err := app.storage.FindConfig(model.ConfigTypeSync, appID, orgID)
			if err != nil {
				log.Printf("error finding sync config for appID %s, orgID %s: %v", appID, orgID, err)
			}

			timeout := defaultConfigSyncTimeout
			var syncConfig *model.SyncConfigData
			if config != nil {
				syncConfig, err = model.GetConfigData[model.SyncConfigData](*config)
				if err != nil {
					log.Printf("error asserting as sync config for appID %s, orgID %s: %v", appID, orgID, err)
				}
				if syncConfig != nil && syncConfig.Timeout > 0 {
					timeout = syncConfig.GroupTimeout
				}
			}
			if group.SyncEndTime == nil {
				if !startTime.After(group.SyncStartTime.Add(time.Minute * time.Duration(timeout))) {
					log.Println("Another Authman sync process is running for group ID " + group.ID)
					return fmt.Errorf("another Authman sync process is running for group ID %s", group.ID)
				}
				log.Printf("Authman sync timed out after %d mins for group ID %s\n", timeout, group.ID)
			}
		}

		group.SyncStartTime = &startTime
		group.SyncEndTime = nil
		err = app.storage.UpdateGroupSyncTimes(context, group)
		if err != nil {
			return fmt.Errorf("error switching to group memberships for Authman %s: %s", *group.AuthmanGroup, err)
		}
		return nil
	}

	err = app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (app *Application) syncAuthmanGroupMemberships(authmanGroup *model.Group, authmanExternalIDs []string) error {
	syncID := uuid.NewString()
	log.Printf("Sync ID %s for Authman %s...\n", syncID, *authmanGroup.AuthmanGroup)

	//TODO: These operations should ideally use a transaction, but the transaction may get too large

	// Get list of all member external IDs (Authman members + admins)
	allExternalIDs := append([]string{}, authmanExternalIDs...)

	// Load existing admins
	adminExternalIDsMap := map[string]bool{}
	adminMembers, err := app.storage.FindGroupMembershipsWithContext(nil, model.MembershipFilter{
		GroupIDs: []string{authmanGroup.ID},
		AppID:    authmanGroup.AppID,
		OrgID:    authmanGroup.OrgID,
		Statuses: []string{"admin"},
	})
	if err != nil {
		return fmt.Errorf("error finding admin memberships in authman %s: %s", *authmanGroup.AuthmanGroup, err)
	}

	for _, adminMember := range adminMembers.Items {
		if len(adminMember.ExternalID) > 0 {
			allExternalIDs = append(allExternalIDs, adminMember.ExternalID)
			adminExternalIDsMap[adminMember.ExternalID] = true
		}
	}

	// Load user records for all members
	localUsersMapping := map[string]model.User{}
	localUsers, err := app.storage.FindUsers(authmanGroup.AppID, authmanGroup.OrgID, allExternalIDs, true)
	if err != nil {
		return fmt.Errorf("error on getting %d users for Authman %s: %s", len(allExternalIDs), *authmanGroup.AuthmanGroup, err)
	}

	for _, user := range localUsers {
		localUsersMapping[user.ExternalID] = user
	}

	missingInfoMembers := []model.GroupMembership{}
	updateOperations := []storage.SingleMembershipOperation{}
	log.Printf("Processing %d current members for Authman %s...\n", len(authmanExternalIDs), *authmanGroup.AuthmanGroup)
	for _, externalID := range authmanExternalIDs {
		status := "member"
		if _, ok := adminExternalIDsMap[externalID]; ok {
			status = "admin"
		}
		var userID *string
		var name *string
		var email *string
		if user, ok := localUsersMapping[externalID]; ok {
			if user.ID != "" {
				userID = &user.ID
			}
			if user.Name != "" {
				name = &user.Name
			}
			if user.Email != "" {
				email = &user.Email
			}
		}

		updateOperations = append(updateOperations, storage.SingleMembershipOperation{
			ExternalID: externalID,
			UserID:     userID,
			Status:     &status,
			Email:      email,
			Name:       name,
			SyncID:     &syncID,
			Answers:    authmanGroup.CreateMembershipEmptyAnswers(),
		})
	}
	if len(updateOperations) > 0 {
		err = app.storage.BulkUpdateGroupMembershipsByExternalID(authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID, updateOperations, false)
		if err != nil {
			log.Printf("Error on bulk saving membership (phase 1) in Authman %s: %s\n", *authmanGroup.AuthmanGroup, err)
		} else {
			log.Printf("Successful bulk saving membership (phase 1) in Authman '%s'", *authmanGroup.AuthmanGroup)
			memberships, _ := app.storage.FindGroupMembershipsWithContext(nil, model.MembershipFilter{
				GroupIDs: []string{authmanGroup.ID},
				AppID:    authmanGroup.AppID,
				OrgID:    authmanGroup.OrgID,
			})
			for _, membership := range memberships.Items {
				if membership.Email == "" || membership.Name == "" {
					missingInfoMembers = append(missingInfoMembers, membership)
				}
			}
		}
	}

	// Update admin user data
	for _, adminMember := range adminMembers.Items {
		var userID *string
		var name *string
		var email *string
		updatedInfo := false
		if mappedUser, ok := localUsersMapping[adminMember.ExternalID]; ok {
			if mappedUser.ID != "" && mappedUser.ID != adminMember.UserID {
				userID = &mappedUser.ID
				updatedInfo = true
			}
			if mappedUser.Name != "" && mappedUser.Name != adminMember.Name {
				name = &mappedUser.Name
				updatedInfo = true
			}
			if mappedUser.Email != "" && mappedUser.Email != adminMember.Email {
				email = &mappedUser.Email
				updatedInfo = true
			}
		}
		if updatedInfo {
			err := app.storage.PerformTransaction(func(context storage.TransactionContext) error {
				err := app.storage.SaveGroupMembershipByExternalID(context, authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID, adminMember.ExternalID, userID, nil, email, name, nil, nil)
				if err != nil {
					return fmt.Errorf("error saving admin membership with missing info for external ID %s in Authman %s: %s", adminMember.ExternalID, *authmanGroup.AuthmanGroup, err)
				}

				log.Printf("Update admin member %s for group '%s'", adminMember.ExternalID, authmanGroup.Title)
				return app.storage.UpdateGroupStats(context, authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID, false, false, true, true)
			})
			if err != nil {
				log.Println(err.Error())
			}
		}
	}

	// Fetch user info for the required users
	log.Printf("Processing %d members missing info for Authman %s...\n", len(missingInfoMembers), *authmanGroup.AuthmanGroup)
	for i := 0; i < len(missingInfoMembers); i += authmanUserBatchSize {
		j := i + authmanUserBatchSize
		if j > len(missingInfoMembers) {
			j = len(missingInfoMembers)
		}
		log.Printf("Processing members missing info %d - %d for Authman %s...\n", i, j, *authmanGroup.AuthmanGroup)
		members := missingInfoMembers[i:j]
		externalIDs := make([]string, j-i)
		for i, member := range members {
			externalIDs[i] = member.ExternalID
		}
		authmanUsers, err := app.authman.RetrieveAuthmanUsers(externalIDs)
		if err != nil {
			log.Printf("error on retrieving missing user info for %d members: %s\n", len(externalIDs), err)
		} else if len(authmanUsers) > 0 {

			updateOperations = []storage.SingleMembershipOperation{}
			for _, member := range members {
				var name *string
				var email *string
				updatedInfo := false
				if mappedUser, ok := authmanUsers[member.ExternalID]; ok {
					if member.Name == "" && mappedUser.Name != "" {
						name = &mappedUser.Name
						updatedInfo = true
					}
					if member.Email == "" && len(mappedUser.AttributeValues) > 0 {
						email = &mappedUser.AttributeValues[0]
						updatedInfo = true
					}
					if !updatedInfo {
						log.Printf("The user has missing info: %+v Group: '%s' Authman Group: '%s'\n", mappedUser, authmanGroup.Title, *authmanGroup.AuthmanGroup)
					}
				}
				if updatedInfo {
					updateOperations = append(updateOperations, storage.SingleMembershipOperation{
						ExternalID: member.ExternalID,
						Email:      email,
						Name:       name,
						SyncID:     &syncID,
					})
				}
			}

			if len(updateOperations) > 0 {
				err = app.storage.BulkUpdateGroupMembershipsByExternalID(authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID, updateOperations, true)
				if err != nil {
					log.Printf("Error on bulk saving membership (phase 2) in Authman '%s': %s\n", *authmanGroup.AuthmanGroup, err)
				} else {
					log.Printf("Successful bulk saving membership (phase 2) in Authman '%s'", *authmanGroup.AuthmanGroup)
				}
			}
		}
	}

	// Delete removed non-admin members
	log.Printf("Deleting removed members for Authman %s...\n", *authmanGroup.AuthmanGroup)
	deleteCount, err := app.storage.DeleteUnsyncedGroupMemberships(authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID, syncID)
	if err != nil {
		log.Printf("Error deleting removed memberships in Authman %s\n", *authmanGroup.AuthmanGroup)
	} else {
		log.Printf("%d memberships removed from Authman %s\n", deleteCount, *authmanGroup.AuthmanGroup)
	}

	err = app.storage.UpdateGroupStats(nil, authmanGroup.AppID, authmanGroup.OrgID, authmanGroup.ID, false, false, true, true)
	if err != nil {
		log.Printf("Error updating group stats for '%s' - %s", *authmanGroup.AuthmanGroup, err)
	}

	return nil
}

func (app *Application) sendGroupNotification(appID string, orgID string, notification model.GroupNotification) error {
	memberStatuses := notification.MemberStatuses
	if len(memberStatuses) == 0 {
		memberStatuses = []string{"admin", "member"}
	}

	members, err := app.findGroupMemberships(model.MembershipFilter{
		GroupIDs: []string{notification.GroupID},
		AppID:    appID,
		OrgID:    orgID,
		UserIDs:  notification.Members.ToUserIDs(),
		Statuses: memberStatuses,
	})

	if err != nil {
		return err
	}

	app.sendNotification(members.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
		return true, true // Should it be a separate notification preference?
	}), notification.Topic, notification.Subject, notification.Body, notification.Data, appID, orgID)

	return nil
}

func (app *Application) sendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string) {
	app.notifications.SendNotification(recipients, topic, title, text, data, nil, appID, orgID)
}

func (app *Application) getResearchProfileUserCount(current *model.User, researchProfile map[string]map[string][]string) (int64, error) {
	searchParams := app.formatCoreAccountSearchParams(researchProfile)
	return app.corebb.GetAccountsCount(searchParams, &current.AppID, &current.OrgID)
}

func (app *Application) formatCoreAccountSearchParams(researchProfile map[string]map[string][]string) map[string]interface{} {
	searchParams := map[string]interface{}{}
	for k1, v1 := range researchProfile {
		for k2, v2 := range v1 {
			searchParams["profile.unstructured_properties.research_questionnaire_answers."+k1+"."+k2] = map[string]interface{}{"operation": "any", "value": v2}
		}
	}
	// If empty profile is provided, find all users that have filled out the profile
	//TODO: Handle filled out profile search better
	if len(searchParams) == 0 {
		searchParams["profile.unstructured_properties.research_questionnaire_answers.demographics"] = "$exists"
	}

	return searchParams
}
