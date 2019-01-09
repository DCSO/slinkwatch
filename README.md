# 🔃 slinkwatch [![CircleCI](https://circleci.com/gh/DCSO/slinkwatch.svg?style=svg)](https://circleci.com/gh/DCSO/slinkwatch)

slinkwatch is the *Suricata Link Watcher*, a tool to dynamically maintain `interface` entries in Suricata's configuration file, depending on what network interfaces are connected. It is meant to ease deployment of identical sensor installations at many heterogenous sites, allowing to make full use of the sensor resources in the light of varying monitoring volume.

## Interaction with Suricata

In order to propagate changed interface configuration to Suricata, one would need to configure Suricata in such a way that the
section of the configuration YAML created by the slinkwatch template is included from a separate file, e.g. via

```yaml
...

include: interfaces.yaml

...
```

in, for example, `/etc/suricata/suricata.yaml` and then specifying `/etc/suricata/interfaces.yaml` as the `--target-file` in the slinkwatch call. Note that it must be writable by the slinkwatch process!

After modifying this file whenever a status change occurs, slinkwatch will attempt to restart Suricata. For systems that run [systemd](https://www.freedesktop.org/wiki/Software/systemd/), slinkwatch will always try to restart the service given by the `--service-name` option (default is `suricata.service`). On non-systemd systems, it will run a simple command to restart Suricata (default is `/etc/init.d/suricata restart`). Whether systemd is available will be checked at runtime.

## Resource assignment

Support for interfaces going online and offline at runtime obviously raises the question of how to assign computing resources (i.e. detection threads) to the individual interfaces. Unless all interfaces on a sensor are completely  identical in terms of supported bandwidth and traffic, it is not sufficient to simply assign equal amounts of threads to all interfaces that are connected at a given point in time. For example, one might have a 10Gbit interface `eth1` and a 1Gbit interface `eth2` on such a machine, and most certainly one would want more threads  assigned to the former than to the latter. Even more importantly, how would an existing assignment be changed to be most efficient when another 10Gbit interface, say `eth3` gets a link?

We address this issue using _thread weights_. That is, each interface _i_ is assigned a integer value _w<sub>i</sub>_ to denote its resource allocation importance. For example, in the case above we could assign _w<sub>eth1</sub>_ = _w<sub>eth3</sub>_ = 10 and _w<sub>eth2</sub>_ = 2. We also consider the _active set_ of interfaces _A_ as the set of all interfaces that have an active link. On a machine with _n_ threads available for detection, we can then assign to each interface _i_ the value 

![t_i=\lceil n \frac{w_i}{\sum_{j \in A}w_j} \rceil](https://latex.codecogs.com/png.latex?t_i%3D%5Clceil%20n%20%5Cfrac%7Bw_i%7D%7B%5Csum_%7Bj%20%5Cin%20A%7Dw_j%7D%20%5Crceil)

as the number of threads allocated to detection for traffic on interface _i_. For example, for _n_ = 40 we would then set

![t_{\textup{eth1}} = t_{\textup{eth3}} = \lceil 40 \frac{10}{22} \rceil = 19](https://latex.codecogs.com/png.latex?t_%7B%5Ctextup%7Beth1%7D%7D%20%3D%20t_%7B%5Ctextup%7Beth3%7D%7D%20%3D%20%5Clceil%2040%20%5Cfrac%7B10%7D%7B22%7D%20%5Crceil%20%3D%2019)

and

![t_{\textup{eth2}} = \lceil 40 \frac{2}{22} \rceil = 4](https://latex.codecogs.com/png.latex?t_%7B%5Ctextup%7Beth2%7D%7D%20%3D%20%5Clceil%2040%20%5Cfrac%7B2%7D%7B22%7D%20%5Crceil%20%3D%204)

We aim for slight overcommitment of CPU hyperthreads to avoid idling CPUs as much as possible.

## Installing dependencies via `dep`

The component of slinkwatch talking to systemd requires a specific godbus version. Please make sure to run

```
$ dep ensure
```

before building to make sure the correct version constraints apply.

## Usage

Define the interfaces that are available to assign in a YAML file, together with their weights:

```yaml
# Interfaces available for Suricata
--- 
ifaces:
  eth1: 
    clusterid: 98
    threadweight: 10
  eth2: 
    clusterid: 97
    threadweight: 2
  eth3: 
    clusterid: 96
    threadweight: 10
```

Adjust the template to fit the desired configuration format:

```
%YAML 1.1
---
af-packet:{{ range $iface, $vals := . }}
  - interface: {{ $iface }}
    threads: {{ $vals.Threads }}
    cluster-id: {{ $vals.ClusterID }}
    cluster-type: cluster_flow
    defrag: yes
    rollover: yes
    use-mmap: yes
    tpacket-v3: yes
    use-emergency-flush: yes
    buffer-size: 128000
{{ else }}
  - interface: default
    threads: auto
    use-mmap: yes
    rollover: yes
    tpacket-v3: yes
{{ end }}

```

Finally, run slinkwatch, preferably in the background (there's also a systemd service unit file in the repo).

```
$ slinkwatch run 
```

It is possible to specify the locations of the template and config file using the command line parameters:

```
$ slinkwatch run --help
Run the slinkwatch service

Usage:
  slinkwatch run [flags]

Flags:
  -c, --config string            Configuration file (default "config.yaml")
  -h, --help                     help for run
  -i, --interfaces string        Template file for interfaces (default "interfaces.tmpl")
  -p, --poll-interval duration   poll time for interface changes (default 5s)
  -r, --restart-command string   Suricata restart command (default "/etc/init.d/suricata restart")
  -s, --service-name string      systemd service name for Suricata service (default "suricata.service")
  -t, --target-file string       Target YAML file with interface information (default "/etc/suricata/interfaces.yaml")
```

## Other commands

 - `slinkwatch make-config` creates an initial YAML file with skeleton config entries for local interfaces (or a subset defined by a regular expression)
 - `slinkwatch makeman` creates a set of man pages for the tool
 - `slinkwatch show-active` lists the currently active set of interfaces (useful for debugging)

## Dependencies/requirements

Needs [ifplugo](http://github.com/satta/ifplugo) for network change notifications. This introduces a runtime dependency on libdaemon.
It is also highly recommended to use systemd.

## Authors

Sascha Steinbiss

## License

GPL2 (due to ifplugo being GPL2).
