---
title: Hooks
weight: 40
---

Warchaeology provides the ability to add hooks - executable files that run before and after opening/closing input and output files.

Hooks allow users to add custom logic and functionality to the file handling process and can be used to perform any necessary setup or validation before a file is accessed by the tool, or to perform any cleanup or post-processing tasks.

### Execution

> Please note that the executable files used as hooks should be able to run without any user interaction.

They should also exit with a status code of 0 to indicate success. Any other exit status will be considered as an error by the tool. See details below for the exit codes used by each hook type.

The executable file must either be in the system's path or have a valid path. The program won't find a file using a path relative to the current directory. To use a program in the current directory, use `./program` or the full path.

Context for hooks is provided through environment variables, which are made available to the executable files. See details below for the environment variables available to each hook type.

## Types of hooks

### *OpenInputFileHook*

The `OpenInputFileHook` is an executable file that runs before an input file is opened.
This hook can be used to perform any necessary setup or validation before the file is accessed by the tool.
It can also be used to skip the file if it does not meet certain criteria.

To use this hook, specify the path to the executable file using the `--open-input-file-hook` flag when running the tool.

#### Environment Variables

* `WARC_COMMAND`: Contains the subcommand name.
* `WARC_HOOK_TYPE`: Contains the hook type: OpenInputFile.
* `WARC_FILE_NAME`: Contains the file name of the input file.

#### Exit codes

     1: The hook should exit with a status of 1 in case of an error.
    10: The hook should exit with a status of 10 to indicate that the file should be skipped.

### *CloseInputFileHook*

The `CloseInputFileHook` is an executable file that runs after an input file has been fully read and has been closed.
This hook can be used to perform any cleanup or post-processing tasks.

To use this hook, specify the path to the executable file using the `--close-input-file-hook` flag when running the tool.

#### Environment Variables

* `WARC_COMMAND`: Contains the subcommand name.
* `WARC_HOOK_TYPE`: Contains the hook type: CloseInputFile.
* `WARC_FILE_NAME`: Contains the file name of the input file.
* `WARC_ERROR`: Contains the error message if there was an error processing the file.
* `WARC_ERROR_COUNT`: Contains the number of errors found processing the file.
* `WARC_HASH`: Contains the hash of the input file if it was computed.

#### Exit codes

     1: The hook should exit with a status of 1 in case of an error.

### *OpenOutputFileHook*

The `OpenOutputFileHook` is an executable file that runs before an output file is opened for writing.
This hook can be used to perform any necessary setup or validation before the file is accessed by the tool.

To use this hook, specify the path to the executable file using the `--open-output-file-hook` flag when running the tool.

#### Environment Variables

* `WARC_COMMAND`: Contains the subcommand name.
* `WARC_HOOK_TYPE`: Contains the hook type: OpenOutputFile.
* `WARC_FILE_NAME`: Contains the file name of the output file.
* `WARC_SRC_FILE_NAME`: Contains the file name of the input file from which the output file is being created.

#### Exit codes

     1: The hook should exit with a status of 1 in case of an error.

### *CloseOutputFileHook*

The `CloseOutputFileHook` is an executable file that runs after an output file has been fully written to and has been closed.
This hook can be used to perform any cleanup or post-processing tasks.

To use this hook, specify the path to the executable file using the `--close-output-file-hook` flag when running the tool.

#### Environment Variables

* `WARC_COMMAND`: Contains the subcommand name.
* `WARC_HOOK_TYPE`: Contains the hook type: CloseOutputFile.
* `WARC_FILE_NAME`: Contains the file name of the output file.
* `WARC_SRC_FILE_NAME`: Contains the file name of the input file from which the output file was created.
* `WARC_SIZE`: Contains the size of the output file in bytes.
* `WARC_INFO_ID`: Contains the ID of the WarcInfo-record in the output file if it exists.
