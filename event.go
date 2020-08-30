package engine

func (cli *Client) appEventLoop() {
	for {
		select {
		case ev := <-cli.eventCh:
			ret := ev.Handler(ev)
			if ev.Rc != nil {
				ev.Rc <- ret
			}
		}
	}
}

func (cli *Client) postTaskEvent(in interface{}, h TaskHandler, waitRet bool) (chan *TaskResult, error) {
	tv := &TaskEvent{
		In:      in,
		Handler: h,
	}
	if waitRet {
		tv.Rc = make(chan *TaskResult, 1)
	}

	select {
	case cli.eventCh <- tv:
	}

	return tv.Rc, nil
}

type TaskEvent struct {
	In      interface{}
	Rc      chan *TaskResult
	Handler TaskHandler
}

type TaskResult struct {
	Out interface{}
	Err error
}

type TaskHandler func(*TaskEvent) *TaskResult
type PostTaskEventFunc func(interface{}, TaskHandler, bool) (chan *TaskResult, error)
