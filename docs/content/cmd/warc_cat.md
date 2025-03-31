---
date: 2025-03-31T14:23:40+02:00
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

# Print all content from a WARC file (in principle the same as zcat)
warc cat file1.warc.gz

# Pipe the payload of the 4th record into the image viewer feh
warc cat -n4 -P file1.warc.gz | feh -
```

### Options

```
  -z, --compress                  compress output (per record)
      --continue-on-error         continue on error. Will continue processing files and directories in spite of errors.
  -f, --force                     force the record iterator to continue regardless of errors.
      --ftp-pool-size int32       size of the ftp pool (default 1)
  -w, --header                    show WARC header
  -h, --help                      help for cat
      --id strings                filter record ID's. For more than one, repeat flag or use comma separated list.
  -i, --input-file string         input file (system). Default is to use OS file system.
                                  Legal values:
                                  	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                  	ftp://user/pass@host:port
                                  
      --lenient                   sets the parser to do as little validation as possible.
  -l, --limit int                 limit the number of records to process. If the -n option is specified the limit is ignored.
  -m, --mime-type strings         filter records with given mime-types. For more than one, repeat flag or use a comma separated list.
  -n, --nth int                   only process the n'th record. Only records that are not filtered out by other options are counted.
  -o, --offset int                start processing from this byte offset in file. Defaults to 0.
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
      --strict                    sets the parser to fail on first validation error.
      --suffixes strings          filter files by suffix (default [.warc,.warc.gz])
  -s, --symlinks                  follow symlinks
      --tmp-dir string            directory to use for temporary files (default "/tmp")
```

### Options inherited from parent commands

```
      --config string       config file. If not set $XDG_CONFIG_DIRS, /etc/xdg/warc $XDG_CONFIG_HOME/warc and the current directory will be searched for a file named 'config.yaml'
  -O, --log-file string     log to file (default "-")
      --log-format string   log format. Valid values: text, json (default "text")
      --log-level string    log level. Valid values: debug, info, warn, error (default "info")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

