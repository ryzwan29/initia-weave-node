package service

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

type Service interface {
	Create(binaryVersion, appHome string) error
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

func NonDaemonStart(s Service) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalChan)

	go func() {
		err := s.Start()
		if err != nil {
			_ = s.Stop()
			panic(err)
		}
		_ = s.Log(100)
	}()

	<-signalChan
	return s.Stop()
}
