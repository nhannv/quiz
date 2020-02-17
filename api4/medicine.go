// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"regexp"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	medicineIDsParamPattern = "[^a-zA-Z0-9,]*"
)

var medicineIDsQueryParamRegex *regexp.Regexp

func init() {
	medicineIDsQueryParamRegex = regexp.MustCompile(medicineIDsParamPattern)
}

func (api *API) InitMedicine() {
	api.BaseRoutes.KidMedicines.Handle("", api.ApiSessionRequired(createMedicineRequest)).Methods("POST")
	api.BaseRoutes.KidMedicines.Handle("", api.ApiSessionRequired(getMedicinesForKid)).Methods("GET")
	//api.BaseRoutes.Medicine.Handle("", api.ApiSessionRequired(updateMedicine)).Methods("PUT")
	//api.BaseRoutes.Medicine.Handle("/patch", api.ApiSessionRequired(patchMedicine)).Methods("PUT")
}

func createMedicineRequest(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	medicineRequest := model.MedicineRequestFromJson(r.Body)

	if medicineRequest == nil {
		c.SetInvalidParam("medicineRequest")
		return
	}
	medicineRequest.KidId = c.Params.KidId

	if !c.App.SessionHasPermissionToKid(*c.App.Session(), c.Params.KidId, model.PERMISSION_MANAGE_KID) {
		c.Err = model.NewAppError("createMedicine", "api.medicine.is_kid_manage_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rmedicine, err := c.App.CreateMedicineRequest(medicineRequest)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the medicine here since the user will be a medicine admin and their session won't reflect that yet

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rmedicine.ToJson()))
}

func getMedicinesForKid(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireKidId()
	if c.Err != nil {
		return
	}

	medicines, err := c.App.GetMedicineRequestsByKid(c.Params.KidId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.MedicineRequestListToJson(medicines)))
}

/*
func updateMedicine(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMedicineId()
	if c.Err != nil {
		return
	}

	medicine := model.MedicineFromJson(r.Body)

	if medicine == nil {
		c.SetInvalidParam("medicine")
		return
	}

	// The medicine being updated in the payload must be the same one as indicated in the URL.
	if medicine.Id != c.Params.MedicineId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	updatedMedicine, err := c.App.UpdateMedicine(medicine)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(updatedMedicine.ToJson()))
}

func patchMedicine(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMedicineId()
	if c.Err != nil {
		return
	}

	medicine := model.MedicinePatchFromJson(r.Body)

	if medicine == nil {
		c.SetInvalidParam("medicine")
		return
	}

	if !c.App.SessionHasPermissionToSchool(*c.App.Session(), c.App.Session().SchoolId, model.PERMISSION_MANAGE_CLASS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_CLASS)
		return
	}

	patchedMedicine, err := c.App.PatchMedicine(c.Params.MedicineId, medicine)

	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(patchedMedicine.ToJson()))
}
*/
