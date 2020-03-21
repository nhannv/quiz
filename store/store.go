//go:generate go run layer_generators/main.go

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package store

import (
	"github.com/nhannv/quiz/v5/model"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type Store interface {
	Audit() AuditStore
	User() UserStore
	ClusterDiscovery() ClusterDiscoveryStore
	Session() SessionStore
	OAuth() OAuthStore
	System() SystemStore
	Preference() PreferenceStore
	License() LicenseStore
	Token() TokenStore
	Emoji() EmojiStore
	Status() StatusStore
	FileInfo() FileInfoStore
	Reaction() ReactionStore
	Role() RoleStore
	Scheme() SchemeStore
	Job() JobStore
	UserAccessToken() UserAccessTokenStore
	LinkMetadata() LinkMetadataStore
	MarkSystemRanUnitTests()
	Close()
	LockToMaster()
	UnlockFromMaster()
	DropAllTables()
	GetCurrentSchemaVersion() string
	TotalMasterDbConnections() int
	TotalReadDbConnections() int
	TotalSearchDbConnections() int
	CheckIntegrity() <-chan IntegrityCheckResult
}

type AuditStore interface {
	Save(audit *model.Audit) *model.AppError
	Get(user_id string, offset int, limit int) (model.Audits, *model.AppError)
	PermanentDeleteByUser(userId string) *model.AppError
}

type UserStore interface {
	Save(user *model.User) (*model.User, *model.AppError)
	Update(user *model.User, allowRoleUpdate bool) (*model.UserUpdate, *model.AppError)
	UpdateLastPictureUpdate(userId string) *model.AppError
	ResetLastPictureUpdate(userId string) *model.AppError
	UpdatePassword(userId, newPassword string) *model.AppError
	UpdateUpdateAt(userId string) (int64, *model.AppError)
	UpdateAuthData(userId string, service string, authData *string, email string, resetMfa bool) (string, *model.AppError)
	UpdateMfaSecret(userId, secret string) *model.AppError
	UpdateMfaActive(userId string, active bool) *model.AppError
	Get(id string) (*model.User, *model.AppError)
	GetAll() ([]*model.User, *model.AppError)
	ClearCaches()
	GetProfilesByUsernames(usernames []string) ([]*model.User, *model.AppError)
	GetAllProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfileByIds(userIds []string, options *UserGetByIdsOpts, allowFromCache bool) ([]*model.User, *model.AppError)
	InvalidateProfileCacheForUser(userId string)
	GetByEmail(email string) (*model.User, *model.AppError)
	GetByAuth(authData *string, authService string) (*model.User, *model.AppError)
	GetAllUsingAuthService(authService string) ([]*model.User, *model.AppError)
	GetByUsername(username string) (*model.User, *model.AppError)
	GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) (*model.User, *model.AppError)
	VerifyEmail(userId, email string) (string, *model.AppError)
	GetEtagForAllProfiles() string
	UpdateFailedPasswordAttempts(userId string, attempts int) *model.AppError
	GetSystemAdminProfiles() (map[string]*model.User, *model.AppError)
	PermanentDelete(userId string) *model.AppError
	AnalyticsActiveCount(time int64, options model.UserCountOptions) (int64, *model.AppError)
	AnalyticsGetInactiveUsersCount() (int64, *model.AppError)
	AnalyticsGetSystemAdminCount() (int64, *model.AppError)
	ClearAllCustomRoleAssignments() *model.AppError
	InferSystemInstallDate() (int64, *model.AppError)
	GetAllAfter(limit int, afterId string) ([]*model.User, *model.AppError)
	Count(options model.UserCountOptions) (int64, *model.AppError)
	DeactivateGuests() ([]string, *model.AppError)
}

type SessionStore interface {
	Get(sessionIdOrToken string) (*model.Session, *model.AppError)
	Save(session *model.Session) (*model.Session, *model.AppError)
	GetSessions(userId string) ([]*model.Session, *model.AppError)
	GetSessionsWithActiveDeviceIds(userId string) ([]*model.Session, *model.AppError)
	Remove(sessionIdOrToken string) *model.AppError
	RemoveAllSessions() *model.AppError
	PermanentDeleteSessionsByUser(teamId string) *model.AppError
	UpdateLastActivityAt(sessionId string, time int64) *model.AppError
	UpdateRoles(userId string, roles string) (string, *model.AppError)
	UpdateDeviceId(id string, deviceId string, expiresAt int64) (string, *model.AppError)
	UpdateProps(session *model.Session) *model.AppError
	AnalyticsSessionCount() (int64, *model.AppError)
	Cleanup(expiryTime int64, batchSize int64)
}

type ClusterDiscoveryStore interface {
	Save(discovery *model.ClusterDiscovery) *model.AppError
	Delete(discovery *model.ClusterDiscovery) (bool, *model.AppError)
	Exists(discovery *model.ClusterDiscovery) (bool, *model.AppError)
	GetAll(discoveryType, clusterName string) ([]*model.ClusterDiscovery, *model.AppError)
	SetLastPingAt(discovery *model.ClusterDiscovery) *model.AppError
	Cleanup() *model.AppError
}

type OAuthStore interface {
	SaveApp(app *model.OAuthApp) (*model.OAuthApp, *model.AppError)
	UpdateApp(app *model.OAuthApp) (*model.OAuthApp, *model.AppError)
	GetApp(id string) (*model.OAuthApp, *model.AppError)
	GetAppByUser(userId string, offset, limit int) ([]*model.OAuthApp, *model.AppError)
	GetApps(offset, limit int) ([]*model.OAuthApp, *model.AppError)
	GetAuthorizedApps(userId string, offset, limit int) ([]*model.OAuthApp, *model.AppError)
	DeleteApp(id string) *model.AppError
	SaveAuthData(authData *model.AuthData) (*model.AuthData, *model.AppError)
	GetAuthData(code string) (*model.AuthData, *model.AppError)
	RemoveAuthData(code string) *model.AppError
	PermanentDeleteAuthDataByUser(userId string) *model.AppError
	SaveAccessData(accessData *model.AccessData) (*model.AccessData, *model.AppError)
	UpdateAccessData(accessData *model.AccessData) (*model.AccessData, *model.AppError)
	GetAccessData(token string) (*model.AccessData, *model.AppError)
	GetAccessDataByUserForApp(userId, clientId string) ([]*model.AccessData, *model.AppError)
	GetAccessDataByRefreshToken(token string) (*model.AccessData, *model.AppError)
	GetPreviousAccessData(userId, clientId string) (*model.AccessData, *model.AppError)
	RemoveAccessData(token string) *model.AppError
	RemoveAllAccessData() *model.AppError
}

type SystemStore interface {
	Save(system *model.System) *model.AppError
	SaveOrUpdate(system *model.System) *model.AppError
	Update(system *model.System) *model.AppError
	Get() (model.StringMap, *model.AppError)
	GetByName(name string) (*model.System, *model.AppError)
	PermanentDeleteByName(name string) (*model.System, *model.AppError)
}

type PreferenceStore interface {
	Save(preferences *model.Preferences) *model.AppError
	GetCategory(userId string, category string) (model.Preferences, *model.AppError)
	Get(userId string, category string, name string) (*model.Preference, *model.AppError)
	GetAll(userId string) (model.Preferences, *model.AppError)
	Delete(userId, category, name string) *model.AppError
	DeleteCategory(userId string, category string) *model.AppError
	DeleteCategoryAndName(category string, name string) *model.AppError
	PermanentDeleteByUser(userId string) *model.AppError
	CleanupFlagsBatch(limit int64) (int64, *model.AppError)
}

type LicenseStore interface {
	Save(license *model.LicenseRecord) (*model.LicenseRecord, *model.AppError)
	Get(id string) (*model.LicenseRecord, *model.AppError)
}

type TokenStore interface {
	Save(recovery *model.Token) *model.AppError
	Delete(token string) *model.AppError
	GetByToken(token string) (*model.Token, *model.AppError)
	Cleanup()
	RemoveAllTokensByType(tokenType string) *model.AppError
}

type EmojiStore interface {
	Save(emoji *model.Emoji) (*model.Emoji, *model.AppError)
	Get(id string, allowFromCache bool) (*model.Emoji, *model.AppError)
	GetByName(name string, allowFromCache bool) (*model.Emoji, *model.AppError)
	GetMultipleByName(names []string) ([]*model.Emoji, *model.AppError)
	GetList(offset, limit int, sort string) ([]*model.Emoji, *model.AppError)
	Delete(emoji *model.Emoji, time int64) *model.AppError
	Search(name string, prefixOnly bool, limit int) ([]*model.Emoji, *model.AppError)
}

type StatusStore interface {
	SaveOrUpdate(status *model.Status) *model.AppError
	Get(userId string) (*model.Status, *model.AppError)
	GetByIds(userIds []string) ([]*model.Status, *model.AppError)
	ResetAll() *model.AppError
	GetTotalActiveUsersCount() (int64, *model.AppError)
	UpdateLastActivityAt(userId string, lastActivityAt int64) *model.AppError
}

type FileInfoStore interface {
	Save(info *model.FileInfo) (*model.FileInfo, *model.AppError)
	Get(id string) (*model.FileInfo, *model.AppError)
	GetByPath(path string) (*model.FileInfo, *model.AppError)
	GetForUser(userId string) ([]*model.FileInfo, *model.AppError)
	GetWithOptions(page, perPage int, opt *model.GetFileInfosOptions) ([]*model.FileInfo, *model.AppError)
	PermanentDelete(fileId string) *model.AppError
	PermanentDeleteBatch(endTime int64, limit int64) (int64, *model.AppError)
	PermanentDeleteByUser(userId string) (int64, *model.AppError)
	ClearCaches()
}

type ReactionStore interface {
	Save(reaction *model.Reaction) (*model.Reaction, *model.AppError)
	Delete(reaction *model.Reaction) (*model.Reaction, *model.AppError)
	DeleteAllWithEmojiName(emojiName string) *model.AppError
	PermanentDeleteBatch(endTime int64, limit int64) (int64, *model.AppError)
}

type JobStore interface {
	Save(job *model.Job) (*model.Job, *model.AppError)
	UpdateOptimistically(job *model.Job, currentStatus string) (bool, *model.AppError)
	UpdateStatus(id string, status string) (*model.Job, *model.AppError)
	UpdateStatusOptimistically(id string, currentStatus string, newStatus string) (bool, *model.AppError)
	Get(id string) (*model.Job, *model.AppError)
	GetAllPage(offset int, limit int) ([]*model.Job, *model.AppError)
	GetAllByType(jobType string) ([]*model.Job, *model.AppError)
	GetAllByTypePage(jobType string, offset int, limit int) ([]*model.Job, *model.AppError)
	GetAllByStatus(status string) ([]*model.Job, *model.AppError)
	GetNewestJobByStatusAndType(status string, jobType string) (*model.Job, *model.AppError)
	GetCountByStatusAndType(status string, jobType string) (int64, *model.AppError)
	Delete(id string) (string, *model.AppError)
}

type UserAccessTokenStore interface {
	Save(token *model.UserAccessToken) (*model.UserAccessToken, *model.AppError)
	DeleteAllForUser(userId string) *model.AppError
	Delete(tokenId string) *model.AppError
	Get(tokenId string) (*model.UserAccessToken, *model.AppError)
	GetAll(offset int, limit int) ([]*model.UserAccessToken, *model.AppError)
	GetByToken(tokenString string) (*model.UserAccessToken, *model.AppError)
	GetByUser(userId string, page, perPage int) ([]*model.UserAccessToken, *model.AppError)
	Search(term string) ([]*model.UserAccessToken, *model.AppError)
	UpdateTokenEnable(tokenId string) *model.AppError
	UpdateTokenDisable(tokenId string) *model.AppError
}

type RoleStore interface {
	Save(role *model.Role) (*model.Role, *model.AppError)
	Get(roleId string) (*model.Role, *model.AppError)
	GetAll() ([]*model.Role, *model.AppError)
	GetByName(name string) (*model.Role, *model.AppError)
	GetByNames(names []string) ([]*model.Role, *model.AppError)
	Delete(roldId string) (*model.Role, *model.AppError)
	PermanentDeleteAll() *model.AppError
}

type SchemeStore interface {
	Save(scheme *model.Scheme) (*model.Scheme, *model.AppError)
	Get(schemeId string) (*model.Scheme, *model.AppError)
	GetByName(schemeName string) (*model.Scheme, *model.AppError)
	GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, *model.AppError)
	Delete(schemeId string) (*model.Scheme, *model.AppError)
	PermanentDeleteAll() *model.AppError
}

type LinkMetadataStore interface {
	Save(linkMetadata *model.LinkMetadata) (*model.LinkMetadata, *model.AppError)
	Get(url string, timestamp int64) (*model.LinkMetadata, *model.AppError)
}

type UserGetByIdsOpts struct {
	// IsAdmin tracks whether or not the request is being made by an administrator. Does nothing when provided by a client.
	IsAdmin bool

	// Restrict to search in a list. Does nothing when provided by a client.
	ViewRestrictions *model.ViewUsersRestrictions

	// Since filters the users based on their UpdateAt timestamp.
	Since int64
}

type OrphanedRecord struct {
	ParentId *string
	ChildId  *string
}

type RelationalIntegrityCheckData struct {
	ParentName   string
	ChildName    string
	ParentIdAttr string
	ChildIdAttr  string
	Records      []OrphanedRecord
}

type IntegrityCheckResult struct {
	Data interface{}
	Err  error
}
