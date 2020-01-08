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
	MEDICINE_NOTE_MAX_LENGTH = 128
	MEDICINE_NAME_MAX_LENGTH = 128
	MEDICINE_NAME_MIN_LENGTH = 2
)

type Medicine struct {
	Id                string `json:"id"`
	CreateAt          int64  `json:"create_at"`
	UpdateAt          int64  `json:"update_at"`
	DeleteAt          int64  `json:"delete_at"`
	Name              string `json:"subject"`
	Note              string `json:"note"`
	Dosage            string `json:"dosage"`
	MedicineRequestId string `json:"request_id"`
}

type MedicinesWithCount struct {
	Medicines  []*Medicine `json:"medicines"`
	TotalCount int64       `json:"total_count"`
}

func (o *Medicine) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func MedicineFromJson(data io.Reader) *Medicine {
	var o *Medicine
	json.NewDecoder(data).Decode(&o)
	return o
}

func MedicineMapToJson(u map[string]*Medicine) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func MedicineMapFromJson(data io.Reader) map[string]*Medicine {
	var medicines map[string]*Medicine
	json.NewDecoder(data).Decode(&medicines)
	return medicines
}

func MedicinesWithCountToJson(tlc *MedicineRequestsWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func MedicinesWithCountFromJson(data io.Reader) *MedicineRequestsWithCount {
	var swc *MedicineRequestsWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Medicine) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.MedicineRequestId) != 26 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.request_id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Name) == 0 || utf8.RuneCountInString(o.Name) > MEDICINE_NAME_MAX_LENGTH {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Dosage) > MEDICINE_NOTE_MAX_LENGTH {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.dosage.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Note) > MEDICINE_NOTE_MAX_LENGTH {
		return NewAppError("Medicine.IsValid", "model.medicine.is_valid.note.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Medicine) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Medicine) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IsReservedMedicineName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidMedicineName(s string) bool {

	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < MEDICINE_NAME_MIN_LENGTH {
		return false
	}

	return true
}

var validMedicineNameCharacter = regexp.MustCompile(`^[a-z0-9-]$`)

func CleanMedicineName(s string) string {
	s = strings.ToLower(strings.Replace(s, " ", "-", -1))

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			s = strings.Replace(s, value, "", -1)
		}
	}

	s = strings.TrimSpace(s)

	for _, c := range s {
		char := fmt.Sprintf("%c", c)
		if !validMedicineNameCharacter.MatchString(char) {
			s = strings.Replace(s, char, "", -1)
		}
	}

	s = strings.Trim(s, "-")

	if !IsValidMedicineName(s) {
		s = NewId()
	}

	return s
}

func MedicineListToJson(s []*Medicine) string {
	b, _ := json.Marshal(s)
	return string(b)
}
