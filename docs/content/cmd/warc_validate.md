---
date: 2024-02-22T10:56:13+01:00
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
  -c, --concurrency int                 number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores> (default 24)
  -h, --help                            help for validate
  -i, --index-dir string                directory to store indexes (default "/home/johnh/.cache/warc")
  -k, --keep-index                      true to keep index on disk so that the next run will continue where the previous run left off
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
  -r, --recursive                       walk directories recursively
      --source-file-list string         a file containing a list of files to process, one file per line
      --source-filesystem string        the source filesystem to use for input files. Default is to use OS file system. Legal values:
                                          ftp://user/pass@host:port
                                          tar://path/to/archive.tar
                                          tgz://path/to/archive.tar.gz
                                        
      --suffixes strings                filter files by suffixes (default [.warc,.warc.gz])
  -s, --symlinks                        follow symlinks
      --warc-dir string                 output directory for validated warc files. If not empty this enables copying of input file. Directory must exist.
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

