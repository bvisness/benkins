package main

import (
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"

	"github.com/frc-2175/roboci/pkg/project"
	"github.com/frc-2175/roboci/pkg/runners"

	"io/ioutil"
)

type Config struct {
	Projects map[string]project.Project `toml:"projects"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "roboci",
		Short: "A simple CI server",
		Long:  `A simple CI server`,
		Run: func(cmd *cobra.Command, args []string) {
			configFile, _ := cmd.PersistentFlags().GetString("config")

			configDoc, err := ioutil.ReadFile(configFile)
			if err != nil {
				panic(err)
			}

			var config Config
			toml.Unmarshal(configDoc, &config)

			//for _, p := range config.Projects {
			//	p.Run("")
			//}

			server := runners.RunnerServer{}
			server.Boot()

			//server.Boot(config)
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file to load project info from")
	rootCmd.MarkFlagRequired("config")

	rootCmd.Execute()
}
