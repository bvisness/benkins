package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func FileIndex(r *gin.Engine, loader Loader) gin.HandlerFunc {
	return func(c *gin.Context) {
		commit, err := loader.Commit(c.Param("project"), c.Param("hash"))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		path := filepath.Join(commit.Filepath, c.Param("file"))

		_, err = os.Stat(path)
		if err != nil {
			code := http.StatusInternalServerError
			if os.IsNotExist(err) {
				code = http.StatusNotFound
			}

			c.AbortWithError(code, err)
			return
		}

		c.Header("Content-Type", "text/plain")
		c.File(path)
		c.AbortWithStatus(http.StatusOK)
	}
}
