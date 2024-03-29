---
date: 2024-02-21T15:43:33+01:00
title: "warc dedup"
slug: warc_dedup
url: /cmd/warc_dedup/
---
## warc dedup

Deduplicate WARC files

### Synopsis

Deduplicate WARC files.

NOTE: The filtering options only decides which records are candidates for deduplication.
The remaining records are written as is.

```
warc dedup [flags]
```

### Options

```
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
  -z, --compress                        use gzip compression for WARC files
      --compression-level               the gzip compression level to use (value between 1 and 9)
  -c, --concurrency int                 number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores> (default 24)
  -C, --concurrent-writers int          maximum concurrent WARC writers. This is the number of WARC-files simultaneously written to.
                                        A consequence is that at least this many WARC files are created even if there is only one input file. (default 16)
  -S, --file-size string                The maximum size for WARC files (default "1GB")
      --flush                           if true, sync WARC file to disk after writing each record
  -h, --help                            help for dedup
      --id stringArray                  filter record ID's. For more than one, repeat flag or comma separated list.
  -i, --index-dir string                directory to store indexes (default "/home/johnh/.cache/warc")
  -k, --keep-index                      true to keep index on disk so that the next run will continue where the previous run left off
  -m, --mime-type strings               filter records with given mime-types. For more than one, repeat flag or comma separated list.
      --min-free-disk string            minimum free space on disk to allow WARC writing (default "256MB")
  -g, --min-size-gain string            minimum bytes one must earn to perform a deduplication (default "2KB")
  -n, --name-generator string           the name generator to use. By setting this to 'identity', the input filename will also be used as
                                        output file name (prefix and suffix might still change). In this mode exactly one file is generated for every input file (default "default")
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
  -p, --prefix string                   filename prefix for WARC files
  -t, --record-type strings             filter record types. For more than one, repeat flag or comma separated list.
                                        Legal values: warcinfo,request,response,metadata,revisit,resource,continuation,conversion (default [response])
  -r, --recursive                       walk directories recursively
  -R, --repair                          try to fix errors in records
  -e, --response-code string            filter records with given http response codes. Format is 'from-to' where from is inclusive and to is exclusive.
                                        Examples:
                                        '200': only records with 200 response
                                        '200-300': all records with response code between 200(inclusive) and 300(exclusive)
                                        '-400': all response codes below 400
                                        '500-': all response codes from 500 and above
      --source-filesystem string        the source filesystem to use for input files. Default is to use OS file system. Legal values:
                                          ftp://user/pass@host:port
                                          tar://path/to/archive.tar
                                          tgz://path/to/archive.tar.gz
                                        
      --subdir-pattern string           a pattern to use for generating subdirectories.
                                        / in pattern separates subdirectories on all platforms
                                        {YYYY} is replaced with a 4 digit year
                                        {YY} is replaced with a 2 digit year
                                        {MM} is replaced with a 2 digit month
                                        {DD} is replaced with a 2 digit day
                                        The date used is the WARC date of each record. Therefore a input file might be split into 
                                        WARC files in different subdirectories. If NameGenerator is 'identity' only the first record
                                        of each file's date is used to keep the file as one.
      --suffixes strings                filter files by suffixes (default [.warc,.warc.gz])
  -s, --symlinks                        follow symlinks
  -w, --warc-dir string                 output directory for generated warc files. Directory must exist. (default ".")
```

### Options inherited from parent commands

```
      --config string           config file. If not set, /etc/xdg/warc, /home/johnh/.config/warc and the current directory will be searched for a file named 'config.yaml'
      --log-console strings     the kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings        the kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string    a file to write log output. Empty for no log file
      --max-buffer-mem string   the maximum bytes of memory allowed for each buffer before overflowing to disk (default "1MB")
      --tmpdir string           directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

