// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *App) CreateEvent(event *model.Event) (*model.Event, *model.AppError) {
	class, err := a.GetClass(event.ClassId)
	if err != nil {
		return nil, err
	}

	if err = class.IsBelongToSchool(a.Session.SchoolId); err != nil {
		return nil, model.NewAppError("CreateEvent", "api.event.create_event.invalid_class.app_error", nil, "", http.StatusBadRequest)
	}

	revent, err := a.Srv.Store.Event().Save(event)
	if err != nil {
		return nil, err
	}

	return revent, nil
}

func (a *App) UpdateEvent(event *model.Event) (*model.Event, *model.AppError) {
	oldEvent, err := a.GetEvent(event.Id)
	if err != nil {
		return nil, err
	}

	oldEvent.Title = event.Title
	oldEvent.Description = event.Description
	oldEvent.Note = event.Note
	oldEvent.Picture = event.Picture
	oldEvent.Fee = event.Fee
	oldEvent.RegisterExpired = event.RegisterExpired
	oldEvent.IsAllClass = event.IsAllClass
	oldEvent.StartTime = event.StartTime
	oldEvent.EndTime = event.EndTime
	oldEvent.Active = event.Active

	oldEvent, err = a.updateEventUnsanitized(oldEvent)
	if err != nil {
		return event, err
	}

	return oldEvent, nil
}

func (a *App) updateEventUnsanitized(event *model.Event) (*model.Event, *model.AppError) {
	return a.Srv.Store.Event().Update(event)
}

func (a *App) PatchEvent(eventId string, patch *model.EventPatch) (*model.Event, *model.AppError) {
	event, err := a.GetEvent(eventId)
	if err != nil {
		return nil, err
	}

	event.Patch(patch)

	updatedEvent, err := a.UpdateEvent(event)
	if err != nil {
		return nil, err
	}

	return updatedEvent, nil
}

func (a *App) GetEvent(eventId string) (*model.Event, *model.AppError) {
	return a.Srv.Store.Event().Get(eventId)
}

func (a *App) GetEvents(classId string) ([]*model.Event, *model.AppError) {
	return a.Srv.Store.Event().GetByClass(classId)
}
