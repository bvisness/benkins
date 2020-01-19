package shared

import "fmt"

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
