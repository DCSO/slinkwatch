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
	"fmt"
	"net"

	"github.com/ryanuber/columnize"
	"github.com/satta/ifplugo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/DCSO/slinkwatch"
)

// showactCmd represents the show-active command
var showactCmd = &cobra.Command{
	Use:   "show-active [options]",
	Short: "List active set of interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		showAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatal(err)
		}

		var interfaces []string
		if showAll {
			interfaces = make([]string, 0)
			netIfs, err := net.Interfaces()
			if err != nil {
				log.Fatal(err)
			}
			for _, i := range netIfs {
				interfaces = append(interfaces, i.Name)
			}
		} else {
			var conf slinkwatch.Config
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				log.Fatal(err)
			}
			err = conf.LoadConfig(configFile)
			if err != nil {
				log.Fatalf("Error loading config: %s", err.Error())
			}
			interfaces = conf.GetWatchedInterfaces()
		}

		output := make([]string, 0)
		for _, i := range interfaces {
			status, err := ifplugo.GetLinkStatus(i)
			if err != nil {
				continue
			}
			line := fmt.Sprintf("%s|%s\n", i, status)
			output = append(output, line)
		}
		result := columnize.SimpleFormat(output)
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(showactCmd)
	showactCmd.Flags().BoolP("all", "a", false, "Show status for all interfaces")
	showactCmd.Flags().StringP("config", "c", "config.yaml", "Configuration file")
}
