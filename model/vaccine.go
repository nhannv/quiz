// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type VaccineBook struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Times       int    `json:"times"`
}

type Vaccine struct {
	Id            string `json:"id"`
	CreateAt      int64  `json:"create_at"`
	UpdateAt      int64  `json:"update_at"`
	DeleteAt      int64  `json:"delete_at"`
	VaccineBookId int    `json:"vaccine_book_id"`
	VaccineName   string `json:"vaccine_name"`
	Time          int    `json:"time"`
	Date          int64  `json:"date"`
	Place         string `json:"place"`
	KidId         string `json:"kid_id"`
}

type VaccinePatch struct {
	VaccineBookId *int    `json:"vaccine_book_id"`
	VaccineName   *string `json:"vaccine_name"`
	Time          *int    `json:"time"`
	Date          *int64  `json:"date"`
	Place         *string `json:"place"`
}

type VaccinesWithCount struct {
	Vaccines   []*Vaccine `json:"vaccines"`
	TotalCount int64      `json:"total_count"`
}

func (o *Vaccine) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func VaccineFromJson(data io.Reader) *Vaccine {
	var o *Vaccine
	json.NewDecoder(data).Decode(&o)
	return o
}

func VaccineMapToJson(u map[string]*Vaccine) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func VaccineMapFromJson(data io.Reader) map[string]*Vaccine {
	var vaccines map[string]*Vaccine
	json.NewDecoder(data).Decode(&vaccines)
	return vaccines
}

func VaccinesWithCountToJson(tlc *VaccinesWithCount) []byte {
	b, _ := json.Marshal(tlc)
	return b
}

func VaccinesWithCountFromJson(data io.Reader) *VaccinesWithCount {
	var swc *VaccinesWithCount
	json.NewDecoder(data).Decode(&swc)
	return swc
}

func (o *Vaccine) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.VaccineBookId == 0 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.vaccine_book_id.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.VaccineName) == 0 || len(o.VaccineName) > 100 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.vaccine_name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.Time == 0 || o.Time > 20 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.time.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.Date == 0 {
		return NewAppError("Vaccine.IsValid", "model.vaccine.is_valid.date.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Vaccine) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *Vaccine) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (t *Vaccine) Patch(patch *VaccinePatch) {
	if patch.VaccineBookId != nil {
		t.VaccineBookId = *patch.VaccineBookId
	}
	if patch.VaccineName != nil {
		t.VaccineName = *patch.VaccineName
	}
	if patch.Time != nil {
		t.Time = *patch.Time
	}
	if patch.Date != nil {
		t.Date = *patch.Date
	}
	if patch.Place != nil {
		t.Place = *patch.Place
	}
}

func (t *VaccinePatch) ToJson() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func VaccineListToJson(s []*Vaccine) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func VaccinePatchFromJson(data io.Reader) *VaccinePatch {
	decoder := json.NewDecoder(data)
	var vaccine VaccinePatch
	err := decoder.Decode(&vaccine)
	if err != nil {
		return nil
	}

	return &vaccine
}

// VaccineBookData represents master data
var VaccineBookData = []VaccineBook{
	VaccineBook{
		Id:          1,
		Title:       "Lao (Tuberculosis)",
		Description: "- Sơ sinh",
		Times:       1,
	},
	VaccineBook{
		Id:    2,
		Title: "Viêm gan siêu vi B (Hepatitis B)",
		Description: `- Từ lúc mới sinh
		- Các liều kế tiếp cách nhau 1-2 tháng
		- Có thể nhắc lại theo chỉ định của bác sĩ`,
		Times: 6,
	},
	VaccineBook{
		Id:    3,
		Title: "Bạch hầu - Uốn ván - Ho gà (Diphtheria - Tetanus - Pertussis)",
		Description: `- Từ 2 tháng: 3 liều liên tiếp cách nhau 1-2 tháng
		- Nhắc lại 1 liều lúc 16-18 tháng
		- Nhắc lại 1 liều lúc 4-6 tuổi (IP_DTaP)`,
		Times: 5,
	},
	VaccineBook{
		Id:    4,
		Title: "Bại liệt (Poliovirus)",
		Description: `- Uống hoặc chích
		- Từ 2 tháng: 3 liều liên tiếp cách nhau 1-2 tháng
		- Nhắc lại 1 liều lúc 16-18 tháng
		- Nhắc lại 1 liều lúc 4-6 tuổi (IP_DTaP)`,
		Times: 5,
	},
	VaccineBook{
		Id:    5,
		Title: "Viêm màng não mủ do Haemophilus Influenza Type B",
		Description: `- Nếu từ 2-6 tháng: 3 liều liên tiếp cách nhau 1-2 tháng; nhắc lại 1 liều lúc 16-18 tháng
		- Nếu sau 6 tháng: 2 liều liên tiếp cách nhau 1-2 tháng; nhắc lại 1 liều lúc 12 tháng
		- Nếu sau 1 tuổi: 1 liều duy nhất`,
		Times: 4,
	},
	VaccineBook{
		Id:          6,
		Title:       "Tiêu chảy do Rotavirus",
		Description: `- Trẻ từ 2 tháng: các liều uống cách nhau 1-2 tháng`,
		Times:       3,
	},
	VaccineBook{
		Id:          7,
		Title:       "Viêm màng não, viêm phổi, viêm tai giữa do phế cầu (Pneumococcal)",
		Description: `- Mũi 1 từ 2 tháng: (Pneumococcal conjugate vaccine PCV)`,
		Times:       4,
	},
	VaccineBook{
		Id:          8,
		Title:       "Viêm màng não do não mô cầu BC (Meningococal BC)",
		Description: `- 3 tháng: 2 liều cách nhau tối thiểu 8 tuần`,
		Times:       2,
	},
	VaccineBook{
		Id:    9,
		Title: "Cúm (Influenza)",
		Description: `- Từ 6 tháng - 9 tuổi:
		+ Lần đầu tiêm 2 mũi cách nhau tối thiểu 1 tháng
		+ Nhắc lại mỗi năm 1 liều
		- Trẻ từ 9 tuổi: mỗi năm chích 1 liều`,
		Times: 10,
	},
	VaccineBook{
		Id:          10,
		Title:       "Sởi (Measles)",
		Description: `- Lúc 9 tháng`,
		Times:       1,
	},
	VaccineBook{
		Id:    11,
		Title: "Sởi - Quai bị - Rubella (Measles - Mumps - Rebella)",
		Description: `- Tiêm lần 1: từ 12 tháng
		- Nhắc lại lúc 4-6 tuổi (có thể tiêm trước 4 tuổi)`,
		Times: 2,
	},
	VaccineBook{
		Id:    12,
		Title: "Thủy đậu - Trái rạ (Varicella)",
		Description: `- Từ 12 tháng đến 12 tuổi: 2 liều cách nhau trên 3 tháng
		- Từ 12 tuổi: 2 liều cách nhau trên 6 tuần`,
		Times: 2,
	},
	VaccineBook{
		Id:    13,
		Title: "Viêm não nhật bản B (Japanese Encephalitis)",
		Description: `- Từ 12 tháng: 2 liều liên tiếp cách nhau 1-2 tuần
		- Nhắc lại 1 liều 1 năm sau tiêm lần 1
		- Sau đó nhắc lại 1 liều mỗi 3 năm đến 15 tuổi`,
		Times: 6,
	},
	VaccineBook{
		Id:          14,
		Title:       "Viêm gan A (Hepatitis A)",
		Description: `- Từ 12 tháng: 2 liều cách nhau 6-18 tháng`,
		Times:       2,
	},
	VaccineBook{
		Id:          15,
		Title:       "Viêm màng não do mô cầu AC (Meningococcal AC)",
		Description: `- Từ 24 tháng: 2 liều cách nhau mỗi 3 năm`,
		Times:       7,
	},
	VaccineBook{
		Id:          16,
		Title:       "Thương hàn (Typhoid)",
		Description: `- Bắt đầu từ 2 tuổi: 2 liều cách nhau 3 năm`,
		Times:       3,
	},
	VaccineBook{
		Id:    17,
		Title: "Ung tư cổ tử cung",
		Description: `- Liều 1: từ 10 tuổi
		- Liều 2: cách liều 1 một tháng
		- Liều 3: cách liều 1 sáu tháng`,
		Times: 3,
	},
	VaccineBook{
		Id:          18,
		Title:       "Sốt xuất huyết (Dengue Fever)",
		Description: "",
		Times:       3,
	},
}
