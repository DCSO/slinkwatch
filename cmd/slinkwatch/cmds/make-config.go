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
	"regexp"

	"github.com/DCSO/slinkwatch"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// confCmd represents the make-config command
var confCmd = &cobra.Command{
	Use:   "make-config [options]",
	Short: "Create initial configuration file from local interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		pattern, err := cmd.Flags().GetString("pattern")
		if err != nil {
			log.Fatal(err)
		}
		pat, err := regexp.CompilePOSIX(pattern)
		if err != nil {
			log.Fatal(err)
		}

		netIfs, err := net.Interfaces()
		consideredIfs := make([]string, 0)
		if err != nil {
			log.Fatal(err)
		}
		for _, i := range netIfs {
			if pat.Match([]byte(i.Name)) {
				consideredIfs = append(consideredIfs, i.Name)
			}
		}

		cfg := slinkwatch.Config{
			Ifaces: make(map[string]slinkwatch.ConfIface),
		}
		cid := 99
		for _, i := range consideredIfs {
			cif := slinkwatch.ConfIface{
				ThreadWeight: 1,
				ClusterID:    cid,
			}
			cid--
			cfg.Ifaces[i] = cif
		}
		out, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("# Interfaces available for Suricata")
		fmt.Println("---")
		fmt.Println(string(out[:]))
	},
}

func init() {
	rootCmd.AddCommand(confCmd)
	confCmd.Flags().StringP("pattern", "p", ".*",
		"Selection pattern for interfaces to include")
}
