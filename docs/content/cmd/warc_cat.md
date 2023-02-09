---
date: 2023-02-09T06:46:00+01:00
title: "warc cat"
slug: warc_cat
url: /cmd/warc_cat/
---
## warc cat

Concatenate and print warc files

```
warc cat [flags]
```

### Examples

```
# Print all content from a WARC file
warc cat file1.warc.gz

# Pipe payload from record #4 into the image viewer feh
warc cat -n4 -P file1.warc.gz | feh -
```

### Options

```
  -w, --header                 show WARC header
  -h, --help                   help for cat
      --id stringArray         filter record ID's. For more than one, repeat flag or comma separated list.
  -m, --mime-type strings      filter records with given mime-types. For more than one, repeat flag or comma separated list.
  -n, --num int                print the n'th record. This is applied after records are filtered out by other options (default -1)
  -o, --offset int             print record at offset bytes (default -1)
  -P, --payload                show payload
  -p, --protocol-header        show protocol header
  -c, --record-count int       The maximum number of records to show. Defaults to show all records except if -o or -n option is set, then default is one.
  -t, --record-type strings    filter record types. For more than one, repeat flag or comma separated list.
                               Legal values: warcinfo,request,response,metadata,revisit,resource,continuation,conversion
  -S, --response-code string   filter records with given http response codes. Format is 'from-to' where from is inclusive and to is exclusive.
                               Examples:
                               '200': only records with 200 response
                               '200-300': all records with response code between 200(inclusive) and 300(exclusive)
                               '-400': all response codes below 400
                               '500-': all response codes from 500 and above
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

