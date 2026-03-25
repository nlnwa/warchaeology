---
title: Configuration
---

Warchaeology commands can be configured through flags, environment variables, and config files.

Flag values override environment variables and config-file values.

## Environment variables

To convert a parameter name into an environment variable:

1. Convert the parameter name to uppercase.
2. Replace `-` with `_`.
3. Prefix with `WARC_`.

Example: `record-count` becomes `WARC_RECORD_COUNT`.

## Configuration file

Config files use YAML.

```yaml
delimiter: "\t"
record-count: 2
```

### Search order for `config.yaml`

1. Current working directory
2. Standard global config locations
3. Standard user-specific config locations

To use a specific file, pass `--config <path>`.
