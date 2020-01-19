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
		commit, err := loader.Commit(c.Param("project"), c.Param("hash"))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		projectName := shared.Base64Decode(c.Param("project"))

		c.HTML(http.StatusOK, "commit", v{
			"projectName": projectName,
			"commit":      commit,
		})
	}
}
