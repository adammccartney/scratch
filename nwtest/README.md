nwtest
======

This program was written to debug the network setup at a multi-site HPC
center. We noticed a problem with a `prometheus-cvmfs-exporter`. The exporter
is installed via the package manager.

```
Name         : prometheus-cvmfs-exporter
Version      : 1.0.0
Release      : 2
Architecture : noarch
Size         : 32 k
Source       : prometheus-cvmfs-exporter-1.0.0-2.src.rpm
Repository   : @System
From repo    : cernvm
Summary      : Prometheus exporter for CVMFS client monitoring
URL          : https://github.com/cvmfs-contrib/prometheus-cvmfs
License      : BSD-3-Clause
Description  : A Prometheus exporter for monitoring CVMFS (CernVM File System) clients.
             : This package provides a script that collects metrics from CVMFS repositories
             : and exposes them in Prometheus format, along with systemd service files
             : for running the exporter as a service.
             :
             : The exporter collects various metrics including:
             : - Cache hit rates and sizes
             : - Download statistics
             : - Repository status and configuration
             : - Proxy usage and performance
             : - System resource usage by CVMFS processes
```

The package installs a single shell script that can be used to generate a http
response. The script is managed by a systemd service [^1] and provided over a
socket [^2]. This can then be scraped via a http GET to
`$COMPUTE_NODE:9868`.

This call works fine on the site specific network, but as soon as we attempt to
route the request via wireguard, it fails with: 

`curl: (56) Recv failure: Connection reset by peer`


As a first step to debug this issue, we tried to rule out that it was
an issue being caused by CVMFS. To do this we set up a simple webserver
that sends an equivalent number of bytes over the network on the same
port. In the context of a successful call ~40Kb are sent. 

`nwtest` is a program to run a simple web server written in go that
serves a html page with ~40Kb of content.

To run the test, we disable the systemd socket that serves the
prometheus exporter, and instead run `nwtest` listening to the same port
on all interfaces.

The `nwtest` run manages to successfully transfer the ~40Kb payload after a curl
call:

```
> curl -v $COMPUTE_NODE:9868 -w %{size_download}
...
* Connection #0 to host $COMPUTE_NODE left intact
40429
```

After googling a bit for the `Recv failure: Connection reset by peer` error
types, one answer mentions the MTU value as a possible cause [^3]
Examining the network setup on the compute node, we see MTU is set to 9000 for
the musica network interface

```
4: eno2np0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9000 qdisc mq state UP group default qlen 1000
```

This means that the initial segments sent from the compute node are chunked
into 9000 byte units, then sent on their way. We know from our setup that this
MTU size is not normalized across the networks. This can lead to problems with 
Path MTU Discovery.[^4]

[^1]: <https://raw.githubusercontent.com/cvmfs-contrib/prometheus-cvmfs/refs/heads/main/systemd/cvmfs-client-prometheus%40.service>
[^2]: <https://raw.githubusercontent.com/cvmfs-contrib/prometheus-cvmfs/refs/heads/main/systemd/cvmfs-client-prometheus.socket>
[^3]: <https://stackoverflow.com/questions/10285700/curl-error-recv-failure-connection-reset-by-peer>
[^4]: <https://en.wikipedia.org/wiki/Path_MTU_Discovery>

