// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"

	"github.com/nhannv/quiz/v5/model"
	"github.com/pkg/errors"
)

const permissionsExportBatchSize = 100
const systemSchemeName = "00000000-0000-0000-0000-000000000000" // Prevents collisions with user-created schemes.

func (a *App) ResetPermissionsSystem() *model.AppError {

	// Reset all Custom Role assignments to Users.
	if err := a.Srv().Store.User().ClearAllCustomRoleAssignments(); err != nil {
		return err
	}

	// Purge all schemes from the database.
	if err := a.Srv().Store.Scheme().PermanentDeleteAll(); err != nil {
		return err
	}

	// Purge all roles from the database.
	if err := a.Srv().Store.Role().PermanentDeleteAll(); err != nil {
		return err
	}

	// Remove the "System" table entry that marks the advanced permissions migration as done.
	if _, err := a.Srv().Store.System().PermanentDeleteByName(ADVANCED_PERMISSIONS_MIGRATION_KEY); err != nil {
		return err
	}

	// Now that the permissions system has been reset, re-run the migration to reinitialise it.
	a.DoAppMigrations()

	return nil
}

func rollback(a *App, createdSchemeIDs []string) {
	for _, schemeID := range createdSchemeIDs {
		a.DeleteScheme(schemeID)
	}
}

func updateRole(a *App, sc *model.SchemeConveyor, roleCreatedName, defaultRoleName string) error {
	var err *model.AppError

	roleCreated, err := a.GetRoleByName(roleCreatedName)
	if err != nil {
		return errors.New(err.Message)
	}

	var roleIn *model.Role
	for _, role := range sc.Roles {
		if role.Name == defaultRoleName {
			roleIn = role
			break
		}
	}

	roleCreated.DisplayName = roleIn.DisplayName
	roleCreated.Description = roleIn.Description
	roleCreated.Permissions = roleIn.Permissions

	_, err = a.UpdateRole(roleCreated)
	if err != nil {
		return errors.New(fmt.Sprintf("%v: %v\n", err.Message, err.DetailedError))
	}

	return nil
}
