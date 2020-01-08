// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *App) CreateMedicineRequest(medicineRequest *model.MedicineRequest) (*model.MedicineRequest, *model.AppError) {
	rmedicine, err := a.Srv.Store.Medicine().SaveRequest(medicineRequest)
	if err != nil {
		return nil, err
	}

	return rmedicine, nil
}

func (a *App) UpdateMedicineRequest(medicineRequest *model.MedicineRequest) (*model.MedicineRequest, *model.AppError) {
	oldMedicineRequest, err := a.GetMedicineRequest(medicineRequest.Id)
	if err != nil {
		return nil, err
	}

	oldMedicineRequest.FromDate = medicineRequest.FromDate
	oldMedicineRequest.ToDate = medicineRequest.ToDate

	oldMedicineRequest, err = a.updateMedicineRequestUnsanitized(oldMedicineRequest)
	if err != nil {
		return medicineRequest, err
	}

	return oldMedicineRequest, nil
}

func (a *App) updateMedicineRequestUnsanitized(medicineRequest *model.MedicineRequest) (*model.MedicineRequest, *model.AppError) {
	return a.Srv.Store.Medicine().UpdateRequest(medicineRequest)
}

func (a *App) PatchMedicineRequest(medicineRequestId string, patch *model.MedicineRequestPatch) (*model.MedicineRequest, *model.AppError) {
	medicineRequest, err := a.GetMedicineRequest(medicineRequestId)
	if err != nil {
		return nil, err
	}

	medicineRequest.Patch(patch)

	updatedMedicineRequest, err := a.UpdateMedicineRequest(medicineRequest)
	if err != nil {
		return nil, err
	}

	return updatedMedicineRequest, nil
}

func (a *App) GetMedicineRequest(medicineRequestId string) (*model.MedicineRequest, *model.AppError) {
	return a.Srv.Store.Medicine().GetRequest(medicineRequestId)
}

func (a *App) GetMedicineRequestsByKid(kidId string) ([]*model.MedicineRequest, *model.AppError) {
	return a.Srv.Store.Medicine().GetRequestsByKid(kidId)
}
