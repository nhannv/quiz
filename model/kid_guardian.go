// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type KidGuardian struct {
	KidId    string `json:"kid_id"`
	UserId   string `json:"user_id"`
	IsParent bool   `json:"is_parent"`
	DeleteAt int64  `json:"delete_at"`
}

type KidParentForExport struct {
	KidGuardian
	KidName string
}

func (o *KidGuardian) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func KidParentFromJson(data io.Reader) *KidGuardian {
	var o *KidGuardian
	json.NewDecoder(data).Decode(&o)
	return o
}

func KidParentsToJson(o []*KidGuardian) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func KidParentsFromJson(data io.Reader) []*KidGuardian {
	var o []*KidGuardian
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *KidGuardian) IsValid() *AppError {

	if len(o.KidId) != 26 {
		return NewAppError("KidGuardian.IsValid", "model.kid_parent.is_valid.kid_id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.UserId) != 26 {
		return NewAppError("KidGuardian.IsValid", "model.kid_parent.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (o *KidGuardian) PreUpdate() {
}
