#!/bin/sh

set -e
rm --recursive --force completions
mkdir completions
for sh in bash zsh fish; do
	go run main.go completion "$sh" >"completions/warc.$sh"
done
