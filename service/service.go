package service

import (
	"fmt"
	"runtime"
)

type CommandName string

const (
	Initia  CommandName = "initia"
	Minitia CommandName = "minitia"
)

type Service interface {
	Create(binaryVersion string) error
	Log(n int) error
	Start() error
	Stop() error
	Restart() error
}

func NewService(commandName CommandName) (Service, error) {
	switch runtime.GOOS {
	case "linux":
		return NewSystemd(commandName), nil
	case "darwin":
		return NewLaunchd(commandName), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
