package main

import (
	"io/ioutil"

	"github.com/frc-2175/benkins/app"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

type Config struct {
	Name           string
	ServerUrl      string
	Password       string
	SlackToken     string
	SlackChannelId string
	RepoUrl        string
}

func main() {
	var config Config

	if configBytes, err := ioutil.ReadFile("config.toml"); err == nil {
		err = toml.Unmarshal(configBytes, &config)
		if err != nil {
			panic(err)
		}
	}

	cmd := &cobra.Command{
		Use: "benkins-app",
		Run: func(cmd *cobra.Command, args []string) {
			app.Main(config.Name, config.ServerUrl, config.Password, config.SlackToken, config.SlackChannelId, config.RepoUrl)
		},
	}
	cmd.Flags().StringVar(&config.Name, "name", config.Name, "The name to use to identify this client")
	cmd.Flags().StringVar(&config.ServerUrl, "serverUrl", config.ServerUrl, "The url of the Benkins server")
	cmd.Flags().StringVar(&config.Password, "password", config.Password, "The Password used for client authentication")
	cmd.Flags().StringVar(&config.SlackToken, "slackToken", config.SlackToken, "The OAuth token for Slack")
	cmd.Flags().StringVar(&config.SlackChannelId, "slackChannelId", config.SlackChannelId, "The Slack channel ID (NOT the channel name)")
	cmd.Flags().StringVar(&config.RepoUrl, "repoUrl", config.RepoUrl, "The HTTPS URL of the Git repo to watch")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
