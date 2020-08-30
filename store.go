package engine

import (
	"errors"
)

var (
	ErrStoreAppNoFound        = errors.New("store application no found")
	ErrStoreAppExisted        = errors.New("store application Existed")
	ErrStoreAppRuntimeExisted = errors.New("store application runtime existed")
	ErrStoreAppRuntimeNoFound = errors.New("store application runtime no found")
	ErrStoreConfigNoFound     = errors.New("store config no found")
)

type Store interface {
	AddApplication(*Application) error
	RemoveApplication(*ApplicationTag) error
	UpdateApplication(*ApplicationTag, func(*Application)) error
	HasApplication(*ApplicationTag) (bool, error)
	GetApplication(*ApplicationTag) (*Application, error)
	ListApplications(int, string) ([]*ApplicationTag, string, error)

	AddApplicationRunTime(*ApplicationRuntime) error
	RemoveApplicationRunTime(string) error
	UpdateApplicationRuntime(string, func(*ApplicationRuntime) error) error
	GetApplicationRuntime(string) (*ApplicationRuntime, error)
	ListApplicationRunTimes(int, string) ([]*ApplicationRuntime, string, error)
	ForeachApplicationRunTime(func(*ApplicationRuntime)) error

	AddConfig(*Config) error
	RemoveConfig(string) error
	HasConfig(string) (bool, error)
	GetConfig(string) (*Config, error)
}
