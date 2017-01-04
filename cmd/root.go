// Copyright © 2017 HANAI Tohru <tohru@reedom.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	version  string
	revision string
	initErr  error
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "refcode",
	Short: "Reference Code Mapper utility",
	Long: `refcode is a CLI library for programmers that maps reference codes
and embbeds them into your source code files.
The reference codes are for the users of your project; a reference code
will be shown to the user when any error happens. The user can inquiry about
the error with the code. And you can research find the place where the error
occuured by the code.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("v%v\n", version)
			return nil
		}

		if initErr != nil {
			return initErr
		}

		cmd.Usage()
		return nil
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Command.PersistentFlags() is for command-global flag settings.
	// Command.Flags() is for command-local flag settings.
	// viper.BindPFlag() will overwrite viper config entry by the specified command flag.

	// [global]
	// | arg long  | short | config           |
	// |-----------|-------|------------------|
	// | config    | c     |                  |
	// | codespace | z     | codespace        |
	// |           |       | dataDir          |
	// | endpoint  | u     | remote.endpoint  |
	// | secretKey | k     | remote.secretKey |
	//
	// [local]
	// | arg long  | short | config           |
	// |-----------|-------|------------------|
	// | version   | v     |                  |

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	curdir, err := os.Getwd()
	if err != nil {
		initErr = err
		return
	}
	defaultCodespace := filepath.Base(curdir)
	RootCmd.PersistentFlags().StringP("codespace", "z", defaultCodespace,
		"codespace, reference code are managed under each codespace separately")
	viper.BindPFlag("codespace", RootCmd.PersistentFlags().Lookup("codespace"))

	viper.SetDefault("dataDir", "~/.refcode")

	RootCmd.PersistentFlags().StringP("endpoint", "u", "",
		"refcode management server endpoint (ex. http://localhost:4454/)")
	viper.BindPFlag("remote.endpoint", RootCmd.PersistentFlags().Lookup("endpoint"))

	RootCmd.PersistentFlags().StringP("secret", "k", "",
		"refcode management server secret key")
	viper.BindPFlag("remote.secretKey", RootCmd.PersistentFlags().Lookup("secret"))

	RootCmd.Flags().BoolP("version", "v", false, "show version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	if cfgFile != "" { // enable ability to specify config file via flag
		if cfgFile, err = homedir.Expand(cfgFile); err == nil {
			viper.SetConfigFile(cfgFile)
		}
	}

	viper.SetConfigName(".refcode") // name of config file (without extension)
	viper.AddConfigPath("$HOME")    // adding home directory as first search path
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("REFCODE")

	// If a config file is found, read it in.
	if err = viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else if viper.ConfigFileUsed() != "" {
		fmt.Fprintln(os.Stderr, "Error: reading config file", viper.ConfigFileUsed())
		initErr = err
		return
	}

	dataDir, err := homedir.Expand(strings.Replace(viper.GetString("dataDir"), "$HOME", "~", 1))
	if err != nil {
		initErr = err
		return
	}
	viper.Set("dataDir", dataDir)
}
