package main

import (
	"fmt"
	"io"

	"github.com/frc-2175/roboci/runnerserver"
	"github.com/frc-2175/roboci/server"
	"github.com/spf13/cobra"

	"io/ioutil"
)

import toml "github.com/pelletier/go-toml"

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

			server := runnerserver.RunnerServer{
				Routes: map[string]runnerserver.RequestHandler{
					"foo": func(r *runnerserver.Request) {
						for {
							body, err := r.ReadBody()
							if err == io.EOF {
								fmt.Printf("client closed connection\n")
								break
							}
							if err != nil {
								fmt.Printf("handler error: %v\n", err)
								break
							}

							fmt.Printf("from handler: %s\n", string(body))
						}
					},
				},
			}

			server.Boot()

			//server.Boot(config)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file to load project info from")
	rootCmd.MarkFlagRequired("config")

	rootCmd.Execute()
}
