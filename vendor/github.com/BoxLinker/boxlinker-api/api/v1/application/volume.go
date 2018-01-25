package application

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/boxlinker-api/controller/models"
	"github.com/gorilla/mux"
	apiv1 "k8s.io/api/core/v1"
)

type VolumeForm struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

func (a *Api) CreateVolume(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	form := &VolumeForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}
	claim, err := a.manager.CreateVolume(user.Name, &models.Volume{
		Name: form.Name,
		Size: form.Size,
	})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, claim)
}
func (a *Api) DeleteVolume(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	name := mux.Vars(r)["name"]
	if err := a.manager.DeleteVolume(user.Name, name); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}
func (a *Api) QueryVolume(w http.ResponseWriter, r *http.Request){
	user := a.getUserInfo(r)
	pc := boxlinker.ParsePageConfig(r)
	claims, err := a.manager.QueryVolume(user.Name, pc)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	l := len(claims)
	//var start, end int
	//if pc.Offset() >= l {
	//	start = 0
	//	end = l
	//} else {
	//	start = pc.Offset()
	//	if pc.Offset() + pc.Limit() >= l {
	//		end = l
	//	} else {
	//		end = pc.Offset() + pc.Limit()
	//	}
	//}
	pc.TotalCount = l
	output := make([]*VolumeForm, 0)

	//listOut := claims[start:end]
	//logrus.Debugf("listOut:>\n%+v", listOut)
	//logrus.Debugf("==========")
	for _, item := range claims {
		//logrus.Debugf("item:>\n%+v", item.ObjectMeta.Name, item)
		//logrus.Debugf("==========")
		capacity := item.Status.Capacity[apiv1.ResourceStorage]
		output = append(output, &VolumeForm{
			Name: item.ObjectMeta.Name,
			Size: (&capacity).String(),
		})
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, pc.FormatOutput(output))
}
