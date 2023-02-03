---
date: 2023-02-03T12:57:12+01:00
title: "warc completion"
slug: warc_completion
url: /cmd/warc_completion/
---
## warc completion

Generate completion script

### Synopsis

To load completions:

Bash:

    $ source <(warc completion bash)

    # To load completions for each session, execute once:
    # Linux:
    $ warc completion bash > /etc/bash_completion.d/warc
    # macOS:
    $ warc completion bash > /usr/local/etc/bash_completion.d/warc

Zsh:

    # If shell completion is not already enabled in your environment,
    # you will need to enable it.  You can execute the following once:

    $ echo "autoload -U compinit; compinit" >> ~/.zshrc

    # To load completions for each session, execute once:
    $ warc completion zsh > "${fpath[1]}/_warc"

    # You will need to start a new shell for this setup to take effect.

fish:

    $ warc completion fish | source

    # To load completions for each session, execute once:
    $ warc completion fish > ~/.config/fish/completions/warc.fish

PowerShell:

    PS> warc completion powershell | Out-String | Invoke-Expression

    # To load completions for every new session, run:
    PS> warc completion powershell > warc.ps1
    # and source this file from your PowerShell profile.


```
warc completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config string          config file. If not set, /etc/warc/, $HOME/.warc/ and current working dir will be searched for file config.yaml
      --log-console strings    the kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings       the kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string   a file to write log output. Empty for no log file
      --tmpdir string          directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc](../warc/)	 - A tool for handling warc files

