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
		projectName := shared.NewProjectNameFromEncoded(c.Param("project"))

		commits, err := loader.ProjectCommits(projectName)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.HTML(http.StatusOK, "project", v{
			"projectName": projectName,
			"commits":     commits,
			"branches":    loader.Branches(commits),
		})
	}
}
