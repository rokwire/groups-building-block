package core

import "groups/core/model"

func (app *Application) getTODO() error {
	//TODO
	return nil
}

func (app *Application) updateConfig(config *model.GroupsConfig) error {
	err := app.storage.SaveConfig(config)
	if err != nil {
		return err
	}

	return nil
}
func (app *Application) getConfig() (*model.GroupsConfig, error) {
	config, err := app.storage.ReadConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}
