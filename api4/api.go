// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nhannv/quiz/v5/app"
	"github.com/nhannv/quiz/v5/model"
	"github.com/nhannv/quiz/v5/services/configservice"
	"github.com/nhannv/quiz/v5/web"

	_ "github.com/mattermost/go-i18n/i18n"
)

type Routes struct {
	Root    *mux.Router // ''
	ApiRoot *mux.Router // 'api/v4'

	Users          *mux.Router // 'api/v4/users'
	User           *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}'
	UserByUsername *mux.Router // 'api/v4/users/username/{username:[A-Za-z0-9_-\.]+}'
	UserByEmail    *mux.Router // 'api/v4/users/email/{email}'

	Bots *mux.Router // 'api/v4/bots'
	Bot  *mux.Router // 'api/v4/bots/{bot_user_id:[A-Za-z0-9]+}'

	Files *mux.Router // 'api/v4/files'
	File  *mux.Router // 'api/v4/files/{file_id:[A-Za-z0-9]+}'

	PublicFile *mux.Router // 'files/{file_id:[A-Za-z0-9]+}/public'

	Commands *mux.Router // 'api/v4/commands'
	Command  *mux.Router // 'api/v4/commands/{command_id:[A-Za-z0-9]+}'

	OAuth     *mux.Router // 'api/v4/oauth'
	OAuthApps *mux.Router // 'api/v4/oauth/apps'
	OAuthApp  *mux.Router // 'api/v4/oauth/apps/{app_id:[A-Za-z0-9]+}'

	OpenGraph *mux.Router // 'api/v4/opengraph'

	SAML       *mux.Router // 'api/v4/saml'
	Compliance *mux.Router // 'api/v4/compliance'
	Cluster    *mux.Router // 'api/v4/cluster'

	Image *mux.Router // 'api/v4/image'

	LDAP *mux.Router // 'api/v4/ldap'

	Elasticsearch *mux.Router // 'api/v4/elasticsearch'

	DataRetention *mux.Router // 'api/v4/data_retention'

	Brand *mux.Router // 'api/v4/brand'

	System *mux.Router // 'api/v4/system'

	Jobs *mux.Router // 'api/v4/jobs'

	Preferences *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/preferences'

	License *mux.Router // 'api/v4/license'

	Public *mux.Router // 'api/v4/public'

	Reactions *mux.Router // 'api/v4/reactions'

	Roles   *mux.Router // 'api/v4/roles'
	Schemes *mux.Router // 'api/v4/schemes'

	Emojis      *mux.Router // 'api/v4/emoji'
	Emoji       *mux.Router // 'api/v4/emoji/{emoji_id:[A-Za-z0-9]+}'
	EmojiByName *mux.Router // 'api/v4/emoji/name/{emoji_name:[A-Za-z0-9_-\.]+}'

	ReactionByNameForPostForUser *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/posts/{post_id:[A-Za-z0-9]+}/reactions/{emoji_name:[A-Za-z0-9_-+]+}'

}

type API struct {
	ConfigService       configservice.ConfigService
	GetGlobalAppOptions app.AppOptionCreator
	BaseRoutes          *Routes
}

func Init(configservice configservice.ConfigService, globalOptionsFunc app.AppOptionCreator, root *mux.Router) *API {
	api := &API{
		ConfigService:       configservice,
		GetGlobalAppOptions: globalOptionsFunc,
		BaseRoutes:          &Routes{},
	}

	api.BaseRoutes.Root = root
	api.BaseRoutes.ApiRoot = root.PathPrefix(model.API_URL_SUFFIX).Subrouter()

	api.BaseRoutes.Users = api.BaseRoutes.ApiRoot.PathPrefix("/users").Subrouter()
	api.BaseRoutes.User = api.BaseRoutes.ApiRoot.PathPrefix("/users/{user_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.UserByUsername = api.BaseRoutes.Users.PathPrefix("/username/{username:[A-Za-z0-9\\_\\-\\.]+}").Subrouter()
	api.BaseRoutes.UserByEmail = api.BaseRoutes.Users.PathPrefix("/email/{email}").Subrouter()

	api.BaseRoutes.Files = api.BaseRoutes.ApiRoot.PathPrefix("/files").Subrouter()
	api.BaseRoutes.File = api.BaseRoutes.Files.PathPrefix("/{file_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.PublicFile = api.BaseRoutes.Root.PathPrefix("/files/{file_id:[A-Za-z0-9]+}/public").Subrouter()

	api.BaseRoutes.Commands = api.BaseRoutes.ApiRoot.PathPrefix("/commands").Subrouter()
	api.BaseRoutes.Command = api.BaseRoutes.Commands.PathPrefix("/{command_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.SAML = api.BaseRoutes.ApiRoot.PathPrefix("/saml").Subrouter()

	api.BaseRoutes.OAuth = api.BaseRoutes.ApiRoot.PathPrefix("/oauth").Subrouter()
	api.BaseRoutes.OAuthApps = api.BaseRoutes.OAuth.PathPrefix("/apps").Subrouter()
	api.BaseRoutes.OAuthApp = api.BaseRoutes.OAuthApps.PathPrefix("/{app_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Compliance = api.BaseRoutes.ApiRoot.PathPrefix("/compliance").Subrouter()
	api.BaseRoutes.Cluster = api.BaseRoutes.ApiRoot.PathPrefix("/cluster").Subrouter()
	api.BaseRoutes.LDAP = api.BaseRoutes.ApiRoot.PathPrefix("/ldap").Subrouter()
	api.BaseRoutes.Brand = api.BaseRoutes.ApiRoot.PathPrefix("/brand").Subrouter()
	api.BaseRoutes.System = api.BaseRoutes.ApiRoot.PathPrefix("/system").Subrouter()
	api.BaseRoutes.Preferences = api.BaseRoutes.User.PathPrefix("/preferences").Subrouter()
	api.BaseRoutes.License = api.BaseRoutes.ApiRoot.PathPrefix("/license").Subrouter()
	api.BaseRoutes.Public = api.BaseRoutes.ApiRoot.PathPrefix("/public").Subrouter()
	api.BaseRoutes.Reactions = api.BaseRoutes.ApiRoot.PathPrefix("/reactions").Subrouter()
	api.BaseRoutes.Jobs = api.BaseRoutes.ApiRoot.PathPrefix("/jobs").Subrouter()
	api.BaseRoutes.Elasticsearch = api.BaseRoutes.ApiRoot.PathPrefix("/elasticsearch").Subrouter()
	api.BaseRoutes.DataRetention = api.BaseRoutes.ApiRoot.PathPrefix("/data_retention").Subrouter()

	api.BaseRoutes.Emojis = api.BaseRoutes.ApiRoot.PathPrefix("/emoji").Subrouter()
	api.BaseRoutes.Emoji = api.BaseRoutes.ApiRoot.PathPrefix("/emoji/{emoji_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.EmojiByName = api.BaseRoutes.Emojis.PathPrefix("/name/{emoji_name:[A-Za-z0-9\\_\\-\\+]+}").Subrouter()

	api.BaseRoutes.OpenGraph = api.BaseRoutes.ApiRoot.PathPrefix("/opengraph").Subrouter()

	api.BaseRoutes.Roles = api.BaseRoutes.ApiRoot.PathPrefix("/roles").Subrouter()
	api.BaseRoutes.Schemes = api.BaseRoutes.ApiRoot.PathPrefix("/schemes").Subrouter()

	api.BaseRoutes.Image = api.BaseRoutes.ApiRoot.PathPrefix("/image").Subrouter()

	api.InitUser()
	api.InitSystem()
	api.InitLicense()
	api.InitConfig()
	api.InitPreference()
	api.InitSaml()
	api.InitCluster()
	api.InitLdap()
	api.InitElasticsearch()
	api.InitDataRetention()
	api.InitJob()
	api.InitStatus()
	api.InitWebSocket()
	api.InitOAuth()
	api.InitOpenGraph()
	api.InitRole()
	api.InitScheme()
	api.InitImage()

	root.Handle("/api/v4/{anything:.*}", http.HandlerFunc(api.Handle404))

	return api
}

func (api *API) Handle404(w http.ResponseWriter, r *http.Request) {
	web.Handle404(api.ConfigService, w, r)
}

var ReturnStatusOK = web.ReturnStatusOK
