---
date: 2023-05-02T10:06:41+02:00
title: "warc console"
slug: warc_console
url: /cmd/warc_console/
---
## warc console

A shell for working with WARC files

```
warc console <directory> [flags]
```

### Options

```
  -h, --help               help for console
      --suffixes strings   filter files by suffixes (default [.warc,.warc.gz])
```

### Options inherited from parent commands

```
      --config string           config file. If not set, /etc/xdg/warc, /home/johnh/.config/warc and the current directory will be searched for a file named 'config.yaml'
      --log-console strings     the kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings        the kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string    a file to write log output. Empty for no log file
      --max-buffer-mem string   the maximum bytes of memory allowed for each buffer before overflowing to disk (default "1MB")
      --tmpdir string           directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

