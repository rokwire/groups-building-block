package core

import "groups/core/model"

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getGroupCategories() ([]string, error) {
	groupCategories, err := app.storage.ReadAllGroupCategories()
	if err != nil {
		return nil, err
	}
	return groupCategories, nil
}

func (app *Application) createGroup(current model.User, title string, description *string, category string, tags []string, privacy string,
	creatorName string, creatorEmail string, creatorPhotoURL string) (*string, error) {
	insertedID, err := app.storage.CreateGroup(title, description, category, tags, privacy,
		current.ID, creatorName, creatorEmail, creatorPhotoURL)
	if err != nil {
		return nil, err
	}
	return insertedID, nil
}

func (app *Application) getGroups(category *string) ([]map[string]interface{}, error) {

	//TODO - data protection
	groups, err := app.storage.FindGroups(category)
	if err != nil {
		return nil, err
	}

	groupsList := make([]map[string]interface{}, len(groups))

	/*	inrec, err := json.Marshal(groups)
		if err != nil {
			log.Printf("error marshaling the groups - %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var response []map[string]interface{}
		err = json.Unmarshal(inrec, &response)
		if err != nil {
			log.Printf("error unmarshaling the groups - %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, groupMap := range response {
			members := groupMap["members"].([]interface{})
			for _, membersItem := range members {
				delete(membersItem.(map[string]interface{}), "user")
				delete(membersItem.(map[string]interface{}), "group")
			}
		} */

	//TODO
	return groupsList, nil
}
