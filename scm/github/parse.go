package github

import (
	"net/http"
	"github.com/BoxLinker/cicd/models"
	"io"
	"bytes"
	"io/ioutil"
	"encoding/json"
)

const (
	hookEvent = "X-Github-Event"
	hookField = "payload"
	hookDeploy = "deployment"
	hookPush = "push"
	hookPull = "pull_request"

	actionOpen = "opened"
	actionSync = "synchronize"

	stateOpen = "open"
)

// 根据 github 的回调请求，解析并获取对应的 repo 和 build
func parseHook(r *http.Request, merge bool) (*models.Repo, *models.Build, error) {
	var reader io.Reader = r.Body

	if payload := r.FormValue(hookField); payload != "" {
		reader = bytes.NewBufferString(payload)
	}

	raw, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	switch r.Header.Get(hookEvent) {
	case hookPush:
		return parsePushHook(raw)
	case hookDeploy:
		return parseDeployHook(raw)
	case hookPull:
		return parsePullHook(raw, merge)
	}
	return nil, nil, nil
}

func parsePushHook(payload []byte) (*models.Repo, *models.Build, error) {
	hook := new(webhook)
	err := json.Unmarshal(payload, hook)
	if err != nil {
		return nil, nil, err
	}
	if hook.Deleted {
		return nil, nil, err
	}
	return convertRepoHook(hook), convertPushHook(hook), nil
}

func parseDeployHook(payload []byte) (*models.Repo, *models.Build, error) {
	hook := new(webhook)
	if err := json.Unmarshal(payload, hook); err != nil {
		return nil, nil, err
	}
	return convertRepoHook(hook), convertDeployHook(hook), nil
}

func parsePullHook(payload []byte, merge bool) (*models.Repo, *models.Build, error) {
	hook := new(webhook)
	err := json.Unmarshal(payload, hook)
	if err != nil {
		return nil, nil, err
	}

	// ignore these
	if hook.Action != actionOpen && hook.Action != actionSync {
		return nil, nil, nil
	}
	if hook.PullRequest.State != stateOpen {
		return nil, nil, nil
	}
	return convertRepoHook(hook), convertPullHook(hook, merge), nil
}

