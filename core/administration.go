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
	"groups/core/model"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/core-auth-library-go/v2/authutils"
	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logutils"
)

func (app *Application) getTODO() error {
	return nil
}

func (app *Application) getGroupsUnprotected(clientID string, filter model.GroupsFilter) ([]model.Group, error) {
	groups, err := app.storage.FindGroups(clientID, nil, filter)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (app *Application) getManagedGroupConfigs(clientID string) ([]model.ManagedGroupConfigData, error) {
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

func (app *Application) getConfig(configType string, appID string, orgID string) (*model.Config, error) {
	return app.storage.FindConfig(configType, appID, orgID)
}

func (app *Application) admGetConfig(id string, appID string, orgID string, system bool) (*model.Config, error) {
	config, err := app.storage.FindConfigByID(id)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	ok, err := app.checkConfigAccess(config, appID, orgID, system)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}
	if !ok {
		return nil, errors.ErrorAction(logutils.ActionGet, model.TypeConfig, logutils.StringArgs("invalid claims"))
	}

	return config, nil
}

func (app *Application) admGetConfigs(configType *string, appID string, orgID string, system bool) ([]model.Config, error) {
	configs, err := app.storage.FindConfigs(configType)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	if !system {
		allowedConfigs := make([]model.Config, 0)
		for _, config := range configs {
			ok, _ := app.checkConfigAccess(&config, appID, orgID, system)
			if ok {
				allowedConfigs = append(allowedConfigs, config)
			}
		}
		return allowedConfigs, nil
	}

	return configs, nil
}

func (app *Application) admCreateConfig(config model.Config, appID string, orgID string, system bool) (*model.Config, error) {
	_, err := app.checkConfigAccess(&config, appID, orgID, system)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}

	config.ID = uuid.NewString()
	config.DateCreated = time.Now().UTC()
	err = app.storage.InsertConfig(config)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}
	return &config, nil
}

func (app *Application) admUpdateConfig(config model.Config, appID string, orgID string, system bool) error {
	oldConfig, err := app.storage.FindConfig(config.Type, config.AppID, config.OrgID)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}
	if oldConfig == nil {
		return errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"type": config.Type, "app_id": config.AppID, "org_id": config.OrgID})
	}

	if !system && oldConfig.System {
		return errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
	}
	_, err = app.checkConfigAccess(&config, appID, orgID, system)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}

	now := time.Now().UTC()
	config.DateUpdated = &now

	err = app.storage.UpdateConfig(config)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeConfig, nil, err)
	}
	return nil
}

func (app *Application) admDeleteConfig(id string, appID string, orgID string, system bool) error {
	config, err := app.storage.FindConfigByID(id)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	ok, err := app.checkConfigAccess(config, appID, orgID, system)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionValidate, "config access", nil, err)
	}
	if !ok {
		return errors.ErrorAction(logutils.ActionDelete, model.TypeConfig, logutils.StringArgs("invalid claims"))
	}

	err = app.storage.DeleteConfig(id)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeConfig, nil, err)
	}
	return nil
}

// TODO: how to determine config access without application, organization data?
func (app *Application) checkConfigAccess(config *model.Config, appIDClaim string, orgIDClaim string, systemClaim bool) (bool, error) {
	if config == nil {
		return false, errors.ErrorData(logutils.StatusMissing, model.TypeConfig, nil)
	}

	claimsMatch := true
	sysOrgIDClaim := systemClaim
	if !systemClaim {
		// org admins: cannot manage system configs, can only manage configs for their orgID
		if config.System {
			return false, errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
		}
		if config.OrgID != orgIDClaim {
			return false, errors.ErrorData(logutils.StatusInvalid, model.TypeConfig, &logutils.FieldArgs{"org_id": config.OrgID})
		}
		sysOrgIDClaim = false
	} else {
		// system admins: configs access allowed for any orgID (can use "all" when using the system organization)
		organization, err := app.storage.FindOrganization(orgIDClaim)
		if err != nil {
			return false, errors.WrapErrorAction(logutils.ActionFind, model.TypeOrganization, &logutils.FieldArgs{"id": orgIDClaim}, err)
		}
		if organization == nil {
			return false, errors.ErrorData(logutils.StatusMissing, model.TypeOrganization, &logutils.FieldArgs{"id": orgIDClaim})
		}

		sysOrgIDClaim = organization.System
		if config.OrgID != authutils.AllOrgs {
			if config.OrgID != orgIDClaim {
				claimsMatch = false
			}
		} else if !organization.System {
			return false, errors.ErrorData(logutils.StatusInvalid, model.TypeConfig, &logutils.FieldArgs{"org_id": authutils.AllOrgs, "org.system": false})
		}
	}

	// all admins (including system admins): configs access allowed for any appID (can use "all" when using an admin application)
	application, err := app.storage.FindApplication(nil, appIDClaim)
	if err != nil {
		return false, errors.WrapErrorAction(logutils.ActionFind, model.TypeApplication, &logutils.FieldArgs{"id": appIDClaim}, err)
	}
	if application == nil {
		return false, errors.ErrorData(logutils.StatusMissing, model.TypeApplication, &logutils.FieldArgs{"id": appIDClaim})
	}
	if config.AppID != authutils.AllApps {
		// give access to any appID to system admins using the system organization and an admin application
		if (!application.Admin || !sysOrgIDClaim) && config.AppID != appIDClaim {
			claimsMatch = false
		}
	} else if !application.Admin {
		return false, errors.ErrorData(logutils.StatusInvalid, model.TypeConfig, &logutils.FieldArgs{"app_id": authutils.AllApps, "app.admin": false})
	}

	return claimsMatch, nil
}
