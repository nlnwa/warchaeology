---
title: Contributing
weight: 50
---

## Getting Started

Warchaeology is written in the [golang](https://go.dev) programming language.

### Clone the repository

```shell
git clone https://github.com/nlnwa/warchaeology.git
cd warchaeology
```

### Build

```shell
go build -o warc
```

## Generate distribution files

This step is automatically run by the CI pipeline when a release is done. The following description is for testing goreleaser configuration.

Install [goreleaser](https://goreleaser.com/). Then run:

```shell
goreleaser release --clean --snapshot
```

## Generate documentation

Documentation is generated with [Hugo](https://gohugo.io/). A local installation is not needed unless
you want to view generated documentation on localhost.

Source files for documentation are located in `docs/content`. The files in `docs/content/cmd` (except `_index.md`)
are generated from the source and should not be edited.

Generate documentation from source code by running:

```shell
go generate ./docs
```

Start a webserver serving documentation at [http://localhost:1313/](http://localhost:1313/)

```shell
./docs/serve.sh
```
