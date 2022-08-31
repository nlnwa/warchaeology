---
date: 2022-08-31T12:06:31+02:00
title: "warc ls"
slug: warc_ls
url: /cmd/warc_ls/
---
## warc ls

List warc file contents

### Synopsis

List information about records in one or more warc files.

Several options exist to influence what to output.
  --delimiter accepts a string to be used as the output field delimiter.
  --format specifies one of the predefined output formats (only cdxj is supported at the moment).
  --fields specifies which fields to include in output. Field specification letters are mostly the same as the fields in
           the CDX file specification (https://iipc.github.io/warc-specifications/specifications/cdx-format/cdx-2015/).
           The following fields are supported:
             a - original URL
             b - date in 14 digit format
             B - date in RFC3339 format
             e - IP
             g - file name
             h - original host
             i - record id
             k - checksum
             m - document mime type
             s - http response code
             S - record size in WARC file
             T - record type
             V - Offset in WARC file
           A number after the field letter restricts the field length. By adding a + or - sign before the number the field is
           padded to have the exact length. + is right aligned and - is left aligned.

```
warc ls <files/dirs> [flags]
```

### Options

```
  -c, --concurrency int       number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores> (default 1)
  -d, --delimiter string      use string instead of SPACE for field delimiter (default " ")
  -f, --fields string         which fields to include. See 'warc help ls' for a description
  -h, --help                  help for ls
      --id stringArray        specify record ids to ls
  -o, --offset int            record offset (default -1)
  -n, --record-count int      The maximum number of records to show
  -t, --record-type strings   which record types to include. For more than one, repeat flag or comma separated list.
                              Legal values: warcinfo,request,response,metadata,revisit,resource,continuation,conversion (defaults to all record types)
  -r, --recursive             walk directories recursively
      --strict                strict parsing
      --suffixes strings      filter files by suffixes (default [.warc,.warc.gz])
  -s, --symlinks              follow symlinks
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

