package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/frc-2175/benkins/shared"
	"github.com/pelletier/go-toml"
)

type Commit struct {
	Hash       string
	BranchName string
	Message    string
	Time       time.Time
	Success    bool
	Filepath   string
	Files      []string
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

func (l *Loader) LoadProjects() (map[shared.ProjectName][]Commit, error) {
	result := map[shared.ProjectName][]Commit{}

	projectInfos, err := ioutil.ReadDir(l.BasePath)
	if err != nil {
		return nil, err
	}

	for _, projectInfo := range projectInfos {
		if !projectInfo.IsDir() {
			continue
		}

		projectName := shared.NewProjectNameFromEncoded(projectInfo.Name())

		commits, err := l.ProjectCommits(projectName)
		if err != nil {
			return nil, err
		}

		result[projectName] = commits
	}

	return result, nil
}

func (l *Loader) ProjectCommits(name shared.ProjectName) ([]Commit, error) {
	commitInfos, err := ioutil.ReadDir(filepath.Join(l.BasePath, name.Encoded()))
	if err != nil {
		return nil, err
	}

	var commits []Commit

	for _, commitInfo := range commitInfos {
		commit, err := l.Commit(name, commitInfo.Name())
		if err != nil {
			fmt.Printf("WARNING: %v\n", err)
			continue
		}

		commits = append(commits, commit)
	}

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	return commits, nil
}

func (l *Loader) Commit(projectName shared.ProjectName, hash string) (Commit, error) {
	folderPath := filepath.Join(l.BasePath, projectName.Encoded(), hash)

	info, err := os.Stat(folderPath)
	if err != nil {
		return Commit{}, err
	}

	resultBytes, err := ioutil.ReadFile(filepath.Join(folderPath, shared.ResultsFilename))
	if err != nil {
		return Commit{}, fmt.Errorf("failed to read %s for %s commit %s: %v", shared.ResultsFilename, projectName.Decoded(), hash, err)
	}

	var results shared.JobResults
	err = toml.Unmarshal(resultBytes, &results)
	if err != nil {
		return Commit{}, fmt.Errorf("failed to decode %s for %s commit %s: %v", shared.ResultsFilename, projectName.Decoded(), hash, err)
	}

	fileInfos, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return Commit{}, err
	}

	var files []string
	for _, info := range fileInfos {
		files = append(files, info.Name())
	}

	return Commit{
		Hash:       hash,
		BranchName: results.BranchName,
		Message:    results.CommitMessage,
		Time:       info.ModTime(),
		Success:    results.Success,
		Filepath:   filepath.Join(l.BasePath, projectName.Encoded(), hash),
		Files:      files,
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

		return result[i].Commits[0].Time.After(result[j].Commits[0].Time)
	})

	return result
}
