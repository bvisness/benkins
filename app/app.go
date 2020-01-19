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
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/frc-2175/benkins/shared"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Config struct {
	Script    string
	Artifacts []string
}

func Main() {
	reader := bufio.NewReader(os.Stdin)

	var password string

	var serverUrl string
	for {
		fmt.Print("Enter the Benkins server URL: ")
		url, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		url = strings.TrimSpace(url)

		fmt.Print("Enter the password for the server: ")
		passwordBytes, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		fmt.Println()

		passwordString := strings.TrimSpace(string(passwordBytes))

		res, err := authedGet(BuildUrl(url, "api"), passwordString)
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
		password = passwordString

		break
	}

	var slack *SlackClient
	var slackChannelId string
	for {
		fmt.Print("Enter the Slack OAuth token: ")
		tokenBytes, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		fmt.Println()
		tokenString := strings.TrimSpace(string(tokenBytes))

		fmt.Print("Enter the Slack channel ID (NOT the channel name): ")
		channelId, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		channelId = strings.TrimSpace(channelId)

		slack = NewSlackClient(tokenString)
		slackChannelId = channelId

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

				branchesToRun = append(branchesToRun, remoteRef)
			}
		}()

		for _, branch := range branchesToRun {
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						fmt.Fprintf(os.Stderr, "PANIC RECOVERED: %v", recovered)
					}
				}()

				outputBuffer := &bytes.Buffer{}

				stdout := io.MultiWriter(os.Stdout, outputBuffer)
				stderr := io.MultiWriter(os.Stderr, outputBuffer)

				branchName := branch.Name().Short()
				hash := branch.Hash().String()
				color.New(color.Bold).Fprintf(stdout, "Running for branch %v (commit %v)\n", branchName, hash)

				// Check if the server has already run for this commit
				res, err := authedGet(BuildUrl(serverUrl, "api", shared.Base64Encode(projectName), hash), password)
				if err != nil {
					fmt.Fprintf(stderr, "WARNING: failed to check if this commit has already run: %v\n", err)
					fmt.Fprintf(stderr, "Skipping job.\n")
					return
				}

				if !((200 <= res.StatusCode && res.StatusCode <= 299) || res.StatusCode == http.StatusNotFound) {
					fmt.Fprintf(stderr, "WARNING: got unexpected status code when checking if this commit has already run: %v\n", res.StatusCode)
					dump, _ := httputil.DumpResponse(res, true)
					fmt.Fprintf(stderr, string(dump)+"\n")
					return
				}

				if res.StatusCode == http.StatusOK {
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

				jobResults := shared.JobResults{
					BranchName: branchName,
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
					err := cmd.Wait()
					if err != nil {
						if _, isExitError := err.(*exec.ExitError); !isExitError {
							panic(err)
						}
					}

					if cmd.ProcessState.Success() {
						color.New(color.FgGreen, color.Bold).Fprintf(stdout, "Script executed successfully.\n")
					} else {
						color.New(color.FgRed, color.Bold).Fprintf(stderr, "Script failed with exit code %v.\n", cmd.ProcessState.ExitCode())
					}

					jobResults.Success = cmd.ProcessState.Success()
				}()

				// Upload the artifacts
				func() {
					requestBody := &bytes.Buffer{}
					writer := multipart.NewWriter(requestBody)

					err := WriteMultipartFile(writer, "benkins-execution-log.txt", outputBuffer)
					if err != nil {
						fmt.Printf("WARNING: Failed to add execution log as artifact")
					}

					err = WriteMultipartFile(writer, "benkins-results.toml", bytes.NewBufferString(jobResults.ToTOML()))
					if err != nil {
						fmt.Printf("WARNING: Failed to add job results as an artifact")
					}

					for _, artifactName := range config.Artifacts {
						func() {
							file, err := os.Open(filepath.Join(dir, artifactName))
							if os.IsNotExist(err) {
								fmt.Fprintf(stderr, "WARNING: Failed to read artifact '%v'\n", artifactName)
								return
							}
							defer file.Close()

							err = WriteMultipartFile(writer, artifactName, file)
							if err != nil {
								fmt.Fprintf(stderr, "ERROR adding artifact to request: %v\n", err)
								return
							}
						}()
					}

					err = writer.Close()
					if err != nil {
						fmt.Fprintf(stderr, "ERROR: Failed to close multipart write for artifacts")
						return
					}

					res, err := authedPost(
						BuildUrl(serverUrl, "api", shared.Base64Encode(ProjectName(repoUrl)), hash, "artifacts"),
						writer.FormDataContentType(),
						password,
						requestBody,
					)
					if err != nil {
						fmt.Fprintf(stderr, "ERROR uploading artifacts to server: %v\n", err)
					}
					if res.StatusCode < 200 || 299 < res.StatusCode {
						fmt.Fprintf(stderr, "ERROR: did not receive success from server when uploading artifacts: \n")
						dump, _ := httputil.DumpResponse(res, true)
						fmt.Fprintf(stderr, string(dump)+"\n")
					}
				}()

				// Notify us on Slack
				{
					notificationText := ""

					if notificationBytes, err := ioutil.ReadFile(filepath.Join(dir, "benkins-notification.txt")); err == nil {
						notificationText = string(notificationBytes)
					} else {
						if os.IsNotExist(err) {
							fmt.Println("No custom notification text.")
						} else {
							fmt.Fprintf(stderr, "WARNING: error while reading custom notification text")
						}
					}

					successEmoji := ":white_check_mark:"
					successString := "Success!"
					if !jobResults.Success {
						successEmoji = ":x:"
						successString = "Failure"
					}

					_, err := slack.SlackPostMessage(SlackMessageRequest{
						Channel: slackChannelId,
						Text:    fmt.Sprintf("%s Branch %s (Commit %s) %s", successEmoji, branchName, hash[0:7], successString),
						Blocks: []*SlackBlock{
							TextBlock("*%s Branch %s (Commit %s) %s*", successEmoji, branchName, hash[0:7], successString),
							TextBlock(notificationText),
						},
					})
					if err == nil {
						fmt.Fprintf(stdout, "Successfully posted message to Slack.\n")
					} else {
						fmt.Fprintf(stderr, "ERROR posting message to Slack: %v\n", err)
					}
				}

				// TODO: Update CI status on GitHub

				fmt.Fprintf(stdout, "Done.\n")
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

var serverClient = &http.Client{}

func authedGet(url, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", password)

	return serverClient.Do(req)
}

func authedPost(url, contentType, password string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", password)

	return serverClient.Do(req)
}
