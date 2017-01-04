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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mapCmd represents the map command
var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Map reference code into source code files.",
	Long: `Traverse source code files and replace refcode skeleton string
with real reference code.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initErr
	},

	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	RootCmd.AddCommand(mapCmd)

	// Command.PersistentFlags() is for command-global flag settings.
	// Command.Flags() is for command-local flag settings.
	// viper.BindPFlag() will overwrite viper config entry by the specified command flag.

	// [local]
	// | arg long         | short | config                |
	// |------------------|-------|-----------------------|
	// | dryrun           | n     |                       |
	// | pattern          | p     | mapper.pattern        |
	// | replace          | r     | mapper.replace        |
	// | includes         | i     | files.includes        |
	// | excludes         | x     | files.excludes        |
	// | global-gitignore |       | files.globalGitIgnore |
	// | symlink          |       | files.followSymlink   |

	mapCmd.Flags().BoolP("dryrun", "n", false, "dry run")

	mapCmd.Flags().StringP("pattern", "p", "", "source pattern in regular expression")
	viper.BindPFlag("mapper.pattern", mapCmd.Flags().Lookup("pattern"))

	mapCmd.Flags().StringP("replare", "r", "", "replace template")
	viper.BindPFlag("mapper.replace", mapCmd.Flags().Lookup("replace"))

	mapCmd.Flags().StringSliceP("includes", "i", nil, "include file patterns, delimited by comma(,)")
	// viper.BindPFlag("files.includes", mapCmd.Flags().Lookup("includes"))

	mapCmd.Flags().StringSliceP("excludes", "x", nil, "exclude file patterns, this exceeds includes")
	viper.BindPFlag("files.excludes", mapCmd.Flags().Lookup("excludes"))

	mapCmd.Flags().BoolP("global-gitignore", "", false, "whether apply ~/.gitignore to file include/exclude pattern")
	viper.BindPFlag("files.globalGitIgnore", mapCmd.Flags().Lookup("global-gitignore"))

	mapCmd.Flags().BoolP("symlink", "", false, "whether follow symlink")
	viper.BindPFlag("files.followSymlink", mapCmd.Flags().Lookup("symlink"))
}
