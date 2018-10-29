package cmds

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
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/DCSO/slinkwatch"

	"github.com/satta/ifplugo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func runMain(cmd *cobra.Command, args []string) {
	ifaceFile, err := cmd.Flags().GetString("interfaces")
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.ParseFiles(ifaceFile)
	log.Infof("Parsing template file %s", ifaceFile)
	if err != nil {
		log.Fatalf("Invalid template: %s", err.Error())
	}
	log.Infof("Parsing of %s done, success", ifaceFile)

	var conf slinkwatch.Config
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Loading config file %s", configFile)
	err = conf.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %s", err.Error())
	}
	log.Infof("Config file %s loaded successfully", configFile)
	log.Infof("Watching interfaces %s", strings.Join(conf.GetWatchedInterfaces(), ","))

	restartCmd, err := cmd.Flags().GetString("restart-command")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Suricata restart command set to '%s'", restartCmd)
	serviceName, err := cmd.Flags().GetString("service-name")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Suricata service name command set to '%s'", restartCmd)

	restartMan, err := slinkwatch.MakeRestartManager(restartCmd, serviceName)
	if err != nil {
		log.Fatal(err)
	}

	targetFile, err := cmd.Flags().GetString("target-file")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Target file set to %s", targetFile)

	pollInterval, err := cmd.Flags().GetDuration("poll-interval")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Poll interval set to %v", pollInterval)

	deltaBytes, err := cmd.Flags().GetUint("delta-bytes")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Delta bytes set to %v", deltaBytes)

	activeSet := slinkwatch.MakeActiveSet(&conf)
	notices := make(chan ifplugo.LinkStatusSample, 10)
	mon := ifplugo.MakeLinkStatusMonitor(pollInterval,
		conf.GetWatchedInterfaces(), notices)
	if deltaBytes > 0 {
		mon.CheckIncomingDelta(true, uint64(deltaBytes))
	}
	mon.Run()

	go func() {
		for n := range notices {
			changed := false
			activeSet.Reset()
			for i, v := range n.Ifaces {
				if v == ifplugo.InterfaceUp {
					activeSet.Add(i)
					changed = true
				}
			}
			out, err := activeSet.ToYAML(tmpl, &conf)
			if err != nil {
				log.Fatalf("Error rendering to YAML: %s", err.Error())
			}
			if changed {
				log.Infof("interface state change detected, active set is %s",
					activeSet)
			} else {
				log.Infof("interface state change detected, active set is empty")
			}
			err = ioutil.WriteFile(targetFile, []byte(out), 0644)
			if err != nil {
				log.Println(err)
			} else {
				err = restartMan.RestartSuricata()
				if err != nil {
					log.Println(err)
				} else {
					log.Println("restart requested successfully")
				}
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for sig := range c {
			if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				log.Info("received SIGTERM, terminating")
				close(notices)
				os.Exit(1)
			}
		}
	}()

	select {}
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the slinkwatch service",
	Run: func(cmd *cobra.Command, args []string) {
		runMain(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("config", "c", "config.yaml", "Configuration file")
	runCmd.Flags().StringP("interfaces", "i", "interfaces.tmpl", "Template file for interfaces")
	runCmd.Flags().StringP("target-file", "t", "/etc/suricata/interfaces.yaml", "Target YAML file with interface information")
	runCmd.Flags().StringP("restart-command", "r", "/etc/init.d/suricata restart", "Suricata restart command")
	runCmd.Flags().DurationP("poll-interval", "p", 5*time.Second, "poll time for interface changes")
	runCmd.Flags().StringP("service-name", "s", "suricata.service", "systemd service name for Suricata service")
	runCmd.Flags().UintP("delta-bytes", "d", 100, "threshold of bytes to be exceeded on interface to be marked as up")
}
