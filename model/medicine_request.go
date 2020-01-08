// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type MedicineRequest struct {
	Id        string `json:"id"`
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
	DeleteAt  int64  `json:"delete_at"`
	CreateBy  string `json:"create_by"`
	KidId     string `json:"kid_id"`
	FromDate  int64  `json:"from_date"`
	ToDate    int64  `json:"to_date"`
	Confirmed bool   `json:"confirmed"`
	ConfirmBy string `json:"confirmBy"`
}

type MedicineRequestPatch struct {
	FromDate  *int64  `json:"from_date"`
	ToDate    *int64  `json:"to_date"`
	Confirmed *bool   `json:"confirmed"`
	ConfirmBy *string `json:"confirmBy"`
}

type MedicineRequestsWithCount struct {
	MedicineRequests []*MedicineRequest `json:"medicine_requests"`
	TotalCount       int64              `json:"total_count"`
}

func (o *MedicineRequest) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func MedicineRequestFromJson(data io.Reader) *MedicineRequest {
	var o *MedicineRequest
	json.NewDecoder(data).Decode(&o)
	return o
}

func MedicineRequestMapToJson(u map[string]*MedicineRequest) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func MedicineRequestMapFromJson(data io.Reader) map[string]*MedicineRequest {
	var medicines map[string]*MedicineRequest
	json.NewDecoder(data).Decode(&medicines)
	return medicines
}

func MedicineRequestsWithCountToJson(tlc *MedicineRequestsWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func MedicineRequestsWithCountFromJson(data io.Reader) *MedicineRequestsWithCount {
	var swc *MedicineRequestsWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *MedicineRequest) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.CreateBy) == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.create_by.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.FromDate == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.from_date.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.ToDate == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.to_date.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *MedicineRequest) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *MedicineRequest) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (t *MedicineRequest) Patch(patch *MedicineRequestPatch) {
	if patch.FromDate != nil {
		t.FromDate = *patch.FromDate
	}
	if patch.ToDate != nil {
		t.ToDate = *patch.ToDate
	}
	if patch.Confirmed != nil {
		t.Confirmed = *patch.Confirmed
	}
	if patch.ConfirmBy != nil {
		t.ConfirmBy = *patch.ConfirmBy
	}
}

func (t *MedicineRequestPatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func MedicineRequestListToJson(s []*MedicineRequest) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func MedicineRequestPatchFromJson(data io.Reader) *MedicineRequestPatch {
	decoder := json.NewDecoder(data)
	var medicine MedicineRequestPatch
	err := decoder.Decode(&medicine)
	if err != nil {
		return nil
	}

	return &medicine
}
