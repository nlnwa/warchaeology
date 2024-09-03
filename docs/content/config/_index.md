---
title: Configuration
weight: 30
---

## Configuration parameters

Warchaeology commands can be configured by specifying parameters.
There are several options for specifying parameters where using command line flags
is the easiest. But if you find yourself always setting a specific flag it might be better
to add a configuration file or use environment variables.

Flags set on the command line takes precedence over configuration files and environment variables.

Parameter documentation can be found in the *options* section for each [command](/cmd).
The parameter name is the long flag name with the dashes removed.

## Environment variables

Environment variables can be used to set parameters. Use the following steps to convert
a parameter name to an environment variable name:

* converting the parameter name to upper case
* replace '-' with '_'
* prefix with `WARC_`

> Setting the environment variable **WARC_RECORD_COUNT=2** is equal to specify the flag `--record-count=2`.

Environment variables takes precedence over parameters in config files.

## Configuration File

Parameters can also be set in configuration files. The configuration file format is YAML.

### File structure

To set a configuration parameter use the parameter name as key and then the value:

```yaml
delimiter: "\t"
record-count: 2
```

### Config file location

The standard configuration files are named `config.yaml` and are searched for in
system default directories.

The directories are looked up in the following order:

1. Working directory
   * The directory warc was started from

2. Standard Global Configuration Paths
   * *Linux*: $XDG_CONFIG_DIRS or "/etc/xdg/warc"
   * *Windows*: %PROGRAMDATA% or "C:\\ProgramData/warc"
   * *macOS*: /Library/Application Support/warc

3. Standard User-Specific Configuration Paths
   * *Linux*: $XDG_CONFIG_HOME or "$HOME/.config/warc"
   * *Windows*: %APPDATA% or "C:\\Users\\%USER%\\AppData\\Roaming\\warc"
   * *macOS*: $HOME/Library/Application Support/warc

The file found first will be used.

By setting the command line flag `--config` to a file name, the user can override the default
config with a user specified config file.
