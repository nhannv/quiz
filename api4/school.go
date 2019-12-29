// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	MAX_ADD_SCHOOL_MEMBERS_BATCH    = 20
	MAXIMUM_SCHOOL_BULK_IMPORT_SIZE = 10 * 1024 * 1024
	schoolIDsParamPattern           = "[^a-zA-Z0-9,]*"
)

var schoolIDsQueryParamRegex *regexp.Regexp

func init() {
	schoolIDsQueryParamRegex = regexp.MustCompile(schoolIDsParamPattern)
}

func (api *API) InitSchool() {
	api.BaseRoutes.Schools.Handle("", api.ApiSessionRequired(createSchool)).Methods("POST")

	api.BaseRoutes.School.Handle("", api.ApiSessionRequired(getSchool)).Methods("GET")
	api.BaseRoutes.School.Handle("", api.ApiSessionRequired(updateSchool)).Methods("PUT")
	api.BaseRoutes.School.Handle("/patch", api.ApiSessionRequired(patchSchool)).Methods("PUT")
	// api.BaseRoutes.School.Handle("/regenerate_invite_id", api.ApiSessionRequired(regenerateSchoolInviteId)).Methods("POST")

	api.BaseRoutes.Branches.Handle("", api.ApiSessionRequired(getBranches)).Methods("GET")
	api.BaseRoutes.Branches.Handle("", api.ApiSessionRequired(addBranch)).Methods("POST")
	api.BaseRoutes.Branch.Handle("", api.ApiSessionRequired(removeBranch)).Methods("DELETE")
	api.BaseRoutes.Branch.Handle("", api.ApiSessionRequired(getBranch)).Methods("GET")

	api.BaseRoutes.Classes.Handle("", api.ApiSessionRequired(getClasses)).Methods("GET")
	api.BaseRoutes.Classes.Handle("", api.ApiSessionRequired(addClass)).Methods("POST")
	api.BaseRoutes.Class.Handle("", api.ApiSessionRequired(removeClass)).Methods("DELETE")
	api.BaseRoutes.Class.Handle("", api.ApiSessionRequired(getClass)).Methods("GET")
}

func createSchool(c *Context, w http.ResponseWriter, r *http.Request) {
	school := model.SchoolFromJson(r.Body)
	if school == nil {
		c.SetInvalidParam("school")
		return
	}

	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_CREATE_TEAM) {
		c.Err = model.NewAppError("createSchool", "api.school.is_school_creation_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rschool, err := c.App.CreateSchoolWithUser(school, c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the school here since the user will be a school admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rschool.ToJson()))
}

func getSchool(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	school, err := c.App.GetSchool(c.Params.SchoolId)
	if err != nil {
		c.Err = err
		return
	}

	// TODO  check school type
	// if (!school.AllowOpenInvite || school.Type != model.TEAM_OPEN) && !c.App.SessionHasPermissionToSchool(c.App.Session, school.Id, model.PERMISSION_VIEW_TEAM) {
	// 	c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
	// 	return
	// }

	c.App.SanitizeSchool(c.App.Session, school)
	w.Write([]byte(school.ToJson()))
}

func updateSchool(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	school := model.SchoolFromJson(r.Body)

	if school == nil {
		c.SetInvalidParam("school")
		return
	}

	// The school being updated in the payload must be the same one as indicated in the URL.
	if school.Id != c.Params.SchoolId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_TEAM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_TEAM)
		return
	}

	updatedSchool, err := c.App.UpdateSchool(school)
	if err != nil {
		c.Err = err
		return
	}

	c.App.SanitizeSchool(c.App.Session, updatedSchool)
	w.Write([]byte(updatedSchool.ToJson()))
}

func patchSchool(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	school := model.SchoolPatchFromJson(r.Body)

	if school == nil {
		c.SetInvalidParam("school")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_TEAM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_TEAM)
		return
	}

	patchedSchool, err := c.App.PatchSchool(c.Params.SchoolId, school)

	if err != nil {
		c.Err = err
		return
	}

	c.App.SanitizeSchool(c.App.Session, patchedSchool)

	c.LogAudit("")
	w.Write([]byte(patchedSchool.ToJson()))
}

func getBranches(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	branches, err := c.App.GetBranches(c.Params.SchoolId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.BranchesToJson(branches)))
}

func getBranch(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId().RequireBranchId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	branch, err := c.App.GetBranch(c.Params.BranchId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(branch.ToJson()))
}

func addBranch(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	var err *model.AppError
	branch := model.BranchFromJson(r.Body)
	branch.SchoolId = c.Params.SchoolId

	branch.CreatorId = c.App.Session.UserId
	if !c.App.SessionHasPermissionToSchool(c.App.Session, branch.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	result, err := c.App.SaveBranch(branch)

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result.ToJson()))
}

func removeBranch(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId().RequireBranchId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	if err := c.App.RemoveBranch(c.Params.BranchId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

// Classes api

func getClasses(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_VIEW_TEAM) {
		c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
		return
	}

	classes, err := c.App.GetClasses(c.Params.SchoolId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.ClassesToJson(classes)))
}

func getClass(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId().RequireClassId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	class, err := c.App.GetClass(c.Params.ClassId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(class.ToJson()))
}

func addClass(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId()
	if c.Err != nil {
		return
	}

	var err *model.AppError
	class := model.ClassFromJson(r.Body)
	class.SchoolId = c.Params.SchoolId

	class.CreatorId = c.App.Session.UserId
	if !c.App.SessionHasPermissionToSchool(c.App.Session, class.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	result, err := c.App.SaveClass(class)

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result.ToJson()))
}

func removeClass(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireSchoolId().RequireClassId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.Params.SchoolId, model.PERMISSION_MANAGE_SCHOOL) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SCHOOL)
		return
	}

	if err := c.App.RemoveClass(c.Params.ClassId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}
