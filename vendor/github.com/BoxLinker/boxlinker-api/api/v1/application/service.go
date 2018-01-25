package application

import (
	"fmt"
	"net/http"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServicePortForm struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
}

type ServiceForm struct {
	Name   string             `json:"name"`
	Image  string             `json:"image"`
	Memory string             `json:"memory"`
	CPU    string             `json:"cpu"`
	Ports  []*ServicePortForm `json:"ports"`
}

func getDeployByName(name string, list *appsv1beta1.DeploymentList) *appsv1beta1.Deployment {
	for _, item := range list.Items {
		if item.Name == name {
			return &item
		}
	}
	return nil
}

func getIngByName(name string, list *extv1beta1.IngressList) *extv1beta1.Ingress {
	for _, item := range list.Items {
		if item.Name == name {
			return &item
		}
	}
	return nil
}

func (a *Api) IsServiceExist(w http.ResponseWriter, r *http.Request) {
	svcName := mux.Vars(r)["name"]
	user := a.getUserInfo(r)
	namespace := user.Name
	found, err, _, _, _ := a.manager.GetServiceByName(namespace, svcName)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	if !found {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil)
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}

func (a *Api) DeleteService(w http.ResponseWriter, r *http.Request) {
	svcName := mux.Vars(r)["name"]
	user := a.getUserInfo(r)
	namespace := user.Name
	deployOperator := a.clientSet.AppsV1beta1().Deployments(namespace)
	deploy, err := deployOperator.Get(svcName, metav1.GetOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	svcOperator := a.clientSet.CoreV1().Services(namespace)
	svc, err := svcOperator.Get(svcName, metav1.GetOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("service not found (%s/%s)", namespace, svcName))
		return
	}
	ingOperator := a.clientSet.ExtensionsV1beta1().Ingresses(namespace)
	ing, err := ingOperator.Get(svcName, metav1.GetOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("ingress not found (%s/%s)", namespace, svcName))
		return
	}

	deletePolicy := metav1.DeletePropagationForeground
	if err := deployOperator.Delete(deploy.Name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, err.Error())
		return
	}
	if err := svcOperator.Delete(svc.Name, &metav1.DeleteOptions{}); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, err.Error())
		return
	}
	if err := ingOperator.Delete(ing.Name, &metav1.DeleteOptions{}); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, err.Error())
		return
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}

func (a *Api) UpdateService(w http.ResponseWriter, r *http.Request) {
	svcName := mux.Vars(r)["name"]
	user := a.getUserInfo(r)
	namespace := user.Name
	form := &ServiceForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}
	deployOperator := a.clientSet.AppsV1beta1().Deployments(namespace)
	deploy, err := deployOperator.Get(svcName, metav1.GetOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	svcOperator := a.clientSet.CoreV1().Services(namespace)
	svc, err := svcOperator.Get(svcName, metav1.GetOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("service not found (%s/%s)", namespace, svcName))
		return
	}
	ingOperator := a.clientSet.ExtensionsV1beta1().Ingresses(namespace)
	ing, err := ingOperator.Get(svcName, metav1.GetOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("ingress not found (%s/%s)", namespace, svcName))
		return
	}

	var container *apiv1.Container
	containers := deploy.Spec.Template.Spec.Containers
	if len(containers) == 1 {
		container = &containers[0]

	}

	// update image
	if form.Image != "" {
		if form.Image != container.Image {
			logrus.Debugf("Update deploy %s/%s with new image (%s)", user.Name, svcName, form.Image)
			container.Image = form.Image
		}
	}

	// update memory
	if form.Memory != "" {
		memory, err := resource.ParseQuantity(form.Memory)
		if err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, fmt.Sprintf("memory param (%s) is invalid", form.Memory))
			return
		}
		logrus.Debugf("Update deploy %s/%s with new memory (%s)", user.Name, svcName, form.Memory)
		container.Resources.Limits[apiv1.ResourceMemory] = memory
		container.Resources.Requests[apiv1.ResourceMemory] = memory
	}

	// update cpu
	if form.CPU != "" {
		cpu, err := resource.ParseQuantity(form.CPU)
		if err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, fmt.Sprintf("cpu param (%s) is invalid", form.CPU))
			return
		}
		logrus.Debugf("Update deploy %s/%s with new cpu (%s)", user.Name, svcName, form.CPU)
		container.Resources.Limits[apiv1.ResourceCPU] = cpu
		container.Resources.Requests[apiv1.ResourceCPU] = cpu
	}

	// update ports/path
	ports := make([]apiv1.ContainerPort, 0)
	svcPorts := make([]apiv1.ServicePort, 0)
	paths := make([]extv1beta1.HTTPIngressPath, 0)
	if len(form.Ports) > 0 {
		for _, port := range form.Ports {
			ports = append(ports, FormatContainerPort(port.Name, port.Protocol, port.Port))
			svcPorts = append(svcPorts, FormatServicePort(port.Name, port.Protocol, port.Port))
			paths = append(paths, FormatIngressPath(port.Path, svcName, port.Port))
		}
		container.Ports = ports
		logrus.Debugf("updated deployment (%s/%s) ports: ->\n\t%+v", namespace, svcName, svcPorts)

		svc.Spec.Ports = svcPorts
		logrus.Debugf("updated service (%s/%s) ports: ->\n\t%+v", namespace, svcName, svcPorts)

		rules := ing.Spec.Rules
		if len(rules) > 0 {
			rule := rules[0]
			rule.HTTP.Paths = paths
			logrus.Debugf("updated ingress (%s/%s) paths: ->\n\t%+v", namespace, svcName, paths)
		}
	}

	// 处理 ingress
	if _, err := ingOperator.Update(ing); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, fmt.Sprintf("update ingress (%s/%s) error: %v", namespace, svcName, err))
		return
	}

	// 处理 service
	if _, err := svcOperator.Update(svc); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, fmt.Sprintf("update service (%s/%s) error: %v", namespace, svcName, err))
		return
	}

	// 处理 deployment
	if _, err := deployOperator.Update(deploy); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, deploy)
}

func (a *Api) QueryService(w http.ResponseWriter, r *http.Request) {
	user := a.getUserInfo(r)
	pc := boxlinker.ParsePageConfig(r)
	deploys, err := a.clientSet.AppsV1beta1().Deployments(user.Name).List(metav1.ListOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	ings, err := a.clientSet.ExtensionsV1beta1().Ingresses(user.Name).List(metav1.ListOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	svcs, err := a.clientSet.CoreV1().Services(user.Name).List(metav1.ListOptions{})
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	output := make([]*ServiceForm, 0)
	var start, end int
	l := len(svcs.Items)
	// todo 这里得判断 ings deploys 和 svcs 长度是否一样
	pc.TotalCount = l
	if pc.Offset() >= l {
		start = 0
		end = l
	} else {
		start = pc.Offset()
		if pc.Offset()+pc.Limit() >= l {
			end = l
		} else {
			end = pc.Offset() + pc.Limit()
		}
	}
	listOut := svcs.Items[start:end]
	for _, item := range listOut {
		deploy := getDeployByName(item.Name, deploys)
		ing := getIngByName(item.Name, ings)
		line := &ServiceForm{
			Name: item.Name,
		}
		if deploy != nil {
			containers := deploy.Spec.Template.Spec.Containers
			if len(containers) == 1 {
				container := containers[0]
				line.Image = container.Image
				line.Memory = container.Resources.Limits.Memory().String()
				line.CPU = container.Resources.Limits.Cpu().String()
			} else {
				logrus.Warnf("Found Service contains more than one container: (%s)", item.Name)
			}
		}
		ports := item.Spec.Ports
		portsF := make([]*ServicePortForm, 0)
		if len(ports) > 0 {
			for _, port := range ports {
				svcPortForm := &ServicePortForm{
					Name: port.Name,
					// todo 转化 ServicePort Protocol 为 字符串
					Protocol: "tcp",
					Port:     int(port.Port),
				}
				if ing != nil {
					rules := ing.Spec.Rules
					if len(rules) > 0 {
						paths := rules[0].HTTP.Paths
						for _, path := range paths {
							if path.Backend.ServiceName == item.Name && path.Backend.ServicePort.IntVal == port.Port {
								svcPortForm.Path = path.Path
							}
						}
					}
				}
				portsF = append(portsF, svcPortForm)
			}
			line.Ports = portsF
		}
		output = append(output, line)
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, map[string]interface{}{
		"pagination": pc.PaginationJSON(),
		"data":       output,
	})
}

func (a *Api) CreateService(w http.ResponseWriter, r *http.Request) {
	user := a.getUserInfo(r)
	form := &ServiceForm{}
	if err := boxlinker.ReadRequestBody(r, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}

	logrus.Debugf("Create Service form: (%+v)", form)
	memoryQuantity, err := resource.ParseQuantity(form.Memory)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, fmt.Sprintf("memory param (%s) is invalid", form.Memory))
		return
	}
	cpuQuantity, err := resource.ParseQuantity(form.CPU)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, nil, fmt.Sprintf("cpu param (%s) is invalid", form.CPU))
		return
	}
	// registry key
	registryKey := make([]apiv1.LocalObjectReference, 0)
	registryKey = append(registryKey, apiv1.LocalObjectReference{Name: "registry-key"})

	// ports
	ports := make([]apiv1.ContainerPort, 0)
	portsF := form.Ports
	if len(portsF) > 0 {
		for _, port := range portsF {
			ports = append(ports, FormatContainerPort(port.Name, port.Protocol, port.Port))
		}
	}

	// create deployment
	deploymentsClient := a.clientSet.AppsV1beta1().Deployments(user.Name)
	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: form.Name,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": form.Name,
					},
				},
				Spec: apiv1.PodSpec{
					ImagePullSecrets: registryKey,
					Containers: []apiv1.Container{
						{
							Name:  fmt.Sprintf("%s-%s", form.Name, "container"),
							Image: form.Image,
							Ports: ports,
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: memoryQuantity,
									apiv1.ResourceCPU:    cpuQuantity,
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: memoryQuantity,
									apiv1.ResourceCPU:    cpuQuantity,
								},
							},
						},
					},
				},
			},
		},
	}
	logrus.Debugf("Create Deployment %s/%s (%+v)", user.Name, form.Name, deployment)
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	logrus.Debugf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	/**
	 *	如果没有暴露 port ，那么就没必要生成 svc 和 ingress 了， 直接返回
	 */
	if len(portsF) <= 0 {
		boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
		return
	}

	// create service
	svcPorts := make([]apiv1.ServicePort, 0)
	// ports
	for _, port := range portsF {
		svcPorts = append(svcPorts, FormatServicePort(port.Name, port.Protocol, port.Port))
	}

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: form.Name,
		},
		Spec: apiv1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"app": form.Name,
			},
		},
	}
	logrus.Debugf("Create Svc %s/%s (%+v)", user.Name, form.Name, service)
	svc, err := a.clientSet.CoreV1().Services(user.Name).Create(service)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, "", err.Error())
		return
	}
	logrus.Debugf("Created Svc %q.\n", svc.GetObjectMeta().GetName())

	// create ingress
	paths := make([]extv1beta1.HTTPIngressPath, 0)
	for _, port := range portsF {
		paths = append(paths, FormatIngressPath(port.Path, form.Name, port.Port))
	}
	rules := make([]extv1beta1.IngressRule, 0)
	rules = append(rules, extv1beta1.IngressRule{
		Host: fmt.Sprintf("%s.%s.boxlinker.com", form.Name, "lb1"),
		IngressRuleValue: extv1beta1.IngressRuleValue{
			HTTP: &extv1beta1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	})
	ingress := &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: form.Name,
		},
		Spec: extv1beta1.IngressSpec{
			Rules: rules,
		},
	}
	logrus.Debugf("Create Ingress %s/%s (%+v)", user.Name, form.Name, ingress)
	ing, err := a.clientSet.ExtensionsV1beta1().Ingresses(user.Name).Create(ingress)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FAILED, "", err.Error())
		return
	}
	logrus.Debugf("Created ingress %q.\n", ing.GetObjectMeta().GetName())

	boxlinker.Resp(w, boxlinker.STATUS_OK, nil)
}
