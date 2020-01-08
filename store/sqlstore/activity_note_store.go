// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlActivityNoteStore struct {
	SqlStore
}

func NewSqlActivityNoteStore(sqlStore SqlStore) store.ActivityNoteStore {
	s := &SqlActivityNoteStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.ActivityNote{}, "ActivityNotes").SetKeys(false, "KidId", "ActivityId")
		table.ColMap("ActivityId").SetMaxSize(26)
		table.ColMap("Note").SetMaxSize(128)
		table.ColMap("Type").SetMaxSize(1)
		table.ColMap("KidId").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
	}

	return s
}

func (s SqlActivityNoteStore) CreateIndexesIfNotExists() {
	s.RemoveIndexIfExists("idx_activityNotes_note", "ActivityNotes")
	s.CreateIndexIfNotExists("idx_activityNotes_update_at", "ActivityNotes", "UpdateAt")
	s.CreateIndexIfNotExists("idx_activityNotes_create_at", "ActivityNotes", "CreateAt")
	s.CreateIndexIfNotExists("idx_activityNotes_delete_at", "ActivityNotes", "DeleteAt")
	s.CreateIndexIfNotExists("idx_activityNotes_kid_id", "ActivityNotes", "KidId")
	s.CreateIndexIfNotExists("idx_activityNotes_activity_id", "ActivityNotes", "ActivityId")
	s.CreateIndexIfNotExists("idx_activityNotes_type", "ActivityNotes", "Type")
}

func (s SqlActivityNoteStore) Save(activityNote *model.ActivityNote) (*model.ActivityNote, *model.AppError) {
	activityNote.PreSave()

	if err := activityNote.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(activityNote); err != nil {
		return nil, model.NewAppError("SqlActivityNoteStore.Save", "store.sql_activityNote.save.app_error", nil, "id="+activityNote.ActivityId+", "+err.Error(), http.StatusInternalServerError)
	}
	return activityNote, nil
}
