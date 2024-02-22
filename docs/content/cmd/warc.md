---
date: 2024-02-22T10:56:13+01:00
title: "warc"
slug: warc
url: /cmd/warc/
---
## warc

A tool for handling warc files

### Options

```
      --config string           config file. If not set, /etc/xdg/warc, /home/johnh/.config/warc and the current directory will be searched for a file named 'config.yaml'
  -h, --help                    help for warc
      --log-console strings     the kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings        the kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string    a file to write log output. Empty for no log file
      --max-buffer-mem string   the maximum bytes of memory allowed for each buffer before overflowing to disk (default "1MB")
      --tmpdir string           directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc cat](../warc_cat/)	 - Concatenate and print warc files
* [warc completion](../warc_completion/)	 - Generate completion script
* [warc console](../warc_console/)	 - A shell for working with WARC files
* [warc convert](../warc_convert/)	 - Convert web archives to warc files. Use subcommands for the supported formats
* [warc dedup](../warc_dedup/)	 - Deduplicate WARC files
* [warc ls](../warc_ls/)	 - List warc file contents
* [warc validate](../warc_validate/)	 - Validate warc files

