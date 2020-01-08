// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlMedicineStore struct {
	SqlStore
}

func NewSqlMedicineStore(sqlStore SqlStore) store.MedicineStore {
	s := &SqlMedicineStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.MedicineRequest{}, "MedicineRequests").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("KidId").SetMaxSize(26)
		table.ColMap("FromDate").SetMaxSize(20)
		table.ColMap("ToDate").SetMaxSize(20)
		table.ColMap("CreateBy").SetMaxSize(26)
		table.ColMap("Confirmed").SetMaxSize(1)
		table.ColMap("ConfirmBy").SetMaxSize(26)

		tablem := db.AddTableWithName(model.Medicine{}, "Medicines").SetKeys(false, "Id", "MedicineRequestId")
		tablem.ColMap("Id").SetMaxSize(26)
		tablem.ColMap("MedicineRequestId").SetMaxSize(26)
		tablem.ColMap("Name").SetMaxSize(128)
		tablem.ColMap("Dosage").SetMaxSize(128)
		tablem.ColMap("Note").SetMaxSize(128)
	}

	return s
}

func (s SqlMedicineStore) CreateIndexesIfNotExists() {
	s.RemoveIndexIfExists("idx_medicineRequests_description", "MedicineRequests")
	s.CreateIndexIfNotExists("idx_medicineRequests_update_at", "MedicineRequests", "UpdateAt")
	s.CreateIndexIfNotExists("idx_medicineRequests_create_at", "MedicineRequests", "CreateAt")
	s.CreateIndexIfNotExists("idx_medicineRequests_delete_at", "MedicineRequests", "DeleteAt")
	s.CreateIndexIfNotExists("idx_medicineRequests_kid_id", "MedicineRequests", "KidId")
	s.CreateIndexIfNotExists("idx_medicineRequests_from_date", "MedicineRequests", "FromDate")
	s.CreateIndexIfNotExists("idx_medicineRequests_to_date", "MedicineRequests", "ToDate")
	s.CreateIndexIfNotExists("idx_medicines_request_id", "Medicines", "MedicineRequestId")
}

func (s SqlMedicineStore) SaveRequest(medicineRequest *model.MedicineRequest) (*model.MedicineRequest, *model.AppError) {
	if len(medicineRequest.Id) > 0 {
		return nil, model.NewAppError("SqlMedicineStore.Save",
			"store.sql_medicineRequest.save.existing.app_error", nil, "id="+medicineRequest.Id, http.StatusBadRequest)
	}

	medicineRequest.PreSave()

	if err := medicineRequest.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(medicineRequest); err != nil {
		return nil, model.NewAppError("SqlMedicineStore.Save", "store.sql_medicineRequest.save.app_error", nil, "id="+medicineRequest.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return medicineRequest, nil
}

func (s SqlMedicineStore) UpdateRequest(medicineRequest *model.MedicineRequest) (*model.MedicineRequest, *model.AppError) {

	medicineRequest.PreUpdate()

	if err := medicineRequest.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.MedicineRequest{}, medicineRequest.Id)
	if err != nil {
		return nil, model.NewAppError("SqlMedicineStore.Update", "store.sql_medicineRequest.update.finding.app_error", nil, "id="+medicineRequest.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlMedicineStore.Update", "store.sql_medicineRequest.update.find.app_error", nil, "id="+medicineRequest.Id, http.StatusBadRequest)
	}

	oldMedicineRequest := oldResult.(*model.MedicineRequest)
	medicineRequest.CreateAt = oldMedicineRequest.CreateAt
	medicineRequest.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(medicineRequest)
	if err != nil {
		return nil, model.NewAppError("SqlMedicineStore.Update", "store.sql_medicineRequest.update.updating.app_error", nil, "id="+medicineRequest.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlMedicineStore.Update", "store.sql_medicineRequest.update.app_error", nil, "id="+medicineRequest.Id, http.StatusInternalServerError)
	}

	return medicineRequest, nil
}

func (s SqlMedicineStore) GetRequest(id string) (*model.MedicineRequest, *model.AppError) {
	obj, err := s.GetReplica().Get(model.MedicineRequest{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlMedicineStore.Get", "store.sql_medicineRequest.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlMedicineStore.Get", "store.sql_medicineRequest.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.MedicineRequest), nil
}

func (s SqlMedicineStore) GetRequestsByClass(classId string) ([]*model.MedicineRequest, *model.AppError) {
	var medicineRequests []*model.MedicineRequest
	if _, err := s.GetReplica().Select(&medicineRequests,
		`SELECT MedicineRequests.* FROM MedicineRequests
			INNER JOIN Kids ON MedicineRequests.KidId = Kids.Id
			WHERE Kids.ClassId = :ClassId AND MedicineRequests.DeleteAt = 0`,
		map[string]interface{}{"ClassId": classId}); err != nil {
		return nil, model.NewAppError("SqlMenuStore.GetMenusByUserId", "store.sql_menu.get_requests_by_class.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return medicineRequests, nil
}

func (s SqlMedicineStore) GetRequestsByKid(kidId string) ([]*model.MedicineRequest, *model.AppError) {
	var medicineRequests []*model.MedicineRequest
	if _, err := s.GetReplica().Select(&medicineRequests, "SELECT MedicineRequests.* FROM MedicineRequests WHERE MedicineRequests.KidId = :KidId AND MedicineRequests.DeleteAt = 0", map[string]interface{}{"KidId": kidId}); err != nil {
		return nil, model.NewAppError("SqlMenuStore.GetMenusByUserId", "store.sql_menu.get_requests_by_kid.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return medicineRequests, nil
}

func (s SqlMedicineStore) GetMedicinesByRequest(requestId string) ([]*model.Medicine, *model.AppError) {
	var medicines []*model.Medicine
	if _, err := s.GetReplica().Select(&medicines, "SELECT Medicines.* FROM Medicines WHERE Medicines.RequestId = :RequestId AND Medicines.DeleteAt = 0", map[string]interface{}{"RequestId": requestId}); err != nil {
		return nil, model.NewAppError("SqlMenuStore.GetMenusByUserId", "store.sql_menu.get_medicines_by_request.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return medicines, nil
}

func (s SqlMedicineStore) SaveMedicine(medicine *model.Medicine) (*model.Medicine, *model.AppError) {
	if len(medicine.Id) > 0 {
		return nil, model.NewAppError("SqlMedicineStore.Save",
			"store.sql_medicineRequest.save.existing.app_error", nil, "id="+medicine.Id, http.StatusBadRequest)
	}
	if len(medicine.MedicineRequestId) != 26 {
		return nil, model.NewAppError("SqlMedicineStore.Save",
			"store.sql_medicineRequest.save.invalid_request_id.app_error", nil, "id="+medicine.Id, http.StatusBadRequest)
	}

	medicine.PreSave()

	if err := medicine.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(medicine); err != nil {
		return nil, model.NewAppError("SqlMedicineStore.Save", "store.sql_medicineRequest.save.app_error", nil, "id="+medicine.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return medicine, nil
}
