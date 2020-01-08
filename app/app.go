package app

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/src-d/go-git.v4"
)

func Main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	_ = home
	configBytes, err := ioutil.ReadFile(filepath.Join(".", "roboci.txt")) // TODO: use home directory
	// TODO: detect missing file
	if err != nil {
		panic(err)
	}

	repos := strings.Split(string(configBytes), "\n")
	log.Print(repos)

	var wg sync.WaitGroup

	for _, repo := range repos {
		if strings.TrimSpace(repo) == "" {
			continue
		}

		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			watchRepo(repo)
		}(repo)
	}

	wg.Wait()
}

func watchRepo(repo string) {
	log.Print("Starting watch for " + repo)
	ticker := time.NewTicker(time.Second * 15)

	for {
		log.Print("Checking out " + repo)
		dir, cleanup := temporaryCheckout(repo)
		defer cleanup()

		files, _ := ioutil.ReadDir(dir)
		for _, f := range files {
			log.Print(f.Name())
		}

		<-ticker.C
	}
}

func temporaryCheckout(url string) (dir string, cleanup func()) {
	tmpdir, _ := ioutil.TempDir("", "")

	r, _ := git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	wt, _ := r.Worktree()
	wt.Checkout(&git.CheckoutOptions{})

	return tmpdir, func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			panic(err)
		}
	}
}
