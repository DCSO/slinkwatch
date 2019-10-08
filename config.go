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

	yaml "gopkg.in/yaml.v2"
)

// ConfIface represents the configured options associated with a single
// interface.
type ConfIface struct {
	ThreadWeight int
	ClusterID    int
	ExtraOptions map[string]string `yaml:"extraopts,omitempty"`
}

// Config represents basic configuration parsed from a file, such as interfaces
// to be watched.
type Config struct {
	Ifaces map[string]ConfIface
}

// LoadConfig populates a Config from a file.
func (c *Config) LoadConfig(path string) error {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, c)
	return err
}

// GetWatchedInterfaces returns a slice of strings, listing all interfaces
// mentioned in the configuration file.
func (c *Config) GetWatchedInterfaces() []string {
	out := make([]string, 0)
	for k := range c.Ifaces {
		out = append(out, k)
	}
	return out
}
