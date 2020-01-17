package server

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const LineWidth = 100

func Main() {
	reader := bufio.NewReader(os.Stdin)

	var basePath string

	for {
		fmt.Print("Enter the path where you would like to serve from: ")
		path, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %v", err)
			continue
		}
		path = strings.TrimSpace(path)

		for {
			basePath, _ = filepath.Abs(path)
			fmt.Printf("Checking if Benkins can be served from %v...\n", basePath)

			confirmPath := filepath.Join(basePath, "benkins-confirm")
			_, err := os.Stat(confirmPath)
			if os.IsNotExist(err) {
				fmt.Print(wrap(strings.Join([]string{
					"It doesn't look like you've used this folder for Benkins before.",
					"To confirm that you want to serve Benkins in this folder, create a file",
					fmt.Sprintf("in %s", basePath),
					"called \"benkins-confirm\", then type y to confirm: ",
				}, " ")))
				reader.ReadString('\n')
				continue
			}

			break
		}

		break
	}

	r := gin.Default()

	// TODO: AUTH!!

	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	})
	r.GET(":project/:hash", func(c *gin.Context) {
		project := c.Param("project")
		hash := c.Param("hash")

		_, err := os.Stat(filepath.Join(basePath, artifactPath(project, hash)))
		if os.IsNotExist(err) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.AbortWithStatus(http.StatusOK)
	})
	r.POST(":project/:hash/artifacts", func(c *gin.Context) {
		project := c.Param("project")
		hash := c.Param("hash")

		form, err := c.MultipartForm()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("form was no good: %v", err))
			return
		}

		files := form.File["files"]

		for _, file := range files {
			dstDir := filepath.Join(basePath, artifactPath(project, hash))
			err := os.MkdirAll(dstDir, 0755)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("error creating artifact directory: %v", err))
				return
			}

			dst := filepath.Join(dstDir, filepath.Base(file.Filename))
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %v", err))
				return
			}
		}

		c.String(http.StatusOK, "Artifacts uploaded successfully.")
	})

	r.Run(":8080")
}

func wrap(text string) string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return text
	}
	wrapped := words[0]
	spaceLeft := LineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + word
			spaceLeft = LineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return wrapped
}

func artifactPath(project, hash string) string {
	return filepath.Join(project, hash)
}
