// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	EVENT_DESCRIPTION_MAX_LENGTH = 255
	EVENT_NAME_MAX_LENGTH        = 128
	EVENT_NAME_MIN_LENGTH        = 2
)

type Event struct {
	Id              string  `json:"id"`
	CreateAt        int64   `json:"create_at"`
	UpdateAt        int64   `json:"update_at"`
	DeleteAt        int64   `json:"delete_at"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Note            string  `json:"note"`
	Picture         string  `json:"picture"`
	Fee             float32 `json:"fee"`
	StartTime       int64   `json:"start_time"`
	EndTime         int64   `json:"end_time"`
	RegisterExpired int64   `json:"register_expired"`
	IsAllClass      bool    `json:"is_all_class"`
	Active          bool    `json:"active"`
	ClassId         string  `json:"class_id"`
}

type EventPatch struct {
	Title           *string  `json:"title"`
	Description     *string  `json:"description"`
	Note            *string  `json:"note"`
	Picture         *string  `json:"picture"`
	Fee             *float32 `json:"fee"`
	StartTime       *int64   `json:"start_time"`
	EndTime         *int64   `json:"end_time"`
	RegisterExpired *int64   `json:"register_expired"`
	IsAllClass      *bool    `json:"is_all_class"`
	Active          *bool    `json:"active"`
}

type EventsWithCount struct {
	Events     []*Event `json:"events"`
	TotalCount int64    `json:"total_count"`
}

func (o *Event) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func EventFromJson(data io.Reader) *Event {
	var o *Event
	json.NewDecoder(data).Decode(&o)
	return o
}

func EventMapToJson(u map[string]*Event) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func EventMapFromJson(data io.Reader) map[string]*Event {
	var events map[string]*Event
	json.NewDecoder(data).Decode(&events)
	return events
}

func EventsWithCountToJson(tlc *EventsWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func EventsWithCountFromJson(data io.Reader) *EventsWithCount {
	var swc *EventsWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Event) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Event.IsValid", "model.event.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.Id) != 26 && !o.IsAllClass {
		return NewAppError("Event.IsValid", "model.event.is_valid.class_id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Event.IsValid", "model.event.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Event.IsValid", "model.event.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.StartTime == 0 {
		return NewAppError("Event.IsValid", "model.event.is_valid.start_time.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.EndTime == 0 {
		return NewAppError("Event.IsValid", "model.event.is_valid.end_time.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Title) == 0 || utf8.RuneCountInString(o.Title) > EVENT_NAME_MAX_LENGTH {
		return NewAppError("Event.IsValid", "model.event.is_valid.title.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > EVENT_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Event.IsValid", "model.event.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Event) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.Active = true
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Event) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedEventName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidEventName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < EVENT_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validEventNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanEventName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validEventNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidEventName(s) {
		s = NewId()
	}

	return s
}

func (t *Event) Patch(patch *EventPatch) {
	if patch.Title != nil {
		t.Title = *patch.Title
	}
	if patch.Note != nil {
		t.Note = *patch.Note
	}
	if patch.Picture != nil {
		t.Picture = *patch.Picture
	}
	if patch.Fee != nil {
		t.Fee = *patch.Fee
	}
	if patch.StartTime != nil {
		t.StartTime = *patch.StartTime
	}
	if patch.EndTime != nil {
		t.EndTime = *patch.EndTime
	}
	if patch.RegisterExpired != nil {
		t.RegisterExpired = *patch.RegisterExpired
	}
	if patch.IsAllClass != nil {
		t.IsAllClass = *patch.IsAllClass
	}
	if patch.Description != nil {
		t.Description = *patch.Description
	}
	if patch.Active != nil {
		t.Active = *patch.Active
	}
}

func (t *EventPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func EventListToJson(s []*Event) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func EventPatchFromJson(data io.Reader) *EventPatch {
	decoder := json.NewDecoder(data)
	var event EventPatch
	err := decoder.Decode(&event)
	if err != nil {
		return nil
	}

	return &event
}
