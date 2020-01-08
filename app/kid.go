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

func (a *App) CreateKid(kid *model.Kid) (*model.Kid, *model.AppError) {
	kid.InviteId = ""

	class, err := a.GetClass(kid.ClassId)
	if err != nil {
		return nil, err
	}

	if err = class.IsBelongToSchool(a.Session.SchoolId); err != nil {
		return nil, model.NewAppError("GetKidAvatar", "api.kid.create_kid.invalid_class.app_error", nil, "", http.StatusBadRequest)
	}

	rkid, err := a.Srv.Store.Kid().Save(kid)
	if err != nil {
		return nil, err
	}

	return rkid, nil
}

func (a *App) CreateKidWithUser(kid *model.Kid, userId string) (*model.Kid, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}

	rkid, err := a.CreateKid(kid)
	if err != nil {
		return nil, err
	}

	if err = a.JoinUserToKid(rkid, user, ""); err != nil {
		return nil, err
	}

	return rkid, nil
}

func (a *App) UpdateKid(kid *model.Kid) (*model.Kid, *model.AppError) {
	oldKid, err := a.GetKid(kid.Id)
	if err != nil {
		return nil, err
	}

	oldKid.FirstName = kid.FirstName
	oldKid.LastName = kid.LastName
	oldKid.NickName = kid.NickName
	oldKid.Avatar = kid.Avatar
	oldKid.Cover = kid.Cover
	oldKid.Description = kid.Description
	oldKid.Dob = kid.Dob
	oldKid.Gender = kid.Gender

	oldKid, err = a.updateKidUnsanitized(oldKid)
	if err != nil {
		return kid, err
	}

	return oldKid, nil
}

func (a *App) updateKidUnsanitized(kid *model.Kid) (*model.Kid, *model.AppError) {
	return a.Srv.Store.Kid().Update(kid)
}

func (a *App) PatchKid(kidId string, patch *model.KidPatch) (*model.Kid, *model.AppError) {
	kid, err := a.GetKid(kidId)
	if err != nil {
		return nil, err
	}

	kid.Patch(patch)

	updatedKid, err := a.UpdateKid(kid)
	if err != nil {
		return nil, err
	}

	return updatedKid, nil
}

func (a *App) GetKid(kidId string) (*model.Kid, *model.AppError) {
	return a.Srv.Store.Kid().Get(kidId)
}

func (a *App) GetKidsForUser(userId string) ([]*model.Kid, *model.AppError) {
	return a.Srv.Store.Kid().GetKidsByUserId(userId)
}

func (a *App) GetKidAvatar(kid *model.Kid) ([]byte, *model.AppError) {
	if len(*a.Config().FileSettings.DriverName) == 0 {
		return nil, model.NewAppError("GetKidAvatar", "api.kid.get_kid_icon.filesettings_no_driver.app_error", nil, "", http.StatusNotImplemented)
	}

	path := "kids/" + kid.Id + "/kidAvatar.png"
	data, err := a.ReadFile(path)
	if err != nil {
		return nil, model.NewAppError("GetKidAvatar", "api.kid.get_kid_icon.read_file.app_error", nil, err.Error(), http.StatusNotFound)
	}

	return data, nil
}

func (a *App) SetKidAvatar(kidId string, imageData *multipart.FileHeader) *model.AppError {
	file, err := imageData.Open()
	if err != nil {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.open.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	defer file.Close()
	return a.SetKidAvatarFromMultiPartFile(kidId, file)
}

func (a *App) SetKidAvatarFromMultiPartFile(kidId string, file multipart.File) *model.AppError {
	kid, getKidErr := a.GetKid(kidId)

	if getKidErr != nil {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.get_kid.app_error", nil, getKidErr.Error(), http.StatusBadRequest)
	}

	if len(*a.Config().FileSettings.DriverName) == 0 {
		return model.NewAppError("setKidAvatar", "api.kid.set_kid_icon.storage.app_error", nil, "", http.StatusNotImplemented)
	}

	// Decode image config first to check dimensions before loading the whole thing into memory later on
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.decode_config.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	if config.Width*config.Height > model.MaxImageSize {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.too_large.app_error", nil, "", http.StatusBadRequest)
	}

	file.Seek(0, 0)

	return a.SetKidAvatarFromFile(kid, file)
}

func (a *App) SetKidAvatarFromFile(kid *model.Kid, file io.Reader) *model.AppError {
	// Decode image into Image object
	img, _, err := image.Decode(file)
	if err != nil {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.decode.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	orientation, _ := getImageOrientation(file)
	img = makeImageUpright(img, orientation)

	// Scale kid icon
	kidAvatarWidthAndHeight := 128
	img = imaging.Fill(img, kidAvatarWidthAndHeight, kidAvatarWidthAndHeight, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.encode.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path := "kids/" + kid.Id + "/kidAvatar.png"

	if _, err := a.WriteFile(buf, path); err != nil {
		return model.NewAppError("SetKidAvatar", "api.kid.set_kid_icon.write_file.app_error", nil, "", http.StatusInternalServerError)
	}

	return nil
}

// Returns three values:
// 1. a pointer to the kid guardian, if successful
// 2. a boolean: true if the user has a non-deleted kid guardian for that kid already, otherwise false.
// 3. a pointer to an AppError if something went wrong.
func (a *App) joinUserToKid(kid *model.Kid, user *model.User) (*model.KidGuardian, bool, *model.AppError) {
	sm := &model.KidGuardian{
		KidId:  kid.Id,
		UserId: user.Id,
	}

	rsm, err := a.Srv.Store.Kid().GetGuardian(kid.Id, user.Id)
	if err != nil {
		// Guardianship appears to be missing. Lets try to add.
		var smr *model.KidGuardian
		smr, err = a.Srv.Store.Kid().SaveGuardian(sm)
		if err != nil {
			return nil, false, err
		}
		return smr, false, nil
	}

	// Guardianship already exists.  Check if deleted and update, otherwise do nothing
	// Do nothing if already added
	if rsm.DeleteAt == 0 {
		return rsm, true, nil
	}

	guardiansCount, err := a.Srv.Store.Kid().GetActiveGuardianCount(sm.KidId)
	if err != nil {
		return nil, false, err
	}

	if guardiansCount >= 3 {
		return nil, false, model.NewAppError("joinUserToKid", "app.kid.join_user_to_kid.max_accounts.app_error", nil, "kidId="+sm.KidId, http.StatusBadRequest)
	}

	guardian, err := a.Srv.Store.Kid().UpdateGuardian(sm)
	if err != nil {
		return nil, false, err
	}

	return guardian, false, nil
}

func (a *App) JoinUserToKid(kid *model.Kid, user *model.User, userRequestorId string) *model.AppError {
	_, alreadyAdded, err := a.joinUserToKid(kid, user)
	if err != nil {
		return err
	}
	if alreadyAdded {
		return nil
	}

	if _, err := a.Srv.Store.User().UpdateUpdateAt(user.Id); err != nil {
		return err
	}

	a.ClearSessionCacheForUser(user.Id)
	a.InvalidateCacheForUser(user.Id)

	return nil
}
