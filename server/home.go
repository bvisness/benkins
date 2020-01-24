package server

import (
	"net/http"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func Home(r *gin.Engine, loader Loader) gin.HandlerFunc {
	r.HTMLRender.(multitemplate.Renderer).AddFromFilesFuncs("home", TemplateFuncs, "server/tmpl/base.html", "server/tmpl/home.html")

	return func(c *gin.Context) {
		projects, err := loader.LoadProjects()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.HTML(http.StatusOK, "home", v{
			"projects":   projects,
			"heartbeats": heartbeats,
		})
	}
}
