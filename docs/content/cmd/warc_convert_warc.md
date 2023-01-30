---
date: 2023-01-30T10:17:46+01:00
title: "warc convert warc"
slug: warc_convert_warc
url: /cmd/warc_convert_warc/
---
## warc convert warc

Convert WARC file into WARC file

### Synopsis

The WARC to WARC converter can be used to reorganize, convert or repair WARC-records.
This is an experimental feature.

```
warc convert warc <files/dirs> [flags]
```

### Options

```
  -z, --compress                 use gzip compression for WARC files
      --compression-level        the gzip compression level to use (value between 1 and 9)
  -c, --concurrency int          number of input files to process simultaneously. The default value is 1.5 x <number of cpu cores> (default 24)
  -C, --concurrent-writers int   maximum concurrent WARC writers. This is the number of WARC-files simultaneously written to.
                                 A consequence is that at least this many WARC files are created even if there is only one input file. (default 1)
  -t, --default-date string      fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date (default "2023-1-30")
  -S, --file-size int            The maximum size for WARC files (default 1073741824)
      --flush                    if true, sync WARC file to disk after writing each record
  -h, --help                     help for warc
  -i, --index-dir string         directory to store indexes (default "/home/johnh/.cache/warc")
  -k, --keep-index               true to keep index on disk so that the next run will continue where the previous run left off
      --min-free-disk string     minimum free space on disk to allow WARC writing (default "256MB")
  -n, --name-generator string    the name generator to use. By setting this to 'identity', the input filename will also be used as
                                 output file name (prefix and suffix might still change). In this mode exactly one file is generated for every input file (default "default")
  -K, --new-index                true to start from a fresh index, deleting eventual index from last run
  -p, --prefix string            filename prefix for WARC files
  -r, --recursive                walk directories recursively
  -R, --repair                   try to fix errors in records
      --subdir-pattern string    a pattern to use for generating subdirectories.
                                 / in pattern separates subdirectories on all platforms
                                 {YYYY} is replaced with a 4 digit year
                                 {YY} is replaced with a 2 digit year
                                 {MM} is replaced with a 2 digit month
                                 {DD} is replaced with a 2 digit day
                                 The date used is the WARC date of each record. Therefore a input file might be split into 
                                 WARC files in different subdirectories. If NameGenerator is 'identity' only the first record
                                 of each file's date is used to keep the file as one.
      --suffixes strings         filter files by suffixes (default [.warc,.warc.gz])
  -s, --symlinks                 follow symlinks
  -w, --warc-dir string          output directory for generated warc files. Directory must exist. (default ".")
      --warc-version string      the WARC version to use for created files (default "1.1")
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

* [warc convert](../warc_convert/)	 - Convert web archives to warc files. Use subcommands for the supported formats

