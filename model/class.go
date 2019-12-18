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
	CLASS_CONTACT_NAME_MAX_LENGTH = 64
	CLASS_DESCRIPTION_MAX_LENGTH  = 255
	CLASS_DISPLAY_NAME_MAX_RUNES  = 128
	CLASS_NAME_MIN_LENGTH         = 2
)

type Class struct {
	Id              string `json:"id"`
	CreateAt        int64  `json:"create_at"`
	UpdateAt        int64  `json:"update_at"`
	DeleteAt        int64  `json:"delete_at"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	SchoolId        string `json:"school_id"`
	BranchId        string `json:"branch_id"`
	InviteId        string `json:"invite_id"`
	AllowOpenInvite bool   `json:"allow_open_invite"`
	CreatorId       string `json:"creator_id"`
}

type ClassPatch struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	BranchId        *string `json:"branch_id"`
	AllowOpenInvite *bool   `json:"allow_open_invite"`
}

type ClassInvites struct {
	Invites []map[string]string `json:"invites"`
}

type ClassesWithCount struct {
	Classes    []*Class `json:"classes"`
	TotalCount int64    `json:"total_count"`
}

func ClassInvitesFromJson(data io.Reader) *ClassInvites {
	var o *ClassInvites
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *ClassInvites) ToEmailList() []string {
	emailList := make([]string, len(o.Invites))
	for _, invite := range o.Invites {
		emailList = append(emailList, invite["email"])
	}
	return emailList
}

func (o *ClassInvites) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Class) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func ClassFromJson(data io.Reader) *Class {
	var o *Class
	json.NewDecoder(data).Decode(&o)
	return o
}

func ClassesToJson(o []*Class) string {
	if c, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(c)
	}
}

func ClassMapToJson(u map[string]*Class) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func ClassMapFromJson(data io.Reader) map[string]*Class {
	var classes map[string]*Class
	json.NewDecoder(data).Decode(&classes)
	return classes
}

func ClassesWithCountToJson(tlc *ClassesWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func ClassesWithCountFromJson(data io.Reader) *ClassesWithCount {
	var swc *ClassesWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Class) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Class.IsValid", "model.class.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Class.IsValid", "model.class.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Class.IsValid", "model.class.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Name) == 0 || utf8.RuneCountInString(o.Name) > CLASS_DISPLAY_NAME_MAX_RUNES {
		return NewAppError("Class.IsValid", "model.class.is_valid.name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > CLASS_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Class.IsValid", "model.class.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.InviteId) == 0 {
		return NewAppError("Class.IsValid", "model.class.is_valid.invite_id.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Class) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt

	if len(o.InviteId) == 0 {
		o.InviteId = NewId()
	}
}

func (o *Class) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedClassName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidClassName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < CLASS_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validClassNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanClassName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validClassNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidClassName(s) {
		s = NewId()
	}

	return s
}

func (o *Class) Sanitize() {
	o.InviteId = ""
}

func (t *Class) Patch(patch *ClassPatch) {
	if patch.Name != nil {
		t.Name = *patch.Name
	}

	if patch.Description != nil {
		t.Description = *patch.Description
	}

	if patch.BranchId != nil {
		t.BranchId = *patch.BranchId
	}

}

func (t *ClassPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func ClassPatchFromJson(data io.Reader) *ClassPatch {
	decoder := json.NewDecoder(data)
	var class ClassPatch
	err := decoder.Decode(&class)
	if err != nil {
		return nil
	}

	return &class
}
