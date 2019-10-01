package slinkwatch

import (
	"io/ioutil"
	"os"
	"testing"
	"text/template"
)

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

func mkConfig() *Config {
	return &Config{
		Ifaces: map[string]ConfIface{
			"eth0": ConfIface{
				ThreadWeight: 2,
				ClusterID:    99,
			},
			"eth1": ConfIface{
				ThreadWeight: 2,
				ClusterID:    99,
				ExtraOptions: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
}

func mkConfigFile(t *testing.T) *os.File {
	cfgString := `# Interfaces available for Suricata
---
ifaces:
  eth0:
    threadweight: 2
    clusterid: 99
    extraopts:
      cluster-type: cluster_qm
      foo: barbaz
      hasi: 234
  eth1:
    threadweight: 2
    clusterid: 99
    extraopts:
      cluster-type: cluster_flow
      captain: jack`
	tempFile, err := ioutil.TempFile("", "slinkwatch")
	if err != nil {
		t.Fatal(err)
	}
	tempFile.WriteString(cfgString)
	return tempFile
}

func mkEmptyConfig() *Config {
	return &Config{}
}

func mkTemplate(t *testing.T) *template.Template {
	tmpl := `%YAML 1.1
---
af-packet:{{ range $iface, $vals := . }}
  - interface: {{ $iface }}
    threads: {{ $vals.Threads }}
    cluster-id: {{ $vals.ClusterID }}{{ range $extrakey, $extraval := $vals.ExtraOptions }}
    {{ $extrakey }}: {{ $extraval}}{{ end }}
{{ else }}
  - interface: default
    threads: auto
{{ end }}`
	tmplCompiled, err := template.New("test").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	return tmplCompiled
}

func TestActiveSetEmptyIfaces(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)

	if as.Len() != 0 {
		t.Fail()
	}
	if as.String() != "<empty>" {
		t.Fail()
	}
}

func TestActiveSetNoMatchConfig(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)
	as.Add("eth4")

	if as.Len() != 0 {
		t.Fail()
	}
}

func TestActiveSetOneIface(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)
	as.Add("eth0")

	if as.Len() != 1 {
		t.Fail()
	}
}

func TestActiveSetOneIfaceTwice(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)
	as.Add("eth0")
	as.Add("eth0")

	if as.Len() != 1 {
		t.Fail()
	}
}

func TestActiveSetTwoIfaces(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)
	as.Add("eth0")
	as.Add("eth1")

	if as.Len() != 2 {
		t.Fail()
	}

	as.Reset()
	if as.Len() != 0 {
		t.Fail()
	}
}

func TestActiveSetTwoIfacesToString(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)
	as.Add("eth0")
	as.Add("eth1")

	if as.String() == "<empty>" {
		t.Fail()
	}
	as.Reset()
	if as.String() != "<empty>" {
		t.Fail()
	}
}

func TestActiveSetToYAML(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)
	as.Add("eth0")
	as.Add("eth1")

	y, err := as.ToYAML(mkTemplate(t), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(y) == 0 {
		t.Fatal(err)
	}
}

func TestActiveSetEmptyToYAML(t *testing.T) {
	cfg := mkConfig()
	as := MakeActiveSet(cfg)

	y, err := as.ToYAML(mkTemplate(t), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(y) == 0 {
		t.Fatal(err)
	}
}

func TestActiveSetConfigFileToYAML(t *testing.T) {
	cfg := &Config{}
	as := MakeActiveSet(cfg)
	tf := mkConfigFile(t)
	defer os.Remove(tf.Name())
	err := cfg.LoadConfig(tf.Name())
	if err != nil {
		t.Fatal(err)
	}
	as.Add("eth0")
	as.Add("eth1")

	y, err := as.ToYAML(mkTemplate(t), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(y) == 0 {
		t.Fatal(err)
	}
}
