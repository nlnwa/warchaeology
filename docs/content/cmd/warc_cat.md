---
date: 2022-01-07T16:26:57+01:00
title: "warc cat"
slug: warc_cat
url: /cmd/warc_cat/
---
## warc cat

Concatenate and print warc files

```
warc cat [flags]
```

### Options

```
  -h, --help               help for cat
      --id stringArray     id
  -o, --offset int         record offset (default -1)
  -c, --record-count int   The maximum number of records to show
  -s, --strict             strict parsing
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

