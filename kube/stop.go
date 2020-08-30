package kube

import (
	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
)

func (cli *Client) StopApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("stop kube application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("stop kube application[%s]......", tag.Tag())

	// hek and get appliation runtime
	rt, err := cli.store.GetApplicationRuntime(tag.Name)
	if err != nil {
		log.Warnf("stop kube application[%s] error: %s", tag.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	// check application runtime version
	if rt.Version != tag.Version || !rt.IsStarted {
		log.Warnf("stop kube application[%s] error: %s", tag.Tag(), engine.ErrApplicationNotStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationNotStarted,
		}
	}

	// remove application runtime
	cli.store.RemoveApplicationRunTime(tag.Name)

	if err := cli.deleteService(tag.Name); err != nil {
		log.Warnf("stop kube application[%s] error: %s", tag.Tag(), err.Error())
	}
	if err := cli.kubeCli.AppsV1().Deployments(cli.ns).Delete(tag.Name, nil); err != nil {
		log.Warnf("stop kube application[%s] error: %s", tag.Tag(), err.Error())
	}

	log.Debugf("stop kube application[%s] finished: %s", tag.Tag())

	return &engine.TaskResult{}
}
