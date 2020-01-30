// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlHealthStore struct {
	SqlStore
}

func NewSqlHealthStore(sqlStore SqlStore) store.HealthStore {
	s := &SqlHealthStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Health{}, "Healths").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Height").SetMaxSize(4)
		table.ColMap("Weight").SetMaxSize(4)
		table.ColMap("MeasureAt").SetMaxSize(20)
		table.ColMap("KidId").SetMaxSize(26)
	}

	return s
}

func (s SqlHealthStore) CreateIndexesIfNotExists() {
	s.RemoveIndexIfExists("idx_healths_description", "Healths")
	s.CreateIndexIfNotExists("idx_healths_update_at", "Healths", "UpdateAt")
	s.CreateIndexIfNotExists("idx_healths_create_at", "Healths", "CreateAt")
	s.CreateIndexIfNotExists("idx_healths_delete_at", "Healths", "DeleteAt")
	s.CreateIndexIfNotExists("idx_healths_kid_id", "Healths", "KidId")
}

func (s SqlHealthStore) Save(health *model.Health) (*model.Health, *model.AppError) {
	if len(health.Id) > 0 {
		return nil, model.NewAppError("SqlHealthStore.Save",
			"store.sql_health.save.existing.app_error", nil, "id="+health.Id, http.StatusBadRequest)
	}

	health.PreSave()

	if err := health.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(health); err != nil {
		return nil, model.NewAppError("SqlHealthStore.Save", "store.sql_health.save.app_error", nil, "id="+health.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return health, nil
}

func (s SqlHealthStore) Update(health *model.Health) (*model.Health, *model.AppError) {

	health.PreUpdate()

	if err := health.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.Health{}, health.Id)
	if err != nil {
		return nil, model.NewAppError("SqlHealthStore.Update", "store.sql_health.update.finding.app_error", nil, "id="+health.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlHealthStore.Update", "store.sql_health.update.find.app_error", nil, "id="+health.Id, http.StatusBadRequest)
	}

	oldHealth := oldResult.(*model.Health)
	health.CreateAt = oldHealth.CreateAt
	health.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(health)
	if err != nil {
		return nil, model.NewAppError("SqlHealthStore.Update", "store.sql_health.update.updating.app_error", nil, "id="+health.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlHealthStore.Update", "store.sql_health.update.app_error", nil, "id="+health.Id, http.StatusInternalServerError)
	}

	return health, nil
}

func (s SqlHealthStore) Get(id string) (*model.Health, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Health{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlHealthStore.Get", "store.sql_health.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlHealthStore.Get", "store.sql_health.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Health), nil
}

func (s SqlHealthStore) GetAll(kidId string) ([]*model.Health, *model.AppError) {
	var healths []*model.Health
	if _, err := s.GetReplica().Select(&healths,
		`SELECT Healths.* FROM Healths
		WHERE Healths.KidId = :KidId AND Healths.DeleteAt = 0`,
		map[string]interface{}{"KidId": kidId}); err != nil {
		return nil, model.NewAppError("SqlHealthStore.GetAll", "store.sql_health.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return healths, nil
}
