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
	BRANCH_CONTACT_NAME_MAX_LENGTH = 64
	BRANCH_DESCRIPTION_MAX_LENGTH  = 255
	BRANCH_DISPLAY_NAME_MAX_RUNES  = 128
	BRANCH_NAME_MIN_LENGTH         = 2
)

type Branch struct {
	Id          string `json:"id"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchoolId    string `json:"school_id"`
	CreatorId   string `json:"creator_id"`
}

type BranchPatch struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type BranchesWithCount struct {
	Branches   []*Branch `json:"branches"`
	TotalCount int64     `json:"total_count"`
}

func (o *Branch) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func BranchFromJson(data io.Reader) *Branch {
	var o *Branch
	json.NewDecoder(data).Decode(&o)
	return o
}

func BranchMapToJson(u map[string]*Branch) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func BranchMapFromJson(data io.Reader) map[string]*Branch {
	var branches map[string]*Branch
	json.NewDecoder(data).Decode(&branches)
	return branches
}

func BranchesToJson(o []*Branch) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func BranchesWithCountToJson(tlc *BranchesWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func BranchesWithCountFromJson(data io.Reader) *BranchesWithCount {
	var swc *BranchesWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Branch) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Branch.IsValid", "model.branch.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Branch.IsValid", "model.branch.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Branch.IsValid", "model.branch.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Name) == 0 || utf8.RuneCountInString(o.Name) > BRANCH_DISPLAY_NAME_MAX_RUNES {
		return NewAppError("Branch.IsValid", "model.branch.is_valid.name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > BRANCH_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Branch.IsValid", "model.branch.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Branch) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt

}

func (o *Branch) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedBranchName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidBranchName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < BRANCH_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validBranchNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanBranchName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validBranchNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidBranchName(s) {
		s = NewId()
	}

	return s
}

func (t *Branch) Patch(patch *BranchPatch) {
	if patch.Name != nil {
		t.Name = *patch.Name
	}

	if patch.Description != nil {
		t.Description = *patch.Description
	}

}

func (t *BranchPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func BranchPatchFromJson(data io.Reader) *BranchPatch {
	decoder := json.NewDecoder(data)
	var branch BranchPatch
	err := decoder.Decode(&branch)
	if err != nil {
		return nil
	}

	return &branch
}
