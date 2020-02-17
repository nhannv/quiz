// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *App) CreateMenu(menu *model.Menu) (*model.Menu, *model.AppError) {
	class, err := a.GetClass(menu.ClassId)
	if err != nil {
		return nil, err
	}

	if err = class.IsBelongToSchool(a.Session().SchoolId); err != nil {
		return nil, model.NewAppError("CreateMenu", "api.menu.create_menu.invalid_class.app_error", nil, "", http.StatusBadRequest)
	}

	rmenu, err := a.Srv().Store.Menu().Save(menu)
	if err != nil {
		return nil, err
	}

	return rmenu, nil
}

func (a *App) CreateActivityNote(note *model.ActivityNote) (*model.ActivityNote, *model.AppError) {
	rnote, err := a.Srv().Store.ActivityNote().Save(note)
	if err != nil {
		return nil, err
	}

	return rnote, nil
}

func (a *App) UpdateMenu(menu *model.Menu) (*model.Menu, *model.AppError) {
	oldMenu, err := a.GetMenu(menu.Id)
	if err != nil {
		return nil, err
	}

	oldMenu.Week = menu.Week
	oldMenu.Year = menu.Year
	oldMenu.FoodName = menu.FoodName
	oldMenu.Description = menu.Description
	oldMenu.WeekDay = menu.WeekDay
	oldMenu.StartTime = menu.StartTime

	oldMenu, err = a.updateMenuUnsanitized(oldMenu)
	if err != nil {
		return menu, err
	}

	return oldMenu, nil
}

func (a *App) updateMenuUnsanitized(menu *model.Menu) (*model.Menu, *model.AppError) {
	return a.Srv().Store.Menu().Update(menu)
}

func (a *App) PatchMenu(menuId string, patch *model.MenuPatch) (*model.Menu, *model.AppError) {
	menu, err := a.GetMenu(menuId)
	if err != nil {
		return nil, err
	}

	menu.Patch(patch)

	updatedMenu, err := a.UpdateMenu(menu)
	if err != nil {
		return nil, err
	}

	return updatedMenu, nil
}

func (a *App) GetMenu(menuId string) (*model.Menu, *model.AppError) {
	return a.Srv().Store.Menu().Get(menuId)
}

func (a *App) GetMenus(week int, year int) ([]*model.Menu, *model.AppError) {
	return a.Srv().Store.Menu().GetByWeek(week, year)
}
