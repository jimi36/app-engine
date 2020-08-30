package native

import (
	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
)

func (cli *Client) CreateApplication(ev *engine.TaskEvent) *engine.TaskResult {
	app, ok := ev.In.(*engine.Application)
	if !ok {
		log.Fatalf("create native application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("create native application[%s]......", app.Tag())

	if err := cli.store.AddApplication(app); err != nil {
		log.Warnf("create native application[%s] error: %s", app.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("create native application[%s] finished", app.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) RemoveApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("remove native application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("remove native application[%s]......", tag.Tag())

	if ins, found := cli.appInstances[tag.Name]; found && ins.Version == tag.Version {
		// stop application instance
		if err := ins.Stop(); err != nil {
			log.Warnf("remove native application[%s] error: %s", tag.Tag(), err.Error())
			return &engine.TaskResult{
				Err: err,
			}
		}
		// remove application runtime
		cli.store.RemoveApplicationRunTime(tag.Name)
	}

	if err := cli.store.RemoveApplication(tag); err != nil {
		log.Warnf("remove native application[%s] error: %s", tag.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("remove native application[%s] finished", tag.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) ListApplications(ev *engine.TaskEvent) *engine.TaskResult {
	opt, ok := ev.In.(*engine.ListApplicationOption)
	if !ok {
		log.Fatalf("list native applications error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("list native applications......")

	tags, _, err := cli.store.ListApplications(opt.Size, opt.LastPos)
	if err != nil {
		log.Warnf("list native applications error: %s", err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("list native applications finished")

	return &engine.TaskResult{
		Out: tags,
	}
}

func (cli *Client) GetApplicationStates(ev *engine.TaskEvent) *engine.TaskResult {
	tags, ok := ev.In.([]*engine.ApplicationTag)
	if !ok {
		log.Fatalf("get native application states error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("get native application states......")

	var appStates []*engine.ApplicationState
	for _, tag := range tags {
		state := &engine.ApplicationState{
			Name:      tag.Name,
			Version:   tag.Version,
			ToStart:   false,
			IsStarted: false,
		}

		if rt, _ := cli.store.GetApplicationRuntime(tag.Name); rt != nil && rt.Version == tag.Version {
			state.Version = rt.Version
			state.ToStart = rt.ToStart
			state.IsStarted = rt.IsStarted
			state.Err = rt.Err
		}

		if ins, found := cli.appInstances[tag.Name]; found && ins.Version == tag.Version {
			if insState, _ := ins.GetState(); insState != nil {
				state.Instances = append(state.Instances, *insState)
			}
		}

		appStates = append(appStates, state)
	}

	log.Debugf("get native application finished")

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
