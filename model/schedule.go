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
	SCHEDULE_DESCRIPTION_MAX_LENGTH = 128
	SCHEDULE_NAME_MAX_LENGTH        = 24
	SCHEDULE_NAME_MIN_LENGTH        = 2
)

type Schedule struct {
	Id          string `json:"id"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
	Week        int    `json:"week"`
	Year        int    `json:"year"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	WeekDay     int    `json:"week_day"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Active      bool   `json:"active"`
	ClassId     string `json:"class_id"`
}

type SchedulePatch struct {
	Week        *int    `json:"week"`
	Year        *int    `json:"year"`
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	WeekDay     *int    `json:"week_day"`
	StartTime   *int64  `json:"start_time"`
	EndTime     *int64  `json:"end_time"`
	Active      *bool   `json:"active"`
}

type SchedulesWithCount struct {
	Schedules  []*Schedule `json:"schedules"`
	TotalCount int64       `json:"total_count"`
}

func (o *Schedule) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func ScheduleFromJson(data io.Reader) *Schedule {
	var o *Schedule
	json.NewDecoder(data).Decode(&o)
	return o
}

func ScheduleMapToJson(u map[string]*Schedule) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func ScheduleMapFromJson(data io.Reader) map[string]*Schedule {
	var schedules map[string]*Schedule
	json.NewDecoder(data).Decode(&schedules)
	return schedules
}

func SchedulesWithCountToJson(tlc *SchedulesWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func SchedulesWithCountFromJson(data io.Reader) *SchedulesWithCount {
	var swc *SchedulesWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Schedule) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.ClassId) != 26 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.class_id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.StartTime == 0 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.start_time.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.EndTime == 0 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.end_time.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.WeekDay < 1 && o.WeekDay > 7 {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.week_day.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Subject) == 0 || utf8.RuneCountInString(o.Subject) > SCHEDULE_NAME_MAX_LENGTH {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.first_name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > SCHEDULE_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Schedule.IsValid", "model.schedule.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Schedule) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.Active = true
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Schedule) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedScheduleName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidScheduleName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < SCHEDULE_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validScheduleNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanScheduleName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validScheduleNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidScheduleName(s) {
		s = NewId()
	}

	return s
}

func (t *Schedule) Patch(patch *SchedulePatch) {
	if patch.Week != nil {
		t.Week = *patch.Week
	}
	if patch.Year != nil {
		t.Year = *patch.Year
	}
	if patch.Subject != nil {
		t.Subject = *patch.Subject
	}
	if patch.WeekDay != nil {
		t.WeekDay = *patch.WeekDay
	}
	if patch.StartTime != nil {
		t.StartTime = *patch.StartTime
	}
	if patch.EndTime != nil {
		t.EndTime = *patch.EndTime
	}
	if patch.Description != nil {
		t.Description = *patch.Description
	}
	if patch.Active != nil {
		t.Active = *patch.Active
	}
}

func (t *SchedulePatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func ScheduleListToJson(s []*Schedule) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func SchedulePatchFromJson(data io.Reader) *SchedulePatch {
	decoder := json.NewDecoder(data)
	var schedule SchedulePatch
	err := decoder.Decode(&schedule)
	if err != nil {
		return nil
	}

	return &schedule
}
