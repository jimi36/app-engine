package kube

import (
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
)

func (cli *Client) CreateConfig(ev *engine.TaskEvent) *engine.TaskResult {
	config, ok := ev.In.(*engine.Config)
	if !ok {
		log.Fatalf("create kube config error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("create kube config[%s]......", config.Name)

	if err := cli.store.AddConfig(config); err != nil {
		log.Warnf("create native config[%s] error: %s", config.Name, err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	kubeConfig := toKubeConfigMap(config)
	if _, err := cli.kubeCli.CoreV1().ConfigMaps(cli.ns).Create(kubeConfig); err != nil {
		log.Warnf("create kube config[%s] error: %s", config.Name, err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("create kube config[%s] finished", config.Name)

	return &engine.TaskResult{}
}

func (cli *Client) RemoveConfig(ev *engine.TaskEvent) *engine.TaskResult {
	name, ok := ev.In.(string)
	if !ok {
		log.Fatalf("remove kube config error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("remove kube config[%s]......", name)

	if err := cli.kubeCli.CoreV1().ConfigMaps(cli.ns).Delete(name, nil); err != nil {
		log.Warnf("remove kube config[%s] error: %s", name, engine.ErrConfigNoExisted.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	if err := cli.store.RemoveConfig(name); err != nil {
		log.Warnf("remove kube config[%s] error: %s", name, err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("remove kube config[%s] finished", name)

	return &engine.TaskResult{}
}

func toKubeConfigMap(config *engine.Config) *coreV1.ConfigMap {
	return &coreV1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.Name,
			Labels: config.Labels,
		},
		Data: config.Data,
	}
}
