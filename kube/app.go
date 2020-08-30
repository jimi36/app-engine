package kube

import (
	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
)

func (cli *Client) CreateApplication(ev *engine.TaskEvent) *engine.TaskResult {
	app, ok := ev.In.(*engine.Application)
	if !ok {
		log.Fatalf("create kube application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("create kube application[%s]......", app.Tag())

	if err := cli.store.AddApplication(app); err != nil {
		log.Warnf("create kube application[%s] error: %s", app.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("create kube application[%s] finished", app.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) RemoveApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("remove kube application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("remove kube application[%s]......", tag.Tag())

	// check application runtime
	if rt, _ := cli.store.GetApplicationRuntime(tag.Name); rt != nil && rt.Version == tag.Version {
		// stop application instance
		cli.StopApplication(&engine.TaskEvent{In: tag})
		// remove application runtime
		cli.store.RemoveApplicationRunTime(tag.Name)
	}

	// remove application
	if err := cli.store.RemoveApplication(tag); err != nil {
		log.Warnf("remove kube application[%s] error: %s", tag.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("remove kube application[%s] finished", tag.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) ListApplications(ev *engine.TaskEvent) *engine.TaskResult {
	opt, ok := ev.In.(*engine.ListApplicationOption)
	if !ok {
		log.Fatalf("list kube applications error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("list kube applications......")

	tags, _, err := cli.store.ListApplications(opt.Size, opt.LastPos)
	if err != nil {
		log.Warnf("list kube applications error: %s", err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("list kube applications finished")

	return &engine.TaskResult{
		Out: tags,
	}
}

func (cli *Client) GetApplicationStates(ev *engine.TaskEvent) *engine.TaskResult {
	tags, ok := ev.In.([]*engine.ApplicationTag)
	if !ok {
		log.Fatalf("get kube application states error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("get kube application states......")

	var appStates []*engine.ApplicationState
	for _, tag := range tags {
		state := &engine.ApplicationState{
			Name:      tag.Name,
			Version:   tag.Version,
			ToStart:   false,
			IsStarted: false,
		}

		if rt, err := cli.store.GetApplicationRuntime(tag.Name); err == nil && rt.Version == tag.Version {
			state.ToStart = rt.ToStart
			state.IsStarted = rt.IsStarted
			state.Err = rt.Err
		}

		if state.IsStarted {
			state.Instances = cli.getInstanceStates(tag.Name)
		}

		appStates = append(appStates, state)
	}

	log.Debugf("get kube application states finished")

	return &engine.TaskResult{
		Out: appStates,
	}
}

func (cli *Client) GetStartedApplications(ev *engine.TaskEvent) *engine.TaskResult {
	var rts []*engine.ApplicationRuntime
	cli.store.ForeachApplicationRunTime(func(rt *engine.ApplicationRuntime) {
		if rt.ToStart {
			rts = append(rts, rt)
		}
	})
	return &engine.TaskResult{
		Out: rts,
	}
}
