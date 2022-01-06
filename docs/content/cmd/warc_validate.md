---
date: 2022-01-06T15:44:11+01:00
title: "warc validate"
slug: warc_validate
url: /cmd/warc_validate/
---
## warc validate

Validate warc files

```
warc validate <files/dirs> [flags]
```

### Options

```
  -c, --concurrency int    number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores> (default 24)
  -h, --help               help for validate
  -r, --recursive          walk directories recursively
      --suffixes strings   filter by suffixes (default [.warc,.warc.gz])
  -s, --symlinks           follow symlinks
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

