package shared

import (
	"github.com/pelletier/go-toml"
)

const (
	ExecutionLogFilename = "benkins-execution-log.txt"
	ResultsFilename      = "benkins-results.toml"
	NotificationFilename = "benkins-notification.txt"
)

type JobResults struct {
	Success       bool
	CommitMessage string
	BranchName    string
}

func (r JobResults) ToTOML() string {
	rBytes, _ := toml.Marshal(r)

	return string(rBytes)
}
