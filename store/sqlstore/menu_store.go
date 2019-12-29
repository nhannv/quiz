// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlMenuStore struct {
	SqlStore
}

func NewSqlMenuStore(sqlStore SqlStore) store.MenuStore {
	s := &SqlMenuStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Menu{}, "Menus").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Week").SetMaxSize(2)
		table.ColMap("Year").SetMaxSize(4)
		table.ColMap("FoodName").SetMaxSize(24)
		table.ColMap("Description").SetMaxSize(128)
		table.ColMap("WeekDay").SetMaxSize(1)
		table.ColMap("StartTime").SetMaxSize(20)
		table.ColMap("Active").SetMaxSize(1)
		table.ColMap("ClassId").SetMaxSize(26)
	}

	return s
}

func (s SqlMenuStore) CreateIndexesIfNotExists() {
	s.RemoveIndexIfExists("idx_menus_description", "Menus")
	s.CreateIndexIfNotExists("idx_menus_update_at", "Menus", "UpdateAt")
	s.CreateIndexIfNotExists("idx_menus_create_at", "Menus", "CreateAt")
	s.CreateIndexIfNotExists("idx_menus_delete_at", "Menus", "DeleteAt")
	s.CreateIndexIfNotExists("idx_menus_week_id", "Menus", "Week")
	s.CreateIndexIfNotExists("idx_menus_year_id", "Menus", "Year")
	s.CreateIndexIfNotExists("idx_menus_class_id", "Menus", "ClassId")
}

func (s SqlMenuStore) Save(menu *model.Menu) (*model.Menu, *model.AppError) {
	if len(menu.Id) > 0 {
		return nil, model.NewAppError("SqlMenuStore.Save",
			"store.sql_menu.save.existing.app_error", nil, "id="+menu.Id, http.StatusBadRequest)
	}

	menu.PreSave()

	if err := menu.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(menu); err != nil {
		return nil, model.NewAppError("SqlMenuStore.Save", "store.sql_menu.save.app_error", nil, "id="+menu.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return menu, nil
}

func (s SqlMenuStore) Update(menu *model.Menu) (*model.Menu, *model.AppError) {

	menu.PreUpdate()

	if err := menu.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.Menu{}, menu.Id)
	if err != nil {
		return nil, model.NewAppError("SqlMenuStore.Update", "store.sql_menu.update.finding.app_error", nil, "id="+menu.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlMenuStore.Update", "store.sql_menu.update.find.app_error", nil, "id="+menu.Id, http.StatusBadRequest)
	}

	oldMenu := oldResult.(*model.Menu)
	menu.CreateAt = oldMenu.CreateAt
	menu.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(menu)
	if err != nil {
		return nil, model.NewAppError("SqlMenuStore.Update", "store.sql_menu.update.updating.app_error", nil, "id="+menu.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlMenuStore.Update", "store.sql_menu.update.app_error", nil, "id="+menu.Id, http.StatusInternalServerError)
	}

	return menu, nil
}

func (s SqlMenuStore) Get(id string) (*model.Menu, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Menu{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlMenuStore.Get", "store.sql_menu.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlMenuStore.Get", "store.sql_menu.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Menu), nil
}

func (s SqlMenuStore) GetByWeek(week int, year int) ([]*model.Menu, *model.AppError) {
	var menus []*model.Menu
	if _, err := s.GetReplica().Select(&menus, "SELECT Menus.* FROM Menus WHERE Menus.Week = :Week AND Menus.Year = :Year AND Menus.DeleteAt = 0", map[string]interface{}{"Week": week, "Year": year}); err != nil {
		return nil, model.NewAppError("SqlMenuStore.GetMenusByUserId", "store.sql_menu.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return menus, nil
}
