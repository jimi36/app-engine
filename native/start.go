package native

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/jimi36/app-engine/utils"
)

func (cli *Client) RestartApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("restart native application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("restart native application[%s]......", tag.Tag())

	// check application instance
	if _, found := cli.appInstances[tag.Name]; found {
		log.Warnf("start native application[%s] error: %s", tag.Tag(), engine.ErrApplicationStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationStarted,
		}
	}

	ret := &engine.TaskResult{}
	if rt, _ := cli.store.GetApplicationRuntime(tag.Name); rt != nil {
		ins, _ := CreateInstance(tag.Name, tag.Version, cli.basePath)
		if err := ins.Bind(rt.Pid); err == nil {
			cli.store.UpdateApplicationRuntime(tag.Name, func(runtime *engine.ApplicationRuntime) error {
				runtime.ToStart = true
				runtime.IsStarted = true
				runtime.Err = ""
				return nil
			})
			cli.appInstances[rt.Name] = ins
			cli.monitorInstance(ins)
		} else {
			ret = cli.StartApplication(ev)
		}
	} else {
		ret = cli.StartApplication(ev)
	}

	log.Debugf("restart native application[%s] finished", tag.Tag())

	return ret
}

func (cli *Client) StartApplication(ev *engine.TaskEvent) *engine.TaskResult {
	tag, ok := ev.In.(*engine.ApplicationTag)
	if !ok {
		log.Fatalf("start native application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("start native application[%s]......", tag.Tag())

	// check and get application
	app, err := cli.store.GetApplication(tag)
	if err != nil {
		log.Warnf("start native application[%s] error: %s", tag.Tag(), err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	// check application instance
	if _, found := cli.appInstances[tag.Name]; found {
		log.Warnf("start native application[%s] error: %s", tag.Tag(), engine.ErrApplicationStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationStarted,
		}
	}

	// get application runtime, but maybe not existed
	rt, err := cli.store.GetApplicationRuntime(tag.Name)
	if err != nil && err != engine.ErrStoreAppRuntimeNoFound {
		log.Warnf("start native application[%s] error: %s", tag.Tag(), engine.ErrApplicationStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationStarted,
		}
	}

	if rt != nil {
		// if existed, update application runtime
		cli.store.UpdateApplicationRuntime(tag.Name, func(runtime *engine.ApplicationRuntime) error {
			runtime.Version = tag.Version
			runtime.ToStart = true
			runtime.IsStarted = false
			runtime.Pid = -1
			runtime.Err = ""
			return nil
		})
	} else {
		// if not existed, add application runtime
		cli.store.AddApplicationRunTime(&engine.ApplicationRuntime{
			ApplicationTag: *tag,
			ToStart:        true,
			IsStarted:      false,
			Pid:            -1,
		})
	}

	// create application instance
	ins, err := CreateInstance(app.Name, app.Version, cli.basePath)
	if err != nil {
		log.Warnf("start native application[%s] error: %s", tag.Tag(), err.Error())
		// update application runtime with error
		cli.store.UpdateApplicationRuntime(tag.Name, func(rt *engine.ApplicationRuntime) error {
			rt.ToStart = false
			rt.Err = "create instance error"
			return nil
		})
		return &engine.TaskResult{
			Err: err,
		}
	}
	cli.appInstances[app.Name] = ins
	cli.monitorInstance(ins)

	// download application
	go cli.downloadApplication(ins.Context(), app)

	log.Debugf("start native application[%s] finished", tag.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) downloadApplication(ctx context.Context, app *engine.Application) {
	log.Debugf("download native application[%s]......", app.Tag())

	// download application resource
	rc := app.NativeSpec.Rc
	if rc != nil {
		fileUrl := app.NativeSpec.Rc.Url
		fileMd5 := app.NativeSpec.Rc.Md5
		filePath := filepath.Join(cli.basePath, app.Name, app.Version, app.NativeSpec.Rc.FileName)
		if err := httpDownloadFile(ctx, fileUrl, filePath, fileMd5); err != nil {
			log.Warnf("download native application[%s] error: %s", app.Tag(), err.Error())
			if _, err := cli.postTaskEvent(app, cli.downloadApplicationFailed, false); err != nil {
				log.Fatalf("download native application[%s] error: %s", app.Tag(), err.Error())
			}
			return
		}
	}

	if _, err := cli.postTaskEvent(app, cli.runApplication, false); err != nil {
		log.Fatalf("download native application[%s] error: %s", app.Tag(), err.Error())
		return
	}

	log.Debugf("download native application[%s] finished", app.Tag())
}

func (cli *Client) downloadApplicationFailed(ev *engine.TaskEvent) *engine.TaskResult {
	app, ok := ev.In.(*engine.Application)
	if !ok {
		log.Fatalf("download native application failed error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("download native application[%s] failed......", app.Tag())

	// check and get application instance
	ins, found := cli.appInstances[app.Name]
	if found {
		log.Warnf("download native application[%s] failed error: %s", app.Tag(), engine.ErrApplicationNotStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationNotStarted,
		}
	}

	// update application runtime with error
	cli.store.UpdateApplicationRuntime(app.Name, func(rt *engine.ApplicationRuntime) error {
		rt.ToStart = false
		rt.Err = "download application failed"
		return nil
	})

	if err := ins.Stop(); err != nil {
		log.Warnf("download native application[%s] failed error: %s", app.Tag(), err.Error())
	}

	log.Debugf("download native application[%s] failed finished", app.Tag())

	return &engine.TaskResult{}
}

func (cli *Client) runApplication(ev *engine.TaskEvent) *engine.TaskResult {
	app, ok := ev.In.(*engine.Application)
	if !ok {
		log.Fatalf("run native application error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("run native application[%s]......", app.Tag())

	// check and get application instance
	ins, found := cli.appInstances[app.Name]
	if !found {
		log.Warnf("run native application[%s] error: %s", app.Tag(), engine.ErrApplicationStarted.Error())
		return &engine.TaskResult{
			Err: engine.ErrApplicationStarted,
		}
	}

	// start application instance
	if err := ins.Start(app); err != nil {
		log.Warnf("run native application[%s] error: %s", app.Tag(), err.Error())
		// update application runtime with error
		cli.store.UpdateApplicationRuntime(app.Name, func(rt *engine.ApplicationRuntime) error {
			rt.ToStart = false
			rt.Err = "start instance err: " + err.Error()
			return nil
		})
		if err1 := ins.Stop(); err1 != nil {
			log.Warnf("run native application[%s] error: %s", app.Tag(), err1.Error())
		}
		return &engine.TaskResult{
			Err: err,
		}
	}

	// update application runtime
	cli.store.UpdateApplicationRuntime(app.Name, func(rt *engine.ApplicationRuntime) error {
		rt.IsStarted = true
		rt.Pid = ins.Pid()
		return nil
	})

	log.Debugf("run native application[%s] finished", app.Tag())

	return &engine.TaskResult{}
}

func fileMD5(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	sum := md5.Sum(data)

	return hex.EncodeToString(sum[:]), nil
}

func httpDownloadFile(ctx context.Context, url, filePath, md5 string) error {
	if utils.IsExistedPath(filePath) {
		if fileMd5, err := fileMD5(filePath); err == nil && fileMd5 == md5 {
			return nil
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}

	if fileMd5, _ := fileMD5(filePath); fileMd5 != md5 {
		return errors.New("file md5 not match")
	}

	return nil
}
