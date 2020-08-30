package native

import (
	"os"
	"strings"

	engine "github.com/jimi36/app-engine"
)

func Store(store engine.Store) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.store = store
		return nil
	}
}

func BasePath(basePath string) engine.Option {
	return func(cli engine.ClientImpl) error {
		c, ok := cli.(*Client)
		if !ok {
			return engine.ErrOptionInvalid
		}
		c.basePath = basePath
		return nil
	}
}

func EnvVariable(key, value string) engine.Option {
	return func(cli engine.ClientImpl) error {
		if _, ok := cli.(*Client); !ok {
			return engine.ErrOptionInvalid
		}

		if len(key) == 0 || len(value) == 0 {
			return nil
		}

		if key == "PATH" {
			value = strings.Join([]string{os.Getenv("PATH"), value}, ";")
		}
		os.Setenv(key, value)

		return nil
	}
}

func applyOptions(cli *Client, opts []engine.Option) error {
	for _, opt := range opts {
		if err := opt(cli); err != nil {
			return err
		}
	}

	return nil
}
