package server

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gin-contrib/multitemplate"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/gin-gonic/gin"
)

const LineWidth = 100

type v map[string]interface{}

// TODO: Sanitize dots in filepath stuff everywhere

func Main(basePath, password string) {
	reader := bufio.NewReader(os.Stdin)

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

	for {
		fmt.Print("Enter the password you would like the server to use: ")
		passwordBytes, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}

		password = strings.TrimSpace(string(passwordBytes))

		break
	}

	loader := NewLoader(basePath)

	r := gin.Default()
	r.HTMLRender = multitemplate.NewRenderer()

	r.Static("/static", "server/static")

	r.GET("/", Home(r, loader))
	r.GET("p/:project", ProjectIndex(r, loader))
	r.GET("p/:project/:hash", CommitIndex(r, loader))
	r.GET("p/:project/:hash/logs", LogsIndex(r, loader))
	r.GET("p/:project/:hash/f/:file", FileIndex(r, loader))

	api := r.Group("api", func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth != password {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	})
	{
		api.GET("/", func(c *gin.Context) {
			c.AbortWithStatus(http.StatusOK)
		})
		api.GET(":project/:hash", func(c *gin.Context) {
			projectEncoded := c.Param("project")
			hash := c.Param("hash")

			_, err := os.Stat(filepath.Join(basePath, artifactPath(projectEncoded, hash)))
			if os.IsNotExist(err) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.AbortWithStatus(http.StatusOK)
		})
		api.POST(":project/:hash/artifacts", func(c *gin.Context) {
			projectEncoded := c.Param("project")
			hash := c.Param("hash")

			form, err := c.MultipartForm()
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("form was no good: %v", err))
				return
			}

			files := form.File["files"]

			for _, file := range files {
				dstDir := filepath.Join(basePath, artifactPath(projectEncoded, hash))
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
	}

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
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

func artifactPath(projectEncoded, hash string) string {
	return filepath.Join(projectEncoded, hash)
}
