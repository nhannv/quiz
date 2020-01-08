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
	KID_DESCRIPTION_MAX_LENGTH = 255
	KID_NAME_MAX_LENGTH        = 24
	KID_NAME_MIN_LENGTH        = 2
)

type Kid struct {
	Id          string `json:"id"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	NickName    string `json:"nick_name"`
	Avatar      string `json:"avatar"`
	Cover       string `json:"cover"`
	Description string `json:"description"`
	Dob         int64  `json:"dob"`
	Gender      bool   `json:"gender"`
	ClassId     string `json:"class_id"`
	ClassName   string `json:"class_name"`
	InviteId    string `json:"invite_id"`
}

type KidPatch struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	NickName    *string `json:"nick_name"`
	Avatar      *string `json:"avatar"`
	Cover       *string `json:"cover"`
	Description *string `json:"description"`
	Dob         *int64  `json:"dob"`
	Gender      *bool   `json:"gender"`
	ClassId     *string `json:"class_id"`
}

type ParentInvites struct {
	Invites []map[string]string `json:"invites"`
}

func ParentInvitesFromJson(data io.Reader) *ParentInvites {
	var o *ParentInvites
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *ParentInvites) InvitesToEmailList() []string {
	emailList := make([]string, len(o.Invites))
	for _, invite := range o.Invites {
		emailList = append(emailList, invite["email"])
	}
	return emailList
}

func (o *ParentInvites) InvitesToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

type KidsWithCount struct {
	Kids       []*Kid `json:"kids"`
	TotalCount int64  `json:"total_count"`
}

func (o *Kid) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func KidFromJson(data io.Reader) *Kid {
	var o *Kid
	json.NewDecoder(data).Decode(&o)
	return o
}

func KidMapToJson(u map[string]*Kid) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func KidMapFromJson(data io.Reader) map[string]*Kid {
	var kids map[string]*Kid
	json.NewDecoder(data).Decode(&kids)
	return kids
}

func KidsWithCountToJson(tlc *KidsWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func KidsWithCountFromJson(data io.Reader) *KidsWithCount {
	var swc *KidsWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Kid) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.ClassId) != 26 {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.class_id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.InviteId) == 0 {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.invite_id.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.FirstName) == 0 || utf8.RuneCountInString(o.FirstName) > KID_NAME_MAX_LENGTH {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.first_name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.LastName) == 0 || utf8.RuneCountInString(o.LastName) > KID_NAME_MAX_LENGTH {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.first_name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > KID_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Kid.IsValid", "model.kid.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Kid) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	if len(o.InviteId) == 0 {
		o.InviteId = NewId()
	}
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Kid) Sanitize() {
	o.InviteId = ""
}

func (o *Kid) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedKidName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidKidName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < KID_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validKidNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanKidName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validKidNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidKidName(s) {
		s = NewId()
	}

	return s
}

func (t *Kid) Patch(patch *KidPatch) {
	if patch.FirstName != nil {
		t.FirstName = *patch.FirstName
	}

	if patch.Description != nil {
		t.Description = *patch.Description
	}

	if patch.NickName != nil {
		t.NickName = *patch.NickName
	}
	if patch.Avatar != nil {
		t.Avatar = *patch.Avatar
	}
	if patch.Cover != nil {
		t.Cover = *patch.Cover
	}
	if patch.Description != nil {
		t.Description = *patch.Description
	}
	if patch.Dob != nil {
		t.Dob = *patch.Dob
	}
	if patch.Gender != nil {
		t.Gender = *patch.Gender
	}
	if patch.ClassId != nil {
		t.ClassId = *patch.ClassId
	}
}

func (t *KidPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func KidPatchFromJson(data io.Reader) *KidPatch {
	decoder := json.NewDecoder(data)
	var kid KidPatch
	err := decoder.Decode(&kid)
	if err != nil {
		return nil
	}

	return &kid
}

func KidListToJson(u []*Kid) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func KidListFromJson(data io.Reader) []*Kid {
	var kids []*Kid
	json.NewDecoder(data).Decode(&kids)
	return kids
}
