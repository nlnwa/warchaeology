---
title: Configuration
weight: 30
---

## Configuration parameters

Warchaeology commands can be configured by specifying parameters.
There are several options for specifying parameters where using command line flags
is the easiest. But if you find yourself always setting a specific flag it might be better
to add a configuration file or environment variable.

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

#### File structure

To set a configuration parameter use the parameter name as key and then the value:

```yaml
delimiter: "\t"
record-count: 2
```

If you want to have a global default, but override the parameter for a specific command
you can do so by adding a section with the command as key.

```yaml
delimiter: "\t"
record-count: 2
ls:
  record-count: 5
convert:
  tmpdir: mydir
  arc:
    tmpdir: anotherdir
```
This config file gives the following values

{{< table style="table-striped" >}}
| Command           | parameter name | parameter value |
|-------------------|----------------|-----------------|
| warc cat          | record-count   | 2               |
| warc ls           | record-count   | 5               |
| warc ls           | tmpdir         | /tmp (default)  |
| warc convert warc | tmpdir         | mydir           |
| warc convert arc  | tmpdir         | anotherdir      |
{{< /table >}}

#### Config file location

The standard configuration files are named `config.yaml` and are searched for in
system default directories.

The directories are looked up in the following order:

1. Standard Global Configuration Paths
   * _Linux_: $XDG_CONFIG_DIRS or "/etc/xdg/warc"
   * _Windows_: %PROGRAMDATA% or "C:\\ProgramData/warc"
   * _macOS_: /Library/Application Support/warc

2. Standard User-Specific Configuration Paths
   * _Linux_: $XDG_CONFIG_HOME or "$HOME/.config/warc"
   * _Windows_: %APPDATA% or "C:\\Users\\%USER%\\AppData\\Roaming\\warc"
   * _macOS_: $HOME/Library/Application Support/warc

3. Working directory
   * The directory warc was started from

All steps are searched for a file named `config.yaml` and if found,
values in a later file will override values in the files before it.

By setting the command line flag `--config` to a file name, the user can override the default
config with a user specified config file.
