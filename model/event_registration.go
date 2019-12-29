// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type EventRegistration struct {
	EventId    string `json:"event_id"`
	KidId      string `json:"kid_id"`
	Paid       bool   `json:"paid"`
	RegisterBy string `json:"register_by"`
	CreateAt   int64  `json:"create_at"`
	UpdateAt   int64  `json:"update_at"`
	DeleteAt   int64  `json:"delete_at"`
}

func (o *EventRegistration) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func EventRegistrationFromJson(data io.Reader) *EventRegistration {
	var o *EventRegistration
	json.NewDecoder(data).Decode(&o)
	return o
}

func EventRegistrationsToJson(o []*EventRegistration) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func EventRegistrationsFromJson(data io.Reader) []*EventRegistration {
	var o []*EventRegistration
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *EventRegistration) IsValid() *AppError {

	if len(o.EventId) != 26 {
		return NewAppError("EventRegistration.IsValid", "model.event_registration.is_valid.event_id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.KidId) != 26 {
		return NewAppError("EventRegistration.IsValid", "model.event_registration.is_valid.kid_id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.RegisterBy) != 26 {
		return NewAppError("EventRegistration.IsValid", "model.event_registration.is_valid.register_by.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (o *EventRegistration) PreUpdate() {
}
