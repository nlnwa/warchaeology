---
date: 2022-01-05T14:00:02+01:00
title: "Installation"
slug: installation
url: /installation
weight: 1
---

## Linux

### RPM package

```shell
curl -LO https://github.com/nationallibraryofnorway/warchaeology/releases/latest/download/warchaeology-{{< version ".rpm" >}}.x86_64.rpm
sudo rpm -Uvh warchaeology-{{< version ".rpm" >}}.x86_64.rpm
```

### Debian package

```shell
curl -LO https://github.com/nationallibraryofnorway/warchaeology/releases/latest/download/warchaeology_{{< version >}}_amd64.deb
sudo dpkg -i warchaeology_{{< version >}}_amd64.deb
```

### Binary download

```shell
curl -LO https://github.com/nationallibraryofnorway/warchaeology/releases/latest/download/warchaeology_Linux_x86_64.tar.gz
tar zxvf warchaeology_Linux_x86_64.tar.gz
sudo install warc /usr/local/bin/warc
```

For a local installation (only for the current user) execute `install warc $HOME/.local/bin`.

#### Command completion

When using .deb or .rpm packages, command completion is automatically configured.
With other installation methods command completion scripts can be generated with
the [`warc completion`](../cmd/warc_completion) command.

## Mac

```shell
curl -LO https://github.com/nationallibraryofnorway/warchaeology/releases/latest/download/warchaeology_Darwin_x86_64.tar.gz
tar zxvf warchaeology_Darwin_x86_64.tar.gz
sudo install warc /usr/local/bin/warc
```
