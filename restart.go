package slinkwatch

// This file is part of slinkwatch.
//
// slinkwatch is free software; you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// slinkwatch is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with slinkwatch; if not, write to the Free Software Foundation,
// Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307 USA.

import (
	"fmt"
	"os/exec"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/util"
	log "github.com/sirupsen/logrus"
)

// RestartManager encapsulates interaction with the init system to restart the
// Suricata process on request.
type RestartManager struct {
	IsSystemd      bool
	RestartCommand string
	ServiceName    string
}

// MakeRestartManager returns a new RestartManager with restart command and
// service name preconfigured.
func MakeRestartManager(restartCommand string, serviceName string) (*RestartManager, error) {
	r := &RestartManager{
		IsSystemd:      util.IsRunningSystemd(),
		RestartCommand: restartCommand,
		ServiceName:    serviceName,
	}
	if !r.IsSystemd {
		log.Infoln("Not running systemd, will rely on restart command")
	} else {
		log.Infoln("systemd is available")
	}
	if !r.IsSystemd && len(restartCommand) == 0 {
		return nil, fmt.Errorf("Not running systemd but no restart command given")
	}
	return r, nil
}

// RestartSuricata triggers a Suricata restart, using the most appropriate
// method.
func (rm *RestartManager) RestartSuricata() error {
	if !rm.IsSystemd {
		cmd := exec.Command("sh", "-c", rm.RestartCommand)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Cannot restart Suricata: %v", err)
		}
		return nil
	}

	conn, err := dbus.New()
	if err != nil {
		return err
	}
	finishChan := make(chan string)
	_, err = conn.RestartUnit(rm.ServiceName, "replace", finishChan)
	if err != nil {
		return err
	}
	go func(ch chan string) {
		log.Infof("restart result: %s", <-ch)

	}(finishChan)
	return nil
}
