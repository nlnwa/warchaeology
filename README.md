# warchaeology
Command line tool for digging into WARC files

# Installation
## Linux
### RPM package
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_0.1.0-RC.4_x86_64.rpm
sudo rpm -Uvh warchaeology_0.1.0-RC.4_x86_64.rpm
```

### Debian package
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warchaeology_0.1.0-RC.4_amd64.deb
sudo dpkg -i warchaeology_0.1.0-RC.4_amd64.deb
```

### Binary download
```
curl -LO https://github.com/nlnwa/warchaeology/releases/latest/download/warc_linux_x86_64
sudo install warc_linux_x86_64 /usr/local/bin/warc
```

#### Command completion
bash
```
sudo sh -c "/usr/local/bin/warc completion bash > /etc/bash_completion.d/warc"
```

zshell
```
source <(warc completion zsh)
```
