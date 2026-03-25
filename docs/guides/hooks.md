---
title: Hooks
---

Hooks are executable commands that run before or after input/output file handling.

- `OpenInputFileHook`
- `CloseInputFileHook`
- `OpenOutputFileHook`
- `CloseOutputFileHook`

Hook context is passed through environment variables such as:

- `WARC_COMMAND`
- `WARC_HOOK_TYPE`
- `WARC_FILE_NAME`
- `WARC_SRC_FILE_NAME` (output hooks)
- `WARC_SIZE`, `WARC_INFO_ID`, `WARC_HASH`, `WARC_ERROR_COUNT` (close hooks when available)

A non-zero exit code is treated as an error.
