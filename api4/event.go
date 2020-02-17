// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	eventIDsParamPattern = "[^a-zA-Z0-9,]*"
)

var eventIDsQueryParamRegex *regexp.Regexp

func init() {
	eventIDsQueryParamRegex = regexp.MustCompile(eventIDsParamPattern)
}

func (api *API) InitEvent() {
	api.BaseRoutes.Events.Handle("", api.ApiSessionRequired(createEvent)).Methods("POST")
	api.BaseRoutes.Events.Handle("", api.ApiSessionRequired(getEvents)).Methods("GET")
	//api.BaseRoutes.Event.Handle("/register", api.ApiSessionRequired(registerEvent)).Methods("POST")
	api.BaseRoutes.Event.Handle("/{kid_id:[A-Za-z0-9]+}/paid", api.ApiSessionRequired(updatePaid)).Methods("PUT")
	api.BaseRoutes.Event.Handle("", api.ApiSessionRequired(getEvent)).Methods("GET")
	api.BaseRoutes.Event.Handle("", api.ApiSessionRequired(updateEvent)).Methods("PUT")
	api.BaseRoutes.Event.Handle("/patch", api.ApiSessionRequired(patchEvent)).Methods("PUT")
}

func createEvent(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClassId()
	event := model.EventFromJson(r.Body)

	if event == nil {
		c.SetInvalidParam("event")
		return
	}
	event.ClassId = c.Params.ClassId

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.Err = model.NewAppError("createEvent", "api.event.is_class_manage_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	revent, err := c.App.CreateEvent(event)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the event here since the user will be a event admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(revent.ToJson()))
}

func getEvent(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireEventId()
	if c.Err != nil {
		return
	}

	event, err := c.App.GetEvent(c.Params.EventId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(event.ToJson()))
}

func getEvents(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClassId()
	if c.Err != nil {
		return
	}

	events, err := c.App.GetEvents(c.Params.ClassId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.EventListToJson(events)))
}

func updateEvent(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireEventId()
	if c.Err != nil {
		return
	}

	event := model.EventFromJson(r.Body)

	if event == nil {
		c.SetInvalidParam("event")
		return
	}

	// The event being updated in the payload must be the same one as indicated in the URL.
	if event.Id != c.Params.EventId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	updatedEvent, err := c.App.UpdateEvent(event)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(updatedEvent.ToJson()))
}

func patchEvent(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireEventId()
	if c.Err != nil {
		return
	}

	event := model.EventPatchFromJson(r.Body)

	if event == nil {
		c.SetInvalidParam("event")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	patchedEvent, err := c.App.PatchEvent(c.Params.EventId, event)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedEvent.ToJson()))
}

func updatePaid(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireEventId().RequireKidId()
	if c.Err != nil {
		return
	}

	event := model.EventPatchFromJson(r.Body)

	if event == nil {
		c.SetInvalidParam("event")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	patchedEvent, err := c.App.PatchEvent(c.Params.EventId, event)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedEvent.ToJson()))
}
