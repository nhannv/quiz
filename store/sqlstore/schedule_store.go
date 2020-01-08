// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlScheduleStore struct {
	SqlStore
}

func NewSqlScheduleStore(sqlStore SqlStore) store.ScheduleStore {
	s := &SqlScheduleStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Schedule{}, "Schedules").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Week").SetMaxSize(2)
		table.ColMap("Year").SetMaxSize(4)
		table.ColMap("Subject").SetMaxSize(24)
		table.ColMap("Description").SetMaxSize(128)
		table.ColMap("WeekDay").SetMaxSize(1)
		table.ColMap("StartTime").SetMaxSize(20)
		table.ColMap("EndTime").SetMaxSize(20)
		table.ColMap("ClassId").SetMaxSize(26)
	}

	return s
}

func (s SqlScheduleStore) CreateIndexesIfNotExists() {
	s.RemoveIndexIfExists("idx_schedules_description", "Schedules")
	s.CreateIndexIfNotExists("idx_schedules_update_at", "Schedules", "UpdateAt")
	s.CreateIndexIfNotExists("idx_schedules_create_at", "Schedules", "CreateAt")
	s.CreateIndexIfNotExists("idx_schedules_delete_at", "Schedules", "DeleteAt")
	s.CreateIndexIfNotExists("idx_schedules_week_id", "Schedules", "Week")
	s.CreateIndexIfNotExists("idx_schedules_year_id", "Schedules", "Year")
	s.CreateIndexIfNotExists("idx_schedules_class_id", "Schedules", "ClassId")
}

func (s SqlScheduleStore) Save(schedule *model.Schedule) (*model.Schedule, *model.AppError) {
	if len(schedule.Id) > 0 {
		return nil, model.NewAppError("SqlScheduleStore.Save",
			"store.sql_schedule.save.existing.app_error", nil, "id="+schedule.Id, http.StatusBadRequest)
	}

	schedule.PreSave()

	if err := schedule.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(schedule); err != nil {
		return nil, model.NewAppError("SqlScheduleStore.Save", "store.sql_schedule.save.app_error", nil, "id="+schedule.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return schedule, nil
}

func (s SqlScheduleStore) Update(schedule *model.Schedule) (*model.Schedule, *model.AppError) {

	schedule.PreUpdate()

	if err := schedule.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.Schedule{}, schedule.Id)
	if err != nil {
		return nil, model.NewAppError("SqlScheduleStore.Update", "store.sql_schedule.update.finding.app_error", nil, "id="+schedule.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlScheduleStore.Update", "store.sql_schedule.update.find.app_error", nil, "id="+schedule.Id, http.StatusBadRequest)
	}

	oldSchedule := oldResult.(*model.Schedule)
	schedule.CreateAt = oldSchedule.CreateAt
	schedule.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(schedule)
	if err != nil {
		return nil, model.NewAppError("SqlScheduleStore.Update", "store.sql_schedule.update.updating.app_error", nil, "id="+schedule.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlScheduleStore.Update", "store.sql_schedule.update.app_error", nil, "id="+schedule.Id, http.StatusInternalServerError)
	}

	return schedule, nil
}

func (s SqlScheduleStore) Get(id string) (*model.Schedule, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Schedule{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlScheduleStore.Get", "store.sql_schedule.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlScheduleStore.Get", "store.sql_schedule.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Schedule), nil
}

func (s SqlScheduleStore) GetByWeek(week int, year int, classId string) ([]*model.Schedule, *model.AppError) {
	var schedules []*model.Schedule
	if _, err := s.GetReplica().Select(&schedules, "SELECT Schedules.* FROM Schedules WHERE Schedules.Week = :Week AND Schedules.Year = :Year AND Schedules.ClassId = :ClassId AND Schedules.DeleteAt = 0", map[string]interface{}{"Week": week, "Year": year, "ClassId": classId}); err != nil {
		return nil, model.NewAppError("SqlScheduleStore.GetSchedulesByUserId", "store.sql_schedule.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return schedules, nil
}
