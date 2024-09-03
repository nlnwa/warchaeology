---
date: 2024-09-03T09:54:19+02:00
title: "warc validate"
slug: warc_validate
url: /cmd/warc_validate/
---
## warc validate

Validate WARC files

```
warc validate FILE/DIR ... [flags]
```

### Options

```
      --calculate-hash string           calculate hash of output file. The hash is made available to the close output file hook as WARC_HASH. Valid values: md5, sha1, sha256, sha512
      --close-input-file-hook string    a command to run after closing each input file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the input file
                                        	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
      --close-output-file-hook string   a command to run after closing each output file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the output file
                                        	WARC_SIZE contains the size of the output file
                                        	WARC_INFO_ID contains the ID of the output file's WARCInfo-record if created
                                        	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
                                        	WARC_HASH contains the hash of the output file if computed
                                        	WARC_ERROR_COUNT contains the number of errors found if the file was validated and the validation failed
  -c, --concurrency int                 number of input files to process simultaneously. (default 24)
      --ftp-pool-size int32             size of the ftp pool (default 1)
  -h, --help                            help for validate
      --id strings                      filter record ID's. For more than one, repeat flag or comma separated list.
      --index-dir string                directory to store indexes (default "/home/mariusb/.cache/index")
  -i, --input-file string               input file (system). Default is to use OS file system.
                                        Legal values:
                                        	/path/to/archive.( tar | tar.gz | tgz | zip | wacz )
                                        	ftp://user/pass@host:port
                                        
  -k, --keep-index                      true to keep index on disk so that the next run will continue where the previous run left off
  -m, --mime-type strings               filter records with given mime-types. For more than one, repeat flag or comma separated list.
  -K, --new-index                       true to start from a fresh index, deleting eventual index from last run
      --open-input-file-hook string     a command to run before opening each input file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the input file
      --open-output-file-hook string    a command to run before opening each output file. The command has access to data as environment variables.
                                        	WARC_COMMAND contains the subcommand name
                                        	WARC_HOOK_TYPE contains the hook type (OpenInputFile, CloseInputFile, OpenOutputFile, CloseOutputFile)
                                        	WARC_FILE_NAME contains the file name of the output file
                                        	WARC_SRC_FILE_NAME contains the file name of the input file if the output file is generated from an input file
  -o, --output-dir string               output directory for validated warc files. If not empty this enables copying of input file. Directory must exist.
  -t, --record-type strings             filter records by type. For more than one, repeat the flag or use a comma separated list.
                                        Legal values:
                                        	warcinfo, request, response, metadata, revisit, resource, continuation and conversion
  -r, --recursive                       walk directories recursively
  -S, --response-code string            filter records by http response code
                                        Example:
                                        	200	- only records with a 200 response
                                        	200-300	- records with response codes between 200 (inclusive) and 300 (exclusive)
                                        	500-	- response codes from 500 and above
                                        	-400	- all response codes below 400
      --source-file-list string         a file containing a list of files to process, one file per line
      --suffixes strings                filter files by suffix (default [.warc,.warc.gz])
  -s, --symlinks                        follow symlinks
      --tmpdir string                   directory to use for temporary files (default "/tmp")
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

