// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	scheduleIDsParamPattern = "[^a-zA-Z0-9,]*"
)

var scheduleIDsQueryParamRegex *regexp.Regexp

func init() {
	scheduleIDsQueryParamRegex = regexp.MustCompile(scheduleIDsParamPattern)
}

func (api *API) InitSchedule() {
	api.BaseRoutes.Schedules.Handle("", api.ApiSessionRequired(createSchedule)).Methods("POST")
	api.BaseRoutes.Schedules.Handle("", api.ApiSessionRequired(getSchedules)).Methods("GET")
	api.BaseRoutes.Schedule.Handle("", api.ApiSessionRequired(getSchedule)).Methods("GET")
	api.BaseRoutes.Schedule.Handle("", api.ApiSessionRequired(updateSchedule)).Methods("PUT")
	api.BaseRoutes.Schedule.Handle("/patch", api.ApiSessionRequired(patchSchedule)).Methods("PUT")
}

func createSchedule(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClassId()
	schedule := model.ScheduleFromJson(r.Body)

	if schedule == nil {
		c.SetInvalidParam("schedule")
		return
	}
	schedule.ClassId = c.Params.ClassId

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.Err = model.NewAppError("createSchedule", "api.schedule.is_class_manage_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rschedule, err := c.App.CreateSchedule(schedule)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the schedule here since the user will be a schedule admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rschedule.ToJson()))
}

func getSchedule(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireScheduleId()
	if c.Err != nil {
		return
	}

	schedule, err := c.App.GetSchedule(c.Params.ScheduleId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(schedule.ToJson()))
}

func getSchedules(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClassId()
	if c.Err != nil {
		return
	}

	schedules, err := c.App.GetSchedules(c.Params.Week, c.Params.Year, c.Params.ClassId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.ScheduleListToJson(schedules)))
}

func updateSchedule(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireScheduleId()
	if c.Err != nil {
		return
	}

	schedule := model.ScheduleFromJson(r.Body)

	if schedule == nil {
		c.SetInvalidParam("schedule")
		return
	}

	// The schedule being updated in the payload must be the same one as indicated in the URL.
	if schedule.Id != c.Params.ScheduleId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	updatedSchedule, err := c.App.UpdateSchedule(schedule)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(updatedSchedule.ToJson()))
}

func patchSchedule(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireScheduleId()
	if c.Err != nil {
		return
	}

	schedule := model.SchedulePatchFromJson(r.Body)

	if schedule == nil {
		c.SetInvalidParam("schedule")
		return
	}

	if !c.App.SessionHasPermissionToSchool(c.App.Session, c.App.Session.SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	patchedSchedule, err := c.App.PatchSchedule(c.Params.ScheduleId, schedule)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedSchedule.ToJson()))
}
