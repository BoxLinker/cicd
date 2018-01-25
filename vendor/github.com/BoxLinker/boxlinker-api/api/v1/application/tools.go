package application

import (
	apiv1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func FormatIngressPath(path, svcName string, svcPort int) extv1beta1.HTTPIngressPath {
	return extv1beta1.HTTPIngressPath{
		Path: path,
		Backend: extv1beta1.IngressBackend{
			ServiceName: svcName,
			ServicePort: intstr.FromInt(svcPort),
		},
	}
}

func FormatContainerPort(name, protocol string, port int) apiv1.ContainerPort {
	return apiv1.ContainerPort{
		Name: name,
		Protocol: apiv1.ProtocolTCP,
		ContainerPort: int32(port),
	}
}

func FormatServicePort(name, protocol string, port int) apiv1.ServicePort {
	return apiv1.ServicePort{
		Name: name,
		Protocol: apiv1.ProtocolTCP,
		Port: int32(port),
	}
}