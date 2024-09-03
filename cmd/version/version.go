package version

import (
	"encoding/json"
	"fmt"

	"github.com/nlnwa/warchaeology/v3/internal/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type VersionOptions struct {
	output string
}

func NewCmdVersion() *cobra.Command {
	o := &VersionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show extended version information",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "", "One of 'yaml' or 'json'")

	return cmd
}

func (o *VersionOptions) Run() error {
	versionInfo := version.Version

	switch o.output {
	case "":
		fmt.Print(versionInfo)
	case "yaml":
		v, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}
		fmt.Println(string(v))
	case "json":
		v, err := json.Marshal(&versionInfo)
		if err != nil {
			return err
		}
		fmt.Println(string(v))
	default:
		return fmt.Errorf("unsupported output format: %s", o.output)
	}
	return nil
}
