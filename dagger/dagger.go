package dagger

import (
	"fmt"
	"log"
	"strings"
)

type Depender struct {
}

type DependencyFormer interface {
	GetJobs() []*Job
	GetDependencies() []*Job
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

type Job struct {
	ID        int
	Name      string
	DependsOn []*Job
}

func (j *Job) String() string {
	return fmt.Sprintf("Job<%s>(%p)", j.Name, j)
}

//func MakeJobDAG(configs []interface{}) []*Job {
//	var jobConfigs []*JobConfig
//	var groupConfigs []*Group
//	seenConfigs := map[interface{}]struct{}{}
//	var listConfigs func(config interface{})
//	listConfigs = func(config interface{}) {
//		if _, seen := seenConfigs[config]; seen {
//			return // we already saw and processed this one
//		}
//
//		// TODO: Maybe a future interface can keep us from having to type-switch here.
//		switch config := config.(type) {
//		case *JobConfig:
//			jobConfigs = append(jobConfigs, config)
//			for _, dependency := range config.DependsOn {
//				listConfigs(dependency)
//			}
//		case *Group:
//			groupConfigs = append(groupConfigs, config)
//			for _, dependency := range config.Jobs {
//				listConfigs(dependency)
//			}
//			for _, dependency := range config.DependsOn {
//				listConfigs(dependency)
//			}
//		}
//
//		seenConfigs[config] = struct{}{}
//	}
//	for _, config := range configs {
//		listConfigs(config)
//	}
//
//	// Convert all job configs into jobs, without dependencies yet
//	var jobs []*Job
//	jobConfigsToJobs := map[*JobConfig]*Job{}
//	for _, jobConfig := range jobConfigs {
//		job := &Job{
//			Name:      jobConfig.Name,
//			DependsOn: nil, // will be filled out later
//		}
//
//		jobs = append(jobs, job)
//		jobConfigsToJobs[jobConfig] = job
//	}
//
//	// Go through all jobs and add inter-job dependencies
//	//for _, jobConfig := range jobConfigs {
//	//	jobConfig.DependsOn
//	//}
//
//	return jobs
//}

func MakeJobDAG2(config Config) []*Job {
	log.Printf("%+v", config.Jobs)
	log.Printf("%+v", config.Groups)

	for _, group := range config.Groups {
		for _, depGroup := range group.DependsOn {
			for _, thisJob := range group.Jobs {
				for _, thatJob := range depGroup.Jobs {
					thisJob.DependsOn = append(thisJob.DependsOn, thatJob)
				}
			}
		}
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

	log.Print(o.String())

	return o.String()
}
