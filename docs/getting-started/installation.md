---
title: Installation
---

## Linux

### RPM package

Download the latest `.rpm` from the
[latest release page](https://github.com/nationallibraryofnorway/warchaeology/releases/latest), then install it.

### Debian package

Download the latest `.deb` from the
[latest release page](https://github.com/nationallibraryofnorway/warchaeology/releases/latest), then install it.

### Binary download

```shell
curl -LO https://github.com/nationallibraryofnorway/warchaeology/releases/latest/download/warchaeology_Linux_x86_64.tar.gz
tar zxvf warchaeology_Linux_x86_64.tar.gz
sudo install warc /usr/local/bin/warc
```

For a local installation for the current user:

```shell
install warc "$HOME/.local/bin"
```

### Command completion

When using `.deb` or `.rpm` packages, command completion is configured automatically.
For other installation methods, generate completion scripts with:

```shell
warc completion
```

## macOS

```shell
curl -LO https://github.com/nationallibraryofnorway/warchaeology/releases/latest/download/warchaeology_Darwin_x86_64.tar.gz
tar zxvf warchaeology_Darwin_x86_64.tar.gz
sudo install warc /usr/local/bin/warc
```
