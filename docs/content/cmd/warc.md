---
date: 2024-09-03T09:54:19+02:00
title: "warc"
slug: warc
url: /cmd/warc/
---
## warc

A tool for handling warc files

### Options

```
      --config string       config file. If not set, $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'
  -h, --help                help for warc
  -O, --log-file string     log to file (default "-")
      --log-format string   log format. Valid values: text, json (default "text")
      --log-level string    log level. Valid values: debug, info, warn, error (default "info")
```

### SEE ALSO

* [warc cat](../warc_cat/)	 - Concatenate and print warc files
* [warc console](../warc_console/)	 - A shell for working with WARC files
* [warc convert](../warc_convert/)	 - Convert web archives to warc files. Use subcommands for the supported formats
* [warc dedup](../warc_dedup/)	 - Deduplicate WARC files
* [warc ls](../warc_ls/)	 - List WARC record fields
* [warc validate](../warc_validate/)	 - Validate WARC files
* [warc version](../warc_version/)	 - Show extended version information

