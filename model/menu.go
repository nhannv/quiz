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
	MENU_DESCRIPTION_MAX_LENGTH = 128
	MENU_NAME_MAX_LENGTH        = 24
	MENU_NAME_MIN_LENGTH        = 2
)

type Menu struct {
	Id          string `json:"id"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
	Week        int    `json:"week"`
	Year        int    `json:"year"`
	FoodName    string `json:"food_name"`
	Description string `json:"description"`
	WeekDay     int    `json:"week_day"`
	StartTime   int64  `json:"start_time"`
	Note        string `json:"note"`
	ClassId     string `json:"class_id"`
}

type MenuPatch struct {
	Week        *int    `json:"week"`
	Year        *int    `json:"year"`
	FoodName    *string `json:"food_name"`
	Description *string `json:"description"`
	WeekDay     *int    `json:"week_day"`
	StartTime   *int64  `json:"start_time"`
}

type MenusWithCount struct {
	Menus      []*Menu `json:"menus"`
	TotalCount int64   `json:"total_count"`
}

func (o *Menu) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func MenuFromJson(data io.Reader) *Menu {
	var o *Menu
	json.NewDecoder(data).Decode(&o)
	return o
}

func MenuMapToJson(u map[string]*Menu) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func MenuMapFromJson(data io.Reader) map[string]*Menu {
	var menus map[string]*Menu
	json.NewDecoder(data).Decode(&menus)
	return menus
}

func MenusWithCountToJson(tlc *MenusWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func MenusWithCountFromJson(data io.Reader) *MenusWithCount {
	var swc *MenusWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Menu) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.StartTime == 0 {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.start_time.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.WeekDay < 1 && o.WeekDay > 7 {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.week_day.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.FoodName) == 0 || utf8.RuneCountInString(o.FoodName) > MENU_NAME_MAX_LENGTH {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.first_name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > MENU_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Menu.IsValid", "model.menu.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Menu) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Menu) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedMenuName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidMenuName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < MENU_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validMenuNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanMenuName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validMenuNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidMenuName(s) {
		s = NewId()
	}

	return s
}

func (t *Menu) Patch(patch *MenuPatch) {
	if patch.Week != nil {
		t.Week = *patch.Week
	}
	if patch.Year != nil {
		t.Year = *patch.Year
	}
	if patch.FoodName != nil {
		t.FoodName = *patch.FoodName
	}
	if patch.WeekDay != nil {
		t.WeekDay = *patch.WeekDay
	}
	if patch.StartTime != nil {
		t.StartTime = *patch.StartTime
	}
	if patch.Description != nil {
		t.Description = *patch.Description
	}
}

func (t *MenuPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func MenuListToJson(s []*Menu) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func MenuPatchFromJson(data io.Reader) *MenuPatch {
	decoder := json.NewDecoder(data)
	var menu MenuPatch
	err := decoder.Decode(&menu)
	if err != nil {
		return nil
	}

	return &menu
}
