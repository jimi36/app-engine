package engine

import (
	"time"

	"github.com/jimi36/app-engine/log"
)

const (
	TaskHandleTimeout = time.Second * 100
)

type Option func(ClientImpl) error

type ClientImpl interface {
	Init(postFunc PostTaskEventFunc) error

	CreateApplication(*TaskEvent) *TaskResult
	RemoveApplication(*TaskEvent) *TaskResult
	StartApplication(*TaskEvent) *TaskResult
	RestartApplication(*TaskEvent) *TaskResult
	StopApplication(*TaskEvent) *TaskResult
	ListApplications(*TaskEvent) *TaskResult
	GetApplicationStates(*TaskEvent) *TaskResult
	GetStartedApplications(*TaskEvent) *TaskResult

	CreateConfig(*TaskEvent) *TaskResult
	RemoveConfig(*TaskEvent) *TaskResult
}

func NewClient(impl ClientImpl) *Client {
	c := &Client{
		impl:    impl,
		eventCh: make(chan *TaskEvent, 1024),
	}
	return c
}

type Client struct {
	// client impl
	impl ClientImpl
	// task event channel
	eventCh chan *TaskEvent
}

func (c *Client) Start() error {
	c.impl.Init(c.postTaskEvent)

	go c.appEventLoop()

	ret := c.impl.GetStartedApplications(nil)
	rts := ret.Out.([]*ApplicationRuntime)
	for _, rt := range rts {
		c.restartApplication(&ApplicationTag{rt.Name, rt.Version})
	}

	return nil
}

func (c *Client) CreateApplication(app *Application) error {
	log.Debugf("create application[%s]......", app.Tag())

	rc, err := c.postTaskEvent(app, c.impl.CreateApplication, true)
	if err != nil {
		log.Warnf("create application[%s] error: %s", app.Tag(), err.Error())
		return err
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("create application[%s] timeout", app.Tag())
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("create application[%s] error: %s", app.Tag(), ret.Err.Error())
		return ret.Err
	}

	log.Infof("create application[%s] finished", app.Tag())

	return nil
}

func (c *Client) RemoveApplication(tag *ApplicationTag) error {
	log.Debugf("remove application[%s]......", tag.Tag())

	// stop application
	err := c.StopApplication(tag)
	if err != nil && err != ErrApplicationNotStarted {
		log.Warnf("remove application[%s] error: %s", tag.Tag(), err.Error())
		return err
	}

	rc, err := c.postTaskEvent(tag, c.impl.RemoveApplication, true)
	if err != nil {
		log.Warnf("remove application[%s] error: %s", tag.Tag(), err.Error())
		return err
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("remove application[%s] timeout", tag.Tag())
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("remove application[%s] error: %s", tag.Tag(), ret.Err.Error())
		return ret.Err
	}

	log.Infof("remove application[%s] finished", tag.Tag())

	return nil
}

func (c *Client) restartApplication(tag *ApplicationTag) error {
	log.Debugf("restart application[%s]......", tag.Tag())

	if len(tag.Name) == 0 || len(tag.Version) == 0 {
		log.Warnf("restart application[%s] error: %s", tag.Tag(), ErrParamInvalid.Error())
		return ErrParamInvalid
	}

	// create start application task event
	rc, err := c.postTaskEvent(tag, c.impl.RestartApplication, true)
	if err != nil {
		log.Fatalf("restart application[%s] error: %s", tag.Tag(), err.Error())
		return err
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("restart application[%s] timeout", tag.Tag())
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("restart application[%s] error: %s", tag.Tag(), ret.Err.Error())
		return ret.Err
	}

	log.Debugf("restart application[%s] finished", tag.Tag())

	return nil
}

func (c *Client) StartApplication(tag *ApplicationTag) error {
	log.Debugf("start application[%s]......", tag.Tag())

	if len(tag.Name) == 0 || len(tag.Version) == 0 {
		log.Warnf("start application[%s] error: %s", tag.Tag(), ErrParamInvalid.Error())
		return ErrParamInvalid
	}

	// create start application task event
	rc, err := c.postTaskEvent(tag, c.impl.StartApplication, true)
	if err != nil {
		log.Fatalf("start application[%s] error: %s", tag.Tag(), err.Error())
		return err
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("start application[%s] timeout", tag.Tag())
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("start application[%s] error: %s", tag.Tag(), ret.Err.Error())
		return ret.Err
	}

	log.Debugf("start application[%s] finished", tag.Tag())

	return nil
}

func (c *Client) StopApplication(tag *ApplicationTag) error {
	log.Debugf("stop application[%s]......", tag.Tag())

	rc, err := c.postTaskEvent(tag, c.impl.StopApplication, true)
	if err != nil {
		log.Fatalf("stop application[%s] error: %s", tag.Tag(), err.Error())
		return err
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("stop application[%s] timeout", tag.Tag())
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("stop application[%s] error: %s", tag.Tag(), ret.Err.Error())
		return ret.Err
	}

	log.Infof("stop application[%s] finished", tag.Tag())

	return nil
}

func (c *Client) ListApplications(opt *ListApplicationOption) ([]*ApplicationTag, error) {
	log.Debugf("list applications......")

	rc, err := c.postTaskEvent(opt, c.impl.ListApplications, true)
	if err != nil {
		log.Warnf("list application error: %s", err.Error())
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("list applications timeout")
		return nil, ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("list applications error: %s", ret.Err.Error())
		return nil, ret.Err
	}

	out, ok := ret.Out.([]*ApplicationTag)
	if !ok {
		log.Warnf("list applications error: %s", ErrTaskResultInvalid.Error())
		return nil, ErrTaskResultInvalid
	}

	return out, nil
}

func (c *Client) GetApplicationStates(tags []*ApplicationTag) ([]*ApplicationState, error) {
	log.Debugf("get application states......")

	rc, err := c.postTaskEvent(tags, c.impl.GetApplicationStates, true)
	if err != nil {
		log.Warnf("get application states error: %s", err.Error())
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("get application states timeout")
		return nil, ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("get application states error: %s", ret.Err.Error())
		return nil, ret.Err
	}

	out, ok := ret.Out.([]*ApplicationState)
	if !ok {
		log.Warnf("get application states error: %s", ErrTaskResultInvalid.Error())
		return nil, ErrTaskResultInvalid
	}

	return out, nil
}

func (c *Client) CreateConfig(config *Config) error {
	log.Debugf("create config[%s]......", config.Name)

	rc, err := c.postTaskEvent(config, c.impl.CreateConfig, true)
	if err != nil {
		log.Warnf("create config[%s] error: %s", config.Name, err.Error())
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("create config[%s] timeout", config.Name)
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("create config[%s] error: %s", config.Name, ret.Err.Error())
		return ret.Err
	}

	log.Infof("create config[%s] finished", config.Name)

	return nil
}

func (c *Client) RemoveConfig(name string) error {
	log.Debugf("remove config[%s]......", name)

	rc, err := c.postTaskEvent(name, c.impl.RemoveConfig, true)
	if err != nil {
		log.Warnf("remove config[%s] error: %s", name, err.Error())
	}

	var ret *TaskResult
	select {
	case <-time.After(TaskHandleTimeout):
		log.Warnf("remove config[%s] timeout", name)
		return ErrTimeout
	case ret = <-rc:
	}

	if ret.Err != nil {
		log.Warnf("remove config[%s] error: %s", name, ret.Err.Error())
		return ret.Err
	}

	log.Infof("remove config[%s] finished", name)

	return nil
}
