// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlKidStore struct {
	SqlStore
}

type kidGuardian struct {
	KidId    string
	UserId   string
	IsParent bool
	DeleteAt int64
}

func NewKidGuardianFromModel(tm *model.KidGuardian) *kidGuardian {
	return &kidGuardian{
		KidId:    tm.KidId,
		UserId:   tm.UserId,
		IsParent: tm.IsParent,
		DeleteAt: tm.DeleteAt,
	}
}

type kidGuardianList []kidGuardian

func (db kidGuardian) ToModel() *model.KidGuardian {
	kp := &model.KidGuardian{
		KidId:    db.KidId,
		UserId:   db.UserId,
		IsParent: db.IsParent,
		DeleteAt: db.DeleteAt,
	}
	return kp
}

func (db kidGuardianList) ToModel() []*model.KidGuardian {
	kps := make([]*model.KidGuardian, 0)

	for _, kp := range db {
		kps = append(kps, kp.ToModel())
	}

	return kps
}

func NewSqlKidStore(sqlStore SqlStore) store.KidStore {
	s := &SqlKidStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Kid{}, "Kids").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("FirstName").SetMaxSize(24)
		table.ColMap("LastName").SetMaxSize(24)
		table.ColMap("NickName").SetMaxSize(24)
		table.ColMap("Description").SetMaxSize(255)
		table.ColMap("Avatar").SetMaxSize(1024)
		table.ColMap("Cover").SetMaxSize(1024)
		table.ColMap("Dob").SetMaxSize(1)
		table.ColMap("Gender").SetMaxSize(1)
		table.ColMap("ClassId").SetMaxSize(26)

		tablem := db.AddTableWithName(kidGuardian{}, "KidGuardians").SetKeys(false, "KidId", "UserId")
		tablem.ColMap("KidId").SetMaxSize(26)
		tablem.ColMap("UserId").SetMaxSize(26)
		tablem.ColMap("IsParent").SetMaxSize(1)
	}

	return s
}

func (s SqlKidStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_kids_firstname", "Kids", "FirstName")
	s.CreateIndexIfNotExists("idx_kids_lastname", "Kids", "LastName")
	s.RemoveIndexIfExists("idx_kids_description", "Kids")
	s.CreateIndexIfNotExists("idx_kids_update_at", "Kids", "UpdateAt")
	s.CreateIndexIfNotExists("idx_kids_create_at", "Kids", "CreateAt")
	s.CreateIndexIfNotExists("idx_kids_delete_at", "Kids", "DeleteAt")
	s.CreateIndexIfNotExists("idx_kids_class_id", "Kids", "ClassId")

	s.CreateIndexIfNotExists("idx_kidguardians_kid_id", "KidGuardians", "KidId")
	s.CreateIndexIfNotExists("idx_kidguardians_user_id", "KidGuardians", "UserId")
	s.CreateIndexIfNotExists("idx_kidguardians_delete_at", "KidGuardians", "DeleteAt")
}

func (s SqlKidStore) Save(kid *model.Kid) (*model.Kid, *model.AppError) {
	if len(kid.Id) > 0 {
		return nil, model.NewAppError("SqlKidStore.Save",
			"store.sql_kid.save.existing.app_error", nil, "id="+kid.Id, http.StatusBadRequest)
	}

	kid.PreSave()

	if err := kid.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(kid); err != nil {
		return nil, model.NewAppError("SqlKidStore.Save", "store.sql_kid.save.app_error", nil, "id="+kid.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return kid, nil
}

func (s SqlKidStore) Update(kid *model.Kid) (*model.Kid, *model.AppError) {

	kid.PreUpdate()

	if err := kid.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.Kid{}, kid.Id)
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.Update", "store.sql_kid.update.finding.app_error", nil, "id="+kid.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlKidStore.Update", "store.sql_kid.update.find.app_error", nil, "id="+kid.Id, http.StatusBadRequest)
	}

	oldKid := oldResult.(*model.Kid)
	kid.CreateAt = oldKid.CreateAt
	kid.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(kid)
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.Update", "store.sql_kid.update.updating.app_error", nil, "id="+kid.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlKidStore.Update", "store.sql_kid.update.app_error", nil, "id="+kid.Id, http.StatusInternalServerError)
	}

	return kid, nil
}

func (s SqlKidStore) Get(id string) (*model.Kid, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Kid{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.Get", "store.sql_kid.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlKidStore.Get", "store.sql_kid.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Kid), nil
}

func (s SqlKidStore) GetKidsByUserId(userId string) ([]*model.Kid, *model.AppError) {
	var kids []*model.Kid
	if _, err := s.GetReplica().Select(&kids, "SELECT Kids.* FROM Kids, KidGuardians WHERE KidGuardians.KidId = Kids.Id AND KidGuardians.UserId = :UserId AND KidGuardians.DeleteAt = 0 AND Kids.DeleteAt = 0", map[string]interface{}{"UserId": userId}); err != nil {
		return nil, model.NewAppError("SqlKidStore.GetKidsByUserId", "store.sql_kid.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return kids, nil
}

func (s SqlKidStore) getKidGuardiansSelectQuery() sq.SelectBuilder {
	return s.getQueryBuilder().
		Select(
			"KidGuardians.*",
		).
		From("KidGuardians").
		LeftJoin("Kids ON KidGuardians.KidId = Kids.Id")
}

func (s SqlKidStore) SaveGuardian(guardian *model.KidGuardian) (*model.KidGuardian, *model.AppError) {
	defer s.InvalidateAllKidIdsForUser(guardian.UserId)
	if err := guardian.IsValid(); err != nil {
		return nil, err
	}
	dbGuardian := NewKidGuardianFromModel(guardian)

	count, err := s.GetMaster().SelectInt(
		`SELECT
			COUNT(0)
		FROM
			KidGuardians
		INNER JOIN
			Users
		ON
			KidGuardians.UserId = Users.Id
		WHERE
			KidId = :KidId
			AND KidGuardians.DeleteAt = 0
			AND Users.DeleteAt = 0`, map[string]interface{}{"KidId": guardian.KidId})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.guardian_count.app_error", nil, "kidId="+guardian.KidId+", "+err.Error(), http.StatusInternalServerError)
	}

	if count >= 3 {
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.max_accounts.app_error", nil, "kidId="+guardian.KidId, http.StatusBadRequest)
	}

	if err := s.GetMaster().Insert(dbGuardian); err != nil {
		if IsUniqueConstraintError(err, []string{"KidId", "kidguardians_pkey", "PRIMARY"}) {
			return nil, model.NewAppError("SqlKidStore.SaveGuardian", TEAM_MEMBER_EXISTS_ERROR, nil, "kid_id="+guardian.KidId+", user_id="+guardian.UserId+", "+err.Error(), http.StatusBadRequest)
		}
		return nil, model.NewAppError("SqlKidStore.SaveGuardian", "store.sql_kid.save_guardian.save.app_error", nil, "kid_id="+guardian.KidId+", user_id="+guardian.UserId+", "+err.Error(), http.StatusInternalServerError)
	}

	query := s.getKidGuardiansSelectQuery().
		Where(sq.Eq{"KidGuardians.KidId": dbGuardian.KidId}).
		Where(sq.Eq{"KidGuardians.UserId": dbGuardian.UserId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.SaveGuardian", "store.sql_kid.get_guardian.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var retrievedGuardian kidGuardian
	if err := s.GetMaster().SelectOne(&retrievedGuardian, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlKidStore.SaveGuardian", "store.sql_kid.get_guardian.missing.app_error", nil, "kid_id="+dbGuardian.KidId+"user_id="+dbGuardian.UserId+","+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlKidStore.SaveGuardian", "store.sql_kid.get_guardian.app_error", nil, "kid_id="+dbGuardian.KidId+"user_id="+dbGuardian.UserId+","+err.Error(), http.StatusInternalServerError)
	}

	return retrievedGuardian.ToModel(), nil
}

func (s SqlKidStore) UpdateGuardian(guardian *model.KidGuardian) (*model.KidGuardian, *model.AppError) {
	guardian.PreUpdate()

	if err := guardian.IsValid(); err != nil {
		return nil, err
	}

	if _, err := s.GetMaster().Update(NewKidGuardianFromModel(guardian)); err != nil {
		return nil, model.NewAppError("SqlKidStore.UpdateGuardian", "store.sql_kid.save_guardian.save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	query := s.getKidGuardiansSelectQuery().
		Where(sq.Eq{"KidGuardians.KidId": guardian.KidId}).
		Where(sq.Eq{"KidGuardians.UserId": guardian.UserId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.UpdateGuardian", "store.sql_kid.get_guardian.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var retrievedGuardian kidGuardian
	if err := s.GetMaster().SelectOne(&retrievedGuardian, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlKidStore.UpdateGuardian", "store.sql_kid.get_guardian.missing.app_error", nil, "kid_id="+guardian.KidId+"user_id="+guardian.UserId+","+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlKidStore.UpdateGuardian", "store.sql_kid.get_guardian.app_error", nil, "kid_id="+guardian.KidId+"user_id="+guardian.UserId+","+err.Error(), http.StatusInternalServerError)
	}

	return retrievedGuardian.ToModel(), nil
}

func (s SqlKidStore) GetGuardian(kidId string, userId string) (*model.KidGuardian, *model.AppError) {
	query := s.getKidGuardiansSelectQuery().
		Where(sq.Eq{"KidGuardians.KidId": kidId}).
		Where(sq.Eq{"KidGuardians.UserId": userId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardian", "store.sql_kid.get_guardian.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbGuardian kidGuardian
	err = s.GetReplica().SelectOne(&dbGuardian, queryString, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlKidStore.GetGuardian", "store.sql_kid.get_guardian.missing.app_error", nil, "kidId="+kidId+" userId="+userId+" "+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlKidStore.GetGuardian", "store.sql_kid.get_guardian.app_error", nil, "kidId="+kidId+" userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbGuardian.ToModel(), nil
}

func (s SqlKidStore) GetGuardians(kidId string, offset int, limit int) ([]*model.KidGuardian, *model.AppError) {
	query := s.getKidGuardiansSelectQuery().
		Where(sq.Eq{"KidGuardians.KidId": kidId}).
		Where(sq.Eq{"KidGuardians.DeleteAt": 0}).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardians", "store.sql_kid.get_guardians.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbGuardians kidGuardianList
	_, err = s.GetReplica().Select(&dbGuardians, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardians", "store.sql_kid.get_guardians.app_error", nil, "kidId="+kidId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbGuardians.ToModel(), nil
}

func (s SqlKidStore) GetTotalGuardianCount(kidId string) (int64, *model.AppError) {
	query := s.getQueryBuilder().
		Select("count(DISTINCT KidGuardians.UserId)").
		From("KidGuardians, Users").
		Where("KidGuardians.DeleteAt = 0").
		Where("KidGuardians.UserId = Users.Id").
		Where(sq.Eq{"KidGuardians.KidId": kidId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return int64(0), model.NewAppError("SqlKidStore.GetTotalGuardianCount", "store.sql_kid.get_guardian_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	count, err := s.GetReplica().SelectInt(queryString, args...)
	if err != nil {
		return int64(0), model.NewAppError("SqlKidStore.GetTotalGuardianCount", "store.sql_kid.get_guardian_count.app_error", nil, "kidId="+kidId+" "+err.Error(), http.StatusInternalServerError)
	}
	return count, nil
}

func (s SqlKidStore) GetActiveGuardianCount(kidId string) (int64, *model.AppError) {
	query := s.getQueryBuilder().
		Select("count(DISTINCT KidGuardians.UserId)").
		From("KidGuardians, Users").
		Where("KidGuardians.DeleteAt = 0").
		Where("KidGuardians.UserId = Users.Id").
		Where("Users.DeleteAt = 0").
		Where(sq.Eq{"KidGuardians.KidId": kidId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return 0, model.NewAppError("SqlKidStore.GetActiveGuardianCount", "store.sql_kid.get_active_guardian_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	count, err := s.GetReplica().SelectInt(queryString, args...)
	if err != nil {
		return 0, model.NewAppError("SqlKidStore.GetActiveGuardianCount", "store.sql_kid.get_active_guardian_count.app_error", nil, "kidId="+kidId+" "+err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (s SqlKidStore) GetGuardiansByIds(kidId string, userIds []string) ([]*model.KidGuardian, *model.AppError) {
	if len(userIds) == 0 {
		return nil, model.NewAppError("SqlKidStore.GetGuardiansByIds", "store.sql_kid.get_guardians_by_ids.app_error", nil, "Invalid list of user ids", http.StatusInternalServerError)
	}

	query := s.getKidGuardiansSelectQuery().
		Where(sq.Eq{"KidGuardians.KidId": kidId}).
		Where(sq.Eq{"KidGuardians.UserId": userIds}).
		Where(sq.Eq{"KidGuardians.DeleteAt": 0})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardiansByIds", "store.sql_kid.get_guardians_by_ids.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbGuardians kidGuardianList
	if _, err := s.GetReplica().Select(&dbGuardians, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardiansByIds", "store.sql_kid.get_guardians_by_ids.app_error", nil, "kidId="+kidId+" "+err.Error(), http.StatusInternalServerError)
	}
	return dbGuardians.ToModel(), nil
}

func (s SqlKidStore) GetKidsForUser(userId string) ([]*model.KidGuardian, *model.AppError) {
	query := s.getKidGuardiansSelectQuery().
		Where(sq.Eq{"KidGuardians.UserId": userId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardians", "store.sql_kid.get_guardians.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbGuardians kidGuardianList
	_, err = s.GetReplica().Select(&dbGuardians, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlKidStore.GetGuardians", "store.sql_kid.get_guardians.app_error", nil, "userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbGuardians.ToModel(), nil
}

func (s SqlKidStore) ClearCaches() {}

func (s SqlKidStore) InvalidateAllKidIdsForUser(userId string) {}
