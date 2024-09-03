---
date: 2024-09-03T09:54:19+02:00
title: "warc ls"
slug: warc_ls
url: /cmd/warc_ls/
---
## warc ls

List WARC record fields

### Synopsis

List information about WARC records

```
warc ls FILE/DIR ... [flags]
```

### Options

```
  -c, --concurrency int           number of input files to process simultaneously. (default 24)
  -d, --delimiter string          field delimiter (default " ")
  -F, --fields string             which fields to include in the output
                                  
                                  Field specification letters are mostly the same as the fields in the CDX file specification (https://iipc.github.io/warc-specifications/specifications/cdx-format/cdx-2015/).
                                  
                                  The following fields are supported:
                                  	a - original URL
                                  	b - date in 14 digit format
                                  	B - date in RFC3339 format
                                  	e - IP address
                                  	g - filename
                                  	h - original host
                                  	i - record id
                                  	k - checksum
                                  	m - document mime type
                                  	s - http response code
                                  	S - record size
                                  	T - record type
                                  	V - offset
                                  
                                  A number after the field letter restricts the field length. By adding a + or - sign before the number the field is padded to have the exact length. + is right aligned and - is left aligned.
      --ftp-pool-size int32       size of the ftp pool (default 1)
  -h, --help                      help for ls
      --id strings                filter record ID's. For more than one, repeat flag or comma separated list.
  -i, --input-file string         input file (system). Default is to use OS file system.
                                  Legal values:
                                  	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                  	ftp://user/pass@host:port
                                  
      --json                      output as JSON lines
  -l, --limit int                 The maximum number of records to show. Defaults to show all records.
                                  If -o or -n option is set limit is set to 1.
  -m, --mime-type strings         filter records with given mime-types. For more than one, repeat flag or comma separated list.
  -n, --num int                   print the n'th record. Only records that are not filtered out by other options are counted.
  -o, --offset int                record offset
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
      --strict                    strict parsing
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

