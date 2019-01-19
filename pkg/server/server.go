package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	git "gopkg.in/src-d/go-git.v4"

	"gopkg.in/src-d/go-git.v4/plumbing"
	yaml "gopkg.in/yaml.v2"
)

type Project struct {
	URL string `toml:"url"`
}

type ServerConfig struct {
	Projects map[string]Project `toml:"projects"`
}

type CIConfig struct {
	Stages []struct {
		Name string  `yaml:"name"`
		Jobs []CIJob `yaml:"jobs"`
	} `yaml:"stages"`
}

type CIJob struct {
	Name   string   `yaml:"name"`
	Script []string `yaml:"script"`
}

func RunJob(project Project) {
	GetCIConfig(project, "")
}

func GetCIConfig(p Project, hash string) (CIConfig, error) {
	tmpdir, _ := ioutil.TempDir("", "project")
	defer os.RemoveAll(tmpdir)

	r, _ := git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:      p.URL,
		Progress: os.Stdout,
	})

	if hash != "" {
		wt, _ := r.Worktree()
		wt.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(hash),
		})
	}

	files, _ := ioutil.ReadDir(tmpdir)

	if ciFile := getFile(".roboci.yml", files); ciFile != nil {
		ciFileText, _ := ioutil.ReadFile(filepath.Join(tmpdir, ciFile.Name()))

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

func getFile(name string, files []os.FileInfo) os.FileInfo {
	for _, file := range files {
		if file.Name() == name {
			return file
		}
	}

	return nil
}
