# Secure DNS

> A simple DNS server written in Go using DNS over HTTPS along with some graceful fallback options

## Features
- [x] DNS over HTTPS
- [ ] Fallback over UDP
- [ ] Fallback over TCP
- [ ] Systemd Daemon

## Installation

```bash
go get github.com/rameezk/secure-dns
```

## Usage

### Classic Mode (AKA running from cli)
If your `$GOPATH` is set correctly, you can simply invoke with:
```bash
secdns <args>
```

For a full list of configuration options, use:
```bash
secdns --help
```

NOTE If you're using the default listening address and port (e.g. 127.0.0.1:53), you'll need to run as root
```bash
sudo secdns <args>
```

### systemd
Coming Soon (TM)
