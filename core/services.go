package core

import (
	"fmt"
	"github.com/google/uuid"
	"groups/driven/rewards"
	"sort"
	"time"

	"groups/core/model"
	"groups/driven/notifications"
	"log"

	"strings"
)

func (app *Application) applyDataProtection(current *model.User, group model.Group) map[string]interface{} {
	//1 apply data protection for "anonymous"
	if current == nil || current.IsAnonymous {
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
		item["hidden_for_search"] = group.HiddenForSearch
		item["description"] = group.Description
		item["image_url"] = group.ImageURL
		item["web_url"] = group.WebURL
		item["tags"] = group.Tags
		item["membership_questions"] = group.MembershipQuestions
		item["authman_enabled"] = group.AuthmanEnabled
		item["authman_group"] = group.AuthmanGroup
		item["only_admins_can_create_polls"] = group.OnlyAdminsCanCreatePolls
		item["block_new_membership_requests"] = group.BlockNewMembershipRequests
		item["attendance_group"] = group.AttendanceGroup

		// Unauthenticated users must not see members
		item["members"] = []map[string]interface{}{}

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
		item["hidden_for_search"] = group.HiddenForSearch
		item["description"] = group.Description
		item["image_url"] = group.ImageURL
		item["web_url"] = group.WebURL
		item["tags"] = group.Tags
		item["membership_questions"] = group.MembershipQuestions
		item["authman_enabled"] = group.AuthmanEnabled
		item["authman_group"] = group.AuthmanGroup
		item["only_admins_can_create_polls"] = group.OnlyAdminsCanCreatePolls
		item["block_new_membership_requests"] = group.BlockNewMembershipRequests
		item["attendance_group"] = group.AttendanceGroup

		// Unauthenticated users must not see members
		item["members"] = []map[string]interface{}{}

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
	item["hidden_for_search"] = group.HiddenForSearch
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup
	item["only_admins_can_create_polls"] = group.OnlyAdminsCanCreatePolls
	item["block_new_membership_requests"] = group.BlockNewMembershipRequests
	item["attendance_group"] = group.AttendanceGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			mItem := make(map[string]interface{})
			mItem["id"] = current.ID
			mItem["user_id"] = current.UserID
			mItem["external_id"] = current.ExternalID
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
			mItem["date_attended"] = current.DateAttended

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
	item["hidden_for_search"] = group.HiddenForSearch
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup
	item["only_admins_can_create_polls"] = group.OnlyAdminsCanCreatePolls
	item["block_new_membership_requests"] = group.BlockNewMembershipRequests
	item["attendance_group"] = group.AttendanceGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.Status == "admin" || current.Status == "member" {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
				mItem["user_id"] = current.UserID
				mItem["external_id"] = current.ExternalID
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
	item["hidden_for_search"] = group.HiddenForSearch
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup
	item["only_admins_can_create_polls"] = group.OnlyAdminsCanCreatePolls
	item["block_new_membership_requests"] = group.BlockNewMembershipRequests
	item["attendance_group"] = group.AttendanceGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.UserID == user.ID {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
				mItem["user_id"] = current.UserID
				mItem["external_id"] = current.ExternalID
				mItem["name"] = current.Name
				mItem["email"] = current.Email
				mItem["photo_url"] = current.PhotoURL
				mItem["status"] = current.Status
				membersItems = append(membersItems, mItem)
				break
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
	item["hidden_for_search"] = group.HiddenForSearch
	item["description"] = group.Description
	item["image_url"] = group.ImageURL
	item["web_url"] = group.WebURL
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions
	item["authman_enabled"] = group.AuthmanEnabled
	item["authman_group"] = group.AuthmanGroup
	item["only_admins_can_create_polls"] = group.OnlyAdminsCanCreatePolls
	item["block_new_membership_requests"] = group.BlockNewMembershipRequests
	item["attendance_group"] = group.AttendanceGroup

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.UserID == user.ID {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
				mItem["user_id"] = current.UserID
				mItem["external_id"] = current.ExternalID
				mItem["name"] = current.Name
				mItem["email"] = current.Email
				mItem["photo_url"] = current.PhotoURL
				mItem["status"] = current.Status
				mItem["rejected_reason"] = current.RejectReason
				membersItems = append(membersItems, mItem)
				break
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

func (app *Application) createGroup(clientID string, current *model.User, title string, description *string, category string, tags []string,
	privacy string, hiddenForSearch bool,
	creatorName string, creatorEmail string, creatorPhotoURL string, imageURL *string, webURL *string, membershipQuestions []string, authmanEnabled bool,
	authmanGroup *string, onlyAdminsCanCreatePolls bool, attendanceGroup bool) (*string, *GroupError) {
	insertedID, err := app.storage.CreateGroup(clientID, title, description, category, tags, privacy, hiddenForSearch, current.ID, creatorName,
		creatorEmail, creatorPhotoURL, imageURL, webURL, membershipQuestions, authmanEnabled, authmanGroup, onlyAdminsCanCreatePolls, attendanceGroup)
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

func (app *Application) updateGroup(clientID string, current *model.User, id string, category string, title string,
	privacy string, hiddenForSearch bool, description *string,
	imageURL *string, webURL *string, tags []string, membershipQuestions []string,
	authmanEnabled bool, authmanGroup *string, onlyAdminsCanCreatePolls bool, blockNewMembershipRequests bool, attendanceGroup bool) *GroupError {

	err := app.storage.UpdateGroup(clientID, id, category, title,
		privacy, hiddenForSearch, description, imageURL, webURL, tags, membershipQuestions,
		authmanEnabled, authmanGroup, onlyAdminsCanCreatePolls, blockNewMembershipRequests, attendanceGroup)
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

		if group.Privacy != "private" ||
			group.IsGroupAdminOrMember(current.ID) ||
			(title != nil && strings.EqualFold(group.Title, *title) && !group.HiddenForSearch) {
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

func (app *Application) getAllGroups(clientID string) ([]model.Group, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return groups, nil
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

func (app *Application) createPendingMember(clientID string, current *model.User, groupID string, name string, email string, photoURL string, memberAnswers []model.MemberAnswer) error {
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
						UserID: member.UserID,
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

func (app *Application) deletePendingMember(clientID string, current *model.User, groupID string) error {
	err := app.storage.DeletePendingMember(clientID, groupID, current.ID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) createMember(clientID string, current *model.User, groupID string, member *model.Member) error {
	return app.storage.CreateMember(clientID, current, groupID, member)
}

func (app *Application) deleteMember(clientID string, current *model.User, groupID string) error {
	err := app.storage.DeleteMember(clientID, groupID, current.ID, false)
	if err != nil {
		return err
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
							UserID: member.UserID,
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

func (app *Application) deleteMembership(clientID string, current *model.User, membershipID string) error {
	err := app.storage.DeleteMembership(clientID, current, membershipID)
	if err != nil {
		return err
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
			title := "Illinois"
			body := fmt.Sprintf("New post has been published in '%s' group", group.Title)
			if post.UseAsNotification {
				title = post.Subject
				body = post.Body
			}

			topic := "group.posts"
			err = app.notifications.SendNotification(
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
			if err != nil {
				log.Printf("error while sending notification for new post: %s", err) // dont fail
			}
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

func (app *Application) reportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post) error {

	err := app.storage.ReportPostAsAbuse(clientID, current.ID, group, post)
	if err != nil {
		log.Printf("error while reporting an abuse post: %s", err)
		return fmt.Errorf("error while reporting an abuse post: %s", err)
	}

	subject := "Report Abuse Post"
	body := fmt.Sprintf(`
	Group title: %s
	Post Title: %s
	Post Body: %s
	`, group.Title, post.Subject, post.Body)
	err = app.notifications.SendMail(app.config.ReportAbuseRecipientEmail, subject, body)
	if err != nil {
		log.Printf("error while reporting an abuse post: %s", err)
		return fmt.Errorf("error while reporting an abuse post: %s", err)
	}

	return nil
}

func (app *Application) deletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	return app.storage.DeletePost(clientID, userID, groupID, postID, force)
}

func (app *Application) findPolls(clientID string, current *model.User, groupID string, filterByToMembers bool, offset *int64, limit *int64, order *string) ([]model.Poll, error) {
	return app.storage.FindPolls(clientID, current, groupID, filterByToMembers, offset, limit, order)
}

func (app *Application) createPoll(clientID string, current *model.User, group *model.Group, poll *model.Poll) (*model.Poll, error) {
	poll, err := app.storage.CreatePoll(clientID, current, poll)
	if err != nil {
		return nil, err
	}

	handleNotification := func() {
		var recipients []notifications.Recipient
		if len(poll.ToMembersList) > 0 {
			recipients = poll.GetMembersAsNotificationRecipients(&current.ID)
		} else {
			recipients = group.GetMembersAsNotificationRecipients(&current.ID)
		}
		topic := "group.polls"
		err = app.notifications.SendNotification(
			recipients,
			&topic,
			"Illinois",
			fmt.Sprintf("New poll has been published in '%s' group", group.Title),
			map[string]string{
				"type":        "group",
				"operation":   "poll_created",
				"entity_type": "group",
				"entity_id":   group.ID,
				"entity_name": group.Title,
			},
		)
		if err != nil {
			log.Printf("error while sending notification for new poll: %s", err) // dont fail
		}
	}
	go handleNotification()

	return poll, err
}

func (app *Application) updatePoll(clientID string, current *model.User, group *model.Group, poll *model.Poll) (*model.Poll, error) {
	persistedPoll, err := app.storage.FindPoll(clientID, current, group.ID, poll.PollID, true)
	if persistedPoll != nil && err == nil {
		if group.IsGroupAdmin(current.ID) || persistedPoll.Creator.UserID == current.ID {
			persistedPoll.ToMembersList = poll.ToMembersList
			return app.storage.UpdatePoll(clientID, current, persistedPoll)
		}
		log.Printf("Only group admin or poll creator can delete it")
		return nil, fmt.Errorf("only group admin or poll creator can delete it")
	}
	return nil, err
}

func (app *Application) deletePoll(clientID string, current *model.User, group *model.Group, pollID string) error {
	persistedPoll, err := app.storage.FindPoll(clientID, current, group.ID, pollID, true)
	if persistedPoll != nil && err == nil {
		if group.IsGroupAdmin(current.ID) || persistedPoll.Creator.UserID == current.ID {
			return app.storage.DeletePoll(clientID, current, group.ID, pollID)
		}
		log.Printf("Only group admin or poll creator can delete it")
		return fmt.Errorf("only group admin or poll creator can delete it")
	}
	return err
}

func (app *Application) synchronizeAuthman(clientID string) error {
	log.Printf("Global Authman synchronization started")
	defer log.Printf("Global Authman synchronization finished")

	giesGroups, err := app.authman.RetrieveAuthmanGiesGroups()
	if err != nil {
		return fmt.Errorf("error on requesting Authman for GIES groups: %s", err)
	}

	if giesGroups != nil && len(giesGroups.WsFindGroupsResults.GroupResults) > 0 {
		for _, giesGroup := range giesGroups.WsFindGroupsResults.GroupResults {
			storedGiesGroup, err := app.storage.FindAuthmanGroupByKey(clientID, giesGroup.Name)
			if err != nil {
				return fmt.Errorf("error on requesting Authman for GIES groups: %s", err)
			}

			if storedGiesGroup == nil {
				title := giesGroup.Description
				if len(title) == 0 {
					title = giesGroup.DisplayName
				}
				emptyText := ""
				_, err := app.storage.CreateGroup(clientID, title, &title, "", nil, "private", true,
					"", "", "", "", &emptyText, &emptyText, nil,
					true, &giesGroup.Name, false, false)
				if err != nil {
					return fmt.Errorf("error on create Authman GIES group: '%s' - %s", giesGroup.Name, err)
				}

				log.Printf("Created new `%s` group", title)
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

func (app *Application) synchronizeAuthmanGroup(clientID string, authmanGroup *model.Group) error {

	log.Printf("Authman synchronization for group %s started", authmanGroup.Title)
	defer log.Printf("Authman synchronization for group %s finished", authmanGroup.Title)

	defaultAdminsMapping := map[string]bool{}
	if len(app.config.AuthmanAdminUINList) > 0 {
		for _, externalID := range app.config.AuthmanAdminUINList {
			defaultAdminsMapping[externalID] = true
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
				break // TBD: Why? This flow looks complicated and needs to be revised and redesign.
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
							members[i] = member
							found = true
							log.Printf("set user(%s, %s, %s) to 'admin' in '%s'", member.UserID, member.Name, member.Email, authmanGroup.Title)
							break
						}
					}
					if !found {
						members = append(members, member)
						log.Printf("add remaining admin user(%s, %s, %s) to '%s'", member.UserID, member.Name, member.Email, authmanGroup.Title)
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

func (app *Application) sendNotification(recipients []notifications.Recipient, topic *string, title string, text string, data map[string]string) error {
	return app.notifications.SendNotification(recipients, topic, title, text, data)
}
