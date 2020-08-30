package native

import (
	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/pkg/errors"
)

func (cli *Client) StopApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("stop native application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("stop native application[%s]......", tag.Tag())

	// check and get appliation instance
	ins, found := cli.appInstances[tag.Name]
	if !found {
		log.Warnf("stop native application[%s] error: %s", tag.Tag(), engine.ErrApplicationNotStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationNotStarted,
		}
	}

	// check application instance version
	if ins.Version != tag.Version {
		log.Warnf("stop native application[%s] error: version not match", tag.Tag())
		return &engine.TaskResult{
			Err: engine.ErrApplicationNotStarted,
		}
	}

	// stop appliaction instance
	if err := ins.Stop(); err != nil {
		log.Warnf("stop native application[%s] error: %s", tag.Tag(), err.Error())
	}

	// remove application runtime
	cli.store.RemoveApplicationRunTime(tag.Name)

	log.Debugf("stop native application[%s] finished", tag.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) monitorInstance(ins *Instance) {
	tag := &engine.ApplicationTag{ins.Name, ins.Version}
	go func() {
		select {
		case <-ins.Done():
			if _, err := cli.postTaskEvent(tag, cli.cleanStartedApplicationInfo, false); err != nil {
				log.Fatalf("notify native application[%s] error: %s", tag.Tag(), err.Error())
			}
		}
	}()
}

func (cli *Client) cleanStartedApplicationInfo(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("clean native started application info error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("clean native started application[%s] info......", tag.Tag())

	// update application runtime
	cli.store.UpdateApplicationRuntime(tag.Name, func(rt *engine.ApplicationRuntime) error {
		if rt.Version != tag.Version {
			return errors.New("version not match")
		}
		rt.IsStarted = false
		rt.Pid = -1
		return nil
	})

	delete(cli.appInstances, tag.Name)

	log.Debugf("clean native started application[%s] info finished", tag.Tag())

	return &engine.TaskResult{}
}
