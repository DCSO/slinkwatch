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
	"io/ioutil"
	"syscall"
	"testing"
)

var watchedIfs = []string{"eth0", "eth1", "eth2", "eth3"}
var testcfg = `
# Interfaces available for Suricata
---
ifaces:
  eth0:
    threadweight: 1
    clusterid: 99
  eth1:
    threadweight: 10
    clusterid: 98
  eth2:
    threadweight: 3
    clusterid: 97
  eth3:
    threadweight: 3
    clusterid: 96
`

func TestConfigLoad(t *testing.T) {
	f, err := ioutil.TempFile("", "testmainconf")
	if err != nil {
		panic(err)
	}
	defer syscall.Unlink(f.Name())
	ioutil.WriteFile(f.Name(), []byte(testcfg), 0644)

	var cfg Config
	err = cfg.LoadConfig(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	seenIfs := cfg.GetWatchedInterfaces()

	if len(seenIfs) != len(watchedIfs) {
		t.Fail()
	}

	for _, v := range watchedIfs {
		seen := false
		for _, w := range seenIfs {
			if v == w {
				seen = true
				break
			}
		}
		if !seen {
			t.Fail()
		}

	}

}

func TestConfigLoadFail(t *testing.T) {
	testcfg := `
---
if
`
	f, err := ioutil.TempFile("", "testmainconf")
	if err != nil {
		panic(err)
	}
	defer syscall.Unlink(f.Name())
	ioutil.WriteFile(f.Name(), []byte(testcfg), 0644)

	var cfg Config
	err = cfg.LoadConfig(f.Name())
	if err == nil {
		t.Fatal("no error but was expected")
	}
}
