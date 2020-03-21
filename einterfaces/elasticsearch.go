// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import (
	"time"

	"github.com/nhannv/quiz/v5/model"
)

type ElasticsearchInterface interface {
	Start() *model.AppError
	Stop() *model.AppError
	GetVersion() int
	IndexUser(user *model.User, teamsIds, channelsIds []string) *model.AppError
	DeleteUser(user *model.User) *model.AppError
	TestConfig(cfg *model.Config) *model.AppError
	PurgeIndexes() *model.AppError
	DataRetentionDeleteIndexes(cutoff time.Time) *model.AppError
}
