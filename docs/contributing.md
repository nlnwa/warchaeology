---
title: Contributing
---

## Build

```shell
go build -o warc
```

## Generate release artifacts locally

```shell
goreleaser release --clean --snapshot
```

## Generate CLI docs

```shell
go generate ./docs
```

## Run docs site locally

```shell
cd website
npm install
npm run start
```
