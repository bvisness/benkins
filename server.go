package main

import (
	"os"

	"github.com/frc-2175/benkins/server"
	"github.com/spf13/cobra"
)

func main() {
	var basePath string
	var password string

	if p, ok := os.LookupEnv("BENKINS_PASSWORD"); ok {
		password = p
	}

	cmd := &cobra.Command{
		Use: "benkins-server",
		Run: func(cmd *cobra.Command, args []string) {
			server.Main(basePath, password)
		},
	}
	cmd.Flags().StringVar(&basePath, "basePath", basePath, "The path to serve all files from")
	cmd.Flags().StringVar(&password, "Password", password, "The Password used for client authentication")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
