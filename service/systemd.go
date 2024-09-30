package service

import (
	"fmt"
	"os"
	"os/exec"
)

type Systemd struct {
	commandName CommandName
}

func NewSystemd(commandName CommandName) *Systemd {
	return &Systemd{commandName: commandName}
}

func (j *Systemd) GetServiceName() string {
	return string(j.commandName) + ".service"
}

func (j *Systemd) Log() error {
	fmt.Printf("Streaming logs from systemd %s\n", j.GetServiceName())

	cmd := exec.Command("journalctl", "-f", "-u", j.GetServiceName())

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (j *Systemd) Start() error {
	cmd := exec.Command("systemctl", "start", j.GetServiceName())
	return cmd.Run()
}

func (j *Systemd) Stop() error {
	cmd := exec.Command("systemctl", "stop", j.GetServiceName())
	return cmd.Run()
}

func (j *Systemd) Restart() error {
	cmd := exec.Command("systemctl", "restart", j.GetServiceName())
	return cmd.Run()
}
