package native

import (
	"os"
	"syscall"
)

func StartProcess(cmd string, args, env []string, stdOut, stdErr *os.File) (*os.Process, error) {
	proc, err := os.StartProcess(cmd, args, &os.ProcAttr{
		Files: []*os.File{
			nil, stdOut, stdErr,
		},
		Env: env,
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	})
	if err != nil {
		return nil, err
	}
	return proc, nil
}
