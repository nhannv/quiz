// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type SqlSchoolStore struct {
	SqlStore
}

type schoolMember struct {
	SchoolId      string
	UserId        string
	Roles         string
	DeleteAt      int64
	SchemeParent  sql.NullBool
	SchemeAdmin   sql.NullBool
	SchemeTeacher sql.NullBool
}

func NewSchoolMemberFromModel(tm *model.SchoolMember) *schoolMember {
	return &schoolMember{
		SchoolId:      tm.SchoolId,
		UserId:        tm.UserId,
		Roles:         tm.ExplicitRoles,
		DeleteAt:      tm.DeleteAt,
		SchemeParent:  sql.NullBool{Valid: true, Bool: tm.SchemeParent},
		SchemeTeacher: sql.NullBool{Valid: true, Bool: tm.SchemeTeacher},
		SchemeAdmin:   sql.NullBool{Valid: true, Bool: tm.SchemeAdmin},
	}
}

type schoolMemberWithSchemeRoles struct {
	SchoolId                       string
	UserId                         string
	Roles                          string
	DeleteAt                       int64
	SchemeParent                   sql.NullBool
	SchemeTeacher                  sql.NullBool
	SchemeAdmin                    sql.NullBool
	SchoolSchemeDefaultParentRole  sql.NullString
	SchoolSchemeDefaultTeacherRole sql.NullString
	SchoolSchemeDefaultAdminRole   sql.NullString
}

type schoolMemberWithSchemeRolesList []schoolMemberWithSchemeRoles

func (db schoolMemberWithSchemeRoles) ToModel() *model.SchoolMember {
	var roles []string
	var explicitRoles []string

	// Identify any scheme derived roles that are in "Roles" field due to not yet being migrated, and exclude
	// them from ExplicitRoles field.
	schemeParent := db.SchemeParent.Valid && db.SchemeParent.Bool
	schemeTeacher := db.SchemeTeacher.Valid && db.SchemeTeacher.Bool
	schemeAdmin := db.SchemeAdmin.Valid && db.SchemeAdmin.Bool
	for _, role := range strings.Fields(db.Roles) {
		isImplicit := false
		if role == model.SCHOOL_PARENT_ROLE_ID {
			// We have an implicit role via the system scheme. Override the "schemeGuest" field to true.
			schemeParent = true
			isImplicit = true
		} else if role == model.SCHOOL_TEACHER_ROLE_ID {
			// We have an implicit role via the system scheme. Override the "schemeUser" field to true.
			schemeTeacher = true
			isImplicit = true
		} else if role == model.TEAM_ADMIN_ROLE_ID {
			// We have an implicit role via the system scheme.
			schemeAdmin = true
			isImplicit = true
		}

		if !isImplicit {
			explicitRoles = append(explicitRoles, role)
		}
		roles = append(roles, role)
	}

	// Add any scheme derived roles that are not in the Roles field due to being Implicit from the Scheme, and add
	// them to the Roles field for backwards compatibility reasons.
	var schemeImpliedRoles []string
	if db.SchemeParent.Valid && db.SchemeParent.Bool {
		if db.SchoolSchemeDefaultParentRole.Valid && db.SchoolSchemeDefaultParentRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.SchoolSchemeDefaultParentRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.SCHOOL_PARENT_ROLE_ID)
		}
	}
	if db.SchemeTeacher.Valid && db.SchemeTeacher.Bool {
		if db.SchoolSchemeDefaultTeacherRole.Valid && db.SchoolSchemeDefaultTeacherRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.SchoolSchemeDefaultTeacherRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.SCHOOL_TEACHER_ROLE_ID)
		}
	}
	if db.SchemeAdmin.Valid && db.SchemeAdmin.Bool {
		if db.SchoolSchemeDefaultAdminRole.Valid && db.SchoolSchemeDefaultAdminRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.SchoolSchemeDefaultAdminRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.SCHOOL_ADMIN_ROLE_ID)
		}
	}
	for _, impliedRole := range schemeImpliedRoles {
		alreadyThere := false
		for _, role := range roles {
			if role == impliedRole {
				alreadyThere = true
			}
		}
		if !alreadyThere {
			roles = append(roles, impliedRole)
		}
	}

	tm := &model.SchoolMember{
		SchoolId:      db.SchoolId,
		UserId:        db.UserId,
		Roles:         strings.Join(roles, " "),
		DeleteAt:      db.DeleteAt,
		SchemeParent:  schemeParent,
		SchemeTeacher: schemeTeacher,
		SchemeAdmin:   schemeAdmin,
		ExplicitRoles: strings.Join(explicitRoles, " "),
	}
	return tm
}

func (db schoolMemberWithSchemeRolesList) ToModel() []*model.SchoolMember {
	tms := make([]*model.SchoolMember, 0)

	for _, tm := range db {
		tms = append(tms, tm.ToModel())
	}

	return tms
}

func NewSqlSchoolStore(sqlStore SqlStore) store.SchoolStore {
	s := &SqlSchoolStore{
		sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.School{}, "Schools").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(128)
		table.ColMap("Description").SetMaxSize(255)
		table.ColMap("Address").SetMaxSize(255)
		table.ColMap("Phone").SetMaxSize(11)
		table.ColMap("Email").SetMaxSize(128).SetUnique(true)
		table.ColMap("ContactName").SetMaxSize(64)
		table.ColMap("InviteId").SetMaxSize(32)

		tablem := db.AddTableWithName(schoolMember{}, "SchoolMembers").SetKeys(false, "SchoolId", "UserId")
		tablem.ColMap("SchoolId").SetMaxSize(26)
		tablem.ColMap("UserId").SetMaxSize(26)
		tablem.ColMap("Roles").SetMaxSize(64)

		tablb := db.AddTableWithName(model.Branch{}, "Branches").SetKeys(false, "Id")
		tablb.ColMap("Name").SetMaxSize(128)
		tablb.ColMap("Description").SetMaxSize(255)
		tablb.ColMap("SchoolId").SetMaxSize(26)
		tablb.ColMap("CreatorId").SetMaxSize(26)

		tablc := db.AddTableWithName(model.Class{}, "Classes").SetKeys(false, "Id")
		tablc.ColMap("Name").SetMaxSize(128)
		tablc.ColMap("Description").SetMaxSize(255)
		tablc.ColMap("SchoolId").SetMaxSize(26)
		tablc.ColMap("BranchId").SetMaxSize(26)
		tablc.ColMap("CreatorId").SetMaxSize(26)
		tablc.ColMap("InviteId").SetMaxSize(32)
	}

	return s
}

func (s SqlSchoolStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_schools_email", "Schools", "Email")
	s.RemoveIndexIfExists("idx_schools_description", "Schools")
	s.CreateIndexIfNotExists("idx_schools_invite_id", "Schools", "InviteId")
	s.CreateIndexIfNotExists("idx_schools_update_at", "Schools", "UpdateAt")
	s.CreateIndexIfNotExists("idx_schools_create_at", "Schools", "CreateAt")
	s.CreateIndexIfNotExists("idx_schools_delete_at", "Schools", "DeleteAt")

	s.CreateIndexIfNotExists("idx_schoolmembers_school_id", "SchoolMembers", "SchoolId")
	s.CreateIndexIfNotExists("idx_schoolmembers_user_id", "SchoolMembers", "UserId")
	s.CreateIndexIfNotExists("idx_schoolmembers_delete_at", "SchoolMembers", "DeleteAt")

	s.CreateIndexIfNotExists("idx_branches_school_id", "Branches", "SchoolId")
	s.CreateIndexIfNotExists("idx_branches_delete_at", "Branches", "DeleteAt")

	s.CreateIndexIfNotExists("idx_classes_school_id", "Classes", "SchoolId")
	s.CreateIndexIfNotExists("idx_classes_branch_id", "Classes", "BranchId")
	s.CreateIndexIfNotExists("idx_classes_delete_at", "Classes", "DeleteAt")
}

func (s SqlSchoolStore) Save(school *model.School) (*model.School, *model.AppError) {
	if len(school.Id) > 0 {
		return nil, model.NewAppError("SqlSchoolStore.Save",
			"store.sql_school.save.existing.app_error", nil, "id="+school.Id, http.StatusBadRequest)
	}

	school.PreSave()

	if err := school.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(school); err != nil {
		if IsUniqueConstraintError(err, []string{"Email", "schools_email_key"}) {
			return nil, model.NewAppError("SqlSchoolStore.Save", "store.sql_school.save.email_exists.app_error", nil, "id="+school.Id+", "+err.Error(), http.StatusBadRequest)
		}
		return nil, model.NewAppError("SqlSchoolStore.Save", "store.sql_school.save.app_error", nil, "id="+school.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return school, nil
}

func (s SqlSchoolStore) Update(school *model.School) (*model.School, *model.AppError) {

	school.PreUpdate()

	if err := school.IsValid(); err != nil {
		return nil, err
	}

	oldResult, err := s.GetMaster().Get(model.School{}, school.Id)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.Update", "store.sql_school.update.finding.app_error", nil, "id="+school.Id+", "+err.Error(), http.StatusInternalServerError)

	}

	if oldResult == nil {
		return nil, model.NewAppError("SqlSchoolStore.Update", "store.sql_school.update.find.app_error", nil, "id="+school.Id, http.StatusBadRequest)
	}

	oldSchool := oldResult.(*model.School)
	school.CreateAt = oldSchool.CreateAt
	school.UpdateAt = model.GetMillis()

	count, err := s.GetMaster().Update(school)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.Update", "store.sql_school.update.updating.app_error", nil, "id="+school.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	if count != 1 {
		return nil, model.NewAppError("SqlSchoolStore.Update", "store.sql_school.update.app_error", nil, "id="+school.Id, http.StatusInternalServerError)
	}

	return school, nil
}

func (s SqlSchoolStore) Get(id string) (*model.School, *model.AppError) {
	obj, err := s.GetReplica().Get(model.School{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.Get", "store.sql_school.get.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlSchoolStore.Get", "store.sql_school.get.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.School), nil
}

func (s SqlSchoolStore) GetByInviteId(inviteId string) (*model.School, *model.AppError) {
	school := model.School{}

	err := s.GetReplica().SelectOne(&school, "SELECT * FROM Schools WHERE InviteId = :InviteId", map[string]interface{}{"InviteId": inviteId})
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetByInviteId", "store.sql_school.get_by_invite_id.finding.app_error", nil, "inviteId="+inviteId+", "+err.Error(), http.StatusNotFound)
	}

	if len(inviteId) == 0 || school.InviteId != inviteId {
		return nil, model.NewAppError("SqlSchoolStore.GetByInviteId", "store.sql_school.get_by_invite_id.find.app_error", nil, "inviteId="+inviteId, http.StatusNotFound)
	}
	return &school, nil
}

func (us SqlSchoolStore) UpdateLastSchoolIconUpdate(schoolId string, curTime int64) *model.AppError {
	if _, err := us.GetMaster().Exec("UPDATE Schools SET LastSchoolIconUpdate = :Time, UpdateAt = :Time WHERE Id = :schoolId", map[string]interface{}{"Time": curTime, "schoolId": schoolId}); err != nil {
		return model.NewAppError("SqlSchoolStore.UpdateLastSchoolIconUpdate", "store.sql_school.update_last_school_icon_update.app_error", nil, "school_id="+schoolId, http.StatusInternalServerError)
	}
	return nil
}

func (s SqlSchoolStore) GetSchoolsByUserId(userId string) ([]*model.School, *model.AppError) {
	var schools []*model.School
	if _, err := s.GetReplica().Select(&schools, "SELECT Schools.* FROM Schools, SchoolMembers WHERE SchoolMembers.SchoolId = Schools.Id AND SchoolMembers.UserId = :UserId AND SchoolMembers.DeleteAt = 0 AND Schools.DeleteAt = 0", map[string]interface{}{"UserId": userId}); err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetSchoolsByUserId", "store.sql_school.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return schools, nil
}

func (s SqlSchoolStore) getSchoolMembersWithSchemeSelectQuery() sq.SelectBuilder {
	return s.getQueryBuilder().
		Select(
			"SchoolMembers.*",
			"SchoolScheme.DefaultSchoolParentRole SchoolSchemeDefaultParentRole",
			"SchoolScheme.DefaultSchoolTeacherRole SchoolSchemeDefaultTeacherRole",
			"SchoolScheme.DefaultSchoolAdminRole SchoolSchemeDefaultAdminRole",
		).
		From("SchoolMembers").
		LeftJoin("Schools ON SchoolMembers.SchoolId = Schools.Id").
		LeftJoin("Schemes SchoolScheme ON Schools.SchemeId = SchoolScheme.Id")
}

func (s SqlSchoolStore) SaveMember(member *model.SchoolMember, maxUsersPerSchool int) (*model.SchoolMember, *model.AppError) {
	defer s.InvalidateAllSchoolIdsForUser(member.UserId)
	if err := member.IsValid(); err != nil {
		return nil, err
	}
	dbMember := NewSchoolMemberFromModel(member)

	// TODO check max kids per school
	//if maxUsersPerSchool >= 0 {
	_, err := s.GetMaster().SelectInt(
		`SELECT
					COUNT(0)
				FROM
					SchoolMembers
				INNER JOIN
					Users
				ON
					SchoolMembers.UserId = Users.Id
				WHERE
					SchoolId = :SchoolId
					AND SchoolMembers.DeleteAt = 0
					AND Users.DeleteAt = 0`, map[string]interface{}{"SchoolId": member.SchoolId})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.member_count.app_error", nil, "schoolId="+member.SchoolId+", "+err.Error(), http.StatusInternalServerError)
	}

	//if count >= int64(maxUsersPerSchool) {
	//	return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.max_accounts.app_error", nil, "schoolId="+member.SchoolId, http.StatusBadRequest)
	//}
	//}

	if err := s.GetMaster().Insert(dbMember); err != nil {
		if IsUniqueConstraintError(err, []string{"SchoolId", "schoolmembers_pkey", "PRIMARY"}) {
			return nil, model.NewAppError("SqlSchoolStore.SaveMember", TEAM_MEMBER_EXISTS_ERROR, nil, "school_id="+member.SchoolId+", user_id="+member.UserId+", "+err.Error(), http.StatusBadRequest)
		}
		return nil, model.NewAppError("SqlSchoolStore.SaveMember", "store.sql_school.save_member.save.app_error", nil, "school_id="+member.SchoolId+", user_id="+member.UserId+", "+err.Error(), http.StatusInternalServerError)
	}

	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.SchoolId": dbMember.SchoolId}).
		Where(sq.Eq{"SchoolMembers.UserId": dbMember.UserId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.SaveMember", "store.sql_school.get_member.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var retrievedMember schoolMemberWithSchemeRoles
	if err := s.GetMaster().SelectOne(&retrievedMember, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlSchoolStore.SaveMember", "store.sql_school.get_member.missing.app_error", nil, "school_id="+dbMember.SchoolId+"user_id="+dbMember.UserId+","+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlSchoolStore.SaveMember", "store.sql_school.get_member.app_error", nil, "school_id="+dbMember.SchoolId+"user_id="+dbMember.UserId+","+err.Error(), http.StatusInternalServerError)
	}

	return retrievedMember.ToModel(), nil
}

func (s SqlSchoolStore) UpdateMember(member *model.SchoolMember) (*model.SchoolMember, *model.AppError) {
	member.PreUpdate()

	if err := member.IsValid(); err != nil {
		return nil, err
	}

	if _, err := s.GetMaster().Update(NewSchoolMemberFromModel(member)); err != nil {
		return nil, model.NewAppError("SqlSchoolStore.UpdateMember", "store.sql_school.save_member.save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.SchoolId": member.SchoolId}).
		Where(sq.Eq{"SchoolMembers.UserId": member.UserId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.UpdateMember", "store.sql_school.get_member.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var retrievedMember schoolMemberWithSchemeRoles
	if err := s.GetMaster().SelectOne(&retrievedMember, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlSchoolStore.UpdateMember", "store.sql_school.get_member.missing.app_error", nil, "school_id="+member.SchoolId+"user_id="+member.UserId+","+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlSchoolStore.UpdateMember", "store.sql_school.get_member.app_error", nil, "school_id="+member.SchoolId+"user_id="+member.UserId+","+err.Error(), http.StatusInternalServerError)
	}

	return retrievedMember.ToModel(), nil
}

func (s SqlSchoolStore) GetMember(schoolId string, userId string) (*model.SchoolMember, *model.AppError) {
	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.SchoolId": schoolId}).
		Where(sq.Eq{"SchoolMembers.UserId": userId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMember", "store.sql_school.get_member.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMember schoolMemberWithSchemeRoles
	err = s.GetReplica().SelectOne(&dbMember, queryString, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlSchoolStore.GetMember", "store.sql_school.get_member.missing.app_error", nil, "schoolId="+schoolId+" userId="+userId+" "+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlSchoolStore.GetMember", "store.sql_school.get_member.app_error", nil, "schoolId="+schoolId+" userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMember.ToModel(), nil
}

func (s SqlSchoolStore) GetMembers(schoolId string, offset int, limit int, restrictions *model.ViewUsersRestrictions) ([]*model.SchoolMember, *model.AppError) {
	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.SchoolId": schoolId}).
		Where(sq.Eq{"SchoolMembers.DeleteAt": 0}).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	query = applySchoolMemberViewRestrictionsFilter(query, schoolId, restrictions)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMembers", "store.sql_school.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMembers schoolMemberWithSchemeRolesList
	_, err = s.GetReplica().Select(&dbMembers, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMembers", "store.sql_school.get_members.app_error", nil, "schoolId="+schoolId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMembers.ToModel(), nil
}

func (s SqlSchoolStore) GetTotalMemberCount(schoolId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError) {
	query := s.getQueryBuilder().
		Select("count(DISTINCT SchoolMembers.UserId)").
		From("SchoolMembers, Users").
		Where("SchoolMembers.DeleteAt = 0").
		Where("SchoolMembers.UserId = Users.Id").
		Where(sq.Eq{"SchoolMembers.SchoolId": schoolId})

	query = applySchoolMemberViewRestrictionsFilterForStats(query, schoolId, restrictions)
	queryString, args, err := query.ToSql()
	if err != nil {
		return int64(0), model.NewAppError("SqlSchoolStore.GetTotalMemberCount", "store.sql_school.get_member_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	count, err := s.GetReplica().SelectInt(queryString, args...)
	if err != nil {
		return int64(0), model.NewAppError("SqlSchoolStore.GetTotalMemberCount", "store.sql_school.get_member_count.app_error", nil, "schoolId="+schoolId+" "+err.Error(), http.StatusInternalServerError)
	}
	return count, nil
}

func (s SqlSchoolStore) GetActiveMemberCount(schoolId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError) {
	query := s.getQueryBuilder().
		Select("count(DISTINCT SchoolMembers.UserId)").
		From("SchoolMembers, Users").
		Where("SchoolMembers.DeleteAt = 0").
		Where("SchoolMembers.UserId = Users.Id").
		Where("Users.DeleteAt = 0").
		Where(sq.Eq{"SchoolMembers.SchoolId": schoolId})

	query = applySchoolMemberViewRestrictionsFilterForStats(query, schoolId, restrictions)
	queryString, args, err := query.ToSql()
	if err != nil {
		return 0, model.NewAppError("SqlSchoolStore.GetActiveMemberCount", "store.sql_school.get_active_member_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	count, err := s.GetReplica().SelectInt(queryString, args...)
	if err != nil {
		return 0, model.NewAppError("SqlSchoolStore.GetActiveMemberCount", "store.sql_school.get_active_member_count.app_error", nil, "schoolId="+schoolId+" "+err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (s SqlSchoolStore) GetMembersByIds(schoolId string, userIds []string, restrictions *model.ViewUsersRestrictions) ([]*model.SchoolMember, *model.AppError) {
	if len(userIds) == 0 {
		return nil, model.NewAppError("SqlSchoolStore.GetMembersByIds", "store.sql_school.get_members_by_ids.app_error", nil, "Invalid list of user ids", http.StatusInternalServerError)
	}

	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.SchoolId": schoolId}).
		Where(sq.Eq{"SchoolMembers.UserId": userIds}).
		Where(sq.Eq{"SchoolMembers.DeleteAt": 0})

	query = applySchoolMemberViewRestrictionsFilter(query, schoolId, restrictions)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMembersByIds", "store.sql_school.get_members_by_ids.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMembers schoolMemberWithSchemeRolesList
	if _, err := s.GetReplica().Select(&dbMembers, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMembersByIds", "store.sql_school.get_members_by_ids.app_error", nil, "schoolId="+schoolId+" "+err.Error(), http.StatusInternalServerError)
	}
	return dbMembers.ToModel(), nil
}

func (s SqlSchoolStore) GetSchoolsForUser(userId string) ([]*model.SchoolMember, *model.AppError) {
	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.UserId": userId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMembers", "store.sql_school.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMembers schoolMemberWithSchemeRolesList
	_, err = s.GetReplica().Select(&dbMembers, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetMembers", "store.sql_school.get_members.app_error", nil, "userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMembers.ToModel(), nil
}

func (s SqlSchoolStore) GetSchoolsForUserWithPagination(userId string, page, perPage int) ([]*model.SchoolMember, *model.AppError) {
	query := s.getSchoolMembersWithSchemeSelectQuery().
		Where(sq.Eq{"SchoolMembers.UserId": userId}).
		Limit(uint64(perPage)).
		Offset(uint64(page * perPage))

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetSchoolsForUserWithPagination", "store.sql_school.get_members.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMembers schoolMemberWithSchemeRolesList
	_, err = s.GetReplica().Select(&dbMembers, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetSchoolsForUserWithPagination", "store.sql_school.get_members.app_error", nil, "userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMembers.ToModel(), nil
}

func (s SqlSchoolStore) ClearCaches() {}

func (s SqlSchoolStore) InvalidateAllSchoolIdsForUser(userId string) {}

func applySchoolMemberViewRestrictionsFilter(query sq.SelectBuilder, schoolId string, restrictions *model.ViewUsersRestrictions) sq.SelectBuilder {
	if restrictions == nil {
		return query
	}

	// If you have no access to schools or e, return and empty result.
	if restrictions.Schools != nil && len(restrictions.Schools) == 0 {
		return query.Where("1 = 0")
	}

	schools := make([]interface{}, len(restrictions.Schools))
	for i, v := range restrictions.Schools {
		schools[i] = v
	}

	resultQuery := query.Join("Users ru ON (SchoolMembers.UserId = ru.Id)")
	if restrictions.Schools != nil && len(restrictions.Schools) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("SchoolMembers rtm ON ( rtm.UserId = ru.Id AND rtm.DeleteAt = 0 AND rtm.SchoolId IN (%s))", sq.Placeholders(len(schools))), schools...)
	}

	return resultQuery.Distinct()
}

func applySchoolMemberViewRestrictionsFilterForStats(query sq.SelectBuilder, schoolId string, restrictions *model.ViewUsersRestrictions) sq.SelectBuilder {
	if restrictions == nil {
		return query
	}

	// If you have no access to schools or branches, return and empty result.
	if restrictions.Schools != nil && len(restrictions.Schools) == 0 {
		return query.Where("1 = 0")
	}

	schools := make([]interface{}, len(restrictions.Schools))
	for i, v := range restrictions.Schools {
		schools[i] = v
	}

	resultQuery := query
	if restrictions.Schools != nil && len(restrictions.Schools) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("SchoolMembers rtm ON ( rtm.UserId = Users.Id AND rtm.DeleteAt = 0 AND rtm.SchoolId IN (%s))", sq.Placeholders(len(schools))), schools...)
	}

	return resultQuery
}

/// BRANCHES

func (s SqlSchoolStore) GetBranches(schoolId string) ([]*model.Branch, *model.AppError) {
	var branches []*model.Branch

	_, err := s.GetReplica().Select(&branches, "SELECT * FROM Branches WHERE SchoolId = :SchoolId AND DeleteAt = 0", map[string]interface{}{"SchoolId": schoolId})
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetBranches", "store.sql_school.get_branches.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return branches, nil
}

func (s SqlSchoolStore) GetBranch(id string) (*model.Branch, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Branch{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.Branch", "store.sql_school.get_branch.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlSchoolStore.Branch", "store.sql_school.get_branch.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Branch), nil
}

func (s SqlSchoolStore) SaveBranch(branch *model.Branch) (*model.Branch, *model.AppError) {
	if len(branch.Id) > 0 {
		return nil, model.NewAppError("SqlSchoolStore.SaveBranch",
			"store.sql_school.save_branch.existing.app_error", nil, "id="+branch.Id, http.StatusBadRequest)
	}

	branch.PreSave()

	if err := branch.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(branch); err != nil {
		return nil, model.NewAppError("SqlSchoolStore.SaveBranch", "store.sql_school.save_branch.app_error", nil, "id="+branch.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return branch, nil
}

// RemoveBranch records the given deleted and updated timestamp to the branch in question.
func (s SqlSchoolStore) RemoveBranch(branchId string) *model.AppError {
	now := model.GetMillis()

	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return model.NewAppError("SqlSchoolStore.SetDeleteAt", "store.sql_branch.set_delete_at.open_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer finalizeTransaction(transaction)

	appErr := s.setDeleteAtT(transaction, branchId, now, now)
	if appErr != nil {
		return appErr
	}

	if err := transaction.Commit(); err != nil {
		return model.NewAppError("SqlSchoolStore.SetDeleteAt", "store.sql_branch.set_delete_at.commit_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (s SqlSchoolStore) setDeleteAtT(transaction *gorp.Transaction, branchId string, deleteAt, updateAt int64) *model.AppError {
	_, err := transaction.Exec("Update Branches SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :BranchId", map[string]interface{}{"DeleteAt": deleteAt, "UpdateAt": updateAt, "BranchId": branchId})
	if err != nil {
		return model.NewAppError("SqlSchoolStore.Delete", "store.sql_branch.delete.branch.app_error", nil, "id="+branchId+", err="+err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// Classes logics

func (s SqlSchoolStore) GetClasses(schoolId string) ([]*model.Class, *model.AppError) {
	var classes []*model.Class

	_, err := s.GetReplica().Select(&classes, "SELECT * FROM Classes WHERE SchoolId = :SchoolId AND DeleteAt = 0", map[string]interface{}{"SchoolId": schoolId})
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetClasses", "store.sql_school.get_classes.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return classes, nil
}

func (s SqlSchoolStore) GetClass(id string) (*model.Class, *model.AppError) {
	obj, err := s.GetReplica().Get(model.Class{}, id)
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.Class", "store.sql_school.get_class.finding.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
	}
	if obj == nil {
		return nil, model.NewAppError("SqlSchoolStore.Class", "store.sql_school.get_class.find.app_error", nil, "id="+id, http.StatusNotFound)
	}

	return obj.(*model.Class), nil
}

func (s SqlSchoolStore) GetClassesByBranch(branchId string) ([]*model.Class, *model.AppError) {
	var classes []*model.Class

	_, err := s.GetReplica().Select(&classes, "SELECT * FROM Classes WHERE BranchId = :BranchId AND DeleteAt = 0", map[string]interface{}{"BranchId": branchId})
	if err != nil {
		return nil, model.NewAppError("SqlSchoolStore.GetClasses", "store.sql_school.get_classes_by_branch.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return classes, nil
}

func (s SqlSchoolStore) SaveClass(class *model.Class) (*model.Class, *model.AppError) {
	if len(class.Id) > 0 {
		return nil, model.NewAppError("SqlSchoolStore.SaveClass",
			"store.sql_school.save_class.existing.app_error", nil, "id="+class.Id, http.StatusBadRequest)
	}

	class.PreSave()

	if err := class.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(class); err != nil {
		return nil, model.NewAppError("SqlSchoolStore.SaveClass", "store.sql_school.save_class.app_error", nil, "id="+class.Id+", "+err.Error(), http.StatusInternalServerError)
	}
	return class, nil
}

// RemoveClass records the given deleted and updated timestamp to the class in question.
func (s SqlSchoolStore) RemoveClass(classId string) *model.AppError {
	now := model.GetMillis()

	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return model.NewAppError("SqlSchoolStore.SetDeleteAt", "store.sql_class.set_delete_at.open_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer finalizeTransaction(transaction)

	appErr := s.setDeleteAtT(transaction, classId, now, now)
	if appErr != nil {
		return appErr
	}

	if err := transaction.Commit(); err != nil {
		return model.NewAppError("SqlSchoolStore.SetDeleteAt", "store.sql_class.set_delete_at.commit_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (s SqlSchoolStore) setDeleteClassAtT(transaction *gorp.Transaction, classId string, deleteAt, updateAt int64) *model.AppError {
	_, err := transaction.Exec("Update Classes SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :ClassId", map[string]interface{}{"DeleteAt": deleteAt, "UpdateAt": updateAt, "ClassId": classId})
	if err != nil {
		return model.NewAppError("SqlSchoolStore.Delete", "store.sql_class.delete.class.app_error", nil, "id="+classId+", err="+err.Error(), http.StatusInternalServerError)
	}

	return nil
}
