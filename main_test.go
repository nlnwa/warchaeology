package main_test

import (
	"os"
	"testing"

	"github.com/nlnwa/warchaeology/cmd"
)

func TestEmptyCommandPrompt(t *testing.T) {
	os.Stdout = nil
	os.Stderr = nil

	shell := cmd.NewCommand()
	_ = shell.Execute()

}
