// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Health struct {
	Id        string  `json:"id"`
	CreateAt  int64   `json:"create_at"`
	UpdateAt  int64   `json:"update_at"`
	DeleteAt  int64   `json:"delete_at"`
	Height    float32 `json:"height"`
	Weight    float32 `json:"weight"`
	MeasureAt int64   `json:"measure_at"`
	KidId     string  `json:"kid_id"`
}

type HealthPatch struct {
	Height    *float32 `json:"height"`
	Weight    *float32 `json:"weight"`
	MeasureAt *int64   `json:"measure_at"`
}

type HealthsWithCount struct {
	Healths    []*Health `json:"healths"`
	TotalCount int64     `json:"total_count"`
}

func (o *Health) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func HealthFromJson(data io.Reader) *Health {
	var o *Health
	json.NewDecoder(data).Decode(&o)
	return o
}

func HealthMapToJson(u map[string]*Health) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func HealthMapFromJson(data io.Reader) map[string]*Health {
	var healths map[string]*Health
	json.NewDecoder(data).Decode(&healths)
	return healths
}

func HealthsWithCountToJson(tlc *HealthsWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func HealthsWithCountFromJson(data io.Reader) *HealthsWithCount {
	var swc *HealthsWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Health) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Health.IsValid", "model.health.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Health.IsValid", "model.health.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Health.IsValid", "model.health.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.Height == 0 {
		return NewAppError("Health.IsValid", "model.health.is_valid.height.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.Weight == 0 {
		return NewAppError("Health.IsValid", "model.health.is_valid.weight.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.MeasureAt == 0 {
		return NewAppError("Health.IsValid", "model.health.is_valid.measure_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Health) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Health) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (t *Health) Patch(patch *HealthPatch) {
	if patch.Height != nil {
		t.Height = *patch.Height
	}
	if patch.Weight != nil {
		t.Weight = *patch.Weight
	}
	if patch.MeasureAt != nil {
		t.MeasureAt = *patch.MeasureAt
	}
}

func (t *HealthPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func HealthListToJson(s []*Health) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func HealthPatchFromJson(data io.Reader) *HealthPatch {
	decoder := json.NewDecoder(data)
	var health HealthPatch
	err := decoder.Decode(&health)
	if err != nil {
		return nil
	}

	return &health
}
