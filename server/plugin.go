package main

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-starter-template/server/command"
	"github.com/mattermost/mattermost-plugin-starter-template/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore kvstore.KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// commandClient is the client used to register and execute slash commands.
	commandClient command.Command

	backgroundJob *cluster.Job

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	p.kvstore = kvstore.NewKVStore(p.client)

	p.commandClient = command.NewCommandHandler(p.client)

	job, err := cluster.Schedule(
		p.API,
		"BackgroundJob",
		cluster.MakeWaitForRoundedInterval(1*time.Hour),
		p.runJob,
	)
	if err != nil {
		return errors.Wrap(err, "failed to schedule background job")
	}

	p.backgroundJob = job

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	if p.backgroundJob != nil {
		if err := p.backgroundJob.Close(); err != nil {
			p.API.LogError("Failed to close background job", "err", err)
		}
	}
	return nil
}

// This will execute the commands that were registered in the NewCommandHandler function.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, err := p.commandClient.Handle(args)
	if err != nil {
		return nil, model.NewAppError("ExecuteCommand", "plugin.command.execute_command.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return response, nil
}

// UserWillLogin Hook 실행 (로그인 직전 사용자 정보 수정)
func (p *Plugin) UserWillLogIn(c *plugin.Context, user *model.User) string {
	p.API.LogWarn("UserWillLogin hook triggered", "user_id", user.Id, "original_firstname", user.FirstName)

	formattedFirstName := formatFirstName(user.FirstName)

	// firstname만 변경이 필요한 경우에만 업데이트
	if formattedFirstName != user.FirstName {
		user.FirstName = formattedFirstName
		p.API.LogInfo("Updated firstname", "user_id", user.Id, "new_firstname", formattedFirstName)

		// 변경된 정보 사용자 데이터베이스에 반영
		_, err := p.API.UpdateUser(user)
		if err != nil {
			p.API.LogError("Failed to update user", "user_id", user.Id, "error", err.Error())
			return "Failed to update user"
		}
		p.API.LogInfo("Successfully updated user firstname", "user_id", user.Id, "new_firstname", user.FirstName)
	} else {
		p.API.LogInfo("No update needed for firstname", "user_id", user.Id)
	}

	return ""
}

// `_` → 공백, `/` 전까지 이름만 추출
func formatFirstName(firstname string) string {
	if firstname == "" {
		return firstname
	}

	// "/" 앞부분만 가져옴
	// parts := strings.Split(firstname, "/")
	// cleanName := parts[0]

	// "_"를 공백으로 변환하고 괄호, 쉼표 처리
	formatted := strings.ReplaceAll(firstname, "_", " ")
	formatted = strings.ReplaceAll(formatted, "  ", " ") // 중복 공백 제거

	return strings.TrimSpace(formatted)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
