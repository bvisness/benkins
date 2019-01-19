package main

import (
	"github.com/frc-2175/roboci/pkg/runner"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"

	"github.com/frc-2175/roboci/pkg/server"

	"io/ioutil"
)

var configFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "roboci",
		Short: "A simple CI server",
		Long:  `A simple CI server`,
		Run: func(cmd *cobra.Command, args []string) {
			configDoc, err := ioutil.ReadFile(configFile)
			if err != nil {
				panic(err)
			}

			var config server.ServerConfig
			toml.Unmarshal(configDoc, &config)

			server := runner.RunnerServer{}

			server.Boot()

			//server.Boot(config)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file to load project info from")
	rootCmd.MarkFlagRequired("config")

	rootCmd.Execute()
}
