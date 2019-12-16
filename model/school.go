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
	SCHOOL_CONTACT_NAME_MAX_LENGTH = 64
	SCHOOL_DESCRIPTION_MAX_LENGTH  = 255
	SCHOOL_DISPLAY_NAME_MAX_RUNES  = 128
	SCHOOL_EMAIL_MAX_LENGTH        = 128
	SCHOOL_NAME_MAX_LENGTH         = 64
	SCHOOL_NAME_MIN_LENGTH         = 2
	SCHOOL_ADDRESS_MAX_LENGTH      = 255
	SCHOOL_PHONE_MAX_LENGTH        = 11
)

type School struct {
	Id                   string `json:"id"`
	CreateAt             int64  `json:"create_at"`
	UpdateAt             int64  `json:"update_at"`
	DeleteAt             int64  `json:"delete_at"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	ContactName          string `json:"contact_name"`
	Address              string `json:"address"`
	InviteId             string `json:"invite_id"`
	AllowOpenInvite      bool   `json:"allow_open_invite"`
	LastSchoolIconUpdate int64  `json:"last_school_icon_update,omitempty"`
}

type SchoolPatch struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	ContactName     *string `json:"contact_name"`
	Phone           *string `json:"phone"`
	Address         *string `json:"address"`
	AllowOpenInvite *bool   `json:"allow_open_invite"`
}

type SchoolInvites struct {
	Invites []map[string]string `json:"invites"`
}

type SchoolsWithCount struct {
	Schools    []*School `json:"schools"`
	TotalCount int64     `json:"total_count"`
}

func SchoolInvitesFromJson(data io.Reader) *SchoolInvites {
	var o *SchoolInvites
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *SchoolInvites) ToEmailList() []string {
	emailList := make([]string, len(o.Invites))
	for _, invite := range o.Invites {
		emailList = append(emailList, invite["email"])
	}
	return emailList
}

func (o *SchoolInvites) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *School) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func SchoolFromJson(data io.Reader) *School {
	var o *School
	json.NewDecoder(data).Decode(&o)
	return o
}

func SchoolMapToJson(u map[string]*School) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func SchoolMapFromJson(data io.Reader) map[string]*School {
	var schools map[string]*School
	json.NewDecoder(data).Decode(&schools)
	return schools
}

func SchoolsWithCountToJson(tlc *SchoolsWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func SchoolsWithCountFromJson(data io.Reader) *SchoolsWithCount {
	var swc *SchoolsWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *School) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("School.IsValid", "model.school.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("School.IsValid", "model.school.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("School.IsValid", "model.school.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Email) > SCHOOL_EMAIL_MAX_LENGTH {
		return NewAppError("School.IsValid", "model.school.is_valid.email.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Email) > 0 && !IsValidEmail(o.Email) {
		return NewAppError("School.IsValid", "model.school.is_valid.email.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Name) == 0 || utf8.RuneCountInString(o.Name) > SCHOOL_DISPLAY_NAME_MAX_RUNES {
		return NewAppError("School.IsValid", "model.school.is_valid.name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > SCHOOL_DESCRIPTION_MAX_LENGTH {
		return NewAppError("School.IsValid", "model.school.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Address) > SCHOOL_ADDRESS_MAX_LENGTH {
		return NewAppError("School.IsValid", "model.school.is_valid.address.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Phone) > SCHOOL_PHONE_MAX_LENGTH {
		return NewAppError("School.IsValid", "model.school.is_valid.phone.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.InviteId) == 0 {
		return NewAppError("School.IsValid", "model.school.is_valid.invite_id.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if IsReservedSchoolName(o.Name) {
		return NewAppError("School.IsValid", "model.school.is_valid.reserved.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if !IsValidSchoolName(o.Name) {
		return NewAppError("School.IsValid", "model.school.is_valid.characters.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.ContactName) > SCHOOL_CONTACT_NAME_MAX_LENGTH {
		return NewAppError("School.IsValid", "model.school.is_valid.contact.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *School) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *School) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedSchoolName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidSchoolName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < SCHOOL_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validSchoolNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanSchoolName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validSchoolNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidSchoolName(s) {
		s = NewId()
	}

	return s
}

func (o *School) Sanitize() {
	o.Email = ""
}

func (t *School) Patch(patch *SchoolPatch) {
	if patch.Name != nil {
		t.Name = *patch.Name
	}

	if patch.Description != nil {
		t.Description = *patch.Description
	}

	if patch.ContactName != nil {
		t.ContactName = *patch.ContactName
	}

	if patch.Phone != nil {
		t.Phone = *patch.Phone
	}

	if patch.Address != nil {
		t.Address = *patch.Address
	}
}

func (t *SchoolPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func SchoolPatchFromJson(data io.Reader) *SchoolPatch {
	decoder := json.NewDecoder(data)
	var school SchoolPatch
	err := decoder.Decode(&school)
	if err != nil {
		return nil
	}

	return &school
}
