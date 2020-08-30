package native

import (
	"fmt"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/jimi36/app-engine/utils"
)

func (cli *Client) CreateConfig(ev *engine.TaskEvent) *engine.TaskResult {
	config, ok := ev.In.(*engine.Config)
	if !ok {
		log.Fatalf("create native config error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("create native config[%s]......", config.Name)

	if err := cli.store.AddConfig(config); err != nil {
		log.Warnf("create native config[%s] error: %s", config.Name, err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	configPath := genConfigPath(cli.basePath, config.Name)
	if !utils.IsExistedPath(configPath) {
		if err := utils.CreateFolder(configPath); err != nil {
			log.Warnf("create native config[%s] folder error: %s", config.Name, err.Error())
			return &engine.TaskResult{
				Err: err,
			}
		}
	}

	for k, v := range config.Data {
		filePath := genConfigFilePath(configPath, k)
		if err := utils.CreateFile(filePath, []byte(v)); err != nil {
			log.Warnf("create native config[%s] error: %s", config.Name, err.Error())
			return &engine.TaskResult{
				Err: err,
			}
		}
	}

	log.Debugf("create native config[%s] finished", config.Name)

	return &engine.TaskResult{}
}

func (cli *Client) RemoveConfig(ev *engine.TaskEvent) *engine.TaskResult {
	name, ok := ev.In.(string)
	if !ok {
		log.Fatalf("remove native config error: %s", engine.ErrTaskEventInvalid.Error())
		return &engine.TaskResult{
			Err: engine.ErrTaskEventInvalid,
		}
	}

	log.Debugf("remove native config[%s]......", name)

	configPath := genConfigPath(cli.basePath, name)
	if !utils.IsExistedPath(configPath) {
		log.Warnf("remove native config[%s] error: %s", name, engine.ErrConfigNoExisted.Error())
		return &engine.TaskResult{
			Err: engine.ErrConfigNoExisted,
		}
	}

	if err := utils.RemoveFolder(configPath); err != nil {
		log.Warnf("remove native config[%s] error: %s", name, err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	if err := cli.store.RemoveConfig(name); err != nil {
		log.Warnf("remove native config[%s] error: %s", name, err.Error())
		return &engine.TaskResult{
			Err: err,
		}
	}

	log.Debugf("remove native config[%s] finished", name)

	return &engine.TaskResult{}
}

func genConfigPath(basePath, name string) string {
	return fmt.Sprintf("%s/config/%s", basePath, name)
}

func genConfigFilePath(configPath, name string) string {
	return fmt.Sprintf("%s/%s", configPath, name)
}
