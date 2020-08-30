package kube

import (
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/pkg/errors"
)

type instanceEvent struct {
	name    string
	version string
	evType  watch.EventType
	err     string
}

func (cli *Client) monitorInstance() {
	deployWatcher, err := cli.kubeCli.AppsV1().Deployments(cli.ns).Watch(metaV1.ListOptions{})
	if err != nil {
		log.Warnf("monitor kube application instance error: %s", err.Error())
		return
	}

	for {
		select {
		case e := <-deployWatcher.ResultChan():
			if e.Type == watch.Added || e.Type == watch.Modified {
				if deploy, ok := e.Object.(*appV1.Deployment); ok {
					if deploy.Status.AvailableReplicas > 0 {
						name, _ := deploy.Labels[labelEdgeApp]
						version, _ := deploy.Labels[labelEdgeAppVersion]
						if len(name) != 0 && len(version) != 0 {
							cli.postTaskEvent(&engine.ApplicationTag{name, version}, cli.markApplicationStarted, false)
						}
					}
				}
			} else if e.Type == watch.Deleted || e.Type == watch.Error {
				if deploy, ok := e.Object.(*appV1.Deployment); ok {
					name, _ := deploy.Labels[labelEdgeApp]
					version, _ := deploy.Labels[labelEdgeAppVersion]
					if len(name) != 0 && len(version) != 0 {
						cli.postTaskEvent(&engine.ApplicationTag{name, version}, cli.markApplicationStopped, false)
					}
				}
			}
		}
	}
}

func (cli *Client) markApplicationStarted(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("mark kube application started error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("mark kube application[%s] started......", tag.Tag())

	err := cli.store.UpdateApplicationRuntime(tag.Name, func(rt *engine.ApplicationRuntime) error {
		if rt.IsStarted || rt.Version != tag.Version {
			return errors.New("version not match")
		}
		rt.IsStarted = true
		rt.Err = ""
		return nil
	})
	if err != nil {
		log.Fatalf("mark kube application started: %s", err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("mark kube application[%s] started finished: %s", tag.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) markApplicationStopped(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("mark kube application stopped error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("mark kube application[%s] stopped......", tag.Tag())

	err := cli.store.UpdateApplicationRuntime(tag.Name, func(runtime *engine.ApplicationRuntime) error {
		if runtime.Version != tag.Version {
			return errors.New("version not match")
		}
		runtime.IsStarted = false
		return nil
	})
	if err != nil {
		log.Fatalf("mark kube application stopped: %s", err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	if err := cli.deleteService(tag.Name); err != nil {
		log.Warnf("mark kube application[%s] stopped error: %s", tag.Tag(), err.Error())
	}
	if err := cli.kubeCli.AppsV1().Deployments(cli.ns).Delete(tag.Name, nil); err != nil {
		log.Warnf("mark kube application[%s] stopped error: %s", tag.Tag(), err.Error())
	}

	log.Debugf("mark kube application[%s] stopped finished: %s", tag.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) getInstanceStates(name string) []engine.InstanceState {
	var insStates []engine.InstanceState

	podCli := cli.kubeCli.CoreV1().Pods(cli.ns)
	deployCli := cli.kubeCli.AppsV1().Deployments(cli.ns)
	podMetricsCli := cli.metricsCli.MetricsV1beta1().PodMetricses(cli.ns)

	deploy, err := deployCli.Get(name, metaV1.GetOptions{})
	if err != nil {
		log.Warnf("get kube application[%s] instance state error: %s", name, err.Error())
		return insStates
	}

	selector := labels.SelectorFromSet(deploy.Spec.Selector.MatchLabels)
	pods, err := podCli.List(metaV1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Warnf("get kube application[%s] instance state error: %s", name, err.Error())
		return insStates
	}

	for _, pod := range pods.Items {
		podMetric, err := podMetricsCli.Get(pod.Name, metaV1.GetOptions{})
		if err != nil {
			return insStates
		}
		if len(podMetric.Containers) < 1 {
			insStates = append(insStates, engine.InstanceState{
				Name:    pod.Name,
				Running: pod.Status.Phase == coreV1.PodRunning,
			})
		} else {
			cpu, _ := podMetric.Containers[0].Usage.Cpu().AsInt64()
			mem, _ := podMetric.Containers[0].Usage.Memory().AsInt64()
			insStates = append(insStates, engine.InstanceState{
				Name:    pod.Name,
				Running: pod.Status.Phase == coreV1.PodRunning,
				Cpu:     cpu / 100,  // %
				Mem:     mem / 1024, // kb
			})
		}
	}

	return insStates
}
