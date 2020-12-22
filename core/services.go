package core

import (
	"groups/core/model"
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
		item["members_count"] = group.MembersCount
		item["tags"] = group.Tags
		item["membership_questions"] = group.MembershipQuestions

		//members
		membersCount := len(group.Members)
		var membersItems []map[string]interface{}
		if membersCount > 0 {
			for _, current := range group.Members {
				if current.Status == "admin" || current.Status == "member" {
					mItem := make(map[string]interface{})
					mItem["id"] = current.ID
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
		item["members_count"] = group.MembersCount
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
	item["members_count"] = group.MembersCount
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			mItem := make(map[string]interface{})
			mItem["id"] = current.ID
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
	item["members_count"] = group.MembersCount
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.Status == "admin" || current.Status == "member" {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
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
	item["members_count"] = group.MembersCount
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.User.ID == user.ID {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
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
	item["members_count"] = group.MembersCount
	item["tags"] = group.Tags
	item["membership_questions"] = group.MembershipQuestions

	//members
	membersCount := len(group.Members)
	var membersItems []map[string]interface{}
	if membersCount > 0 {
		for _, current := range group.Members {
			if current.User.ID == user.ID {
				mItem := make(map[string]interface{})
				mItem["id"] = current.ID
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

func (app *Application) createGroup(clientID string, current model.User, title string, description *string, category string, tags []string, privacy string,
	creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error) {
	insertedID, err := app.storage.CreateGroup(clientID, title, description, category, tags, privacy,
		current.ID, creatorName, creatorEmail, creatorPhotoURL)
	if err != nil {
		return nil, err
	}
	return insertedID, nil
}

func (app *Application) updateGroup(clientID string, current *model.User, id string, category string, title string, privacy string, description *string,
	imageURL *string, webURL *string, tags []string, membershipQuestions []string) error {
	err := app.storage.UpdateGroup(clientID, id, category, title, privacy, description, imageURL, webURL, tags, membershipQuestions)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getGroups(clientID string, current *model.User, category *string) ([]map[string]interface{}, error) {
	// find the groups objects
	groups, err := app.storage.FindGroups(clientID, category)
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
	err := app.storage.DeleteMember(clientID, groupID, current.ID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) applyMembershipApproval(clientID string, current model.User, membershipID string, approve bool, rejectReason string) error {
	err := app.storage.ApplyMembershipApproval(clientID, membershipID, approve, rejectReason)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) deleteMembership(current model.User, membershipID string) error {
	err := app.storage.DeleteMembership(current.ID, membershipID)
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

func (app *Application) createEvent(clientID string, current model.User, eventID string, groupID string) error {
	err := app.storage.CreateEvent(clientID, eventID, groupID)
	if err != nil {
		return err
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
