---
date: 2022-01-05T14:00:02+01:00
title: "Installation"
slug: installation
url: /installation
weight: 1
---

## Linux

### RPM package
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology-{{<version>}}.x86_64.rpm
sudo rpm -Uvh warchaeology-{{<version>}}.x86_64.rpm
```

### Debian package
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_{{<version>}}_amd64.deb
sudo dpkg -i warchaeology_{{<version>}}_amd64.deb
```

### Binary download
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_Linux_x86_64.tar.gz
tar zxvf warchaeology_Linux_x86_64.tar.gz
sudo install warc /usr/local/bin/warc
```

#### Command completion
When using .deb or .rpm packages, command completion is automatically configured.
With other installation methods command completion scripts can be generated with
the [`warc completion`](../cmd/warc_completion) command.

## Mac
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_Darwin_x86_64.tar.gz
tar zxvf warchaeology_Darwin_x86_64.tar.gz
sudo install warc /usr/local/bin/warc
```
