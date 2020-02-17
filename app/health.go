// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *App) CreateHealth(health *model.Health) (*model.Health, *model.AppError) {

	rhealth, err := a.Srv().Store.Health().Save(health)
	if err != nil {
		return nil, err
	}

	return rhealth, nil
}

func (a *App) UpdateHealth(health *model.Health) (*model.Health, *model.AppError) {
	oldHealth, err := a.GetHealth(health.Id)
	if err != nil {
		return nil, err
	}

	oldHealth.Height = health.Height
	oldHealth.Weight = health.Weight
	oldHealth.MeasureAt = health.MeasureAt

	oldHealth, err = a.updateHealthUnsanitized(oldHealth)
	if err != nil {
		return health, err
	}

	return oldHealth, nil
}

func (a *App) updateHealthUnsanitized(health *model.Health) (*model.Health, *model.AppError) {
	return a.Srv().Store.Health().Update(health)
}

func (a *App) PatchHealth(healthId string, patch *model.HealthPatch) (*model.Health, *model.AppError) {
	health, err := a.GetHealth(healthId)
	if err != nil {
		return nil, err
	}

	health.Patch(patch)

	updatedHealth, err := a.UpdateHealth(health)
	if err != nil {
		return nil, err
	}

	return updatedHealth, nil
}

func (a *App) GetHealth(healthId string) (*model.Health, *model.AppError) {
	return a.Srv().Store.Health().Get(healthId)
}

func (a *App) GetHealths(kidId string) ([]*model.Health, *model.AppError) {
	return a.Srv().Store.Health().GetAll(kidId)
}
