// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	healthIDsParamPattern = "[^a-zA-Z0-9,]*"
)

var healthIDsQueryParamRegex *regexp.Regexp

func init() {
	healthIDsQueryParamRegex = regexp.MustCompile(healthIDsParamPattern)
}

func (api *API) InitHealth() {
	api.BaseRoutes.Healths.Handle("", api.ApiSessionRequired(createHealth)).Methods("POST")
	api.BaseRoutes.Healths.Handle("", api.ApiSessionRequired(getHealths)).Methods("GET")
	api.BaseRoutes.Health.Handle("", api.ApiSessionRequired(getHealth)).Methods("GET")
	api.BaseRoutes.Health.Handle("", api.ApiSessionRequired(updateHealth)).Methods("PUT")
	api.BaseRoutes.Health.Handle("/patch", api.ApiSessionRequired(patchHealth)).Methods("PUT")
}

func createHealth(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	health := model.HealthFromJson(r.Body)

	if health == nil {
		c.SetInvalidParam("health")
		return
	}
	health.KidId = c.Params.KidId

	if !c.App.SessionHasPermissionToKid(*c.App.Session(), c.Params.KidId, model.PERMISSION_MANAGE_KID) {
		c.Err = model.NewAppError("createHealth", "api.health.is_kid_manage_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rhealth, err := c.App.CreateHealth(health)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the health here since the user will be a health admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rhealth.ToJson()))
}

func getHealth(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireHealthId()
	if c.Err != nil {
		return
	}

	health, err := c.App.GetHealth(c.Params.HealthId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(health.ToJson()))
}

func getHealths(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	if c.Err != nil {
		return
	}

	healths, err := c.App.GetHealths(c.Params.KidId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.HealthListToJson(healths)))
}

func updateHealth(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireHealthId()
	if c.Err != nil {
		return
	}

	health := model.HealthFromJson(r.Body)

	if health == nil {
		c.SetInvalidParam("health")
		return
	}

	// The health being updated in the payload must be the same one as indicated in the URL.
	if health.Id != c.Params.HealthId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	updatedHealth, err := c.App.UpdateHealth(health)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(updatedHealth.ToJson()))
}

func patchHealth(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireHealthId()
	if c.Err != nil {
		return
	}

	health := model.HealthPatchFromJson(r.Body)

	if health == nil {
		c.SetInvalidParam("health")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	patchedHealth, err := c.App.PatchHealth(c.Params.HealthId, health)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedHealth.ToJson()))
}
