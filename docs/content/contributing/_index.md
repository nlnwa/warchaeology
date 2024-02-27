---
title: Contributing
weight: 50
---

## Getting Started

Warchaeology is written in [go](https://go.dev) and requires v1.21 or greater to build.

#### Clone the repository
``` sh
$ git clone https://github.com/nlnwa/warchaeology.git
$ cd warchaeology
```

#### Compile
``` sh
$ go -o warc build .
```

#### Generate dist files
This step is automatically run by Github actions when a release is done. The description here is only for testing
goreleaser configuration.

Install [goreleaser](https://goreleaser.com/). Then run:

``` sh
$ goreleaser release --snapshot --rm-dist
```

## Documentation

Documentation is generated with [Hugo](https://gohugo.io/). A local installation is not needed unless
you want to view generated documentation on localhost. 

Start a webserver serving documentation on [http://localhost:1313/](http://localhost:1313/)
``` sh
$ script/doc/serve.sh
```

Source files for documentation are located in `docs/content`. The files in `docs/content/cmd` (except `_index.md`)
are generated from the src code and should not be edited.

## Release

Before doing a release, ensure that your local repository is on the main branch and is up to date.
``` sh
$ git checkout main
$ git pull
```

Run `script/release/create.sh` with the release version as parameter. The version should NOT contain a 'v' prefix.
For example to release v1.0.2, run:
``` sh
$ script/release/create.sh 1.0.2
```

This will recreate the command line documentation, commit it and tag the release. The next thing you need to do is to
push the release.
``` sh
$ git push https://github.com/nlnwa/warchaeology.git v1.0.2
```
