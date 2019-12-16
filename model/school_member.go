// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type SchoolMember struct {
	SchoolId      string `json:"school_id"`
	UserId        string `json:"user_id"`
	Roles         string `json:"roles"`
	DeleteAt      int64  `json:"delete_at"`
	SchemeParent  bool   `json:"scheme_parent"`
	SchemeTeacher bool   `json:"scheme_teacher"`
	SchemeAdmin   bool   `json:"scheme_admin"`
	ExplicitRoles string `json:"explicit_roles"`
}

type SchoolMemberForExport struct {
	SchoolMember
	SchoolName string
}

func (o *SchoolMember) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func SchoolMemberFromJson(data io.Reader) *SchoolMember {
	var o *SchoolMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func SchoolMembersToJson(o []*SchoolMember) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func SchoolMembersFromJson(data io.Reader) []*SchoolMember {
	var o []*SchoolMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *SchoolMember) IsValid() *AppError {

	if len(o.SchoolId) != 26 {
		return NewAppError("SchoolMember.IsValid", "model.school_member.is_valid.school_id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.UserId) != 26 {
		return NewAppError("SchoolMember.IsValid", "model.school_member.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (o *SchoolMember) PreUpdate() {
}

func (o *SchoolMember) GetRoles() []string {
	return strings.Fields(o.Roles)
}
