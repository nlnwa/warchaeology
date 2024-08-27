---
date: 2024-09-03T09:54:19+02:00
title: "warc console"
slug: warc_console
url: /cmd/warc_console/
---
## warc console

A shell for working with WARC files

```
warc console DIR [flags]
```

### Options

```
  -h, --help               help for console
      --suffixes strings   filter files by suffix (default [.warc,.warc.gz])
```

### Options inherited from parent commands

```
      --config string       config file. If not set, $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'
  -O, --log-file string     log to file (default "-")
      --log-format string   log format. Valid values: text, json (default "text")
      --log-level string    log level. Valid values: debug, info, warn, error (default "info")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

