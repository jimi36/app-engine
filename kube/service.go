package kube

import (
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	engine "github.com/jimi36/app-engine"
)

func (cli *Client) newService(app *engine.Application) error {
	spec := loadServiceSpec(app)
	if spec == nil {
		return nil
	}

	_, err := cli.kubeCli.CoreV1().Services(cli.ns).Create(spec)
	if err != nil {
		return err
	}

	return nil
}

func (cli *Client) deleteService(name string) error {
	if err := cli.kubeCli.CoreV1().Services(cli.ns).Delete(name, nil); err != nil {
		return err
	}
	return nil
}

func loadServiceSpec(app *engine.Application) *coreV1.Service {
	if app.KubeSpec.Service == nil {
		return nil
	}

	var ports []coreV1.ServicePort
	for _, port := range app.KubeSpec.Service.Ports {
		ports = append(ports, coreV1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.Port,
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: port.TargetPort},
			NodePort:   port.NodePort,
		})
	}

	return &coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   app.Name,
			Labels: app.Labels,
		},
		Spec: coreV1.ServiceSpec{
			Selector: app.Labels,
			Type:     app.KubeSpec.Service.Type,
			Ports:    ports,
		},
	}
}
