// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *App) CreateSchedule(schedule *model.Schedule) (*model.Schedule, *model.AppError) {
	class, err := a.GetClass(schedule.ClassId)
	if err != nil {
		return nil, err
	}

	if err = class.IsBelongToSchool(a.Session.SchoolId); err != nil {
		return nil, model.NewAppError("CreateSchedule", "api.schedule.create_schedule.invalid_class.app_error", nil, "", http.StatusBadRequest)
	}

	rschedule, err := a.Srv.Store.Schedule().Save(schedule)
	if err != nil {
		return nil, err
	}

	return rschedule, nil
}

func (a *App) UpdateSchedule(schedule *model.Schedule) (*model.Schedule, *model.AppError) {
	oldSchedule, err := a.GetSchedule(schedule.Id)
	if err != nil {
		return nil, err
	}

	oldSchedule.Week = schedule.Week
	oldSchedule.Year = schedule.Year
	oldSchedule.Subject = schedule.Subject
	oldSchedule.Description = schedule.Description
	oldSchedule.WeekDay = schedule.WeekDay
	oldSchedule.StartTime = schedule.StartTime
	oldSchedule.EndTime = schedule.EndTime
	oldSchedule.Active = schedule.Active

	oldSchedule, err = a.updateScheduleUnsanitized(oldSchedule)
	if err != nil {
		return schedule, err
	}

	return oldSchedule, nil
}

func (a *App) updateScheduleUnsanitized(schedule *model.Schedule) (*model.Schedule, *model.AppError) {
	return a.Srv.Store.Schedule().Update(schedule)
}

func (a *App) PatchSchedule(scheduleId string, patch *model.SchedulePatch) (*model.Schedule, *model.AppError) {
	schedule, err := a.GetSchedule(scheduleId)
	if err != nil {
		return nil, err
	}

	schedule.Patch(patch)

	updatedSchedule, err := a.UpdateSchedule(schedule)
	if err != nil {
		return nil, err
	}

	return updatedSchedule, nil
}

func (a *App) GetSchedule(scheduleId string) (*model.Schedule, *model.AppError) {
	return a.Srv.Store.Schedule().Get(scheduleId)
}

func (a *App) GetSchedules(week int, year int, classId string) ([]*model.Schedule, *model.AppError) {
	return a.Srv.Store.Schedule().GetByWeek(week, year, classId)
}
