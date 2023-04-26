/*
 * Copyright 2023 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package config

import (
	"fmt"
	"github.com/kirsle/configdir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"testing"
)

type config struct {
	k1 string
	k2 string
	k3 string
	k4 string
	k5 string
}
type env struct {
	key string
	val string
}

func TestCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdLine  []string
		env      []env
		expected config
	}{
		{
			name:    "root_plain",
			cmdLine: []string{"root"},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5flagdefault",
			},
		},
		{
			name:    "root_flag",
			cmdLine: []string{"root", "--k5", "v5flag"},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5flag",
			},
		},
		{
			name:    "root_env",
			cmdLine: []string{"root"},
			env:     []env{{"K5", "v5env"}},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5env",
			},
		},
		{
			name:    "root_file",
			cmdLine: []string{"root", "--config", "custom_config.yaml"},
			expected: config{
				"v1flagdefault",
				"v2flagdefault",
				"v3f4l1",
				"v4f4l1",
				"v5flagdefault",
			},
		},
		{
			name:    "root_file_flag",
			cmdLine: []string{"root", "--config", "custom_config.yaml", "--k5", "v5flag"},
			expected: config{
				"v1flagdefault",
				"v2flagdefault",
				"v3f4l1",
				"v4f4l1",
				"v5flag",
			},
		},
		{
			name:    "root_file_env",
			cmdLine: []string{"root", "--config", "custom_config.yaml"},
			env:     []env{{"K5", "v5env"}},
			expected: config{
				"v1flagdefault",
				"v2flagdefault",
				"v3f4l1",
				"v4f4l1",
				"v5env",
			},
		},
		{
			name:    "root_env_flag",
			cmdLine: []string{"root", "--k5", "v5flag"},
			env:     []env{{"K5", "v5env"}},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5flag",
			},
		},
		{
			name:    "root_help",
			cmdLine: []string{"root", "-h"},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5flagdefault",
			},
		},
		{
			name:    "root_flag_help",
			cmdLine: []string{"root", "--k5", "v5flag", "-h"},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5flag",
			},
		},
		{
			name:    "root_env_help",
			cmdLine: []string{"root", "-h"},
			env:     []env{{"K5", "v5env"}},
			expected: config{
				"v1f1l1",
				"v2f2l1",
				"v3f3l1",
				"v4f3l1",
				"v5env",
			},
		},
		{
			name:    "root_file_help",
			cmdLine: []string{"root", "--config", "custom_config.yaml", "-h"},
			expected: config{
				"v1flagdefault",
				"v2flagdefault",
				"v3f4l1",
				"v4f4l1",
				"v5flagdefault",
			},
		},
		{
			name:    "sub_plain",
			cmdLine: []string{"root", "sub1"},
			expected: config{
				"v1f1l2",
				"v2f2l1",
				"v3f3l2",
				"v4f3l1",
				"v5flagdefault",
			},
		},
		{
			name:    "sub_flag",
			cmdLine: []string{"root", "sub1", "--k5", "v5flag"},
			expected: config{
				"v1f1l2",
				"v2f2l1",
				"v3f3l2",
				"v4f3l1",
				"v5flag",
			},
		},
		{
			name:    "sub_file",
			cmdLine: []string{"root", "sub1", "--config", "custom_config.yaml"},
			expected: config{
				"v1flagdefault",
				"v2f4l2",
				"v3f4l1",
				"v4f4l1",
				"v5flagdefault",
			},
		},
		{
			name:    "sub_sub_plain",
			cmdLine: []string{"root", "sub1", "sub_sub1"},
			expected: config{
				"v1f1l2",
				"v2f2l3",
				"v3f3l2",
				"v4f3l1",
				"v5flagdefault",
			},
		},
		{
			name:    "sub_sub_flag",
			cmdLine: []string{"root", "sub1", "sub_sub1", "--k5", "v5flag"},
			expected: config{
				"v1f1l2",
				"v2f2l3",
				"v3f3l2",
				"v4f3l1",
				"v5flag",
			},
		},
		{
			name:    "sub_sub_file",
			cmdLine: []string{"root", "sub1", "sub_sub1", "--config", "custom_config.yaml"},
			expected: config{
				"v1flagdefault",
				"v2f4l2",
				"v3f4l1",
				"v4f4l1",
				"v5flagdefault",
			},
		},
	}

	_ = os.Chdir("testdata")
	wd, _ := os.Getwd()
	_ = os.Setenv("XDG_CONFIG_DIRS", wd+"/etc")
	_ = os.Setenv("XDG_CONFIG_HOME", wd+"/home")
	configdir.Refresh()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			result := config{}
			os.Args = tt.cmdLine
			for _, e := range tt.env {
				_ = os.Setenv(e.key, e.val)
			}

			root := initializeCommands(result)
			InitConfig(root)

			_ = root.Execute()
			if e := validateKey("k1", tt.expected.k1); e != nil {
				t.Error(e)
			}
			if e := validateKey("k2", tt.expected.k2); e != nil {
				t.Error(e)
			}
			if e := validateKey("k3", tt.expected.k3); e != nil {
				t.Error(e)
			}
			if e := validateKey("k4", tt.expected.k4); e != nil {
				t.Error(e)
			}
			if e := validateKey("k5", tt.expected.k5); e != nil {
				t.Error(e)
			}
			for _, e := range tt.env {
				_ = os.Unsetenv(e.key)
			}
		})
	}
}

func validateKey(key, expectedValue string) error {
	v := viper.GetString(key)
	if v != expectedValue {
		return fmt.Errorf("Expected Get(%s) = '%s', got: '%s'", key, expectedValue, v)
	}
	return nil
}

func initializeCommands(result config) *cobra.Command {
	root := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			result.k1 = viper.GetString("k1")
			result.k2 = viper.GetString("k2")
			result.k3 = viper.GetString("k3")
			result.k4 = viper.GetString("k4")
			result.k5 = viper.GetString("k5")
			return nil
		},
	}

	// Flags
	root.PersistentFlags().String("k1", "v1flagdefault", "k1")
	root.PersistentFlags().String("k2", "v2flagdefault", "k2")
	root.PersistentFlags().String("k3", "v3flagdefault", "k3")
	root.PersistentFlags().String("k4", "v4flagdefault", "k4")
	root.PersistentFlags().String("k5", "v5flagdefault", "k5")

	// Subcommands
	sub1 := &cobra.Command{
		Use: "sub1",
		RunE: func(cmd *cobra.Command, args []string) error {
			result.k1 = viper.GetString("k1")
			result.k2 = viper.GetString("k2")
			result.k3 = viper.GetString("k3")
			result.k4 = viper.GetString("k4")
			result.k5 = viper.GetString("k5")
			return nil
		},
	}
	root.AddCommand(sub1)

	subsub1 := &cobra.Command{
		Use: "sub_sub1",
		RunE: func(cmd *cobra.Command, args []string) error {
			result.k1 = viper.GetString("k1")
			result.k2 = viper.GetString("k2")
			result.k3 = viper.GetString("k3")
			result.k4 = viper.GetString("k4")
			result.k5 = viper.GetString("k5")
			return nil
		},
	}
	sub1.AddCommand(subsub1)

	return root
}
