package engine

import "errors"

var (
	ErrTimeout               = errors.New("timeout")
	ErrNotImplement          = errors.New("not implement")
	ErrOptionInvalid         = errors.New("option invalid")
	ErrParamInvalid          = errors.New("param invalid")
	ErrClientNotStarted      = errors.New("cleint is not started")
	ErrApplicationExisted    = errors.New("application is existed")
	ErrApplicationNoExisted  = errors.New("application is not existed")
	ErrApplicationStarted    = errors.New("application is started")
	ErrApplicationNotStarted = errors.New("application is not started")
	ErrConfigExisted         = errors.New("config is existed")
	ErrConfigNoExisted       = errors.New("config is not existed")

	ErrTaskEventInvalid  = errors.New("task event invalid")
	ErrTaskResultInvalid = errors.New("task result invalid")
)
