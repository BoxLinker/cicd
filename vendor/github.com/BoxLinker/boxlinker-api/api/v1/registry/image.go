package registry

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	registryModels "github.com/BoxLinker/boxlinker-api/controller/models/registry"
	userModels "github.com/BoxLinker/boxlinker-api/controller/models/user"
	"strings"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/boxlinker-api/modules/validate_form"
	"strconv"
)

type ImageForm struct {
	validate_form.ValidateForm
	Name string `json:"name" validate:"required"`
	Tag string `json:"tag"`
	Description string `json:"description"`
	HtmlDoc string `json:"html_doc"`
	IsPrivate bool `json:"is_private"`
}

func (a *Api) getUserInfo(r *http.Request) *userModels.User {
	us := r.Context().Value("user")
	if us == nil {
		return nil
	}
	ctx := us.(map[string]interface{})
	if ctx == nil || ctx["uid"] == nil {
		return nil
	}
	return &userModels.User{
		Id: ctx["uid"].(string),
		Name: ctx["username"].(string),
	}
}

func (a *Api) QueryPubImages(w http.ResponseWriter, r *http.Request) {
	var output []map[string]interface{}
	images, err := a.Manager.QueryImagesByPrivilege(false, boxlinker.ParsePageConfig(r))
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	for _, image := range images {
		output = append(output, image.APISimpleJson())
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, output)
}


// GET		/v1/registry/auth/image/list?private=1&current_page=1&page_count=10
// 			获取已登录用户的镜像
func (a *Api) QueryImages(w http.ResponseWriter, r *http.Request) {
	// 获取用户镜像需要验证
	user := a.getUserInfo(r)
	if user == nil || user.Name == ""  {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}

	var output []map[string]interface{}
	var images []*registryModels.Image
	pc := boxlinker.ParsePageConfig(r)
	var err error
	isPrivate := -1
	namespace := user.Name
	logrus.Debugf("username:>> %s", namespace)
	private := r.URL.Query().Get("private")
	if strings.Index("10", private) >= 0 {
		if ii, err := strconv.Atoi(private); err == nil {
			isPrivate = ii
		}
	}

	cond := "namespace = ?"
	var args []interface{}
	args = append(args, namespace)
	if isPrivate > -1 {
		cond += " and is_private = ?"
		args = append(args, isPrivate)
	}
	images, err = a.Manager.QueryImagesByConditions(cond, args, nil ,pc)

	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	for _, image := range images {
		output = append(output, image.APISimpleJson())
	}
	count, err := a.Manager.CountImagesByNamespace(namespace)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	pc.TotalCount = int(count)
	boxlinker.Resp(w, boxlinker.STATUS_OK, pc.FormatOutput(output))
}
// GET		/v1/registry/image/:id
func (a *Api) GetImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	image, err := a.Manager.FindImageById(id)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	if image.IsPrivate {
		user := a.getUserInfo(r)
		if user == nil || user.Name != image.Namespace{
			boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
			return
		}
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, image.APIJson())
}
// POST		/v1/registry/image/new
func (a *Api) SaveImage(w http.ResponseWriter, r *http.Request) {

	form := &ImageForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}

	user := a.getUserInfo(r)
	if user == nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}

	if err := form.Validate(); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}
	image := &registryModels.Image{
		Namespace: user.Name,
		Name: form.Name,
		Tag: form.Tag,
		Description: form.Description,
		HtmlDoc: form.HtmlDoc,
		IsPrivate: form.IsPrivate,
	}

	if err := a.Manager.SaveImage(image); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil, "success")
}

// /v1/registry/image/exists?name={imageName}
func (a *Api) ImageExists(w http.ResponseWriter, r *http.Request) {
	user := a.getUserInfo(r)
	if user == nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil)
		return
	}
	images, err := a.Manager.QueryImagesByConditions("namespace = ? and name = ?", []interface{}{user.Name, name}, []string{"id"}, boxlinker.PageConfig{
		CurrentPage: 1,
		PageCount:1,
	})
	if err != nil || len(images) == 0 {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil)
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}

func (a *Api) imageExistWithUser(id, namespace string) error {
	image, err := a.Manager.FindImageById(id)

	if err != nil  {
		return err
	}
	if image.Namespace != namespace {
		return fmt.Errorf("image (%s) for (%s) not found", id, namespace)
	}
	return nil
}

// PUT		/v1/registry/image/:id/description
func (a *Api) UpdateImageDescription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	form := &ImageForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}

	user := a.getUserInfo(r)
	if user == nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}

	if err := a.imageExistWithUser(id, user.Name); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}

	if err := a.Manager.UpdateImage(id, &registryModels.Image{
		Description: form.Description,
	}); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil, "success")
}
// PUT		/v1/registry/image/:id/html_doc
func (a *Api) UpdateImageHtmlDoc(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user := a.getUserInfo(r)
	if user == nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}

	if err := a.imageExistWithUser(id, user.Name); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}

	form := &ImageForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}
	if err := a.Manager.UpdateImage(id, &registryModels.Image{
		HtmlDoc: form.HtmlDoc,
	}); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil, "success")
}
// PUT		/v1/registry/image/:id/privilege
func (a *Api) UpdateImagePrivilege(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user := a.getUserInfo(r)
	if user == nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}

	if err := a.imageExistWithUser(id, user.Name); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}

	form := &ImageForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}
	if err := a.Manager.UpdateImage(id, &registryModels.Image{
		IsPrivate: form.IsPrivate,
	}, []string{"is_private", "updated_unix"}); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil, "success")
}
// DELETE	/v1/registry/image/:id
func (a *Api) DeleteImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user := a.getUserInfo(r)
	if user == nil {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}

	if err := a.imageExistWithUser(id, user.Name); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}

	if err := a.Manager.DeleteImageById(id); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil, "success")
}

// POST		/v1/registry/event
func (a *Api) RegistryEvent(w http.ResponseWriter, r *http.Request){
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read body: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	events := &RegistryCallback{}
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
		iImage, err := a.Manager.GetImageByIndexKey(parts[0], parts[1], event.Target.Tag)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		image := &registryModels.Image{
			Size: event.Target.Size,
			Digest: event.Target.Digest,
		}
		if iImage != nil {
			if err := a.Manager.UpdateImage(iImage.Id, image); err != nil {
				http.Error(w, fmt.Errorf("db operation err [update image] (%v)", err).Error(), http.StatusInternalServerError)
				return
			}
			logrus.Debugf("update image: %+v -> %+v", iImage, image)
		} else {
			image.Namespace = parts[0]
			image.Name = parts[1]
			image.Tag = event.Target.Tag
			// 镜像库推送过来的默认是私有
			image.IsPrivate = true
			if err := a.Manager.SaveImage(image); err != nil {
				http.Error(w, fmt.Errorf("db operation err [save image] (%v)", err).Error(), http.StatusInternalServerError)
				return
			}
			logrus.Debugf("save image: %+v", image)
		}
	}
	logrus.Debugf("-------------- deal with registry event end----------------")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

