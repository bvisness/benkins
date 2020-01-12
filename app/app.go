package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Config struct {
	Script    string
	Artifacts []string
}

func Main() {
	reader := bufio.NewReader(os.Stdin)

	var repoUrl string
	if len(os.Args) > 1 {
		repoUrl = os.Args[1]
	} else {
		for {
			fmt.Print("Enter a repo URL (HTTPS): ")
			url, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("ERROR: %v", err)
				continue
			}

			repoUrl = strings.TrimSpace(url)
			break
		}
	}

	ticker := time.NewTicker(time.Second * 15)

	branchHashes := map[string]string{}

	for {
		// Check for new commits to run on
		fmt.Printf("Checking for new commits...\n")
		var branchesToRun []*plumbing.Reference
		func() {
			repo, _, cleanup := temporaryCheckout(repoUrl, "", NewColorWriter(os.Stdout, color.New(color.FgHiBlack)))
			defer cleanup()

			err := repo.Fetch(&git.FetchOptions{
				Progress: os.Stdout,
			})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				panic(err)
			}

			remote, err := repo.Remote("origin")
			must(err)
			remoteRefs, err := remote.List(&git.ListOptions{})
			must(err)
			for _, remoteRef := range remoteRefs {
				refName := remoteRef.Name().String()

				if !strings.HasPrefix(refName, "refs/heads/") {
					continue
				}

				name := remoteRef.Name().Short()
				if lastHash, exists := branchHashes[name]; exists {
					// we have seen this branch before
					if lastHash != remoteRef.Hash().String() {
						// new hash means new commit on this branch
						branchesToRun = append(branchesToRun, remoteRef)
					}
				} else {
					// new branch, never seen it, so run the latest
					branchesToRun = append(branchesToRun, remoteRef)
				}
			}
		}()

		for _, branch := range branchesToRun {
			func() {
				outputFile, err := ioutil.TempFile(".", "benkins-execution-log-")
				must(err)
				defer outputFile.Close()

				stdout := io.MultiWriter(os.Stdout, outputFile)
				stderr := io.MultiWriter(os.Stderr, outputFile)

				branchName := branch.Name().Short()
				hash := branch.Hash().String()
				color.New(color.Bold).Fprintf(stdout, "Running for branch %v (commit %v)\n", branchName, hash)
				branchHashes[branchName] = hash // we only want to run this once!

				_, dir, cleanup := temporaryCheckout(repoUrl, hash, nil)
				defer cleanup()

				var config Config

				files, _ := ioutil.ReadDir(dir)
				didParse := false
				for _, f := range files {
					if f.Name() == "benkins.toml" {
						configBytes, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
						if err != nil {
							fmt.Fprintf(stderr, "ERROR reading benkins.toml: %v\n", err)
							return
						}

						_, err = toml.Decode(string(configBytes), &config)
						if err != nil {
							fmt.Fprintf(stderr, "ERROR reading benkins.toml: %v\n", err)
							return
						}

						didParse = true
						break
					}
				}

				if !didParse {
					fmt.Fprintf(stderr, "WARNING: could not find benkins.toml, so not running anything\n")
					return
				}

				if config.Script == "" {
					fmt.Fprintf(stderr, "ERROR: no script was provided\n")
					return
				}

				scriptPath := filepath.Join(dir, config.Script)
				if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
					fmt.Fprintf(stderr, "ERROR: could not find script named '%v'\n", config.Script)
					return
				}

				// Run the script
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
					defer cancel()

					cmd := exec.CommandContext(ctx, scriptPath)
					cmd.Env = append(os.Environ(), // TODO: Environment variables what make sense
						"BENKINS_COMMIT_HASH="+hash,
					)
					cmd.Dir = dir

					cmd.Stdout = stdout
					cmd.Stderr = NewColorWriter(stderr, color.New(color.Bold, color.FgRed))

					must(cmd.Start())
					must(cmd.Wait())

					if cmd.ProcessState.Success() {
						color.New(color.FgGreen, color.Bold).Fprintf(stdout, "Script executed successfully.\n")
					} else {
						color.New(color.FgRed, color.Bold).Fprintf(stderr, "Script failed with exit code %v.\n", cmd.ProcessState.ExitCode())
					}
				}()

				// Upload the artifacts

				fmt.Fprintf(stdout, "Done.\n")

				// Notify us in Slack
			}()
		}

		<-ticker.C
	}
}

func temporaryCheckout(url string, hash string, progress io.Writer) (repo *git.Repository, dir string, cleanup func()) {
	tmpdir, _ := ioutil.TempDir("", "")

	if progress == nil {
		progress = os.Stdout
	}

	r, err := git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:      url,
		Progress: progress,
	})
	must(err)

	wt, err := r.Worktree()
	must(err)

	opts := &git.CheckoutOptions{}
	if hash != "" {
		opts.Hash = plumbing.NewHash(hash)
	}
	must(wt.Checkout(opts))

	return r, tmpdir, func() {
		must(os.RemoveAll(tmpdir))
	}
}

func must(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

type ColorWriter struct {
	W     io.Writer
	Color *color.Color
}

var _ io.Writer = ColorWriter{}

func NewColorWriter(w io.Writer, c *color.Color) ColorWriter {
	return ColorWriter{
		W:     w,
		Color: c,
	}
}

func (w ColorWriter) Write(p []byte) (n int, err error) {
	return w.Color.Fprint(w.W, string(p))
}
