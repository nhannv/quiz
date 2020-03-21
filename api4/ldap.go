// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"

	"github.com/nhannv/quiz/v5/model"
)

type mixedUnlinkedGroup struct {
	Id           *string `json:"mattermost_group_id"`
	DisplayName  string  `json:"name"`
	RemoteId     string  `json:"primary_key"`
	HasSyncables *bool   `json:"has_syncables"`
}

func (api *API) InitLdap() {
	api.BaseRoutes.LDAP.Handle("/sync", api.ApiSessionRequired(syncLdap)).Methods("POST")
	api.BaseRoutes.LDAP.Handle("/test", api.ApiSessionRequired(testLdap)).Methods("POST")

}

func syncLdap(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.License() == nil || !*c.App.License().Features.LDAP {
		c.Err = model.NewAppError("Api4.syncLdap", "api.ldap_groups.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	c.App.SyncLdap()

	ReturnStatusOK(w)
}

func testLdap(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.License() == nil || !*c.App.License().Features.LDAP {
		c.Err = model.NewAppError("Api4.testLdap", "api.ldap_groups.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	if err := c.App.TestLdap(); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}
