// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlEventStore struct {
	SqlStore
}

func NewSqlEventStore(sqlStore SqlStore) store.EventStore {
	s := &SqlEventStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Event{}, "Events").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Title").SetMaxSize(128)
		table.ColMap("Description").SetMaxSize(255)
		table.ColMap("Note").SetMaxSize(255)
		table.ColMap("Picture").SetMaxSize(255)
		table.ColMap("Fee").SetMaxSize(10)
		table.ColMap("IsAllClass").SetMaxSize(1)
		table.ColMap("RegisterExpired").SetMaxSize(20)
		table.ColMap("StartTime").SetMaxSize(20)
		table.ColMap("EndTime").SetMaxSize(20)
		table.ColMap("Active").SetMaxSize(1)
		table.ColMap("ClassId").SetMaxSize(26)

		tabler := db.AddTableWithName(model.EventRegistration{}, "EventRegistrations").SetKeys(false, "EventId", "KidId")
		tabler.ColMap("EventId").SetMaxSize(26)
		tabler.ColMap("KidId").SetMaxSize(26)
		tabler.ColMap("RegisterBy").SetMaxSize(26)
		tabler.ColMap("Paid").SetMaxSize(1)
	}

	return s
}

func (s SqlEventStore) CreateIndexesIfNotExists() {
	s.RemoveIndexIfExists("idx_events_description", "Events")
	s.CreateIndexIfNotExists("idx_events_update_at", "Events", "UpdateAt")
	s.CreateIndexIfNotExists("idx_events_create_at", "Events", "CreateAt")
	s.CreateIndexIfNotExists("idx_events_delete_at", "Events", "DeleteAt")
	s.CreateIndexIfNotExists("idx_events_class_id", "Events", "ClassId")

	s.CreateIndexIfNotExists("idx_eventregistrations_team_id", "EventRegistrations", "EventId")
	s.CreateIndexIfNotExists("idx_eventregistrations_user_id", "EventRegistrations", "KidId")
	s.CreateIndexIfNotExists("idx_eventregistrations_update_at", "EventRegistrations", "UpdateAt")
	s.CreateIndexIfNotExists("idx_eventregistrations_create_at", "EventRegistrations", "CreateAt")
	s.CreateIndexIfNotExists("idx_eventregistrations_delete_at", "EventRegistrations", "DeleteAt")
}

func (s SqlEventStore) Save(event *model.Event) (*model.Event, *model.AppError) {
	if len(event.Id) > 0 {
		return nil, model.NewAppError("SqlEventStore.Save",
			"store.sql_event.save.existing.app_error", nil, "id="+event.Id, http.StatusBadRequest)
	}

	event.PreSave()

	if err := event.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(event); err != nil {
		return nil, model.NewAppError("SqlEventStore.Save", "store.sql_event.save.app_error", nil, "id="+event.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return event, nil
}

func (s SqlEventStore) Update(event *model.Event) (*model.Event, *model.AppError) {

	event.PreUpdate()

	if err := event.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.Event{}, event.Id)
	if err != nil {
		return nil, model.NewAppError("SqlEventStore.Update", "store.sql_event.update.finding.app_error", nil, "id="+event.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlEventStore.Update", "store.sql_event.update.find.app_error", nil, "id="+event.Id, http.StatusBadRequest)
	}

	oldEvent := oldResult.(*model.Event)
	event.CreateAt = oldEvent.CreateAt
	event.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(event)
	if err != nil {
		return nil, model.NewAppError("SqlEventStore.Update", "store.sql_event.update.updating.app_error", nil, "id="+event.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlEventStore.Update", "store.sql_event.update.app_error", nil, "id="+event.Id, http.StatusInternalServerError)
	}

	return event, nil
}

func (s SqlEventStore) Get(id string) (*model.Event, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Event{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlEventStore.Get", "store.sql_event.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlEventStore.Get", "store.sql_event.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Event), nil
}

func (s SqlEventStore) GetByClass(classId string) ([]*model.Event, *model.AppError) {
	var events []*model.Event
	if _, err := s.GetReplica().Select(&events, "SELECT Events.* FROM Events WHERE Events.ClassId = :ClassId AND Events.DeleteAt = 0", map[string]interface{}{"ClassId": classId}); err != nil {
		return nil, model.NewAppError("SqlEventStore.GetEventsByUserId", "store.sql_event.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return events, nil
}

func (s SqlEventStore) getEventRegistrationsSelectQuery() sq.SelectBuilder {
	return s.getQueryBuilder().
		Select(
			"EventRegistrations.*",
		).
		From("EventRegistrations").
		LeftJoin("Events ON EventRegistrations.EventId = Events.Id")
}

func (s SqlEventStore) SaveRegistration(registration *model.EventRegistration) (*model.EventRegistration, *model.AppError) {
	if err := registration.IsValid(); err != nil {
		return nil, err
	}

	if err := registration.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(registration); err != nil {
		if IsUniqueConstraintError(err, []string{"EventId", "eventregistrations_pkey", "PRIMARY"}) {
			return nil, model.NewAppError("SqlEventStore.SaveRegistration", "store.sql_event.save_registration.exists.app_error", nil, "event_id="+registration.EventId+", kid_id="+registration.KidId+", "+err.Error(), http.StatusBadRequest)
		}
		return nil, model.NewAppError("SqlEventStore.SaveRegistration", "store.sql_event.save_registration.save.app_error", nil, "event_id="+registration.EventId+", kid_id="+registration.KidId+", "+err.Error(), http.StatusInternalServerError)
	}

	return registration, nil
}

func (s SqlEventStore) UpdateRegistration(registration *model.EventRegistration) (*model.EventRegistration, *model.AppError) {
	registration.PreUpdate()

	if err := registration.IsValid(); err != nil {
		return nil, err
	}

	query := s.getEventRegistrationsSelectQuery().
		Where(sq.Eq{"EventRegistrations.EventId": registration.EventId}).
		Where(sq.Eq{"EventRegistrations.KidId": registration.KidId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlEventStore.UpdateRegistration", "store.sql_event.get_registration.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var retrievedRegistration model.EventRegistration
	if err := s.GetMaster().SelectOne(&retrievedRegistration, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlEventStore.UpdateRegistration", "store.sql_event.get_registration.missing.app_error", nil, "team_id="+registration.EventId+"kid_id="+registration.KidId+","+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlEventStore.UpdateRegistration", "store.sql_event.get_registration.app_error", nil, "team_id="+registration.EventId+"kid_id="+registration.KidId+","+err.Error(), http.StatusInternalServerError)
	}

	retrievedRegistration.UpdateAt = model.GetMillis()

	_, err = s.GetMaster().Update(retrievedRegistration)

	if _, err := s.GetMaster().Update(registration); err != nil {
		return nil, model.NewAppError("SqlEventStore.UpdateRegistration", "store.sql_event.save_registration.save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return &retrievedRegistration, nil
}
