package app

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path"
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

	var serverUrl string
	for {
		fmt.Print("Enter the Benkins server URL: ")
		url, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		url = strings.TrimSpace(url)

		res, err := http.Get(url)
		if err != nil {
			fmt.Printf("ERROR verifying server URL: %v\n", err)
			continue
		}
		if res.StatusCode != 200 {
			fmt.Printf("ERROR: did not get a 200 response from the server: \n")
			dump, _ := httputil.DumpResponse(res, true)
			fmt.Println(string(dump))
			continue
		}

		serverUrl = url
		break
	}

	var repoUrl string
	if len(os.Args) > 1 {
		repoUrl = os.Args[1]
	} else {
		for {
			fmt.Print("Enter a repo URL (HTTPS): ")
			url, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				continue
			}

			repoUrl = strings.TrimSpace(url)
			break
		}
	}
	projectName := ProjectName(repoUrl)

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
				outputBuffer := &bytes.Buffer{}

				stdout := io.MultiWriter(os.Stdout, outputBuffer)
				stderr := io.MultiWriter(os.Stderr, outputBuffer)

				branchName := branch.Name().Short()
				hash := branch.Hash().String()
				color.New(color.Bold).Fprintf(stdout, "Running for branch %v (commit %v)\n", branchName, hash)
				branchHashes[branchName] = hash // we only want to run this once!

				// Check if the server has already run for this commit
				res, err := http.Get(BuildUrl(serverUrl, projectName, hash))
				if err != nil {
					fmt.Fprintf(stderr, "WARNING: failed to check if this commit has already run: %v\n", err)
					fmt.Fprintf(stderr, "Skipping job.\n")
					return
				}

				if (res.StatusCode < 200 && 299 < res.StatusCode) && res.StatusCode != http.StatusNotFound {
					fmt.Fprintf(stderr, "WARNING: got unexpected status code when checking if this commit has already run: %v\n")
					dump, _ := httputil.DumpResponse(res, true)
					fmt.Fprintf(stderr, string(dump)+"\n")
					return
				}

				if res.StatusCode != http.StatusNotFound {
					fmt.Fprintf(stdout, "This commit has already been run; skipping.\n")
					return
				}

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
				func() {
					requestBody := &bytes.Buffer{}
					writer := multipart.NewWriter(requestBody)

					err := WriteMultipartFile(writer, "benkins-execution-log.txt", outputBuffer)
					if err != nil {
						fmt.Printf("WARNING: Failed to add execution log as artifact")
					}

					for _, artifactName := range config.Artifacts {
						file, err := os.Open(filepath.Join(dir, artifactName))
						if os.IsNotExist(err) {
							fmt.Fprintf(stderr, "WARNING: Failed to read artifact '%v'\n", artifactName)
							continue
						}
						defer file.Close()

						err = WriteMultipartFile(writer, artifactName, file)
						if err != nil {
							fmt.Fprintf(stderr, "ERROR adding artifact to request: %v\n", err)
							continue
						}
					}

					err = writer.Close()
					if err != nil {
						fmt.Fprintf(stderr, "ERROR: Failed to close multipart write for artifacts")
						return
					}

					u, _ := url.Parse(serverUrl)
					u.Path = path.Join(u.Path, url.PathEscape(ProjectName(repoUrl)), hash, "artifacts")
					res, err := http.Post(u.String(), writer.FormDataContentType(), requestBody)
					if err != nil {
						fmt.Fprintf(stderr, "ERROR uploading artifacts to server: %v\n", err)
					}
					if res.StatusCode < 200 || 299 < res.StatusCode {
						fmt.Fprintf(stderr, "ERROR: did not receive success from server when uploading artifacts: \n")
						dump, _ := httputil.DumpResponse(res, true)
						fmt.Fprintf(stderr, string(dump)+"\n")
					}
				}()

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

func WriteMultipartFile(w *multipart.Writer, name string, src io.Reader) error {
	fileWriter, err := w.CreateFormFile("files", name)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileWriter, src)

	return err
}

func ProjectName(repoUrl string) string {
	u, _ := url.Parse(repoUrl)
	return strings.Trim(u.EscapedPath(), "/")
}

func BuildUrl(baseUrl string, components ...string) string {
	u, _ := url.Parse(baseUrl)

	segments := []string{u.Path}
	for i := range components {
		segments = append(segments, url.PathEscape(components[i]))
	}

	u.Path = path.Join(segments...)

	return u.String()
}
