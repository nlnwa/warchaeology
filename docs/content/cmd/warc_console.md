---
date: 2022-01-06T15:51:39+01:00
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
      --suffixes strings   filter by suffixes (default [.warc,.warc.gz])
```

### Options inherited from parent commands

```
      --config string          config file. If not set, /etc/warc/, $HOME/.warc/ and current working dir will be searched for file config.yaml
      --log-console strings    The kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings       The kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string   a file to write log output. Empty for no log file
      --tmpdir string          directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

