// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	ACTIVITY_TYPE_SCHEDULE = "S"
	ACTIVITY_TYPE_MENU     = "F"
	ACTIVITY_TYPE_MEDICINE = "M"
	NOTE_MAX_LENGTH        = 128
)

type ActivityNote struct {
	CreateAt   int64  `json:"create_at"`
	UpdateAt   int64  `json:"update_at"`
	DeleteAt   int64  `json:"delete_at"`
	ActivityId string `json:"activity_id"`
	Note       string `json:"note"`
	Type       string `json:"type"`
	KidId      string `json:"kid_id"`
	UserId     string `json:"user_id"`
}

func (o *ActivityNote) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func ActivityNoteFromJson(data io.Reader) *ActivityNote {
	var o *ActivityNote
	json.NewDecoder(data).Decode(&o)
	return o
}

func ActivityNoteMapToJson(u map[string]*ActivityNote) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func ActivityNoteMapFromJson(data io.Reader) map[string]*ActivityNote {
	var activityNotes map[string]*ActivityNote
	json.NewDecoder(data).Decode(&activityNotes)
	return activityNotes
}

func (o *ActivityNote) IsValid() *AppError {

	if o.CreateAt == 0 {
		return NewAppError("ActivityNote.IsValid", "model.activityNote.is_valid.create_at.app_error", nil, "id="+o.ActivityId, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("ActivityNote.IsValid", "model.activityNote.is_valid.update_at.app_error", nil, "id="+o.ActivityId, http.StatusBadRequest)
	}

	if len(o.Note) > NOTE_MAX_LENGTH {
		return NewAppError("ActivityNote.IsValid", "model.activityNote.is_valid.description.app_error", nil, "id="+o.ActivityId, http.StatusBadRequest)
	}

	if !(o.Type == ACTIVITY_TYPE_SCHEDULE || o.Type == ACTIVITY_TYPE_MENU || o.Type == ACTIVITY_TYPE_MEDICINE) {
		return NewAppError("ActivityNote.IsValid", "model.activityNote.is_valid.type.app_error", nil, "id="+o.ActivityId, http.StatusBadRequest)
	}

	return nil
}

func (o *ActivityNote) PreSave() {
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *ActivityNote) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func ActivityNoteListToJson(s []*ActivityNote) string {
	b, _ := json.Marshal(s)
	return string(b)
}
