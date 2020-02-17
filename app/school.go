// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/disintegration/imaging"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *App) CreateSchool(school *model.School) (*model.School, *model.AppError) {
	school.InviteId = ""
	rschool, err := a.Srv().Store.School().Save(school)
	if err != nil {
		return nil, err
	}

	// TODO should create default branch and class here
	// if _, err := a.CreateDefaultChannels(rschool.Id); err != nil {
	// 	return nil, err
	// }

	return rschool, nil
}

func (a *App) CreateSchoolWithUser(school *model.School, userId string) (*model.School, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}
	school.Email = user.Email

	rschool, err := a.CreateSchool(school)
	if err != nil {
		return nil, err
	}

	if err = a.JoinUserToSchool(rschool, user, ""); err != nil {
		return nil, err
	}

	return rschool, nil
}

func (a *App) UpdateSchool(school *model.School) (*model.School, *model.AppError) {
	oldSchool, err := a.GetSchool(school.Id)
	if err != nil {
		return nil, err
	}

	oldSchool.Name = school.Name
	oldSchool.Description = school.Description
	oldSchool.ContactName = school.ContactName
	oldSchool.Phone = school.Phone
	oldSchool.LastSchoolIconUpdate = school.LastSchoolIconUpdate
	oldSchool.Address = school.Address

	oldSchool, err = a.updateSchoolUnsanitized(oldSchool)
	if err != nil {
		return school, err
	}

	return oldSchool, nil
}

func (a *App) updateSchoolUnsanitized(school *model.School) (*model.School, *model.AppError) {
	return a.Srv().Store.School().Update(school)
}

// RenameSchool is used to rename the school Name and the DisplayName fields
func (a *App) RenameSchool(school *model.School, newSchoolName string, newDisplayName string) (*model.School, *model.AppError) {
	if newSchoolName != "-" {
		school.Name = newSchoolName
	}

	newSchool, err := a.updateSchoolUnsanitized(school)
	if err != nil {
		return nil, err
	}

	return newSchool, nil
}

func (a *App) PatchSchool(schoolId string, patch *model.SchoolPatch) (*model.School, *model.AppError) {
	school, err := a.GetSchool(schoolId)
	if err != nil {
		return nil, err
	}

	school.Patch(patch)

	updatedSchool, err := a.UpdateSchool(school)
	if err != nil {
		return nil, err
	}

	return updatedSchool, nil
}

func (a *App) GetSchool(schoolId string) (*model.School, *model.AppError) {
	return a.Srv().Store.School().Get(schoolId)
}

func (a *App) SanitizeSchool(session model.Session, school *model.School) *model.School {
	if a.SessionHasPermissionToSchool(session, school.Id, model.PERMISSION_MANAGE_SCHOOL) {
		return school
	}
	school.Sanitize()

	return school
}

func (a *App) SanitizeSchools(session model.Session, schools []*model.School) []*model.School {
	for _, school := range schools {
		a.SanitizeSchool(session, school)
	}

	return schools
}

func (a *App) GetSchoolIcon(school *model.School) ([]byte, *model.AppError) {
	if len(*a.Config().FileSettings.DriverName) == 0 {
		return nil, model.NewAppError("GetSchoolIcon", "api.school.get_school_icon.filesettings_no_driver.app_error", nil, "", http.StatusNotImplemented)
	}

	path := "schools/" + school.Id + "/schoolIcon.png"
	data, err := a.ReadFile(path)
	if err != nil {
		return nil, model.NewAppError("GetSchoolIcon", "api.school.get_school_icon.read_file.app_error", nil, err.Error(), http.StatusNotFound)
	}

	return data, nil
}

func (a *App) SetSchoolIcon(schoolId string, imageData *multipart.FileHeader) *model.AppError {
	file, err := imageData.Open()
	if err != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.open.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	defer file.Close()
	return a.SetSchoolIconFromMultiPartFile(schoolId, file)
}

func (a *App) SetSchoolIconFromMultiPartFile(schoolId string, file multipart.File) *model.AppError {
	school, getSchoolErr := a.GetSchool(schoolId)

	if getSchoolErr != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.get_school.app_error", nil, getSchoolErr.Error(), http.StatusBadRequest)
	}

	if len(*a.Config().FileSettings.DriverName) == 0 {
		return model.NewAppError("setSchoolIcon", "api.school.set_school_icon.storage.app_error", nil, "", http.StatusNotImplemented)
	}

	// Decode image config first to check dimensions before loading the whole thing into memory later on
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.decode_config.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	if config.Width*config.Height > model.MaxImageSize {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.too_large.app_error", nil, "", http.StatusBadRequest)
	}

	file.Seek(0, 0)

	return a.SetSchoolIconFromFile(school, file)
}

func (a *App) SetSchoolIconFromFile(school *model.School, file io.Reader) *model.AppError {
	// Decode image into Image object
	img, _, err := image.Decode(file)
	if err != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.decode.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	orientation, _ := getImageOrientation(file)
	img = makeImageUpright(img, orientation)

	// Scale school icon
	schoolIconWidthAndHeight := 128
	img = imaging.Fill(img, schoolIconWidthAndHeight, schoolIconWidthAndHeight, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.encode.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path := "schools/" + school.Id + "/schoolIcon.png"

	if _, err := a.WriteFile(buf, path); err != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.set_school_icon.write_file.app_error", nil, "", http.StatusInternalServerError)
	}

	curTime := model.GetMillis()

	if err := a.Srv().Store.School().UpdateLastSchoolIconUpdate(school.Id, curTime); err != nil {
		return model.NewAppError("SetSchoolIcon", "api.school.school_icon.update.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	// manually set time to avoid possible cluster inconsistencies
	school.LastSchoolIconUpdate = curTime

	return nil
}

func (a *App) RemoveSchoolIcon(schoolId string) *model.AppError {
	school, err := a.GetSchool(schoolId)
	if err != nil {
		return model.NewAppError("RemoveSchoolIcon", "api.school.remove_school_icon.get_school.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	if err := a.Srv().Store.School().UpdateLastSchoolIconUpdate(schoolId, 0); err != nil {
		return model.NewAppError("RemoveSchoolIcon", "api.school.school_icon.update.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	school.LastSchoolIconUpdate = 0

	return nil
}

func (a *App) SaveBranch(branch *model.Branch) (*model.Branch, *model.AppError) {
	rbranch, err := a.Srv().Store.School().SaveBranch(branch)
	if err != nil {
		return nil, err
	}

	return rbranch, nil
}

func (a *App) GetBranches(schoolId string) ([]*model.Branch, *model.AppError) {
	return a.Srv().Store.School().GetBranches(schoolId)
}

func (a *App) GetBranch(branchId string) (*model.Branch, *model.AppError) {
	return a.Srv().Store.School().GetBranch(branchId)
}

func (a *App) RemoveBranch(branchId string) *model.AppError {
	err := a.Srv().Store.School().RemoveBranch(branchId)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) SaveClass(class *model.Class) (*model.Class, *model.AppError) {
	rclass, err := a.Srv().Store.School().SaveClass(class)
	if err != nil {
		return nil, err
	}

	return rclass, nil
}

func (a *App) GetClasses(schoolId string) ([]*model.Class, *model.AppError) {
	return a.Srv().Store.School().GetClasses(schoolId)
}

func (a *App) GetClass(classId string) (*model.Class, *model.AppError) {
	return a.Srv().Store.School().GetClass(classId)
}

func (a *App) GetClassesByBranch(branchId string) ([]*model.Class, *model.AppError) {
	return a.Srv().Store.School().GetClassesByBranch(branchId)
}

func (a *App) RemoveClass(classId string) *model.AppError {
	err := a.Srv().Store.School().RemoveClass(classId)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) GetSchoolsForUser(userId string) ([]*model.School, *model.AppError) {
	return a.Srv().Store.School().GetSchoolsByUserId(userId)
}

// Returns three values:
// 1. a pointer to the school member, if successful
// 2. a boolean: true if the user has a non-deleted school member for that school already, otherwise false.
// 3. a pointer to an AppError if something went wrong.
func (a *App) joinUserToSchool(school *model.School, user *model.User) (*model.SchoolMember, bool, *model.AppError) {
	sm := &model.SchoolMember{
		SchoolId:      school.Id,
		UserId:        user.Id,
		SchemeTeacher: user.IsTeacher(),
		SchemeParent:  user.IsParent(),
		SchemeAdmin:   user.IsSchoolAdmin(),
	}

	if school.Email == user.Email {
		sm.SchemeAdmin = true
	}

	rsm, err := a.Srv().Store.School().GetMember(school.Id, user.Id)
	if err != nil {
		// Membership appears to be missing. Lets try to add.
		var smr *model.SchoolMember
		smr, err = a.Srv().Store.School().SaveMember(sm, 10) // TODO *a.Config().SchoolSettings.MaxUsersPerSchool)
		if err != nil {
			return nil, false, err
		}
		return smr, false, nil
	}

	// Membership already exists.  Check if deleted and update, otherwise do nothing
	// Do nothing if already added
	if rsm.DeleteAt == 0 {
		return rsm, true, nil
	}

	membersCount, err := a.Srv().Store.School().GetActiveMemberCount(sm.SchoolId, nil)
	if err != nil {
		return nil, false, err
	}

	if membersCount >= 10 { // TODO int64(*a.Config().SchoolSettings.MaxUsersPerSchool) {
		return nil, false, model.NewAppError("joinUserToSchool", "app.school.join_user_to_school.max_accounts.app_error", nil, "schoolId="+sm.SchoolId, http.StatusBadRequest)
	}

	member, err := a.Srv().Store.School().UpdateMember(sm)
	if err != nil {
		return nil, false, err
	}

	return member, false, nil
}

func (a *App) JoinUserToSchool(school *model.School, user *model.User, userRequestorId string) *model.AppError {
	_, alreadyAdded, err := a.joinUserToSchool(school, user)
	if err != nil {
		return err
	}
	if alreadyAdded {
		return nil
	}

	if _, err := a.Srv().Store.User().UpdateUpdateAt(user.Id); err != nil {
		return err
	}

	a.ClearSessionCacheForUser(user.Id)
	a.InvalidateCacheForUser(user.Id)

	return nil
}
