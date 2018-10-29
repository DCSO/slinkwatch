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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// mmanCmd represents the makeman command
var mmanCmd = &cobra.Command{
	Use:   "makeman [options]",
	Short: "Create man pages",
	Run: func(cmd *cobra.Command, args []string) {
		targetDir, err := cmd.Flags().GetString("dir")
		if err != nil {
			log.Fatal(err)
		}
		header := &doc.GenManHeader{}
		err = doc.GenManTree(rootCmd, header, targetDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range rootCmd.Commands() {
			err = doc.GenManTree(v, header, targetDir)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(mmanCmd)
	mmanCmd.Flags().StringP("dir", "d", ".", "target directory for man pages")
}
