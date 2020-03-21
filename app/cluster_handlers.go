// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"strings"

	"github.com/nhannv/quiz/v5/model"
)

// RegisterAllClusterMessageHandlers registers the cluster message handlers that are handled by the App layer.
//
// The cluster event handlers are spread across this function and
// NewLocalCacheLayer. Be careful to not have duplicated handlers here and
// there.
func (a *App) registerAllClusterMessageHandlers() {
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_PUBLISH, a.clusterPublishHandler)
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_UPDATE_STATUS, a.clusterUpdateStatusHandler)
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_INVALIDATE_ALL_CACHES, a.clusterInvalidateAllCachesHandler)
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_USER, a.clusterInvalidateCacheForUserHandler)
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_CLEAR_SESSION_CACHE_FOR_USER, a.clusterClearSessionCacheForUserHandler)
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_CLEAR_SESSION_CACHE_FOR_ALL_USERS, a.clusterClearSessionCacheForAllUsersHandler)
	a.Cluster().RegisterClusterMessageHandler(model.CLUSTER_EVENT_BUSY_STATE_CHANGED, a.clusterBusyStateChgHandler)
}

func (a *App) clusterPublishHandler(msg *model.ClusterMessage) {
	event := model.WebSocketEventFromJson(strings.NewReader(msg.Data))
	a.PublishSkipClusterSend(event)
}

func (a *App) clusterUpdateStatusHandler(msg *model.ClusterMessage) {
	status := model.StatusFromJson(strings.NewReader(msg.Data))
	a.AddStatusCacheSkipClusterSend(status)
}

func (a *App) clusterInvalidateAllCachesHandler(msg *model.ClusterMessage) {
	a.InvalidateAllCachesSkipSend()
}

func (a *App) clusterInvalidateCacheForUserHandler(msg *model.ClusterMessage) {
	a.InvalidateCacheForUserSkipClusterSend(msg.Data)
}

func (a *App) clusterClearSessionCacheForUserHandler(msg *model.ClusterMessage) {
	a.ClearSessionCacheForUserSkipClusterSend(msg.Data)
}

func (a *App) clusterClearSessionCacheForAllUsersHandler(msg *model.ClusterMessage) {
	a.ClearSessionCacheForAllUsersSkipClusterSend()
}

func (a *App) clusterBusyStateChgHandler(msg *model.ClusterMessage) {
	a.ServerBusyStateChanged(model.ServerBusyStateFromJson(strings.NewReader(msg.Data)))
}
