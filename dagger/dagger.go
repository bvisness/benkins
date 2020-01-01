package dagger

import (
	"fmt"
	"strings"
)

type Job struct {
	ID        int
	Name      string
	DependsOn []*Job
}

func (j *Job) String() string {
	return fmt.Sprintf("Job<%s>(%p)", j.Name, j)
}

type Group struct {
	Name      string
	Jobs      []*Job
	DependsOn []*Group
}

func (g *Group) String() string {
	var children []string
	for _, job := range g.Jobs {
		children = append(children, fmt.Sprintf("%v", job))
	}

	var deps []string
	for _, dep := range g.DependsOn {
		deps = append(deps, fmt.Sprintf("%v", dep))
	}

	return fmt.Sprintf("Group<%s>(%s)(%s)", g.Name, strings.Join(children, ", "), strings.Join(deps, ", "))
}

func MakeJobDAG(config Config) []*Job {
	for _, group := range config.Groups {
		for _, depGroup := range group.DependsOn {
			for _, thisJob := range group.Jobs {
				for _, thatJob := range depGroup.Jobs {
					thisJob.DependsOn = append(thisJob.DependsOn, thatJob)
				}
			}
		}
	}

	// remove redundant edges
	for _, job := range config.Jobs {
		var deduped []*Job
		deps := map[*Job]struct{}{}
		for _, dep := range job.DependsOn {
			if _, dupe := deps[dep]; dupe {
				continue
			}

			deduped = append(deduped, dep)
			deps[dep] = struct{}{}
		}

		job.DependsOn = deduped
	}

	return config.Jobs
}

func GetJobsGraphviz(jobs []*Job) string {
	var o strings.Builder

	jobKey := func(j *Job) string {
		if j.Name != "" {
			return j.Name
		} else {
			return fmt.Sprintf("%d", j.ID)
		}
	}

	o.WriteString("digraph {\n")
	o.WriteString("rankdir = BT\n")
	for _, this := range jobs {
		o.WriteString(fmt.Sprintf("\"%s\"\n", jobKey(this)))
		for _, that := range this.DependsOn {
			o.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", jobKey(this), jobKey(that)))
		}
	}
	o.WriteString("}\n")

	return o.String()
}
