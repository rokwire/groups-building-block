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
func (app *Application) getCovid19Config() (*model.GroupsConfig, error) {
	/*config, err := app.storage.ReadCovid19Config()
	if err != nil {
		return nil, err
	}*/
	//return , nil
}
