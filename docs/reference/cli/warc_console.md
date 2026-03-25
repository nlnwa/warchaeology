---
title: "warc console"
---
## warc console

A shell for working with WARC files

```
warc console DIR/FILE [flags]
```

### Options

```
  -h, --help               help for console
      --suffixes strings   only process files with these suffixes (default [.warc,.warc.gz])
      --tmp-dir string     directory used for temporary files (default "/tmp")
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

