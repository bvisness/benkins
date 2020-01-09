package app

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"gopkg.in/src-d/go-git.v4"
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

	log.Print("Starting watch for " + repoUrl)
	ticker := time.NewTicker(time.Second * 15)

	previousHash := ""

	for {
		func() {
			dir, hash, cleanup := temporaryCheckout(repoUrl)
			defer cleanup()

			if hash == previousHash {
				return
			}
			defer func() {
				previousHash = hash
			}()

			var config Config

			files, _ := ioutil.ReadDir(dir)
			didParse := false
			for _, f := range files {
				if f.Name() == "roboci.toml" {
					configBytes, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
					if err != nil {
						fmt.Printf("ERROR reading roboci.toml: %v\n", err)
						return
					}

					_, err = toml.Decode(string(configBytes), &config)
					if err != nil {
						fmt.Printf("ERROR reading roboci.toml: %v\n", err)
						return
					}

					didParse = true
					break
				}
			}

			if !didParse {
				fmt.Printf("ERROR: could not find roboci.toml\n")
				return
			}

			if config.Script == "" {
				fmt.Printf("ERROR: no script was provided\n")
				return
			}

			scriptPath := filepath.Join(dir, config.Script)
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				fmt.Printf("ERROR: could not find script named '%v'\n", config.Script)
				return
			}

			// Run the script
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
				defer cancel()

				cmd := exec.CommandContext(ctx, scriptPath)
				cmd.Env = append(os.Environ(), // TODO: Environment variables what make sense
					"ROBOCI_COMMIT_HASH="+hash,
				)
				cmd.Dir = dir

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				must(cmd.Start())
				must(cmd.Wait())

				if cmd.ProcessState.Success() {
					fmt.Printf("Script executed successfully.\n")
				} else {
					fmt.Printf("Script failed with exit code %v.\n", cmd.ProcessState.ExitCode())
				}
			}()

			// Upload the artifacts

			fmt.Printf("Done.\n")
		}()

		<-ticker.C
	}
}

func temporaryCheckout(url string) (dir string, hash string, cleanup func()) {
	tmpdir, _ := ioutil.TempDir("", "")

	r, err := git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	must(err)

	wt, err := r.Worktree()
	must(err)
	wt.Checkout(&git.CheckoutOptions{})

	head, err := r.Head()
	must(err)

	return tmpdir, head.Hash().String(), func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			panic(err)
		}
	}
}

func must(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
