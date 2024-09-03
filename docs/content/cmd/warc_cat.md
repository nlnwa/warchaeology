---
date: 2024-09-03T09:54:19+02:00
title: "warc cat"
slug: warc_cat
url: /cmd/warc_cat/
---
## warc cat

Concatenate and print warc files

```
warc cat FILE/DIR ... [flags]
```

### Examples

```
Print all content from a WARC file
warc cat file1.warc.gz

# Pipe payload from record #4 into the image viewer feh
warc cat -n4 -P file1.warc.gz | feh -
```

### Options

```
  -z, --compress                  output is compressed (per record)
      --ftp-pool-size int32       size of the ftp pool (default 1)
  -w, --header                    show WARC header
  -h, --help                      help for cat
      --id strings                filter record ID's. For more than one, repeat flag or comma separated list.
  -i, --input-file string         input file (system). Default is to use OS file system.
                                  Legal values:
                                  	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                  	ftp://user/pass@host:port
                                  
  -l, --limit int                 The maximum number of records to show. Defaults to show all records.
                                  If -o or -n option is set limit is set to 1.
  -m, --mime-type strings         filter records with given mime-types. For more than one, repeat flag or comma separated list.
  -n, --num int                   print the n'th record. Only records that are not filtered out by other options are counted.
  -o, --offset int                record offset
  -P, --payload                   show payload
  -p, --protocol-header           show protocol header
  -t, --record-type strings       filter records by type. For more than one, repeat the flag or use a comma separated list.
                                  Legal values:
                                  	warcinfo, request, response, metadata, revisit, resource, continuation and conversion
  -r, --recursive                 walk directories recursively
  -S, --response-code string      filter records by http response code
                                  Example:
                                  	200	- only records with a 200 response
                                  	200-300	- records with response codes between 200 (inclusive) and 300 (exclusive)
                                  	500-	- response codes from 500 and above
                                  	-400	- all response codes below 400
      --source-file-list string   a file containing a list of files to process, one file per line
      --suffixes strings          filter files by suffix (default [.warc,.warc.gz])
  -s, --symlinks                  follow symlinks
      --tmpdir string             directory to use for temporary files (default "/tmp")
```

### Options inherited from parent commands

```
      --config string       config file. If not set, $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'
  -O, --log-file string     log to file (default "-")
      --log-format string   log format. Valid values: text, json (default "text")
      --log-level string    log level. Valid values: debug, info, warn, error (default "info")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

