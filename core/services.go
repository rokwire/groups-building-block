package core

import (
	"fmt"
	"groups/core/model"
	"groups/driven/notifications"
	"log"
	"strings"
)

func (app *Application) applyDataProtection(current *model.User, group model.Group) map[string]interface{} {
	//1 apply data protection for "anonymous"
	if current == nil {
		return app.protectDataForAnonymous(group)
	}

	//2 apply data protection for "group admin"
	if group.IsGroupAdmin(current.ID) {
		return app.protectDataForAdmin(group)
	}

	//3 apply data protection for "group member"
	if group.IsGroupMember(current.ID) {
		return app.protectDataForMember(group)
	}

	//4 apply data protection for "group pending"
	if group.IsGroupPending(current.ID) {
		return app.protectDataForPending(*current, group)
	}

	//5 apply data protection for "group rejected"
	if group.IsGroupRejected(current.ID) {
		return app.protectDataForRejected(*current, group)
	}

	//6 apply data protection for "NOT member" - treat it as anonymous user
	return app.protectDataForAnonymous(group)
}

func (app *Application) protectDataForAnonymous(group model.Group) map[string]interface{} {
	switch group.Privacy {
	case "public":
		item := make(map[string]interface{})

		item["id"] = group.ID
		item["category"] = group.Category
		item["title"] = group.Title
		item["privacy"] = group.Privacy
		item["description"] = group.Description
		item["image_url"] = group.ImageURL
		item["web_url"] = group.WebURL
		item["tags"] = group.Tags
		item["membership_questions"] = group.MembershipQuestions
		item["authman_enabled"] = group.AuthmanEnabled
		item["authman_group"] = group.AuthmanGroup

		//members
		membersCount := len(group.Members)
		var membersItems []map[string]interface{}
		if membersCount > 0 {
			for _, current := range group.Members {
				if current.Status == "admin" || current.Status == "member" {
					mItem := make(map[string]interface{})
					mItem["id"] = current.ID
					mItem["user_id"] = current.User.ID
					mItem["name"] = current.Name
					mItem["email"] = current.Email
					mItem["photo_url"] = current.PhotoURL
					mItem["status"] = current.Status
					membersItems = append(membersItems, mItem)
				}
			}
		}
		item["members"] = membersItems

		item["date_created"] = group.DateCreated
		item["date_updated"] = group.DateUpdated

		//TODO add events and posts when they appear
		return item
	case "private":
		//we must protect events, posts and members(only admins are visible)
		item := make(map[string]interface{})

		item["id"] = group.ID
		item["category"] = group.Category
		item["title"] = group.Title
		item["privacy"] = group.Privacy
		item["description"] = group.Description
		item["image_url"] = group.ImageURL
		item["web_url"] = group.WebURL
		item["tags"] = group.Tags
		item["membership_questions"] = group.MembershipQuestions

		//members
		membersCount := len(group.Members)
		var membersItems []map[string]interface{}
		if membersCount > 0 {
			for _, current := range group.Members {
				if current.Status == "admin" {
					mItem := make(map[string]interface{})
					mItem["id"] = current.ID
					mItem["user_id"] = current.User.ID
					mItem["name"] = current.Name
					mItem["email"] = current.Email
					mItem["photo_url"] = current.PhotoURL
					mItem["status"] = current.Status
					membersItems = append(membersItems, mItem)
				}
			}
		}
		item["members"] = membersItems

		item["date_created"] = group.DateCreated
		item["date_updated"] = group.DateUpdated

		return item
	}
	return nil
}

func (app *Application) protectDataForAdmin(group model.Group) map[string]interface{} {
	item := make(map[string]interface{})

	item["id"] = group.ID
	item["category"] = group.Category
	item["title"] = group.Title
	item["privacy"] = group.Privacy
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			mItem := make(map[string]interface{})
			mItem["id"] = current.ID
			mItem["user_id"] = current.User.ID
			mItem["name"] = current.Name
			mItem["email"] = current.Email
			mItem["photo_url"] = current.PhotoURL
			mItem["status"] = current.Status
			mItem["rejected_reason"] = current.RejectReason

			//member answers
			answersCount := len(current.MemberAnswers)
			var answersItems []map[string]interface{}
			if answersCount > 0 {
				for _, cAnswer := range current.MemberAnswers {
					aItem := make(map[string]interface{})
					aItem["question"] = cAnswer.Question
					aItem["answer"] = cAnswer.Answer
					answersItems = append(answersItems, aItem)
				}
			}
			mItem["member_answers"] = answersItems

			mItem["date_created"] = current.DateCreated
			mItem["date_updated"] = current.DateUpdated
			membersItems = append(membersItems, mItem)
		}
	}
	item["members"] = membersItems

	item["date_created"] = group.DateCreated
	item["date_updated"] = group.DateUpdated

	//TODO add events and posts when they appear
	return item
}

func (app *Application) protectDataForMember(group model.Group) map[string]interface{} {
	item := make(map[string]interface{})

	item["id"] = group.ID
	item["category"] = group.Category
	item["title"] = group.Title
	item["privacy"] = group.Privacy
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.Status == "admin" || current.Status == "member" {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
				mItem["user_id"] = current.User.ID
				mItem["name"] = current.Name
				mItem["email"] = current.Email
				mItem["photo_url"] = current.PhotoURL
				mItem["status"] = current.Status
				membersItems = append(membersItems, mItem)
			}
		}
	}
	item["members"] = membersItems

	item["date_created"] = group.DateCreated
	item["date_updated"] = group.DateUpdated

	//TODO add events and posts when they appear
	return item
}

func (app *Application) protectDataForPending(user model.User, group model.Group) map[string]interface{} {
	item := make(map[string]interface{})

	item["id"] = group.ID
	item["category"] = group.Category
	item["title"] = group.Title
	item["privacy"] = group.Privacy
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.User.ID == user.ID {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
				mItem["user_id"] = current.User.ID
				mItem["name"] = current.Name
				mItem["email"] = current.Email
				mItem["photo_url"] = current.PhotoURL
				mItem["status"] = current.Status
				membersItems = append(membersItems, mItem)
			}
		}
	}
	item["members"] = membersItems

	item["date_created"] = group.DateCreated
	item["date_updated"] = group.DateUpdated

	return item
}

func (app *Application) protectDataForRejected(user model.User, group model.Group) map[string]interface{} {
	item := make(map[string]interface{})

	item["id"] = group.ID
	item["category"] = group.Category
	item["title"] = group.Title
	item["privacy"] = group.Privacy
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.User.ID == user.ID {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
				mItem["user_id"] = current.User.ID
				mItem["name"] = current.Name
				mItem["email"] = current.Email
				mItem["photo_url"] = current.PhotoURL
				mItem["status"] = current.Status
				mItem["rejected_reason"] = current.RejectReason
				membersItems = append(membersItems, mItem)
			}
		}
	}
	item["members"] = membersItems

	item["date_created"] = group.DateCreated
	item["date_updated"] = group.DateUpdated

	return item
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

func (app *Application) createGroup(clientID string, current model.User, title string, description *string, category string, tags []string, privacy string,
	creatorName string, creatorEmail string, creatorPhotoURL string, imageURL *string, webURL *string, membershipQuestions []string, authmanEnabled bool, authmanGroup *string) (*string, *GroupError) {
	insertedID, err := app.storage.CreateGroup(clientID, title, description, category, tags, privacy, current.ID, creatorName,
		creatorEmail, creatorPhotoURL, imageURL, webURL, membershipQuestions, authmanEnabled, authmanGroup)
	if err != nil {
		return nil, err
	}
	return insertedID, nil
}

func (app *Application) updateGroup(clientID string, current *model.User, id string, category string, title string, privacy string, description *string,
	imageURL *string, webURL *string, tags []string, membershipQuestions []string, authmanEnabled bool, authmanGroup *string) *GroupError {
	err := app.storage.UpdateGroup(clientID, id, category, title, privacy, description, imageURL, webURL, tags, membershipQuestions, authmanEnabled, authmanGroup)
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

func (app *Application) getGroups(clientID string, current *model.User, category *string, privacy *string, title *string, offset *int64, limit *int64, order *string) ([]map[string]interface{}, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, category, privacy, title, offset, limit, order)
	if err != nil {
		return nil, err
	}

	visibleGroups := make([]model.Group, 0)
	for _, group := range groups {

		if group.Privacy != "private" || group.IsGroupAdminOrMember(current.ID) || (title != nil && strings.EqualFold(group.Title, *title)) {
			visibleGroups = append(visibleGroups, group)
		}
	}

	//apply data protection
	groupsList := make([]map[string]interface{}, len(visibleGroups))
	for i, item := range visibleGroups {
		groupsList[i] = app.applyDataProtection(current, item)
	}

	return groupsList, nil
}

func (app *Application) getUserGroups(clientID string, current *model.User) ([]map[string]interface{}, error) {
	// find the user groups
	groups, err := app.storage.FindUserGroups(clientID, current.ID)
	if err != nil {
		return nil, err
	}

	//apply data protection
	groupsList := make([]map[string]interface{}, len(groups))
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

func (app *Application) getGroup(clientID string, current *model.User, id string) (map[string]interface{}, error) {
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

	return res, nil
}

func (app *Application) createPendingMember(clientID string, current model.User, groupID string, name string, email string, photoURL string, memberAnswers []model.MemberAnswer) error {
	err := app.storage.CreatePendingMember(clientID, groupID, current.ID, name, email, photoURL, memberAnswers)
	if err != nil {
		return err
	}

	group, err := app.storage.FindGroup(clientID, groupID)
	if err == nil && group != nil {
		members := group.Members
		if len(members) > 0 {
			recipients := []notifications.Recipient{}
			for _, member := range members {
				if member.Status == "admin" {
					recipients = append(recipients, notifications.Recipient{
						UserID: member.User.ID,
						Name:   member.Name,
					})
				}
			}
			if len(recipients) > 0 {
				topic := "group.invitations"
				app.notifications.SendNotification(
					recipients,
					&topic,
					"Illinois",
					fmt.Sprintf("New membership request for '%s' group has been submitted", group.Title),
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

	return nil
}

func (app *Application) deletePendingMember(clientID string, current model.User, groupID string) error {
	err := app.storage.DeletePendingMember(clientID, groupID, current.ID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) deleteMember(clientID string, current model.User, groupID string) error {
	err := app.storage.DeleteMember(clientID, groupID, current.ID, false)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) applyMembershipApproval(clientID string, current model.User, membershipID string, approve bool, rejectReason string) error {
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
							UserID: member.User.ID,
							Name:   member.Name,
						},
					},
					&topic,
					"Illinois",
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
							UserID: member.User.ID,
							Name:   member.Name,
						},
					},
					&topic,
					"Illinois",
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

	return nil
}

func (app *Application) deleteMembership(clientID string, current model.User, membershipID string) error {
	err := app.storage.DeleteMembership(clientID, current.ID, membershipID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) updateMembership(clientID string, current model.User, membershipID string, status string) error {
	err := app.storage.UpdateMembership(clientID, current.ID, membershipID, status)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getEvents(clientID string, groupID string) ([]model.Event, error) {
	events, err := app.storage.FindEvents(clientID, groupID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (app *Application) createEvent(clientID string, current model.User, eventID string, group *model.Group) error {
	err := app.storage.CreateEvent(clientID, eventID, group.ID)
	if err != nil {
		return err
	}

	recipients := group.GetMembersAsNotificationRecipients(&current.ID)
	topic := "group.events"
	err = app.notifications.SendNotification(
		recipients,
		&topic,
		"Illinois",
		fmt.Sprintf("New event has been published in '%s' group", group.Title),
		map[string]string{
			"type":        "group",
			"operation":   "event_created",
			"entity_type": "group",
			"entity_id":   group.ID,
			"entity_name": group.Title,
		},
	)
	if err != nil {
		log.Printf("error while sending notification for new event: %s", err) // dont fail
	}
	return nil
}

func (app *Application) deleteEvent(clientID string, current model.User, eventID string, groupID string) error {
	err := app.storage.DeleteEvent(clientID, eventID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getPosts(clientID string, current *model.User, groupID string, filterPrivatePostsValue *bool, offset *int64, limit *int64, order *string) ([]*model.Post, error) {
	return app.storage.FindPosts(clientID, current, groupID, filterPrivatePostsValue, offset, limit, order)
}

func (app *Application) getUserPostCount(clientID string, userID string) (map[string]interface{}, error) {
	return app.storage.GetUserPostCount(clientID, userID)
}

func (app *Application) createPost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {
	post, err := app.storage.CreatePost(clientID, current, post)
	if err != nil {
		return nil, err
	}
	if post.ParentID == nil {
		recipients := group.GetMembersAsNotificationRecipients(&current.ID)
		topic := "group.posts"
		err = app.notifications.SendNotification(
			recipients,
			&topic,
			"Illinois",
			fmt.Sprintf("New post has been published in '%s' group", group.Title),
			map[string]string{
				"type":        "group",
				"operation":   "post_created",
				"entity_type": "group",
				"entity_id":   group.ID,
				"entity_name": group.Title,
			},
		)
		if err != nil {
			log.Printf("error while sending notification for new post: %s", err) // dont fail
		}
	}
	return post, nil
}

func (app *Application) updatePost(clientID string, current *model.User, post *model.Post) (*model.Post, error) {
	return app.storage.UpdatePost(clientID, current.ID, post)
}

func (app *Application) deletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	return app.storage.DeletePost(clientID, userID, groupID, postID, force)
}

func (app *Application) synchronizeAuthman(clientID string) error {
	log.Printf("Authman synchronization started")
	defer log.Printf("Authman synchronization finished")

	authmanGroups, err := app.storage.FindAuthmanGroups(clientID)
	if err != nil {
		return err
	}

	if len(authmanGroups) > 0 {
		for _, authmanGroup := range authmanGroups {
			if authmanGroup.AuthmanGroup != nil {
				authmanExternalIDs, authmanErr := app.authman.RetrieveAuthmanGroupMembers(*authmanGroup.AuthmanGroup)
				if authmanErr != nil {
					log.Printf("Error on requesting Authman for %s: %s", *authmanGroup.AuthmanGroup, authmanErr)
					continue
				}

				userIDMapping := map[string]interface{}{}
				for _, externalID := range authmanExternalIDs {
					user, userErr := app.storage.FindUser(clientID, externalID, true)
					if authmanErr != nil {
						log.Printf("Error on getting user %s for Authman %s: %s", externalID, *authmanGroup.AuthmanGroup, userErr)
						continue
					}

					if user != nil {
						// Add missed members
						member := authmanGroup.GetMemberByUserID(user.ID)
						if member != nil {
							if member.IsPendingMember() || member.IsRejected() {
								delErr := app.storage.DeleteMember(clientID, authmanGroup.ID, member.User.ID, true)
								if delErr != nil {
									log.Printf("Error on deleting user %s from group '%s' for Authman '%s': %s", externalID, authmanGroup.Title, *authmanGroup.AuthmanGroup, delErr)
									continue
								} else {
									log.Printf("User(%s) has been deleted from '%s' successfully", member.User.ID, authmanGroup.Title)
								}

								memberErr := app.storage.CreateMember(clientID, authmanGroup.ID, user.ID, externalID, user.Name, user.Email, "", authmanGroup.CreateMembershipEmptyAnswers())
								if memberErr != nil {
									log.Printf("Error on creating user as member %s from group '%s' for Authman %s: %s", externalID, authmanGroup.Title, *authmanGroup.AuthmanGroup, memberErr)
									continue
								} else {
									log.Printf("User(%s, %s, %s) has been recreated as regular member of '%s' successfully", member.User.ID, user.ExternalID, user.Email, authmanGroup.Title)
								}
							} else {
								log.Printf("User(%s, %s, %s) is already a member or admin of '%s'", user.ID, user.ExternalID, user.Email, authmanGroup.Title)
							}
						} else {
							memberErr := app.storage.CreateMember(clientID, authmanGroup.ID, user.ID, externalID, user.Name, user.Email, "", authmanGroup.CreateMembershipEmptyAnswers())
							if memberErr != nil {
								log.Printf("Error on creating user as member %s from group '%s' for Authman %s: %s", externalID, authmanGroup.Title, *authmanGroup.AuthmanGroup, memberErr)
								continue
							} else {
								log.Printf("User(%s, %s, %s) has been created as regular member of '%s' successfully", externalID, user.Name, user.Email, authmanGroup.Title)
							}
						}

						userIDMapping[user.ID] = true
					} else {
						memberErr := app.storage.CreateMember(clientID, authmanGroup.ID, "", externalID, "", "", "", authmanGroup.CreateMembershipEmptyAnswers())
						if memberErr != nil {
							log.Printf("Error on creating dummy member %s from group %s for Authman %s: %s", externalID, authmanGroup.ID, *authmanGroup.AuthmanGroup, memberErr)
							continue
						} else {
							log.Printf("Empty User(ExternalID: %s) has been created as regular member of '%s' successfully", externalID, authmanGroup.Title)
						}
					}
				}

				// Delete the rest of non mapped members
				if len(authmanGroup.Members) > 0 {
					for _, member := range authmanGroup.Members {
						val := userIDMapping[member.User.ID]
						if val == nil && !member.IsAdmin() {
							err := app.storage.DeleteMember(clientID, authmanGroup.ID, member.User.ID, true)
							if err != nil {
								log.Printf("Error on deleting user as member %s from group %s for Authman %s: %s", member.User.ID, authmanGroup.Title, *authmanGroup.AuthmanGroup, err)
								continue
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (app *Application) sendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string) error {
	return app.notifications.SendNotification(recipients, topic, title, text, data)
}
