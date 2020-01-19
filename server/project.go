package server

import (
	"net/http"

	"github.com/frc-2175/benkins/shared"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func ProjectIndex(r *gin.Engine, loader Loader) gin.HandlerFunc {
	r.HTMLRender.(multitemplate.Renderer).AddFromFilesFuncs("project", TemplateFuncs, "server/tmpl/base.html", "server/tmpl/project.html")

	return func(c *gin.Context) {
		commits, err := loader.ProjectCommits(c.Param("project"))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		projectName := shared.Base64Decode(c.Param("project"))

		c.HTML(http.StatusOK, "project", v{
			"projectName": projectName,
			"commits":     commits,
			"branches":    loader.Branches(commits),
		})
	}
}
