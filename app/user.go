// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nhannv/quiz/v5/einterfaces"
	"github.com/nhannv/quiz/v5/mlog"
	"github.com/nhannv/quiz/v5/model"
	"github.com/nhannv/quiz/v5/services/mfa"
	"github.com/nhannv/quiz/v5/store"
	"github.com/nhannv/quiz/v5/utils"
	"github.com/nhannv/quiz/v5/utils/fileutils"
)

const (
	TOKEN_TYPE_PASSWORD_RECOVERY  = "password_recovery"
	TOKEN_TYPE_VERIFY_EMAIL       = "verify_email"
	TOKEN_TYPE_GUEST_INVITATION   = "guest_invitation"
	PASSWORD_RECOVER_EXPIRY_TIME  = 1000 * 60 * 60      // 1 hour
	INVITATION_EXPIRY_TIME        = 1000 * 60 * 60 * 48 // 48 hours
	IMAGE_PROFILE_PIXEL_DIMENSION = 128
)

func (a *App) CreateUserWithInviteId(user *model.User, inviteId string) (*model.User, *model.AppError) {

	user.EmailVerified = false

	ruser, err := a.CreateUser(user)
	if err != nil {
		return nil, err
	}

	if err := a.SendWelcomeEmail(ruser.Id, ruser.Email, ruser.EmailVerified, ruser.Locale, a.GetSiteURL()); err != nil {
		mlog.Error("Failed to send welcome email on create user with inviteId", mlog.Err(err))
	}

	return ruser, nil
}

func (a *App) CreateUserAsAdmin(user *model.User) (*model.User, *model.AppError) {
	ruser, err := a.CreateUser(user)
	if err != nil {
		return nil, err
	}

	if err := a.SendWelcomeEmail(ruser.Id, ruser.Email, ruser.EmailVerified, ruser.Locale, a.GetSiteURL()); err != nil {
		mlog.Error("Failed to send welcome email on create admin user", mlog.Err(err))
	}

	return ruser, nil
}

func (a *App) CreateUserFromSignup(user *model.User) (*model.User, *model.AppError) {

	user.EmailVerified = false

	ruser, err := a.CreateUser(user)
	if err != nil {
		return nil, err
	}

	if err := a.SendWelcomeEmail(ruser.Id, ruser.Email, ruser.EmailVerified, ruser.Locale, a.GetSiteURL()); err != nil {
		mlog.Error("Failed to send welcome email on create user from signup", mlog.Err(err))
	}

	return ruser, nil
}

func (a *App) IsFirstUserAccount() bool {
	if a.SessionCacheLength() == 0 {
		count, err := a.Srv().Store.User().Count(model.UserCountOptions{IncludeDeleted: true})
		if err != nil {
			mlog.Error("There was a error fetching if first user account", mlog.Err(err))
			return false
		}
		if count <= 0 {
			return true
		}
	}

	return false
}

// CreateUser creates a user and sets several fields of the returned User struct to
// their zero values.
func (a *App) CreateUser(user *model.User) (*model.User, *model.AppError) {
	return a.createUserOrGuest(user, false)
}

// CreateGuest creates a guest and sets several fields of the returned User struct to
// their zero values.
func (a *App) CreateGuest(user *model.User) (*model.User, *model.AppError) {
	return a.createUserOrGuest(user, true)
}

func (a *App) createUserOrGuest(user *model.User, guest bool) (*model.User, *model.AppError) {
	user.Roles = model.SYSTEM_USER_ROLE_ID
	if guest {
		user.Roles = model.SYSTEM_GUEST_ROLE_ID
	}

	if !user.IsLDAPUser() && !user.IsSAMLUser() && !user.IsGuest() && !CheckUserDomain(user, *a.Config().TeamSettings.RestrictCreationToDomains) {
		return nil, model.NewAppError("CreateUser", "api.user.create_user.accepted_domain.app_error", nil, "", http.StatusBadRequest)
	}

	if !user.IsLDAPUser() && !user.IsSAMLUser() && user.IsGuest() && !CheckUserDomain(user, *a.Config().GuestAccountsSettings.RestrictCreationToDomains) {
		return nil, model.NewAppError("CreateUser", "api.user.create_user.accepted_domain.app_error", nil, "", http.StatusBadRequest)
	}

	// Below is a special case where the first user in the entire
	// system is granted the system_admin role
	count, err := a.Srv().Store.User().Count(model.UserCountOptions{IncludeDeleted: true})
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		user.Roles = model.SYSTEM_ADMIN_ROLE_ID + " " + model.SYSTEM_USER_ROLE_ID
	}

	if _, ok := utils.GetSupportedLocales()[user.Locale]; !ok {
		user.Locale = *a.Config().LocalizationSettings.DefaultClientLocale
	}

	ruser, err := a.createUser(user)
	if err != nil {
		return nil, err
	}
	// This message goes to everyone, so the teamId, channelId and userId are irrelevant
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_NEW_USER, "", "", "", nil)
	message.Add("user_id", ruser.Id)
	a.Publish(message)

	return ruser, nil
}

func (a *App) createUser(user *model.User) (*model.User, *model.AppError) {
	user.MakeNonNil()

	if err := a.IsPasswordValid(user.Password); user.AuthService == "" && err != nil {
		return nil, err
	}

	ruser, err := a.Srv().Store.User().Save(user)
	if err != nil {
		mlog.Error("Couldn't save the user", mlog.Err(err))
		return nil, err
	}

	if user.EmailVerified {
		if err := a.VerifyUserEmail(ruser.Id, user.Email); err != nil {
			mlog.Error("Failed to set email verified", mlog.Err(err))
		}
	}

	pref := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_TUTORIAL_STEPS, Name: ruser.Id, Value: "0"}
	if err := a.Srv().Store.Preference().Save(&model.Preferences{pref}); err != nil {
		mlog.Error("Encountered error saving tutorial preference", mlog.Err(err))
	}

	ruser.Sanitize(map[string]bool{})
	return ruser, nil
}

// CheckEmailDomain checks that an email domain matches a list of space-delimited domains as a string.
func CheckEmailDomain(email string, domains string) bool {
	if len(domains) == 0 {
		return true
	}

	domainArray := strings.Fields(strings.TrimSpace(strings.ToLower(strings.Replace(strings.Replace(domains, "@", " ", -1), ",", " ", -1))))

	for _, d := range domainArray {
		if strings.HasSuffix(strings.ToLower(email), "@"+d) {
			return true
		}
	}

	return false
}

// CheckUserDomain checks that a user's email domain matches a list of space-delimited domains as a string.
func CheckUserDomain(user *model.User, domains string) bool {
	return CheckEmailDomain(user.Email, domains)
}

// IsUsernameTaken checks if the username is already used by another user. Return false if the username is invalid.
func (a *App) IsUsernameTaken(name string) bool {
	if !model.IsValidUsername(name) {
		return false
	}

	if _, err := a.Srv().Store.User().GetByUsername(name); err != nil {
		return false
	}

	return true
}

func (a *App) GetUser(userId string) (*model.User, *model.AppError) {
	return a.Srv().Store.User().Get(userId)
}

func (a *App) GetUserByUsername(username string) (*model.User, *model.AppError) {
	result, err := a.Srv().Store.User().GetByUsername(username)
	if err != nil && err.Id == "store.sql_user.get_by_username.app_error" {
		err.StatusCode = http.StatusNotFound
		return nil, err
	}
	return result, nil
}

func (a *App) GetUserByEmail(email string) (*model.User, *model.AppError) {
	user, err := a.Srv().Store.User().GetByEmail(email)
	if err != nil {
		if err.Id == "store.sql_user.missing_account.const" {
			err.StatusCode = http.StatusNotFound
			return nil, err
		}
		err.StatusCode = http.StatusBadRequest
		return nil, err
	}
	return user, nil
}

func (a *App) GetUserByAuth(authData *string, authService string) (*model.User, *model.AppError) {
	return a.Srv().Store.User().GetByAuth(authData, authService)
}

func (a *App) GetUsers(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetAllProfiles(options)
}

func (a *App) GetUsersPage(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsers(options)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersEtag(restrictionsHash string) string {
	return fmt.Sprintf("%v.%v.%v.%v", a.Srv().Store.User().GetEtagForAllProfiles(), a.Config().PrivacySettings.ShowFullName, a.Config().PrivacySettings.ShowEmailAddress, restrictionsHash)
}

func (a *App) GetUsersByIds(userIds []string, options *store.UserGetByIdsOpts) ([]*model.User, *model.AppError) {
	allowFromCache := options.ViewRestrictions == nil

	users, err := a.Srv().Store.User().GetProfileByIds(userIds, options, allowFromCache)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, options.IsAdmin), nil
}

func (a *App) GetUsersByUsernames(usernames []string, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.Srv().Store.User().GetProfilesByUsernames(usernames)
	if err != nil {
		return nil, err
	}
	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) sanitizeProfiles(users []*model.User, asAdmin bool) []*model.User {
	for _, u := range users {
		a.SanitizeProfile(u, asAdmin)
	}

	return users
}

func (a *App) GenerateMfaSecret(userId string) (*model.MfaSecret, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}

	mfaService := mfa.New(a, a.Srv().Store)
	secret, img, err := mfaService.GenerateSecret(user)
	if err != nil {
		return nil, err
	}

	mfaSecret := &model.MfaSecret{Secret: secret, QRCode: b64.StdEncoding.EncodeToString(img)}
	return mfaSecret, nil
}

func (a *App) ActivateMfa(userId, token string) *model.AppError {
	user, err := a.Srv().Store.User().Get(userId)
	if err != nil {
		return err
	}

	if len(user.AuthService) > 0 && user.AuthService != model.USER_AUTH_SERVICE_LDAP {
		return model.NewAppError("ActivateMfa", "api.user.activate_mfa.email_and_ldap_only.app_error", nil, "", http.StatusBadRequest)
	}

	mfaService := mfa.New(a, a.Srv().Store)
	if err := mfaService.Activate(user, token); err != nil {
		return err
	}

	return nil
}

func (a *App) DeactivateMfa(userId string) *model.AppError {
	mfaService := mfa.New(a, a.Srv().Store)
	if err := mfaService.Deactivate(userId); err != nil {
		return err
	}

	return nil
}

func CreateProfileImage(username string, userId string, initialFont string) ([]byte, *model.AppError) {
	colors := []color.NRGBA{
		{197, 8, 126, 255},
		{227, 207, 18, 255},
		{28, 181, 105, 255},
		{35, 188, 224, 255},
		{116, 49, 196, 255},
		{197, 8, 126, 255},
		{197, 19, 19, 255},
		{250, 134, 6, 255},
		{227, 207, 18, 255},
		{123, 201, 71, 255},
		{28, 181, 105, 255},
		{35, 188, 224, 255},
		{116, 49, 196, 255},
		{197, 8, 126, 255},
		{197, 19, 19, 255},
		{250, 134, 6, 255},
		{227, 207, 18, 255},
		{123, 201, 71, 255},
		{28, 181, 105, 255},
		{35, 188, 224, 255},
		{116, 49, 196, 255},
		{197, 8, 126, 255},
		{197, 19, 19, 255},
		{250, 134, 6, 255},
		{227, 207, 18, 255},
		{123, 201, 71, 255},
	}

	h := fnv.New32a()
	h.Write([]byte(userId))
	seed := h.Sum32()

	initial := string(strings.ToUpper(username)[0])

	font, err := getFont(initialFont)
	if err != nil {
		return nil, model.NewAppError("CreateProfileImage", "api.user.create_profile_image.default_font.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	color := colors[int64(seed)%int64(len(colors))]
	dstImg := image.NewRGBA(image.Rect(0, 0, IMAGE_PROFILE_PIXEL_DIMENSION, IMAGE_PROFILE_PIXEL_DIMENSION))
	srcImg := image.White
	draw.Draw(dstImg, dstImg.Bounds(), &image.Uniform{color}, image.ZP, draw.Src)
	size := float64(IMAGE_PROFILE_PIXEL_DIMENSION / 2)

	c := freetype.NewContext()
	c.SetFont(font)
	c.SetFontSize(size)
	c.SetClip(dstImg.Bounds())
	c.SetDst(dstImg)
	c.SetSrc(srcImg)

	pt := freetype.Pt(IMAGE_PROFILE_PIXEL_DIMENSION/5, IMAGE_PROFILE_PIXEL_DIMENSION*2/3)
	_, err = c.DrawString(initial, pt)
	if err != nil {
		return nil, model.NewAppError("CreateProfileImage", "api.user.create_profile_image.initial.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	buf := new(bytes.Buffer)

	if imgErr := png.Encode(buf, dstImg); imgErr != nil {
		return nil, model.NewAppError("CreateProfileImage", "api.user.create_profile_image.encode.app_error", nil, imgErr.Error(), http.StatusInternalServerError)
	}
	return buf.Bytes(), nil
}

func getFont(initialFont string) (*truetype.Font, error) {
	// Some people have the old default font still set, so just treat that as if they're using the new default
	if initialFont == "luximbi.ttf" {
		initialFont = "nunito-bold.ttf"
	}

	fontDir, _ := fileutils.FindDir("fonts")
	fontBytes, err := ioutil.ReadFile(filepath.Join(fontDir, initialFont))
	if err != nil {
		return nil, err
	}

	return freetype.ParseFont(fontBytes)
}

func (a *App) GetProfileImage(user *model.User) ([]byte, bool, *model.AppError) {
	if len(*a.Config().FileSettings.DriverName) == 0 {
		img, appErr := a.GetDefaultProfileImage(user)
		if appErr != nil {
			return nil, false, appErr
		}
		return img, false, nil
	}

	path := "users/" + user.Id + "/profile.png"

	data, err := a.ReadFile(path)
	if err != nil {
		img, appErr := a.GetDefaultProfileImage(user)
		if appErr != nil {
			return nil, false, appErr
		}

		if user.LastPictureUpdate == 0 {
			if _, err := a.WriteFile(bytes.NewReader(img), path); err != nil {
				return nil, false, err
			}
		}
		return img, true, nil
	}

	return data, false, nil
}

func (a *App) GetDefaultProfileImage(user *model.User) ([]byte, *model.AppError) {
	var img []byte
	var appErr *model.AppError

	img, appErr = CreateProfileImage(user.Username, user.Id, *a.Config().FileSettings.InitialFont)
	if appErr != nil {
		return nil, appErr
	}
	return img, nil
}

func (a *App) SetDefaultProfileImage(user *model.User) *model.AppError {
	img, appErr := a.GetDefaultProfileImage(user)
	if appErr != nil {
		return appErr
	}

	path := "users/" + user.Id + "/profile.png"

	if _, err := a.WriteFile(bytes.NewReader(img), path); err != nil {
		return err
	}

	if err := a.Srv().Store.User().ResetLastPictureUpdate(user.Id); err != nil {
		mlog.Error("Failed to reset last picture update", mlog.Err(err))
	}

	a.InvalidateCacheForUser(user.Id)

	updatedUser, appErr := a.GetUser(user.Id)
	if appErr != nil {
		mlog.Error("Error in getting users profile forcing logout", mlog.String("user_id", user.Id), mlog.Err(appErr))
		return nil
	}

	options := a.Config().GetSanitizeOptions()
	updatedUser.SanitizeProfile(options)

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_UPDATED, "", "", "", nil)
	message.Add("user", updatedUser)
	a.Publish(message)

	return nil
}

func (a *App) SetProfileImage(userId string, imageData *multipart.FileHeader) *model.AppError {
	file, err := imageData.Open()
	if err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.open.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	defer file.Close()
	return a.SetProfileImageFromMultiPartFile(userId, file)
}

func (a *App) SetProfileImageFromMultiPartFile(userId string, file multipart.File) *model.AppError {
	// Decode image config first to check dimensions before loading the whole thing into memory later on
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.decode_config.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	if config.Width*config.Height > model.MaxImageSize {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.too_large.app_error", nil, "", http.StatusBadRequest)
	}

	file.Seek(0, 0)

	return a.SetProfileImageFromFile(userId, file)
}

func (a *App) SetProfileImageFromFile(userId string, file io.Reader) *model.AppError {

	// Decode image into Image object
	img, _, err := image.Decode(file)
	if err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.decode.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	orientation, _ := getImageOrientation(file)
	img = makeImageUpright(img, orientation)

	// Scale profile image
	profileWidthAndHeight := 128
	img = imaging.Fill(img, profileWidthAndHeight, profileWidthAndHeight, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.encode.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path := "users/" + userId + "/profile.png"

	if _, err := a.WriteFile(buf, path); err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.upload_profile.app_error", nil, "", http.StatusInternalServerError)
	}

	if err := a.Srv().Store.User().UpdateLastPictureUpdate(userId); err != nil {
		mlog.Error("Error with updating last picture update", mlog.Err(err))
	}
	a.invalidateUserCacheAndPublish(userId)

	return nil
}

func (a *App) UpdatePasswordAsUser(userId, currentPassword, newPassword string) *model.AppError {
	user, err := a.GetUser(userId)
	if err != nil {
		return err
	}

	if user == nil {
		err = model.NewAppError("updatePassword", "api.user.update_password.valid_account.app_error", nil, "", http.StatusBadRequest)
		return err
	}

	if user.AuthData != nil && *user.AuthData != "" {
		err = model.NewAppError("updatePassword", "api.user.update_password.oauth.app_error", nil, "auth_service="+user.AuthService, http.StatusBadRequest)
		return err
	}

	if err := a.DoubleCheckPassword(user, currentPassword); err != nil {
		if err.Id == "api.user.check_user_password.invalid.app_error" {
			err = model.NewAppError("updatePassword", "api.user.update_password.incorrect.app_error", nil, "", http.StatusBadRequest)
		}
		return err
	}

	T := utils.GetUserTranslations(user.Locale)

	return a.UpdatePasswordSendEmail(user, newPassword, T("api.user.update_password.menu"))
}

func (a *App) userDeactivated(userId string) *model.AppError {
	if err := a.RevokeAllSessions(userId); err != nil {
		return err
	}

	a.SetStatusOffline(userId, false)

	return nil
}

func (a *App) UpdateActive(user *model.User, active bool) (*model.User, *model.AppError) {
	user.UpdateAt = model.GetMillis()
	if active {
		user.DeleteAt = 0
	} else {
		user.DeleteAt = user.UpdateAt
	}

	userUpdate, err := a.Srv().Store.User().Update(user, true)
	if err != nil {
		return nil, err
	}
	ruser := userUpdate.New

	if !active {
		if err := a.userDeactivated(ruser.Id); err != nil {
			return nil, err
		}
	}

	a.InvalidateCacheForUser(user.Id)

	a.sendUpdatedUserEvent(*ruser)

	return ruser, nil
}

func (a *App) DeactivateGuests() *model.AppError {
	userIds, err := a.Srv().Store.User().DeactivateGuests()
	if err != nil {
		return err
	}

	for _, userId := range userIds {
		if err := a.userDeactivated(userId); err != nil {
			return err
		}
	}

	a.Srv().Store.User().ClearCaches()

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_GUESTS_DEACTIVATED, "", "", "", nil)
	a.Publish(message)

	return nil
}

func (a *App) GetSanitizeOptions(asAdmin bool) map[string]bool {
	options := a.Config().GetSanitizeOptions()
	if asAdmin {
		options["email"] = true
		options["fullname"] = true
		options["authservice"] = true
	}
	return options
}

func (a *App) SanitizeProfile(user *model.User, asAdmin bool) {
	options := a.GetSanitizeOptions(asAdmin)

	user.SanitizeProfile(options)
}

func (a *App) UpdateUserAsUser(user *model.User, asAdmin bool) (*model.User, *model.AppError) {
	updatedUser, err := a.UpdateUser(user, true)
	if err != nil {
		return nil, err
	}

	a.sendUpdatedUserEvent(*updatedUser)

	return updatedUser, nil
}

func (a *App) PatchUser(userId string, patch *model.UserPatch, asAdmin bool) (*model.User, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}

	user.Patch(patch)

	updatedUser, err := a.UpdateUser(user, true)
	if err != nil {
		return nil, err
	}

	a.sendUpdatedUserEvent(*updatedUser)

	return updatedUser, nil
}

func (a *App) UpdateUserAuth(userId string, userAuth *model.UserAuth) (*model.UserAuth, *model.AppError) {
	if userAuth.AuthData == nil || *userAuth.AuthData == "" || userAuth.AuthService == "" {
		userAuth.AuthData = nil
		userAuth.AuthService = ""

		if err := a.IsPasswordValid(userAuth.Password); err != nil {
			return nil, err
		}
		password := model.HashPassword(userAuth.Password)

		if err := a.Srv().Store.User().UpdatePassword(userId, password); err != nil {
			return nil, err
		}
	} else {
		userAuth.Password = ""

		if _, err := a.Srv().Store.User().UpdateAuthData(userId, userAuth.AuthService, userAuth.AuthData, "", false); err != nil {
			return nil, err
		}
	}

	return userAuth, nil
}

func (a *App) sendUpdatedUserEvent(user model.User) {
	adminCopyOfUser := user.DeepCopy()
	a.SanitizeProfile(adminCopyOfUser, true)
	adminMessage := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_UPDATED, "", "", "", nil)
	adminMessage.Add("user", adminCopyOfUser)
	adminMessage.GetBroadcast().ContainsSensitiveData = true
	a.Publish(adminMessage)

	a.SanitizeProfile(&user, false)
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_UPDATED, "", "", "", nil)
	message.Add("user", &user)
	message.GetBroadcast().ContainsSanitizedData = true
	a.Publish(message)
}

func (a *App) UpdateUser(user *model.User, sendNotifications bool) (*model.User, *model.AppError) {
	prev, err := a.Srv().Store.User().Get(user.Id)
	if err != nil {
		return nil, err
	}

	if !CheckUserDomain(user, *a.Config().GuestAccountsSettings.RestrictCreationToDomains) {
		if prev.IsGuest() && !prev.IsLDAPUser() && !prev.IsSAMLUser() && user.Email != prev.Email {
			return nil, model.NewAppError("UpdateUser", "api.user.update_user.accepted_guest_domain.app_error", nil, "", http.StatusBadRequest)
		}
	}

	// Don't set new eMail on user account if email verification is required, this will be done as a post-verification action
	// to avoid users being able to set non-controlled eMails as their account email
	newEmail := ""
	if *a.Config().EmailSettings.RequireEmailVerification && prev.Email != user.Email {
		newEmail = user.Email

		_, err = a.GetUserByEmail(newEmail)
		if err == nil {
			return nil, model.NewAppError("UpdateUser", "store.sql_user.update.email_taken.app_error", nil, "user_id="+user.Id, http.StatusBadRequest)
		}

	}

	userUpdate, err := a.Srv().Store.User().Update(user, false)
	if err != nil {
		return nil, err
	}

	if sendNotifications {
		if userUpdate.New.Email != userUpdate.Old.Email || newEmail != "" {
			if *a.Config().EmailSettings.RequireEmailVerification {
				a.Srv().Go(func() {
					if err := a.SendEmailVerification(userUpdate.New, newEmail); err != nil {
						mlog.Error("Failed to send email verification", mlog.Err(err))
					}
				})
			} else {
				a.Srv().Go(func() {
					if err := a.SendEmailChangeEmail(userUpdate.Old.Email, userUpdate.New.Email, userUpdate.New.Locale, a.GetSiteURL()); err != nil {
						mlog.Error("Failed to send email change email", mlog.Err(err))
					}
				})
			}
		}

		if userUpdate.New.Username != userUpdate.Old.Username {
			a.Srv().Go(func() {
				if err := a.SendChangeUsernameEmail(userUpdate.Old.Username, userUpdate.New.Username, userUpdate.New.Email, userUpdate.New.Locale, a.GetSiteURL()); err != nil {
					mlog.Error("Failed to send change username email", mlog.Err(err))
				}
			})
		}
	}

	a.InvalidateCacheForUser(user.Id)

	return userUpdate.New, nil
}

func (a *App) UpdateUserActive(userId string, active bool) *model.AppError {
	user, err := a.GetUser(userId)

	if err != nil {
		return err
	}
	if _, err = a.UpdateActive(user, active); err != nil {
		return err
	}

	return nil
}

func (a *App) UpdateUserNotifyProps(userId string, props map[string]string) (*model.User, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}

	user.NotifyProps = props

	ruser, err := a.UpdateUser(user, true)
	if err != nil {
		return nil, err
	}

	return ruser, nil
}

func (a *App) UpdateMfa(activate bool, userId, token string) *model.AppError {
	if activate {
		if err := a.ActivateMfa(userId, token); err != nil {
			return err
		}
	} else {
		if err := a.DeactivateMfa(userId); err != nil {
			return err
		}
	}

	a.Srv().Go(func() {
		user, err := a.GetUser(userId)
		if err != nil {
			mlog.Error("Failed to get user", mlog.Err(err))
			return
		}

		if err := a.SendMfaChangeEmail(user.Email, activate, user.Locale, a.GetSiteURL()); err != nil {
			mlog.Error("Failed to send mfa change email", mlog.Err(err))
		}
	})

	return nil
}

func (a *App) UpdatePasswordByUserIdSendEmail(userId, newPassword, method string) *model.AppError {
	user, err := a.GetUser(userId)
	if err != nil {
		return err
	}

	return a.UpdatePasswordSendEmail(user, newPassword, method)
}

func (a *App) UpdatePassword(user *model.User, newPassword string) *model.AppError {
	if err := a.IsPasswordValid(newPassword); err != nil {
		return err
	}

	hashedPassword := model.HashPassword(newPassword)

	if err := a.Srv().Store.User().UpdatePassword(user.Id, hashedPassword); err != nil {
		return model.NewAppError("UpdatePassword", "api.user.update_password.failed.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdatePasswordSendEmail(user *model.User, newPassword, method string) *model.AppError {
	if err := a.UpdatePassword(user, newPassword); err != nil {
		return err
	}

	a.Srv().Go(func() {
		if err := a.SendPasswordChangeEmail(user.Email, method, user.Locale, a.GetSiteURL()); err != nil {
			mlog.Error("Failed to send password change email", mlog.Err(err))
		}
	})

	return nil
}

func (a *App) ResetPasswordFromToken(userSuppliedTokenString, newPassword string) *model.AppError {
	token, err := a.GetPasswordRecoveryToken(userSuppliedTokenString)
	if err != nil {
		return err
	}
	if model.GetMillis()-token.CreateAt >= PASSWORD_RECOVER_EXPIRY_TIME {
		return model.NewAppError("resetPassword", "api.user.reset_password.link_expired.app_error", nil, "", http.StatusBadRequest)
	}

	tokenData := struct {
		UserId string
		Email  string
	}{}

	err2 := json.Unmarshal([]byte(token.Extra), &tokenData)
	if err2 != nil {
		return model.NewAppError("resetPassword", "api.user.reset_password.token_parse.error", nil, "", http.StatusInternalServerError)
	}

	user, err := a.GetUser(tokenData.UserId)
	if err != nil {
		return err
	}

	if user.Email != tokenData.Email {
		return model.NewAppError("resetPassword", "api.user.reset_password.link_expired.app_error", nil, "", http.StatusBadRequest)
	}

	if user.IsSSOUser() {
		return model.NewAppError("ResetPasswordFromCode", "api.user.reset_password.sso.app_error", nil, "userId="+user.Id, http.StatusBadRequest)
	}

	T := utils.GetUserTranslations(user.Locale)

	if err := a.UpdatePasswordSendEmail(user, newPassword, T("api.user.reset_password.method")); err != nil {
		return err
	}

	if err := a.DeleteToken(token); err != nil {
		mlog.Error("Failed to delete token", mlog.Err(err))
	}

	return nil
}

func (a *App) SendPasswordReset(email string, siteURL string) (bool, *model.AppError) {
	user, err := a.GetUserByEmail(email)
	if err != nil {
		return false, nil
	}

	if user.AuthData != nil && len(*user.AuthData) != 0 {
		return false, model.NewAppError("SendPasswordReset", "api.user.send_password_reset.sso.app_error", nil, "userId="+user.Id, http.StatusBadRequest)
	}

	token, err := a.CreatePasswordRecoveryToken(user.Id, user.Email)
	if err != nil {
		return false, err
	}

	return a.SendPasswordResetEmail(user.Email, token, user.Locale, siteURL)
}

func (a *App) CreatePasswordRecoveryToken(userId, email string) (*model.Token, *model.AppError) {

	tokenExtra := struct {
		UserId string
		Email  string
	}{
		userId,
		email,
	}
	jsonData, err := json.Marshal(tokenExtra)

	if err != nil {
		return nil, model.NewAppError("CreatePasswordRecoveryToken", "api.user.create_password_token.error", nil, "", http.StatusInternalServerError)
	}

	token := model.NewToken(TOKEN_TYPE_PASSWORD_RECOVERY, string(jsonData))

	if err := a.Srv().Store.Token().Save(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (a *App) GetPasswordRecoveryToken(token string) (*model.Token, *model.AppError) {
	rtoken, err := a.Srv().Store.Token().GetByToken(token)
	if err != nil {
		return nil, model.NewAppError("GetPasswordRecoveryToken", "api.user.reset_password.invalid_link.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	if rtoken.Type != TOKEN_TYPE_PASSWORD_RECOVERY {
		return nil, model.NewAppError("GetPasswordRecoveryToken", "api.user.reset_password.broken_token.app_error", nil, "", http.StatusBadRequest)
	}
	return rtoken, nil
}

func (a *App) DeleteToken(token *model.Token) *model.AppError {
	return a.Srv().Store.Token().Delete(token.Token)
}

func (a *App) UpdateUserRoles(userId string, newRoles string, sendWebSocketEvent bool) (*model.User, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		err.StatusCode = http.StatusBadRequest
		return nil, err
	}

	if err := a.CheckRolesExist(strings.Fields(newRoles)); err != nil {
		return nil, err
	}

	user.Roles = newRoles
	uchan := make(chan store.StoreResult, 1)
	go func() {
		userUpdate, err := a.Srv().Store.User().Update(user, true)
		uchan <- store.StoreResult{Data: userUpdate, Err: err}
		close(uchan)
	}()

	schan := make(chan store.StoreResult, 1)
	go func() {
		userId, err := a.Srv().Store.Session().UpdateRoles(user.Id, newRoles)
		schan <- store.StoreResult{Data: userId, Err: err}
		close(schan)
	}()

	result := <-uchan
	if result.Err != nil {
		return nil, result.Err
	}
	ruser := result.Data.(*model.UserUpdate).New

	if result := <-schan; result.Err != nil {
		// soft error since the user roles were still updated
		mlog.Error("Failed during updating user roles", mlog.Err(result.Err))
	}

	a.InvalidateCacheForUser(user.Id)
	a.ClearSessionCacheForUser(user.Id)

	if sendWebSocketEvent {
		message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_ROLE_UPDATED, "", "", user.Id, nil)
		message.Add("user_id", user.Id)
		message.Add("roles", newRoles)
		a.Publish(message)
	}

	return ruser, nil
}

func (a *App) PermanentDeleteUser(user *model.User) *model.AppError {
	mlog.Warn("Attempting to permanently delete account", mlog.String("user_id", user.Id), mlog.String("user_email", user.Email))
	if user.IsInRole(model.SYSTEM_ADMIN_ROLE_ID) {
		mlog.Warn("You are deleting a user that is a system administrator.  You may need to set another account as the system administrator using the command line tools.", mlog.String("user_email", user.Email))
	}

	if _, err := a.UpdateActive(user, false); err != nil {
		return err
	}

	if err := a.Srv().Store.Session().PermanentDeleteSessionsByUser(user.Id); err != nil {
		return err
	}

	if err := a.Srv().Store.UserAccessToken().DeleteAllForUser(user.Id); err != nil {
		return err
	}

	if err := a.Srv().Store.OAuth().PermanentDeleteAuthDataByUser(user.Id); err != nil {
		return err
	}

	if err := a.Srv().Store.Preference().PermanentDeleteByUser(user.Id); err != nil {
		return err
	}

	infos, err := a.Srv().Store.FileInfo().GetForUser(user.Id)
	if err != nil {
		mlog.Warn("Error getting file list for user from FileInfoStore", mlog.Err(err))
	}

	for _, info := range infos {
		res, err := a.FileExists(info.Path)
		if err != nil {
			mlog.Warn(
				"Error checking existence of file",
				mlog.String("path", info.Path),
				mlog.Err(err),
			)
			continue
		}

		if !res {
			mlog.Warn("File not found", mlog.String("path", info.Path))
			continue
		}

		err = a.RemoveFile(info.Path)

		if err != nil {
			mlog.Warn(
				"Unable to remove file",
				mlog.String("path", info.Path),
				mlog.Err(err),
			)
		}
	}

	if _, err := a.Srv().Store.FileInfo().PermanentDeleteByUser(user.Id); err != nil {
		return err
	}

	if err := a.Srv().Store.Audit().PermanentDeleteByUser(user.Id); err != nil {
		return err
	}

	if err := a.Srv().Store.User().PermanentDelete(user.Id); err != nil {
		return err
	}

	mlog.Warn("Permanently deleted account", mlog.String("user_email", user.Email), mlog.String("user_id", user.Id))

	if a.IsESIndexingEnabled() {
		a.Srv().Go(func() {
			if err := a.Elasticsearch().DeleteUser(user); err != nil {
				mlog.Error("Encountered error deleting user", mlog.String("user_id", user.Id), mlog.Err(err))
			}
		})
	}

	return nil
}

func (a *App) PermanentDeleteAllUsers() *model.AppError {
	users, err := a.Srv().Store.User().GetAll()
	if err != nil {
		return err
	}
	for _, user := range users {
		a.PermanentDeleteUser(user)
	}

	return nil
}

func (a *App) SendEmailVerification(user *model.User, newEmail string) *model.AppError {
	token, err := a.CreateVerifyEmailToken(user.Id, newEmail)
	if err != nil {
		return err
	}

	if _, err := a.GetStatus(user.Id); err != nil {
		return a.SendVerifyEmail(newEmail, user.Locale, a.GetSiteURL(), token.Token)
	}
	return a.SendEmailChangeVerifyEmail(newEmail, user.Locale, a.GetSiteURL(), token.Token)
}

func (a *App) VerifyEmailFromToken(userSuppliedTokenString string) *model.AppError {
	token, err := a.GetVerifyEmailToken(userSuppliedTokenString)
	if err != nil {
		return err
	}
	if model.GetMillis()-token.CreateAt >= PASSWORD_RECOVER_EXPIRY_TIME {
		return model.NewAppError("VerifyEmailFromToken", "api.user.verify_email.link_expired.app_error", nil, "", http.StatusBadRequest)
	}

	tokenData := struct {
		UserId string
		Email  string
	}{}

	err2 := json.Unmarshal([]byte(token.Extra), &tokenData)
	if err2 != nil {
		return model.NewAppError("VerifyEmailFromToken", "api.user.verify_email.token_parse.error", nil, "", http.StatusInternalServerError)
	}

	user, err := a.GetUser(tokenData.UserId)
	if err != nil {
		return err
	}

	tokenData.Email = strings.ToLower(tokenData.Email)
	if err := a.VerifyUserEmail(tokenData.UserId, tokenData.Email); err != nil {
		return err
	}

	if user.Email != tokenData.Email {
		a.Srv().Go(func() {
			if err := a.SendEmailChangeEmail(user.Email, tokenData.Email, user.Locale, a.GetSiteURL()); err != nil {
				mlog.Error("Failed to send email change email", mlog.Err(err))
			}
		})
	}

	if err := a.DeleteToken(token); err != nil {
		mlog.Error("Failed to delete token", mlog.Err(err))
	}

	return nil
}

func (a *App) CreateVerifyEmailToken(userId string, newEmail string) (*model.Token, *model.AppError) {
	tokenExtra := struct {
		UserId string
		Email  string
	}{
		userId,
		newEmail,
	}
	jsonData, err := json.Marshal(tokenExtra)

	if err != nil {
		return nil, model.NewAppError("CreateVerifyEmailToken", "api.user.create_email_token.error", nil, "", http.StatusInternalServerError)
	}

	token := model.NewToken(TOKEN_TYPE_VERIFY_EMAIL, string(jsonData))

	if err := a.Srv().Store.Token().Save(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (a *App) GetVerifyEmailToken(token string) (*model.Token, *model.AppError) {
	rtoken, err := a.Srv().Store.Token().GetByToken(token)
	if err != nil {
		return nil, model.NewAppError("GetVerifyEmailToken", "api.user.verify_email.bad_link.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	if rtoken.Type != TOKEN_TYPE_VERIFY_EMAIL {
		return nil, model.NewAppError("GetVerifyEmailToken", "api.user.verify_email.broken_token.app_error", nil, "", http.StatusBadRequest)
	}
	return rtoken, nil
}

// GetTotalUsersStats is used for the DM list total
func (a *App) GetTotalUsersStats(viewRestrictions *model.ViewUsersRestrictions) (*model.UsersStats, *model.AppError) {
	count, err := a.Srv().Store.User().Count(model.UserCountOptions{
		ViewRestrictions: viewRestrictions,
	})
	if err != nil {
		return nil, err
	}
	stats := &model.UsersStats{
		TotalUsersCount: count,
	}
	return stats, nil
}

func (a *App) VerifyUserEmail(userId, email string) *model.AppError {
	_, err := a.Srv().Store.User().VerifyEmail(userId, email)
	if err != nil {
		return err
	}

	user, err := a.GetUser(userId)

	if err != nil {
		return err
	}

	a.sendUpdatedUserEvent(*user)

	return nil
}

func (a *App) UpdateOAuthUserAttrs(userData io.Reader, user *model.User, provider einterfaces.OauthProvider, service string) *model.AppError {
	oauthUser := provider.GetUserFromJson(userData)
	if oauthUser == nil {
		return model.NewAppError("UpdateOAuthUserAttrs", "api.user.update_oauth_user_attrs.get_user.app_error", map[string]interface{}{"Service": service}, "", http.StatusBadRequest)
	}

	userAttrsChanged := false

	if oauthUser.Username != user.Username {
		if existingUser, _ := a.GetUserByUsername(oauthUser.Username); existingUser == nil {
			user.Username = oauthUser.Username
			userAttrsChanged = true
		}
	}

	if oauthUser.GetFullName() != user.GetFullName() {
		user.FirstName = oauthUser.FirstName
		user.LastName = oauthUser.LastName
		userAttrsChanged = true
	}

	if oauthUser.Email != user.Email {
		if existingUser, _ := a.GetUserByEmail(oauthUser.Email); existingUser == nil {
			user.Email = oauthUser.Email
			userAttrsChanged = true
		}
	}

	if user.DeleteAt > 0 {
		// Make sure they are not disabled
		user.DeleteAt = 0
		userAttrsChanged = true
	}

	if userAttrsChanged {
		users, err := a.Srv().Store.User().Update(user, true)
		if err != nil {
			return err
		}

		user = users.New
		a.InvalidateCacheForUser(user.Id)

	}

	return nil
}

// invalidateUserCacheAndPublish Invalidates cache for a user and publishes user updated event
func (a *App) invalidateUserCacheAndPublish(userId string) {
	a.InvalidateCacheForUser(userId)

	user, userErr := a.GetUser(userId)
	if userErr != nil {
		mlog.Error("Error in getting users profile", mlog.String("user_id", userId), mlog.Err(userErr))
		return
	}

	options := a.Config().GetSanitizeOptions()
	user.SanitizeProfile(options)

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_UPDATED, "", "", "", nil)
	message.Add("user", user)
	a.Publish(message)
}
