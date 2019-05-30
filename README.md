# Secure DNS

> A simple DNS server written in Go using DNS over HTTPS along with some graceful fallback options

## Features
- [x] DNS over HTTPS
- [x] Fallback over UDP
- [x] Systemd Daemon

## Installation

```bash
go get github.com/rameezk/secure-dns
```

## Usage

### Classic Mode (AKA running from cli)
If your `$GOPATH` is set correctly, you can simply invoke with:
```bash
secure-dns <args>
```

For a full list of configuration options, use:
```bash
secure-dns --help
```

NOTE If you're using the default listening address and port (e.g. 127.0.0.1:53), you'll need to run as root
```bash
sudo secure-dns <args>
```

### Daemon (via systemd)
The systemd daemon can be installed by running:
```bash
./install-systemd-service.sh
```

The configuration file can be found at `/etc/secure-dns.conf`
