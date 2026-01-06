---
date: 2026-01-06T11:56:59+01:00
title: "warc convert"
slug: warc_convert
url: /cmd/warc_convert/
---
## warc convert

Convert web archive files to WARC files. Use subcommands for the supported formats

### Options

```
  -h, --help   help for convert
```

### Options inherited from parent commands

```
      --config string       config file. If not set $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'
  -O, --log-file string     log to file (default "-")
      --log-format string   log format. Valid values: text, json (default "text")
      --log-level string    log level. Valid values: debug, info, warn, error (default "info")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files
* [warc convert arc](../warc_convert_arc/)	 - Convert ARC to WARC
* [warc convert nedlib](../warc_convert_nedlib/)	 - Convert directory with files harvested with Nedlib into warc files
* [warc convert warc](../warc_convert_warc/)	 - Convert WARC file into WARC file

