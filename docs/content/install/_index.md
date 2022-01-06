---
date: 2022-01-05T14:00:02+01:00
title: "Installation"
slug: installation
url: /installation
---

## Linux

### RPM package
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_{{<version>}}_x86_64.rpm
sudo rpm -Uvh warchaeology_{{<version>}}_x86_64.rpm

curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_rall_x86_64.rpm
sudo rpm -Uvh warchaeology_rall_x86_64.rpm
```

### Debian package
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_rall_amd64.deb
sudo dpkg -i warchaeology_rall_amd64.deb
```

### Binary download
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warc_linux_x86_64
sudo install warc_linux_x86_64 /usr/local/bin/warc
```

#### Command completion
When using .deb or .rpm packages, command completion is automatically configured.

With other installation methods command completion scripts can be generated with
the [`warc completion`](/cmd/warc_completion) command.
