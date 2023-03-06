---
date: 2023-03-06T12:35:33+01:00
title: "warc convert"
slug: warc_convert
url: /cmd/warc_convert/
---
## warc convert

Convert web archives to warc files. Use subcommands for the supported formats

### Options

```
  -h, --help   help for convert
```

### Options inherited from parent commands

```
      --config string           config file. If not set, /etc/warc/, $HOME/.warc/ and current working dir will be searched for file config.yaml
      --log-console strings     the kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings        the kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string    a file to write log output. Empty for no log file
      --max-buffer-mem string   the maximum bytes of memory allowed for each buffer before overflowing to disk (default "1MB")
      --tmpdir string           directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files
* [warc convert arc](../warc_convert_arc/)	 - Convert arc file into warc file
* [warc convert nedlib](../warc_convert_nedlib/)	 - Convert directory with files harvested with Nedlib into warc files
* [warc convert warc](../warc_convert_warc/)	 - Convert WARC file into WARC file

