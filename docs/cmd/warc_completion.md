## warc completion

Generate completion script

### Synopsis

To load completions:

Bash:

  $ source <(yourprogram completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ yourprogram completion bash > /etc/bash_completion.d/yourprogram
  # macOS:
  $ yourprogram completion bash > /usr/local/etc/bash_completion.d/yourprogram

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ yourprogram completion zsh > "${fpath[1]}/_yourprogram"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ yourprogram completion fish | source

  # To load completions for each session, execute once:
  $ yourprogram completion fish > ~/.config/fish/completions/yourprogram.fish

PowerShell:

  PS> yourprogram completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> yourprogram completion powershell > yourprogram.ps1
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
      --log-console strings    The kind of log output to write to console. Valid values: info, error, summary, progress (default [progress,summary])
      --log-file strings       The kind of log output to write to file. Valid values: info, error, summary (default [info,error,summary])
  -L, --log-file-name string   a file to write log output. Empty for no log file
      --tmpdir string          directory to use for temporary files (default "/tmp")
```

### SEE ALSO

* [warc](warc.md)	 - A tool for handling warc files

