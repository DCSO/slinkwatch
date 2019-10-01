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
	"bytes"
	"math"
	"runtime"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

// InterfaceInfo contains basic information about a specific interface entry.
type InterfaceInfo struct {
	Threads      int
	ClusterID    int
	ExtraOptions map[string]string
}

// ActiveSet contains descriptions of the currently connected interfaces.
type ActiveSet struct {
	Config *Config
	Ifaces map[string]InterfaceInfo
}

// MakeActiveSet returns a new empty ActiveSet.
func MakeActiveSet(config *Config) *ActiveSet {
	return &ActiveSet{
		Config: config,
		Ifaces: make(map[string]InterfaceInfo),
	}
}

// Reset clears the interfaces associated with an ActiveSet.
func (a *ActiveSet) Reset() {
	a.Ifaces = make(map[string]InterfaceInfo)
}

// String returns a string representation of an ActiveSet, listing its
// interfaces.
func (a *ActiveSet) String() string {
	out := ""
	i := 0
	if len(a.Ifaces) == 0 {
		return "<empty>"
	}
	for k := range a.Ifaces {
		out += k
		if i != len(a.Ifaces)-1 {
			out += "/"
			i++
		}
	}
	return out
}

// Add adds an interface to a given ActiveSet.
func (a *ActiveSet) Add(iface string) {
	if _, ok := a.Config.Ifaces[iface]; ok {
		a.Ifaces[iface] = InterfaceInfo{}
	}
	// from here we can assume that we only handle the intersection of the
	// configured interfaces and the observed interfaces
}

// Len returns the number of interfaces in the active set.
func (a *ActiveSet) Len() int {
	return len(a.Ifaces)
}

// ToYAML returns the YAML representation of an ActiveSet, given a template and
// a configuration with thread weights.
func (a *ActiveSet) ToYAML(tmpl *template.Template, config *Config) (string, error) {
	threads := runtime.NumCPU()
	totalweight := 0
	for intf := range a.Ifaces {
		totalweight += config.Ifaces[intf].ThreadWeight
	}
	for intf := range a.Ifaces {
		a.Ifaces[intf] = InterfaceInfo{
			Threads:      int(math.Ceil(float64(threads) * (float64(config.Ifaces[intf].ThreadWeight) / float64(totalweight)))),
			ClusterID:    config.Ifaces[intf].ClusterID,
			ExtraOptions: config.Ifaces[intf].ExtraOptions,
		}
	}
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, a.Ifaces)
	if err != nil {
		return "", err
	}
	var t interface{}
	err = yaml.Unmarshal(buf.Bytes(), &t)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
