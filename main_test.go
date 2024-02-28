package main_test

import (
	"testing"

	"github.com/nlnwa/warchaeology/cmd"
)

func TestEmptyCommandPrompt(t *testing.T) {
	shell := cmd.NewCommand()
	_ = shell.Execute()
}
