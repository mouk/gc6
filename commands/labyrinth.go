// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

// Package commands defines and implements command-line commands
// and flags used by laybrinth.
// Commands and flags are implemented using cobra.

package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AuthorName = "Moukarram Kabbash"

var CfgFile string

// Defining the daedalus command.
// This will be called as 'laybrinth'
// The default behavior will be to run both a server (daedalus)
// and a client (icarus) and connect them to each other.
var RootCmd = &cobra.Command{
	Use:   "labyrinth",
	Short: "a labyrinth generator and solver",
	Long: `In Greek mythology, Daedalus was a skillful craftsman and
artist. He is the father of Icarus, and the creator of the Labyrinth.

Daedalus's job is to create a challenging Labyrinth for his opponent
(shadow Icarus) to solve.

Icarus wakes up to find himself in the middle of a Labyrinth. Due to
the darkness of the Labyrinth he can only see his immediate cell and
if there is a wall or not to the top, right, bottom and left. He takes
one step and then can discover if his new cell has walls on each of
the four sides.`,
	Run: func(cmd *cobra.Command, args []string) {
		go RunServer()

		// give server time to start before sending a request.
		// There's a better way to do this, but I'm lazy and this is just for fun.
		time.Sleep(1 * time.Second)

		RunIcarus()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Setting flags here so they can be used by both the root behavior as well as
	// by the indidual behaviors of icarus and daedalus
	RootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is $CWD/config.yaml)")
	RootCmd.PersistentFlags().IntP("port", "p", 8013, "Port run on")
	RootCmd.PersistentFlags().IntP("width", "x", 15, "width of the laybrinth")
	RootCmd.PersistentFlags().IntP("height", "y", 10, "height of the laybrinth") // 'h' is used for help already
	RootCmd.PersistentFlags().IntP("times", "t", 1, "times to solve the laybrinth")
	RootCmd.PersistentFlags().IntP("max-steps", "m", 500, "Maximum steps before giving up")

	// Bind viper to these flags so viper can read flag values along with config, env, etc.
	viper.BindPFlag("width", RootCmd.PersistentFlags().Lookup("width"))
	viper.BindPFlag("height", RootCmd.PersistentFlags().Lookup("height"))
	viper.BindPFlag("port", RootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("times", RootCmd.PersistentFlags().Lookup("times"))
	viper.BindPFlag("max-steps", RootCmd.PersistentFlags().Lookup("max-steps"))
}

// Read in config file and ENV variables if set.
func initConfig() {
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	}

	dir, _ := os.Getwd()
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(dir)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config.yaml file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

//Execute adds all child commands to the root command Labyrinth and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
