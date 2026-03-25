---
title: "warc ls"
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
  -c, --concurrency int           number of input files to process in parallel (default 21)
      --continue-on-error         continue processing remaining files and directories after errors
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
  -f, --force                     continue iterating even when record read errors occur
      --ftp-pool-size int32       size of the FTP connection pool (default 1)
  -h, --help                      help for ls
      --id strings                only process records with these record IDs; repeat the flag or use a comma-separated list
  -i, --input-file string         input filesystem source; default is the local OS filesystem
                                  Legal values:
                                  	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                  	ftp://user/pass@host:port
                                  
      --json                      output as JSON lines
      --lax-host-parsing          allow lenient host parsing in URL parsing
      --lenient                   minimize validation for faster, more permissive parsing
  -l, --limit int                 maximum number of records to process; ignored when --nth is set
  -m, --mime-type strings         only process records with these MIME types; repeat the flag or use a comma-separated list
  -n, --nth int                   process only the n-th record after filtering
  -o, --offset int                start processing at this byte offset in the input file (default: 0)
  -t, --record-type strings       only process records with these record types; repeat the flag or use a comma-separated list.
                                  Legal values:
                                  	warcinfo, request, response, metadata, revisit, resource, continuation and conversion
  -r, --recursive                 walk input directories recursively
  -S, --response-code string      only process records by HTTP response code
                                  Example:
                                  	200	- only records with a 200 response
                                  	200-300	- records with response codes between 200 (inclusive) and 300 (exclusive)
                                  	500-	- response codes from 500 and above
                                  	-400	- all response codes below 400
      --source-file-list string   path to a file listing input paths, one per line
      --strict                    fail on the first validation error
      --suffixes strings          only process files with these suffixes (default [.warc,.warc.gz])
  -s, --symlinks                  follow symbolic links while walking
      --tmp-dir string            directory used for temporary files (default "/tmp")
```

### Options inherited from parent commands

```
      --config string       path to config file; if unset, searches standard XDG config locations and the current directory for config.yaml
  -O, --log-file string     log output destination ('-' for stderr) (default "-")
      --log-format string   log output format (text or json) (default "text")
      --log-level string    minimum log level (debug, info, warn, error) (default "info")
```

### SEE ALSO

* [warc](warc.md)	 - A tool for handling warc files

