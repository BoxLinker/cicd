package registry_watcher

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
	"encoding/json"
	"github.com/BoxLinker/boxlinker-api/pkg/registry"
)



func (a *Api) RegistryEvent(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read body: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	events := &registry.EventCallback{}
	if err := json.Unmarshal(b, events); err != nil {
		http.Error(w, fmt.Sprintf("Unmarshal body: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	// 确认镜像以及 tag 是否存在，如果不存在创建镜像记录
	// 创建 image:tag action 记录
	authorization := r.Header.Get("Authorization")
	if authorization != "just4fun" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	for _, event := range events.Events {
		logrus.Debugf("------------------------------")
		logrus.Debugf("deal with event: %+v", event)
		if event.Action != "push" {
			logrus.Debugf("event: %s detected pass")
			continue
		}
		// image 格式为 {namespace}/{imageName}:{tag}
		parts := strings.Split(event.Target.Repository, "/")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
			return
		}
		if err := a.manager.Publish(&event); err != nil {
			logrus.Errorf("publish err: (%s)", err.Error())
		}
	}
	logrus.Debugf("-------------- deal with registry event end----------------")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
