// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	MAX_ADD_KID_MEMBERS_BATCH    = 20
	MAXIMUM_KID_BULK_IMPORT_SIZE = 10 * 1024 * 1024
	kidIDsParamPattern           = "[^a-zA-Z0-9,]*"
)

var kidIDsQueryParamRegex *regexp.Regexp

func init() {
	kidIDsQueryParamRegex = regexp.MustCompile(kidIDsParamPattern)
}

func (api *API) InitKid() {
	api.BaseRoutes.Kids.Handle("", api.ApiSessionRequired(createKid)).Methods("POST")
	api.BaseRoutes.Kids.Handle("", api.ApiSessionRequired(getMyKids)).Methods("GET")
	api.BaseRoutes.Kids.Handle("/mine", api.ApiSessionRequired(getMyKids)).Methods("GET")

	api.BaseRoutes.Kid.Handle("", api.ApiSessionRequired(getKid)).Methods("GET")
	api.BaseRoutes.Kid.Handle("", api.ApiSessionRequired(updateKid)).Methods("PUT")
	api.BaseRoutes.Kid.Handle("/patch", api.ApiSessionRequired(patchKid)).Methods("PUT")
}

func createKid(c *Context, w http.ResponseWriter, r *http.Request) {
	kid := model.KidFromJson(r.Body)
	if kid == nil {
		c.SetInvalidParam("kid")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_CREATE_KID) {
		c.Err = model.NewAppError("createKid", "api.kid.is_kid_creation_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rkid, err := c.App.CreateKid(kid)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the kid here since the user will be a kid admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rkid.ToJson()))
}

func getKid(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToKid(c.App.Session, c.Params.KidId, model.PERMISSION_MANAGE_KID) {
		c.SetPermissionError(model.PERMISSION_MANAGE_KID)
		return
	}

	kid, err := c.App.GetKid(c.Params.KidId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(kid.ToJson()))
}

func getMyKids(c *Context, w http.ResponseWriter, r *http.Request) {
	kids, err := c.App.GetKidsForUser(c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.KidListToJson(kids)))
}

func updateKid(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	if c.Err != nil {
		return
	}

	kid := model.KidFromJson(r.Body)

	if kid == nil {
		c.SetInvalidParam("kid")
		return
	}

	// The kid being updated in the payload must be the same one as indicated in the URL.
	if kid.Id != c.Params.KidId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToKid(c.App.Session, c.Params.KidId, model.PERMISSION_MANAGE_KID) {
		c.SetPermissionError(model.PERMISSION_MANAGE_KID)
		return
	}

	updatedKid, err := c.App.UpdateKid(kid)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(updatedKid.ToJson()))
}

func patchKid(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	if c.Err != nil {
		return
	}

	kid := model.KidPatchFromJson(r.Body)

	if kid == nil {
		c.SetInvalidParam("kid")
		return
	}

	if !c.App.SessionHasPermissionToKid(c.App.Session, c.Params.KidId, model.PERMISSION_MANAGE_KID) {
		c.SetPermissionError(model.PERMISSION_MANAGE_KID)
		return
	}

	patchedKid, err := c.App.PatchKid(c.Params.KidId, kid)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedKid.ToJson()))
}
