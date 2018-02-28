# enerflux

[![GoDoc](https://godoc.org/github.com/phlipse/enerflux?status.svg)](https://godoc.org/github.com/phlipse/enerflux)
[![Go Report Card](https://goreportcard.com/badge/github.com/phlipse/enerflux)](https://goreportcard.com/report/github.com/phlipse/enerflux)

enerflux retrieves current data from [getfresh.energy](https://www.getfresh.energy) and stores it to influxdb. getfresh.energy does not offer an official API. enerflux uses same method as the their own web application (customer portal) does to retrieve the data.

## Prerequisite
You need an account at [getfresh.energy](https://www.getfresh.energy).

enerflux works on Windows, too. This description only mentions Linux and also works on MacOS and FreeBSD (other BSDs as well). This has the reason that I don't have a windows system for testing.

## Get enerflux
Build it on your own from source or download binary from [latest release](https://github.com/phlipse/enerflux/releases/latest). Release binaries are build with the following command: ```go build -ldflags "-s -w"```

## Usage
To use enerflux, copy the binary for example to */usr/local/bin/* and make it executable.

```
$ sudo cp enerflux /usr/local/bin/
$ sudo chmod +x /usr/local/bin/enerflux

# use flags to provide all needed information
$ enerflux -h
```

**enerflux can and should safely be run as a non-privileged user. Create one or use an existing.**

### Systemd
If you are on Linux and you want to use systemd to maintain enerflux, simply copy the files from *repositories systemd folder* to */etc/systemd/system/* and enable it:

```
# copy service file to systemd folder
$ sudo cp enerflux.service /etc/systemd/system/

# passwords are sensitive information and should only be accessible by root
$ sudo chmod 600 /etc/systemd/system/enerflux.service

# enable and start it
$ sudo systemctl enable enerflux.service
$ sudo systemctl start enerflux.service

# show logs
$ sudo journalctl -u enerflux
```

**Change username, working directory, path to executable and all command line parameters in *enerflux.service* file to your needs.**

## Visualize
Data could for example be visualized with [grafana](https://grafana.com). To get an idea about it, look at the following picture:

![Grafana dashboard energy data](https://github.com/phlipse/enerflux/blob/master/screenshot.png)

## License

Use of this source code is governed by the [MIT License](https://github.com/phlipse/enerflux/blob/master/LICENSE).