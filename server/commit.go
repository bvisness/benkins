package server

import (
	"net/http"

	"github.com/frc-2175/benkins/shared"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func CommitIndex(r *gin.Engine, loader Loader) gin.HandlerFunc {
	r.HTMLRender.(multitemplate.Renderer).AddFromFilesFuncs("commit", TemplateFuncs, "server/tmpl/base.html", "server/tmpl/commit.html")

	return func(c *gin.Context) {
		projectName := shared.NewProjectNameFromEncoded(c.Param("project"))

		commit, err := loader.Commit(projectName, c.Param("hash"))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.HTML(http.StatusOK, "commit", v{
			"projectName": projectName,
			"commit":      commit,
		})
	}
}
