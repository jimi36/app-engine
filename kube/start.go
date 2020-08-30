package kube

import (
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
)

func (cli *Client) RestartApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("restart kube application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("restart kube application[%s]......", tag.Tag())

	// check appliation runtime
	if rt, _ := cli.store.GetApplicationRuntime(tag.Name); rt != nil && rt.IsStarted {
		deployCli := cli.kubeCli.AppsV1().Deployments(cli.ns)
		if _, err := deployCli.Get(tag.Name, metaV1.GetOptions{}); err == nil {
			return &engine.TaskResult{}
		}
	}

	cli.store.UpdateApplicationRuntime(tag.Name, func(runtime *engine.ApplicationRuntime) error {
		runtime.IsStarted = false
		return nil
	})

	ret := cli.StartApplication(ev)

	log.Debugf("restart kube application[%s] finished", tag.Tag())

	return ret
}

func (cli *Client) StartApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("start kube application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("start kube application[%s]......", tag.Tag())

	// check and get appliation runtime
	rt, _ := cli.store.GetApplicationRuntime(tag.Name)
	if rt != nil && rt.IsStarted {
		log.Warnf("start kube application[%s] error: %s", tag.Tag(), engine.ErrApplicationStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationStarted,
		}
	}

	app, err := cli.store.GetApplication(tag)
	if err != nil {
		log.Warnf("start kube application[%s] error: %s", tag.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	if rt != nil {
		// if existed, update application runtime
		cli.store.UpdateApplicationRuntime(tag.Name, func(runtime *engine.ApplicationRuntime) error {
			runtime.Version = tag.Version
			runtime.ToStart = true
			runtime.IsStarted = false
			runtime.Err = ""
			return nil
		})
	} else {
		// if not existed, add application runtime
		cli.store.AddApplicationRunTime(&engine.ApplicationRuntime{
			ApplicationTag: *tag,
			ToStart:        true,
			IsStarted:      false,
		})
	}

	// init appliation labels
	if app.Labels == nil {
		app.Labels = make(map[string]string)
	}
	app.Labels[labelEdgeApp] = app.Name
	app.Labels[labelEdgeAppVersion] = app.Version

	if err := cli.newService(app); err != nil {
		log.Warnf("start kube application[%s] error: %s", tag.Tag(), err.Error())
		cli.store.UpdateApplicationRuntime(tag.Name, func(rt *engine.ApplicationRuntime) error {
			rt.ToStart = false
			rt.IsStarted = false
			rt.Err = err.Error()
			return nil
		})
		return &engine.TaskResult{
			Err: err,
		}
	}

	spec := loadDeploymentSpec(app)
	if _, err := cli.kubeCli.AppsV1().Deployments(cli.ns).Create(spec); err != nil {
		log.Warnf("start kube application[%s] error: %s", tag.Tag(), err.Error())
		cli.deleteService(app.Name)
		cli.store.UpdateApplicationRuntime(tag.Name, func(rt *engine.ApplicationRuntime) error {
			rt.ToStart = false
			rt.IsStarted = false
			rt.Err = err.Error()
			return nil
		})
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("start kube application[%s] finished", tag.Tag())

	return &engine.TaskResult{}
}

func loadDeploymentSpec(app *engine.Application) *appV1.Deployment {
	deploy := &appV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   app.Name,
			Labels: app.Labels,
		},
	}
	deploy.Spec.Selector = &metaV1.LabelSelector{
		MatchLabels: app.Labels,
	}
	deploy.Spec.Template = coreV1.PodTemplateSpec{
		ObjectMeta: metaV1.ObjectMeta{
			Labels: app.Labels,
		},
		Spec: coreV1.PodSpec{
			Containers: loadContainerSpecs(app),
			Volumes:    loadVolumeSpecs(app),
		},
	}
	return deploy
}

func loadContainerSpecs(app *engine.Application) []coreV1.Container {
	var containerPorts []coreV1.ContainerPort
	for _, port := range app.KubeSpec.Ports {
		containerPorts = append(containerPorts, coreV1.ContainerPort{
			Name:          port.Name,
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      port.Protocol,
		})
	}

	var volMounts []coreV1.VolumeMount
	for _, vol := range app.KubeSpec.Volumes {
		volMounts = append(volMounts, coreV1.VolumeMount{
			Name:      vol.Name,
			MountPath: vol.MountPath,
		})
	}

	var env []coreV1.EnvVar
	for name, value := range app.Env {
		env = append(env, coreV1.EnvVar{
			Name:  name,
			Value: value,
		})
	}

	return []coreV1.Container{
		{
			Name:            app.Name,
			Image:           app.KubeSpec.Image,
			Command:         app.KubeSpec.Command,
			Env:             env,
			VolumeMounts:    volMounts,
			Ports:           containerPorts,
			ImagePullPolicy: coreV1.PullIfNotPresent,
		},
	}
}

func loadVolumeSpecs(app *engine.Application) []coreV1.Volume {
	var vols []coreV1.Volume
	for _, v := range app.KubeSpec.Volumes {
		vol := coreV1.Volume{
			Name: v.Name,
		}
		switch {
		case len(v.HostPath) > 0:
			vol.HostPath = &coreV1.HostPathVolumeSource{
				Path: v.HostPath,
				Type: &v.HostPathType,
			}
		case len(v.ConfigName) > 0:
			vol.ConfigMap = &coreV1.ConfigMapVolumeSource{
				LocalObjectReference: coreV1.LocalObjectReference{Name: v.ConfigName},
			}
		case len(v.SecretName) > 0:
			vol.Secret = &coreV1.SecretVolumeSource{
				SecretName: v.SecretName,
			}
		}
		vols = append(vols, vol)
	}
	return vols
}
