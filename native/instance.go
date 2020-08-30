package native

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/jimi36/app-engine/utils"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type Instance struct {
	// app name
	Name string
	// app version
	Version string
	// root path
	basePath string
	// process
	proc *process.Process
	//stopped chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func CreateInstance(name, version, basePath string) (*Instance, error) {
	// create instance folder
	appFolder := filepath.Join(basePath, name, version)
	if !utils.IsExistedPath(appFolder) {
		if err := utils.CreateFolder(appFolder); err != nil {
			log.Debugf("create instance[%s:%s] warn: %s", name, version, err.Error())
			return nil, errors.New("create instance folder error")
		}
	}

	ins := &Instance{
		Name:     name,
		Version:  version,
		basePath: basePath,
	}
	ins.ctx, ins.cancel = context.WithCancel(context.Background())

	return ins, nil
}

func (ins *Instance) Done() <-chan struct{} {
	return ins.ctx.Done()
}

func (ins *Instance) Context() context.Context {
	return ins.ctx
}

func (ins *Instance) Pid() int {
	if ins.proc == nil {
		return -1
	}
	if isRunning, err := ins.proc.IsRunning(); err != nil || !isRunning {
		return -1
	}
	return int(ins.proc.Pid)
}

func (ins *Instance) Start(app *engine.Application) error {
	if ins.proc != nil {
		log.Debugf("start instance[%s] error: already started or stopped", ins.String())
		return errors.New("instance is already started or stopped")
	}

	if err := ins.startProcess(app.NativeSpec.Command, app.Env); err != nil {
		log.Debugf("start instance[%s] initCmd error: %s", ins.String(), err.Error())
		return err
	}

	go ins.monitor()

	return nil
}

func (ins *Instance) Bind(pid int) error {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}
	ins.proc = proc

	if _, err := ins.proc.IsRunning(); err != nil {
		return err
	}

	go ins.monitor()

	return nil
}

func (ins *Instance) startProcess(cmds []string, envmap map[string]string) error {
	stdOut := os.Stdout
	stdErr := os.Stderr

	// open log file
	logfile := filepath.Join(ins.basePath, ins.Name, ins.Version, "app.log")
	logWriter, err := os.Create(logfile)
	if err == nil {
		stdOut = logWriter
		stdErr = logWriter
		defer logWriter.Close()
	} else {
		log.Debugf("create instance[%s] log file error: %s", ins.String(), err.Error())
	}

	var envs []string
	for k, v := range envmap {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}

	cmdBin := filepath.Join(ins.basePath, ins.Name, ins.Version, cmds[0])
	if !utils.IsExistedPath(cmdBin) || utils.IsDir(cmdBin) {
		cmdBin = cmds[0]
	}

	osProc, err := StartProcess(cmdBin, cmds[1:], envs, stdOut, stdErr)
	if err != nil {
		return err
	}

	proc, err := process.NewProcess(int32(osProc.Pid))
	if err != nil {
		return err
	}
	ins.proc = proc

	return nil
}

func (ins *Instance) Stop() error {
	if ins.proc != nil {
		ins.proc.Kill()
	}

	return nil
}

func (ins *Instance) GetState() (*engine.InstanceState, error) {
	state := &engine.InstanceState{
		Name:    ins.Name,
		Running: false,
	}

	if ins.proc == nil {
		return state, nil
	}

	isRunning, _ := ins.proc.IsRunning()
	state.Running = isRunning

	mem, _ := ins.proc.MemoryInfo()
	if mem != nil {
		state.Mem = int64(mem.RSS)
	}

	cpu, _ := ins.proc.CPUPercent()
	state.Cpu = int64(cpu)

	return state, nil
}

func (ins *Instance) monitor() {
	isRunning := true
	for isRunning {
		select {
		case <-time.After(time.Second * 3):
			if proc, _ := os.FindProcess(int(ins.proc.Pid)); proc != nil {
				proc.Wait()
			}
		}
		isRunning, _ = process.PidExists(ins.proc.Pid)
	}

	ins.cancel()
}

func (ins *Instance) String() string {
	return strings.Join([]string{ins.Name, ins.Version}, ":")
}
