package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/frc-2175/benkins/shared"
)

type Commit struct {
	Hash       string
	BranchName string
	Time       time.Time
	Success    bool
}

type Branch struct {
	Name    string
	Commits []Commit
}

type Loader struct {
	BasePath string
}

func NewLoader(path string) Loader {
	return Loader{
		BasePath: path,
	}
}

func (l *Loader) LoadProjects() (map[string][]Commit, error) {
	result := map[string][]Commit{}

	projectInfos, err := ioutil.ReadDir(l.BasePath)
	if err != nil {
		return nil, err
	}

	for _, projectInfo := range projectInfos {
		if !projectInfo.IsDir() {
			continue
		}

		projectName := shared.Base64Decode(projectInfo.Name())

		commits, err := l.ProjectCommits(projectInfo.Name())
		if err != nil {
			return nil, err
		}

		result[projectName] = commits
	}

	return result, nil
}

func (l *Loader) ProjectCommits(encodedName string) ([]Commit, error) {
	commitInfos, err := ioutil.ReadDir(filepath.Join(l.BasePath, encodedName))
	if err != nil {
		return nil, err
	}

	var commits []Commit

	for _, commitInfo := range commitInfos {
		commit, err := l.Commit(encodedName, commitInfo.Name())
		if err != nil {
			fmt.Printf("WARNING: %v\n", err)
			continue
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (l *Loader) Commit(projectNameEncoded, hash string) (Commit, error) {
	folderPath := filepath.Join(l.BasePath, projectNameEncoded, hash)

	info, err := os.Stat(folderPath)
	if err != nil {
		return Commit{}, err
	}

	resultBytes, err := ioutil.ReadFile(filepath.Join(folderPath, "benkins-results.toml"))
	if err != nil {
		return Commit{}, fmt.Errorf("failed to read benkins-results.toml for %s commit %s: %v", shared.Base64Decode(projectNameEncoded), hash, err)
	}

	var results shared.JobResults
	_, err = toml.Decode(string(resultBytes), &results)
	if err != nil {
		return Commit{}, fmt.Errorf("failed to decode benkins-results.toml for %s commit %s: %v", shared.Base64Decode(projectNameEncoded), hash, err)
	}

	return Commit{
		Hash:       hash,
		BranchName: results.BranchName,
		Time:       info.ModTime(),
		Success:    results.Success,
	}, nil
}

func (l *Loader) Branches(commits []Commit) []Branch {
	branchCommits := map[string][]Commit{}

	for _, commit := range commits {
		branchCommits[commit.BranchName] = append(branchCommits[commit.BranchName], commit)
	}

	var result []Branch
	for name, cs := range branchCommits {
		result = append(result, Branch{
			Name:    name,
			Commits: cs,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if len(result[i].Commits) == 0 {
			return false
		}

		if len(result[j].Commits) == 0 {
			return true
		}

		return result[j].Commits[0].Time.After(result[i].Commits[0].Time)
	})

	return result
}
