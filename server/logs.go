package server

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/frc-2175/benkins/ansicolors"
	"github.com/frc-2175/benkins/shared"
	"github.com/gin-gonic/gin"
)

func LogsIndex(r *gin.Engine, loader Loader) gin.HandlerFunc {
	return func(c *gin.Context) {
		commit, err := loader.Commit(c.Param("project"), c.Param("hash"))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		logs, err := ioutil.ReadFile(filepath.Join(commit.Filepath, shared.ExecutionLogFilename))
		if err != nil {
			code := http.StatusInternalServerError
			if os.IsNotExist(err) {
				code = http.StatusNotFound
			}

			c.AbortWithError(code, err)
			return
		}

		blocks := ansicolors.Process(logs)
		_ = blocks

		c.AbortWithStatus(http.StatusOK)
	}
}
