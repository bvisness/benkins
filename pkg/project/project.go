package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	git "gopkg.in/src-d/go-git.v4"

	"gopkg.in/src-d/go-git.v4/plumbing"
	yaml "gopkg.in/yaml.v2"
)

type Project struct {
	URL string `toml:"url"`
}

type CIConfig struct {
	Stages []struct {
		Name string `yaml:"name"`
		Jobs []Job  `yaml:"jobs"`
	} `yaml:"stages"`
}

type Job struct {
	Name   string   `yaml:"name"`
	Script []string `yaml:"script"`
}

type JobResults struct {
	Success bool
}

func (p *Project) Run(hash string) {
	cfg, _ := p.getCIConfig(hash)

	for _, stage := range cfg.Stages {
		fmt.Printf("Running stage %s\n", stage.Name)

		var wg sync.WaitGroup

		for _, job := range stage.Jobs {
			wg.Add(1)
			go func() {
				p.runJob(job, hash)
				wg.Done()
			}()
		}

		wg.Wait()
	}
}

func (p *Project) getCIConfig(hash string) (CIConfig, error) {
	dir, cleanup := temporaryCheckout(p.URL, hash, "project")
	defer cleanup()

	files, _ := ioutil.ReadDir(dir)

	if ciFile := getFile(".roboci.yml", files); ciFile != nil {
		ciFileText, _ := ioutil.ReadFile(filepath.Join(dir, ciFile.Name()))

		var ciConfig CIConfig
		yaml.Unmarshal(ciFileText, &ciConfig)

		fmt.Printf("%d stages:\n", len(ciConfig.Stages))
		for _, stage := range ciConfig.Stages {
			fmt.Printf("- %s (%d jobs)\n", stage.Name, len(stage.Jobs))
		}

		return ciConfig, nil
	} else {
		fmt.Println("No .roboci.yml found.")
		return CIConfig{}, nil
	}
}

func (p *Project) runJob(job Job, hash string) JobResults {
	_, cleanup := temporaryCheckout(p.URL, hash, fmt.Sprintf("job-%s", job.Name))
	defer cleanup()

	result := JobResults{}

	for _, cmd := range job.Script {
		_, err := runCommand(cmd, false)
		if err != nil {
			fmt.Printf("Error running shell command: %v\n", err)
			result.Success = false
			break
		}
	}

	return result
}

func temporaryCheckout(url, hash, uniqueName string) (dir string, cleanup func()) {
	tmpdir, _ := ioutil.TempDir("", uniqueName)

	r, _ := git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	wt, _ := r.Worktree()
	wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(hash),
	})

	return tmpdir, func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			panic(err)
		}
	}
}

func getFile(name string, files []os.FileInfo) os.FileInfo {
	for _, file := range files {
		if file.Name() == name {
			return file
		}
	}

	return nil
}
