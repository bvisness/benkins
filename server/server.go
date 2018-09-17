package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	git "gopkg.in/src-d/go-git.v4"
	// "gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/src-d/go-git.v4/plumbing"
	// "gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/yaml.v2"
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

func Boot(config ServerConfig) {
	for name, project := range config.Projects {
		fmt.Printf("%s: %s\n", name, project.URL)
	}

	ln, _ := net.Listen("tcp", ":8080")
	for {
		conn, _ := ln.Accept()

		buf := make([]byte, 16)
		reqLen, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		fmt.Print(string(buf))

		conn.Write([]byte(fmt.Sprintf("Message received: %d bytes.\n", reqLen)))

		conn.Close()
	}
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
