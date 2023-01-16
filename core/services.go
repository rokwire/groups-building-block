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

func (app *Application) getGroupEntity(clientID string, id string) (*model.Group, error) {
	group, err := app.storage.FindGroup(nil, clientID, id, nil)
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

func (app *Application) isGroupAdmin(clientID string, groupID string, userID string) (bool, error) {
	membership, err := app.storage.FindGroupMembership(clientID, groupID, userID)
	if err != nil {
		return false, err
	}
	if membership == nil || !membership.Admin {
		return false, nil
	}

	return true, nil
}

func (app *Application) createGroup(clientID string, current *model.User, group *model.Group) (*string, *utils.GroupError) {
	insertedID, err := app.storage.CreateGroup(clientID, current, group, nil)
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

	err := app.storage.UpdateGroup(clientID, current, group)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) updateGroupDateUpdated(clientID string, groupID string) error {
	err := app.storage.UpdateGroupDateUpdated(clientID, groupID)
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

func (app *Application) getGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, &current.ID, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getAllGroups(clientID string) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, nil, model.GroupsFilter{})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getUserGroups(clientID string, current *model.User, filter model.GroupsFilter) ([]model.Group, error) {
	// find the user groups
	groups, err := app.storage.FindUserGroups(clientID, current.ID, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) loginUser(clientID string, current *model.User) error {
	return app.storage.LoginUser(clientID, current)
}

func (app *Application) deleteUser(clientID string, current *model.User) error {
	return app.storage.DeleteUser(clientID, current.ID)
}

func (app *Application) getGroup(clientID string, current *model.User, id string) (*model.Group, error) {
	// find the group
	var userID *string
	if current != nil {
		userID = &current.ID
	}

	group, err := app.storage.FindGroup(nil, clientID, id, userID)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (app *Application) applyMembershipApproval(clientID string, current *model.User, membershipID string, approve bool, rejectReason string) error {
	err := app.storage.ApplyMembershipApproval(clientID, membershipID, approve, rejectReason)
	if err != nil {
		return fmt.Errorf("error applying membership approval: %s", err)
	}

	membership, err := app.storage.FindGroupMembershipByID(clientID, membershipID)
	if err == nil && membership != nil {
		group, _ := app.storage.FindGroup(nil, clientID, membership.GroupID, nil)
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
				current.AppID,
				current.OrgID,
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
				current.AppID,
				current.OrgID,
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

func (app *Application) updateMembership(clientID string, current *model.User, membershipID string, status *string, dateAttended *time.Time, notificationsPreferences *model.NotificationsPreferences) error {
	membership, _ := app.storage.FindGroupMembershipByID(clientID, membershipID)
	if membership != nil {
		if status != nil && membership.Status != *status {
			membership.Status = *status
			membership.Admin = membership.IsAdmin()
		}
		if dateAttended != nil && membership.DateAttended == nil {
			membership.DateAttended = dateAttended
		}
		if notificationsPreferences != nil {
			membership.NotificationsPreferences = *notificationsPreferences
		}

		err := app.storage.UpdateMembership(clientID, current, membershipID, membership)
		if err != nil {
			return err
		}
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

	event, err := app.storage.CreateEvent(clientID, eventID, group.ID, toMemberList, creator)
	if err != nil {
		return nil, err
	}

	var userIDs []string
	var recipients []notifications.Recipient
	if len(event.ToMembersList) > 0 {
		userIDs = event.GetMembersAsUserIDs(skipUserID)
	}

	result, _ := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{group.ID},
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
			current.AppID,
			current.OrgID,
		)
	}

	return event, nil
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

func (app *Application) getPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {
	return app.storage.FindPosts(clientID, current, groupID, filterPrivatePostsValue, filterByToMembers, offset, limit, order)
}

func (app *Application) getPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return app.storage.FindPost(nil, clientID, userID, groupID, postID, skipMembershipCheck, filterByToMembers)
}

func (app *Application) getUserPostCount(clientID string, userID string) (*int64, error) {
	return app.storage.GetUserPostCount(clientID, userID)
}

func (app *Application) createPost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {

	if group.Settings != nil && !group.Settings.PostPreferences.CanSendPostToAdmins {
		userIDs := post.GetMembersAsUserIDs(&current.ID)
		memberships, err := app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
			GroupIDs: []string{post.GroupID},
			UserIDs:  userIDs,
			Statuses: []string{"admin"},
		})
		if err != nil {
			return nil, err
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

		recipientsUserIDs, _ := app.getPostNotificationRecipientsAsUserIDs(clientID, post, &current.ID)

		result, _ := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
			GroupIDs: []string{group.ID},
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
					"post_id":      *post.ID,
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

	return post, nil
}

func (app *Application) getPostNotificationRecipientsAsUserIDs(clientID string, post *model.Post, skipUserID *string) ([]string, error) {
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

		post, err = app.storage.FindPost(nil, clientID, nil, post.GroupID, *post.ParentID, true, false)
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

func (app *Application) updatePost(clientID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error) {
	if group.Settings != nil && !group.Settings.PostPreferences.CanSendPostToAdmins {
		userIDs := post.GetMembersAsUserIDs(&current.ID)
		memberships, err := app.Services.FindGroupMemberships(clientID, model.MembershipFilter{
			GroupIDs: []string{post.GroupID},
			UserIDs:  userIDs,
			Statuses: []string{"admin"},
		})
		if err != nil {
			return nil, err
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

	return app.storage.UpdatePost(clientID, current.ID, post)
}

func (app *Application) reactToPost(clientID string, current *model.User, groupID string, postID string, reaction string) error {
	transaction := func(context storage.TransactionContext) error {
		post, err := app.storage.FindPost(context, clientID, &current.ID, groupID, postID, true, false)
		if err != nil {
			return fmt.Errorf("error finding post: %v", err)
		}
		if post == nil {
			return fmt.Errorf("missing post for id %s", postID)
		}

		//get reactions bases on post
		val, ok := post.ReactionStats[reaction]
		if ok && val != 0 {
			res, err := app.storage.FindReactions(context, postID, current.ID)
			if err == nil {
				for i := range res.Reactions {
					if res.Reactions[i] == reaction {
						err = app.storage.ReactToPost(context, current.ID, postID, reaction, false)
						if err != nil {
							return fmt.Errorf("error removing reaction: %v", err)
						}
					}
				}

				return nil
			}
		}

		//New reaction for current user
		err = app.storage.ReactToPost(context, current.ID, postID, reaction, true)
		if err != nil {
			return fmt.Errorf("error adding reaction: %v", err)
		}

		return nil
	}

	return app.storage.PerformTransaction(transaction)
}

func (app *Application) findReactionsByPost(postID string) ([]model.PostReactions, error) {
	res, err := app.storage.FindReactionsByPost(postID)
	if err != nil {
		return nil, fmt.Errorf("error finding reaction stats: %v", err)
	}
	return res, err
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
		subject = "Report of Obscene, Harassing, or Threatening Content to Group Administrators"
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
		result, _ := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
			GroupIDs: []string{group.ID},
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
			"post_id":      *post.ID,
			"post_subject": post.Subject,
			"post_body":    post.Body,
		},
			nil,
			current.AppID,
			current.OrgID)
	}

	return nil
}

func (app *Application) deletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	return app.storage.DeletePost(nil, clientID, userID, groupID, postID, force)
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
				log.Printf("Authman sync past timeout threshold %d mins for client ID %s\n", timeout, clientID)
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
						var memberships []model.GroupMembership
						if len(constructedAdminUINs) > 0 {
							memberships = app.buildMembersByExternalIDs(clientID, constructedAdminUINs, "admin")
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
						}, memberships)
						if err != nil {
							return fmt.Errorf("error on create Authman stem group: '%s' - %s", stemGroup.Name, err)
						}

						log.Printf("Created new `%s` group", title)
					} else {
						missedUINs := []string{}
						groupUpdated := false

						existingAdmins, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
							GroupIDs: []string{storedStemGroup.ID},
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
											member.Admin = true
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
							log.Printf("error rertieving admins for group: %s - %s", stemGroup.Name, err)
						}

						if len(missedUINs) > 0 {
							missedMembers := app.buildMembersByExternalIDs(clientID, missedUINs, "admin")
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
							err := app.storage.UpdateGroupWithMembership(clientID, nil, storedStemGroup, membershipsForUpdate)
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
			err := app.synchronizeAuthmanGroup(clientID, authmanGroup.ID)
			if err != nil {
				fmt.Errorf("error app.synchronizeAuthmanGroup() '%s' - %s", authmanGroup.Title, err)
			}
		}
	}

	return nil
}

func (app *Application) buildMembersByExternalIDs(clientID string, externalIDs []string, memberStatus string) []model.GroupMembership {
	if len(externalIDs) > 0 {
		users, _ := app.storage.FindUsers(clientID, externalIDs, true)
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
func (app *Application) synchronizeAuthmanGroup(clientID string, groupID string) error {
	if groupID == "" {
		return errors.New("Missing group ID")
	}
	var group *model.Group
	var err error
	group, err = app.checkGroupSyncTimes(clientID, groupID)
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
		err = app.storage.UpdateGroupSyncTimes(nil, clientID, group)
		if err != nil {
			log.Printf("Error saving group to end sync for Authman %s: %s\n", *group.AuthmanGroup, err)
			return
		}
		log.Printf("Authman synchronization for group %s finished", *group.AuthmanGroup)
	}
	defer finishAuthmanSync()

	err = app.syncAuthmanGroupMemberships(clientID, group, authmanExternalIDs)
	if err != nil {
		return fmt.Errorf("error updating group memberships for Authman %s: %s", *group.AuthmanGroup, err)
	}

	return nil
}

func (app *Application) checkGroupSyncTimes(clientID string, groupID string) (*model.Group, error) {
	var group *model.Group
	var err error
	startTime := time.Now()
	transaction := func(context storage.TransactionContext) error {
		group, err = app.storage.FindGroupWithContext(context, clientID, groupID, nil)
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
			config, err := app.storage.FindSyncConfig(clientID)
			if err != nil {
				log.Printf("error finding sync configs for clientID %s: %v", clientID, err)
			}
			timeout := defaultConfigSyncTimeout
			if config != nil && config.GroupTimeout > 0 {
				timeout = config.GroupTimeout
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
		err = app.storage.UpdateGroupSyncTimes(context, clientID, group)
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

func (app *Application) syncAuthmanGroupMemberships(clientID string, authmanGroup *model.Group, authmanExternalIDs []string) error {
	syncID := uuid.NewString()
	log.Printf("Sync ID %s for Authman %s...\n", syncID, *authmanGroup.AuthmanGroup)

	// Get list of all member external IDs (Authman members + admins)
	allExternalIDs := append([]string{}, authmanExternalIDs...)
	adminMembers, err := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{authmanGroup.ID},
		Statuses: []string{"admin"},
	})
	if err != nil {
		log.Printf("Error finding admin memberships in Authman %s: %s\n", *authmanGroup.AuthmanGroup, err)
	} else {
		for _, adminMember := range adminMembers.Items {
			if len(adminMember.ExternalID) > 0 {
				allExternalIDs = append(allExternalIDs, adminMember.ExternalID)
			}
		}
	}

	// Load user records for all members
	localUsersMapping := map[string]model.User{}
	localUsers, err := app.storage.FindUsers(clientID, allExternalIDs, true)
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
			ClientID:   clientID,
			GroupID:    authmanGroup.ID,
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
		err = app.storage.BulkUpdateGroupMembershipsByExternalID(clientID, authmanGroup.ID, updateOperations, false)
		if err != nil {
			log.Printf("Error on bulk saving membership (phase 1) in Authman %s: %s\n", *authmanGroup.AuthmanGroup, err)
		} else {
			log.Printf("Successful bulk saving membership (phase 1) in Authman '%s'", *authmanGroup.AuthmanGroup)
			memberships, _ := app.storage.FindGroupMemberships(clientID, model.MembershipFilter{
				GroupIDs: []string{authmanGroup.ID},
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
			_, err := app.storage.SaveGroupMembershipByExternalID(clientID, authmanGroup.ID, adminMember.ExternalID, userID, nil, nil, email, name, nil, nil, true)
			if err != nil {
				log.Printf("Error saving admin membership with missing info for external ID %s in Authman %s: %s\n", adminMember.ExternalID, *authmanGroup.AuthmanGroup, err)
			} else {
				log.Printf("Update admin member %s for group '%s'", adminMember.ExternalID, authmanGroup.Title)
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
						ClientID:   clientID,
						GroupID:    authmanGroup.ID,
						ExternalID: member.ExternalID,
						Email:      email,
						Name:       name,
						SyncID:     &syncID,
					})
				}
			}

			if len(updateOperations) > 0 {
				err = app.storage.BulkUpdateGroupMembershipsByExternalID(clientID, authmanGroup.ID, updateOperations, true)
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
	admin := false
	deleteCount, err := app.storage.DeleteUnsyncedGroupMemberships(clientID, authmanGroup.ID, syncID, &admin)
	if err != nil {
		log.Printf("Error deleting removed memberships in Authman %s\n", *authmanGroup.AuthmanGroup)
	} else {
		log.Printf("%d memberships removed from Authman %s\n", deleteCount, *authmanGroup.AuthmanGroup)
	}

	err = app.storage.UpdateGroupStats(nil, clientID, authmanGroup.ID, false, false, true, true)
	if err != nil {
		log.Printf("Error updating group stats for '%s' - %s", *authmanGroup.AuthmanGroup, err)
	}

	return nil
}

func (app *Application) sendGroupNotification(clientID string, notification model.GroupNotification) error {
	memberStatuses := notification.MemberStatuses
	if len(memberStatuses) == 0 {
		memberStatuses = []string{"admin", "member"}
	}

	members, err := app.findGroupMemberships(clientID, model.MembershipFilter{
		GroupIDs: []string{notification.GroupID},
		UserIDs:  notification.Members.ToUserIDs(),
		Statuses: memberStatuses,
	})

	if err != nil {
		return err
	}

	app.sendNotification(members.GetMembersAsNotificationRecipients(func(member model.GroupMembership) (bool, bool) {
		return true, true // Should it be a separate notification preference?
	}), notification.Topic, notification.Subject, notification.Body, notification.Data, app.config.AppID, app.config.OrgID)

	return nil
}

func (app *Application) sendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string, appID string, orgID string) {
	app.notifications.SendNotification(recipients, topic, title, text, data, nil, appID, orgID)
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

func (app *Application) findGroupMembership(clientID string, groupID string, userID string) (*model.GroupMembership, error) {
	return app.storage.FindGroupMembership(clientID, groupID, userID)
}

func (app *Application) getResearchProfileUserCount(clientID string, current *model.User, researchProfile map[string]map[string][]string) (int64, error) {
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
