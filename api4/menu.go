// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	menuIDsParamPattern = "[^a-zA-Z0-9,]*"
)

var menuIDsQueryParamRegex *regexp.Regexp

func init() {
	menuIDsQueryParamRegex = regexp.MustCompile(menuIDsParamPattern)
}

func (api *API) InitMenu() {
	api.BaseRoutes.Menus.Handle("", api.ApiSessionRequired(createMenu)).Methods("POST")
	api.BaseRoutes.Menus.Handle("", api.ApiSessionRequired(getMenus)).Methods("GET")
	api.BaseRoutes.Menu.Handle("", api.ApiSessionRequired(getMenu)).Methods("GET")
	api.BaseRoutes.Menu.Handle("/note/{kid_id:[A-Za-z0-9]+}", api.ApiSessionRequired(createNote)).Methods("POST")
	api.BaseRoutes.Menu.Handle("", api.ApiSessionRequired(updateMenu)).Methods("PUT")
	api.BaseRoutes.Menu.Handle("/patch", api.ApiSessionRequired(patchMenu)).Methods("PUT")
}

func createMenu(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClassId()
	menu := model.MenuFromJson(r.Body)

	if menu == nil {
		c.SetInvalidParam("menu")
		return
	}
	menu.ClassId = c.Params.ClassId

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.Err = model.NewAppError("createMenu", "api.menu.is_class_manage_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rmenu, err := c.App.CreateMenu(menu)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the menu here since the user will be a menu admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rmenu.ToJson()))
}

func createNote(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMenuId().RequireKidId()
	activityNote := model.ActivityNoteFromJson(r.Body)

	if activityNote == nil {
		c.SetInvalidParam("activityNote")
		return
	}
	activityNote.KidId = c.Params.KidId
	activityNote.Type = model.ACTIVITY_TYPE_MENU
	activityNote.ActivityId = c.Params.MenuId
	activityNote.UserId = c.App.Session.UserId

	if !c.App.SessionHasPermissionTeacherToKid(c.App.Session, c.Params.KidId, model.PERMISSION_MANAGE_ACTIVITY) {
		c.Err = model.NewAppError("createActivityNote", "api.menu.is_kid_manage_activity_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	ractivityNote, err := c.App.CreateActivityNote(activityNote)
	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ractivityNote.ToJson()))
}

func getMenu(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMenuId()
	if c.Err != nil {
		return
	}

	menu, err := c.App.GetMenu(c.Params.MenuId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(menu.ToJson()))
}

func getMenus(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.Err != nil {
		return
	}

	menus, err := c.App.GetMenus(c.Params.Week, c.Params.Year)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.MenuListToJson(menus)))
}

func updateMenu(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMenuId()
	if c.Err != nil {
		return
	}

	menu := model.MenuFromJson(r.Body)

	if menu == nil {
		c.SetInvalidParam("menu")
		return
	}

	// The menu being updated in the payload must be the same one as indicated in the URL.
	if menu.Id != c.Params.MenuId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	updatedMenu, err := c.App.UpdateMenu(menu)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(updatedMenu.ToJson()))
}

func patchMenu(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMenuId()
	if c.Err != nil {
		return
	}

	menu := model.MenuPatchFromJson(r.Body)

	if menu == nil {
		c.SetInvalidParam("menu")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	patchedMenu, err := c.App.PatchMenu(c.Params.MenuId, menu)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedMenu.ToJson()))
}
