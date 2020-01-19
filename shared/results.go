package shared

import "fmt"

const (
	ExecutionLogFilename = "benkins-execution-log.txt"
	ResultsFilename      = "benkins-results.toml"
	NotificationFilename = "benkins-notification.txt"
)

type JobResults struct {
	Success    bool
	BranchName string
}

func (r JobResults) ToTOML() string {
	return fmt.Sprintf(""+
		"Success = %v\n"+
		"BranchName = \"%s\"\n",
		r.Success,
		r.BranchName,
	)
}
